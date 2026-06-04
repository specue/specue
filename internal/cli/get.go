package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

// GetReport is the typed result of `get <resource>`: a flat row per matching node.
// The columns are uniform (id, status, detail) so a human table and a JSON array
// share one shape; detail is a per-type summary (a UC's service, a Need's
// coverage, a port's transport).
type GetReport struct {
	Resource string    `json:"resource"`
	Rows     []nodeRow `json:"rows"`
}

// nodeRow is one node flattened for the listing. Detail is type-specific context
// kept short enough for a table cell; describe is where the full node lives. Type
// is carried for the `get all` listing (which spans types); a single-type listing
// leaves it empty and the renderer drops the column.
type nodeRow struct {
	ID     string `json:"id"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"` // empty for Port/Container/Plan/ADR
	Detail string `json:"detail,omitempty"`
}

// ResourceList is the typed result of a bare `get`: the selectable resources, so a
// caller (human or agent) can discover what `get <resource>` accepts without
// guessing — the kubectl api-resources pattern. Machine-readable via --json.
type ResourceList struct {
	Resources []resourceInfo `json:"resources"`
}

type resourceInfo struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	Type    string   `json:"type"`
}

// runResources returns the resource registry as a typed list. It needs no graph —
// the set is static — so a bare `get` works even outside a spec tree.
//
//specue:req:list-resources
func runResources() ResourceList {
	out := make([]resourceInfo, len(resources))
	for i, r := range resources {
		a := r.aliases
		if a == nil {
			a = []string{}
		}
		out[i] = resourceInfo{Name: r.name, Aliases: a, Type: string(r.typ)}
	}
	return ResourceList{Resources: out}
}

func (l ResourceList) renderHuman(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "RESOURCE\tALIASES\tNODE TYPE"); err != nil {
		return err
	}
	for _, r := range l.Resources {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", r.Name, strings.Join(r.Aliases, ","), r.Type); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (l ResourceList) jsonValue() any { return l }

// runGet builds the graph, filters to the resource's node type, and rows them up.
// An optional id narrows to a single node (still rendered as a one-row table, so
// the output shape is stable whether listing or selecting).
//
//specue:req:list-resources#nodes-of-a-kind
func runGet(ctx Context, word string, id string) (GetReport, *Problem) {
	// "all" is a pseudo-resource: every node, every type (the kubectl `get all`
	// pattern). It carries no single type, so the type filter is skipped and the row
	// keeps its Type for the renderer to show.
	all := word == allResource
	var rsc resource
	if !all {
		r, p := resolveResource(word)
		if p != nil {
			return GetReport{}, p
		}
		rsc = r
	}

	res, p := buildGraph(ctx)
	if p != nil {
		return GetReport{}, p
	}

	var want *model.NodeID
	if id != "" {
		nid, p := parseNodeID(id)
		if p != nil {
			return GetReport{}, p
		}
		want = &nid
	}

	// Resource records the CANONICAL name (rsc.name), so an alias never leaks into
	// output — `get uc` and `get usecase` produce identical results. For the `all`
	// pseudo-resource there is no registry entry, so the word itself is canonical.
	canonical := word
	if !all {
		canonical = rsc.name
	}
	rep := GetReport{Resource: canonical}
	for n := range res.Graph.Nodes() {
		if !all && n.Node().Type != rsc.typ {
			continue
		}
		if want != nil && n.ID() != *want {
			continue
		}
		rep.Rows = append(rep.Rows, rowFor(n, all))
	}
	sort.Slice(rep.Rows, func(i, j int) bool { return rep.Rows[i].ID < rep.Rows[j].ID })

	if want != nil && len(rep.Rows) == 0 {
		p := Errorf(fmt.Sprintf("run `%s %s` to list what exists, or check the module:slug", cmdPath(cmdGet), word),
			"no %s named %s", word, want)
		return GetReport{}, &p
	}
	return rep, nil
}

// rowFor flattens a resolved node into a listing row, choosing the detail column by
// type. withType carries the node's type (for the `all` listing that spans types).
func rowFor(n *compiler.ResolvedNode, withType bool) nodeRow {
	r := nodeRow{
		ID:     n.ID().String(),
		Status: string(n.Status),
		Detail: detailOf(n),
	}
	if withType {
		r.Type = string(n.Node().Type)
	}
	return r
}

// detailOf is the type-specific one-line context for the listing.
func detailOf(n *compiler.ResolvedNode) string {
	b := n.Node().Body
	switch n.Node().Type {
	case model.TypeUseCase:
		if b != nil && b.UseCase != nil {
			return "service " + b.UseCase.Service.String()
		}
	case model.TypeNeed:
		if b != nil && b.Need != nil {
			return fmt.Sprintf("domain %s · %d atom(s)", b.Need.Domain, len(b.Need.Atoms))
		}
	case model.TypePort:
		if b != nil && b.Port != nil {
			return fmt.Sprintf("%s/%s", b.Port.Kind, b.Port.Transport)
		}
	case model.TypeContainer:
		if b != nil && b.Container != nil {
			return string(b.Container.Kind)
		}
	}
	return n.Node().Title
}

// renderHuman writes the listing as an aligned table; an empty result prints a
// terse "none" so the caller never mistakes silence for breakage. The TYPE column
// appears only for `get all` (where rows span types); a single-type listing omits
// it as redundant.
func (r GetReport) renderHuman(w io.Writer) error {
	if len(r.Rows) == 0 {
		_, err := fmt.Fprintf(w, "no %s found\n", r.Resource)
		return err
	}
	withType := r.Resource == allResource
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	header := "ID\tSTATUS\tDETAIL"
	if withType {
		header = "ID\tTYPE\tSTATUS\tDETAIL"
	}
	if _, err := fmt.Fprintln(tw, header); err != nil {
		return err
	}
	for _, row := range r.Rows {
		var line string
		if withType {
			line = fmt.Sprintf("%s\t%s\t%s\t%s", row.ID, row.Type, row.Status, row.Detail)
		} else {
			line = fmt.Sprintf("%s\t%s\t%s", row.ID, row.Status, row.Detail)
		}
		if _, err := fmt.Fprintln(tw, line); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// jsonValue exposes the stable JSON shape (the struct already has json tags, but a
// nil Rows slice should encode as [] not null).
func (r GetReport) jsonValue() any {
	if r.Rows == nil {
		r.Rows = []nodeRow{}
	}
	return r
}
