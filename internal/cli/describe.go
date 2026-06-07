package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

// DescribeReport is the typed result of `describe <module:slug>` (or
// `<module:slug>#<element-id>`): one resolved node, optionally narrowed to a
// single named element. The human rendering is the full narrative; the JSON is
// the node's structured form, with elements filtered to the one named.
type DescribeReport struct {
	node    *compiler.ResolvedNode
	element model.ElementID // empty = whole node
}

// runDescribe resolves a module:slug to a node, or module:slug#element to one
// of its named elements. Identity is module-scoped and the node carries its own
// type, so no resource word is needed — the type is in the output.
//
//specue:req:describe-node
func runDescribe(ctx Context, ref string) (DescribeReport, *Problem) {
	id, elem, p := parseNodeAtElement(ref)
	if p != nil {
		return DescribeReport{}, p
	}
	res, p := buildGraph(ctx)
	if p != nil {
		return DescribeReport{}, p
	}
	n, ok := res.Graph.Node(id)
	if !ok {
		p := Errorf("run `"+usage(cmdGet)+"` to list nodes, then copy a module:slug",
			"no node %s in the landscape", id)
		return DescribeReport{}, &p
	}
	if elem != "" {
		if !nodeHasElement(n, elem) {
			p := Errorf("drop the `#"+string(elem)+"` suffix to see every element, then copy the right id",
				"node %s has no element %q", id, elem)
			return DescribeReport{}, &p
		}
	}
	return DescribeReport{node: n, element: elem}, nil
}

// nodeHasElement reports whether the node has a named element with this id (an
// invariant, variation, or named pre/postcondition) — or, on a Need, a
// named atom.
//
//specue:req:describe-node#element-scoped
func nodeHasElement(n *compiler.ResolvedNode, elem model.ElementID) bool {
	if b := n.Node().Body; b != nil {
		if b.Contract != nil {
			for _, e := range b.Contract.Elements {
				if e.ID == elem {
					return true
				}
			}
		}
		if b.Need != nil {
			for _, a := range b.Need.Atoms {
				if string(a.ID) == string(elem) {
					return true
				}
			}
		}
	}
	return false
}

// renderHuman writes the full node: header (id, type, status), the type-specific
// body, then the derived edges (uses, realizes, satisfies).
func (r DescribeReport) renderHuman(w io.Writer) error {
	n := r.node
	nd := n.Node()
	header := fmt.Sprintf("%s  [%s]", n.ID(), nd.Type)
	if n.Status != "" {
		header += "  " + string(n.Status)
	}
	if _, err := fmt.Fprintf(w, "%s\n", header); err != nil {
		return err
	}
	if nd.Title != "" {
		if _, err := fmt.Fprintf(w, "%s\n", nd.Title); err != nil {
			return err
		}
	}
	if err := renderBody(w, n, r.element); err != nil {
		return err
	}
	if r.element != "" {
		// Element-scoped view: edges live on the element itself, no rolled-up
		// realizes/uses to repeat.
		return nil
	}
	return renderEdges(w, n)
}

// renderBody dispatches to the type-specific section. When elem is non-empty,
// the rendering narrows to that named element (skipping the node-level
// service/trigger header and the other elements).
func renderBody(w io.Writer, n *compiler.ResolvedNode, elem model.ElementID) error {
	b := n.Node().Body
	if b == nil {
		return nil
	}
	switch {
	case b.Contract != nil:
		return renderContract(w, b.Contract, elem)
	case b.Need != nil:
		return renderNeed(w, b.Need, elem)
	case b.Port != nil:
		return renderPort(w, b.Port, n.Topology)
	case b.Container != nil:
		_, err := fmt.Fprintf(w, "\nkind: %s  boundary: %t\n", b.Container.Kind, b.Container.Boundary)
		return err
	case b.Gov != nil:
		_, err := fmt.Fprintf(w, "\nlifecycle: %s  branch: %s\n", b.Gov.Lifecycle, b.Gov.Branch)
		return err
	}
	return nil
}

func renderContract(w io.Writer, c *model.ContractBody, elem model.ElementID) error {
	if elem == "" {
		if _, err := fmt.Fprintf(w, "\nservice: %s  interaction: %s\n",
			c.Service, c.Interaction); err != nil {
			return err
		}
		if c.Trigger != "" {
			if _, err := fmt.Fprintf(w, "trigger: %s\n", c.Trigger); err != nil {
				return err
			}
		}
	}
	for _, e := range c.Elements {
		if elem != "" && e.ID != elem {
			continue
		}
		if err := renderElement(w, e); err != nil {
			return err
		}
	}
	return nil
}

// renderElement prints one invariant: its nature, id, text, and any satisfies it
// discharges. A guarded invariant shows its when condition.
func renderElement(w io.Writer, e model.Element) error {
	id := string(e.ID)
	if id == "" {
		id = "—"
	}

	head := fmt.Sprintf("  • [%s %s]", e.Kind, id)
	if e.Text != "" {
		head += " " + e.Text
	}
	if _, err := fmt.Fprintf(w, "%s\n", head); err != nil {
		return err
	}
	if e.When != "" {
		if _, err := fmt.Fprintf(w, "      when %s\n", e.When); err != nil {
			return err
		}
	}
	for _, s := range e.Satisfies {
		if _, err := fmt.Fprintf(w, "      satisfies %s\n", s); err != nil {
			return err
		}
	}
	return nil
}

func renderNeed(w io.Writer, nd *model.NeedBody, elem model.ElementID) error {
	if elem == "" {
		if _, err := fmt.Fprintf(w, "\ndomain: %s\n", nd.Domain); err != nil {
			return err
		}
		if nd.Consumer != "" {
			if _, err := fmt.Fprintf(w, "consumer: %s\n", nd.Consumer); err != nil {
				return err
			}
		}
		if nd.Description != "" {
			if _, err := fmt.Fprintf(w, "description: %s\n", nd.Description); err != nil {
				return err
			}
		}
	}
	for _, a := range nd.Atoms {
		if elem != "" && string(a.ID) != string(elem) {
			continue
		}
		if _, err := fmt.Fprintf(w, "  • %s: %s\n", a.ID, a.Text); err != nil {
			return err
		}
	}
	return nil
}

func renderPort(w io.Writer, p *model.PortBody, topo compiler.TopologyRoles) error {
	if _, err := fmt.Fprintf(w, "\nkind: %s  transport: %s\n", p.Kind, p.Transport); err != nil {
		return err
	}
	if p.Schema != (model.NodeRef{}) {
		if _, err := fmt.Fprintf(w, "schema: %s\n", p.Schema); err != nil {
			return err
		}
	}
	return renderTopology(w, topo)
}

// renderTopology prints a Port's derived L2 topology — who produces/consumes/
// serves/calls it. Empty roles are skipped (a datastore has no producers).
func renderTopology(w io.Writer, t compiler.TopologyRoles) error {
	for _, role := range []struct {
		label string
		ids   []model.NodeID
	}{
		{"producedBy", t.ProducedBy},
		{"consumedBy", t.ConsumedBy},
		{"servedBy", t.ServedBy},
		{"calledBy", t.CalledBy},
	} {
		if err := edgeList(w, role.label, role.ids); err != nil {
			return err
		}
	}
	return nil
}

// renderEdges prints the derived relationships every node may have: what it uses,
// Needs it realizes, atoms it satisfies.
func renderEdges(w io.Writer, n *compiler.ResolvedNode) error {
	if err := edgeList(w, "uses", n.Uses); err != nil {
		return err
	}
	if err := edgeList(w, "realizes", n.Realizes); err != nil {
		return err
	}
	return nil
}

func edgeList(w io.Writer, label string, ids []model.NodeID) error {
	if len(ids) == 0 {
		return nil
	}
	sortByID(ids)
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = id.String()
	}
	_, err := fmt.Fprintf(w, "%s: %s\n", label, strings.Join(parts, ", "))
	return err
}

// nodeJSON is the designed wire shape for a node — not the raw model.Body dump.
// References render as `module:slug` strings (the form a caller copies back into
// `describe`/`get`), every field is omitempty (no `""`/`null`/`0`/`false` noise),
// and the names are lowercase. The JSON is a render, with its own shape, exactly
// like the human view — so the internal model can change field names freely.
type nodeJSON struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`
	Status      string        `json:"status,omitempty"` // empty for Port/Container/Plan/ADR (status is Contract/Need only)
	Title       string        `json:"title,omitempty"`
	Service     string        `json:"service,omitempty"`
	Domain      string        `json:"domain,omitempty"`
	Consumer    string        `json:"consumer,omitempty"`
	Description string        `json:"description,omitempty"`
	Trigger     string        `json:"trigger,omitempty"`
	Kind        string        `json:"kind,omitempty"`     // port/container kind
	Schema      string        `json:"schema,omitempty"`   // port wire IDL ref
	Elements    []elementJSON `json:"elements,omitempty"` // Contract
	Atoms       []atomJSON    `json:"atoms,omitempty"`    // Need
	Uses        []string      `json:"uses,omitempty"`
	Realizes    []string      `json:"realizes,omitempty"`
	Topology    *topologyJSON `json:"topology,omitempty"` // Port
}

type elementJSON struct {
	Kind      string   `json:"kind,omitempty"` // nature: plain | returns | rejects
	ID        string   `json:"id,omitempty"`
	Text      string   `json:"text,omitempty"`
	When      string   `json:"when,omitempty"`
	Satisfies []string `json:"satisfies,omitempty"`
}

type atomJSON struct {
	ID   string `json:"id"`
	Text string `json:"text,omitempty"`
}

type topologyJSON struct {
	ProducedBy []string `json:"producedBy,omitempty"`
	ConsumedBy []string `json:"consumedBy,omitempty"`
	ServedBy   []string `json:"servedBy,omitempty"`
	CalledBy   []string `json:"calledBy,omitempty"`
}

// jsonValue projects the resolved node onto the designed wire shape. When the
// report carries a non-empty element id, the elements/atoms list is narrowed to
// that one entry, and node-level edges are dropped (they belong to the whole
// node, not to one element).
func (r DescribeReport) jsonValue() any {
	n := r.node
	j := nodeJSON{
		ID:    n.ID().String(),
		Type:  string(n.Node().Type),
		Title: n.Node().Title,
	}
	if r.element == "" {
		j.Status = string(n.Status)
		j.Uses = idStrings(n.Uses)
		j.Realizes = idStrings(n.Realizes)
	}
	if b := n.Node().Body; b != nil {
		fillBodyJSON(&j, b, n, r.element)
	}
	return j
}

func fillBodyJSON(j *nodeJSON, b *model.Body, n *compiler.ResolvedNode, elem model.ElementID) {
	switch {
	case b.Contract != nil:
		c := b.Contract
		if elem == "" {
			j.Service = refStr(c.Service)
			j.Trigger = c.Trigger
		}
		for _, e := range c.Elements {
			if elem != "" && e.ID != elem {
				continue
			}
			j.Elements = append(j.Elements, elementJSON{
				Kind: string(e.Kind), ID: string(e.ID), Text: e.Text,
				When: e.When, Satisfies: atomStrings(e.Satisfies),
			})
		}
	case b.Need != nil:
		nd := b.Need
		if elem == "" {
			j.Domain = refStr(nd.Domain)
			j.Consumer = nd.Consumer
			j.Description = nd.Description
		}
		for _, a := range nd.Atoms {
			if elem != "" && string(a.ID) != string(elem) {
				continue
			}
			j.Atoms = append(j.Atoms, atomJSON{ID: string(a.ID), Text: a.Text})
		}
	case b.Port != nil:
		j.Kind = string(b.Port.Kind)
		j.Schema = refStr(b.Port.Schema)
		j.Topology = topologyJSONOf(n.Topology)
	case b.Container != nil:
		j.Kind = string(b.Container.Kind)
	case b.Gov != nil:
		j.Kind = string(b.Gov.Lifecycle)
	}
}

// topologyJSONOf returns the wire topology, or nil when every role is empty (so an
// omitempty pointer drops it entirely rather than emitting an empty object).
func topologyJSONOf(t compiler.TopologyRoles) *topologyJSON {
	tj := topologyJSON{
		ProducedBy: idStrings(t.ProducedBy), ConsumedBy: idStrings(t.ConsumedBy),
		ServedBy: idStrings(t.ServedBy), CalledBy: idStrings(t.CalledBy),
	}
	if len(tj.ProducedBy)+len(tj.ConsumedBy)+len(tj.ServedBy)+len(tj.CalledBy) == 0 {
		return nil
	}
	return &tj
}

// refStr renders a node ref as module:slug, or "" for the zero ref (so omitempty
// drops an unset optional reference).
func refStr(ref model.NodeRef) string {
	if ref == (model.NodeRef{}) {
		return ""
	}
	return ref.String()
}

func atomStrings(refs []model.AtomRef) []string {
	if len(refs) == 0 {
		return nil
	}
	out := make([]string, len(refs))
	for i, r := range refs {
		out[i] = r.String()
	}
	return out
}

func idStrings(ids []model.NodeID) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}
