package jsonir

import (
	"fmt"
	"slices"
	"sort"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

// buildCommon fills the identity-and-status envelope shared by every node
// type. The body Prose lands in the long-form `body` field; an empty Prose is
// dropped via omitempty.
func buildCommon(n *compiler.ResolvedNode, revisions map[model.ModulePath]string) commonJSON {
	id := n.ID()
	nd := n.Node()
	c := commonJSON{
		ID:           id.String(),
		Type:         string(nd.Type),
		Module:       string(id.Module),
		Slug:         string(id.Slug),
		Title:        nd.Title,
		Status:       string(n.Status),
		Confidence:   string(nd.Confidence),
		Visibility:   string(nd.Visibility),
		RenderedFrom: revisions[id.Module],
	}
	if nd.Body != nil {
		c.Body = nd.Body.Prose
	}
	return c
}

// buildDerived walks the resolved fields the compiler computed and turns them
// into the wire form. Returns nil when nothing derived is present — the file's
// `derived` key is then omitted entirely (no empty stub).
func buildDerived(n *compiler.ResolvedNode) *derivedJSON {
	d := &derivedJSON{
		Uses:     idStrings(n.Uses),
		CoreUses: idStrings(n.CoreUses),
		Realizes: idStrings(n.Realizes),
		Blocked:  n.Blocked,
	}
	if len(n.Satisfies) > 0 {
		d.Satisfies = make([]satisfyJSON, len(n.Satisfies))
		for i, s := range n.Satisfies {
			d.Satisfies[i] = satisfyJSON{Need: s.Need.String(), Atom: string(s.Atom)}
		}
	}
	if n.Node().Type == model.TypePort {
		t := n.Topology
		if hasTopology(t) {
			d.Topology = &topologyJSON{
				ProducedBy: idStrings(t.ProducedBy),
				ConsumedBy: idStrings(t.ConsumedBy),
				ServedBy:   idStrings(t.ServedBy),
				CalledBy:   idStrings(t.CalledBy),
				GrantedBy:  idStrings(t.GrantedBy),
			}
		}
	}
	if d.Uses == nil && d.CoreUses == nil && d.Realizes == nil &&
		d.Satisfies == nil && d.Topology == nil && !d.Blocked {
		return nil
	}
	return d
}

func hasTopology(t compiler.TopologyRoles) bool {
	return len(t.ProducedBy)+len(t.ConsumedBy)+len(t.ServedBy)+len(t.CalledBy)+len(t.GrantedBy) > 0
}

// buildBindings converts the compiler's binding maps into the wire shape.
// Whole-contract bindings (empty ElementID key) emit with no `element` field;
// scoped ones carry the element id. Infra bindings additionally resolve their
// `to` port by walking the matching Dep on the element's parent node.
func buildBindings(n *compiler.ResolvedNode) *bindingsJSON {
	out := &bindingsJSON{
		Req:    flattenBindings(n.ReqElems),
		Covers: flattenBindings(n.CoverElems),
		Infra:  flattenInfra(n),
	}
	if len(out.Req) == 0 && len(out.Covers) == 0 && len(out.Infra) == 0 {
		return nil
	}
	return out
}

func flattenBindings(m map[model.ElementID][]compiler.Binding) []bindingJSON {
	if len(m) == 0 {
		return nil
	}
	keys := make([]model.ElementID, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return string(keys[i]) < string(keys[j]) })
	var out []bindingJSON
	for _, k := range keys {
		for _, b := range m[k] {
			out = append(out, bindingJSON{
				Element:      string(k),
				Loc:          locOf(b),
				SourceModule: string(b.SourceModule),
			})
		}
	}
	return out
}

// flattenInfra turns InfraProof into wire bindings, resolving each (role,
// element) key against the element's authored Deps to recover the port the
// binding points at. When an element-less infra fact is present (empty
// element), `to` is left empty.
func flattenInfra(n *compiler.ResolvedNode) []infraBindingJSON {
	if len(n.InfraProof) == 0 {
		return nil
	}
	keys := make([]compiler.InfraKey, 0, len(n.InfraProof))
	for k := range n.InfraProof {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Role != keys[j].Role {
			return string(keys[i].Role) < string(keys[j].Role)
		}
		return string(keys[i].Element) < string(keys[j].Element)
	})
	var out []infraBindingJSON
	for _, k := range keys {
		to := resolveInfraTarget(n, k)
		for _, b := range n.InfraProof[k] {
			out = append(out, infraBindingJSON{
				Role:         string(k.Role),
				To:           to,
				Element:      string(k.Element),
				Loc:          locOf(b),
				SourceModule: string(b.SourceModule),
			})
		}
	}
	return out
}

// resolveInfraTarget walks the Contract's authored elements to find the Dep
// that matches the InfraKey's (Role, Element) and returns its target as
// "module:slug". Empty when no match (binding without an authored anchor).
func resolveInfraTarget(n *compiler.ResolvedNode, k compiler.InfraKey) string {
	uc := n.Node().Body
	if uc == nil || uc.Contract == nil {
		return ""
	}
	for _, e := range uc.Contract.Elements {
		if e.ID != k.Element {
			continue
		}
		for _, d := range e.Deps {
			if d.Role == k.Role {
				return d.To.String()
			}
		}
	}
	return ""
}

func locOf(b compiler.Binding) string {
	if b.Line > 0 {
		return fmt.Sprintf("%s:%d", string(b.File), b.Line)
	}
	return string(b.File)
}

// buildElements converts a Contract's invariants into the wire shape. Kind
// carries the element's nature (returns/rejects, empty for plain).
func buildElements(els []model.Element) (inv []elemJSON) {
	for _, e := range els {
		inv = append(inv, elemJSON{
			ID:        string(e.ID),
			Kind:      string(e.Kind),
			Text:      e.Text,
			When:      e.When,
			Rev:       e.Rev,
			DependsOn: depsToJSON(e.Deps),
			Satisfies: satisfiesToJSON(e.Satisfies),
			DecidedBy: refsToJSON(e.DecidedBy),
		})
	}
	return
}

func depsToJSON(ds []model.Dep) []depJSON {
	if len(ds) == 0 {
		return nil
	}
	out := make([]depJSON, len(ds))
	for i, d := range ds {
		out[i] = depJSON{
			To:     d.To.String(),
			Role:   string(d.Role),
			Branch: d.Branch,
		}
		if d.Carries != (model.NodeRef{}) {
			out[i].Carries = d.Carries.String()
		}
	}
	return out
}

func satisfiesToJSON(ss []model.AtomRef) []satisfyJSON {
	if len(ss) == 0 {
		return nil
	}
	out := make([]satisfyJSON, len(ss))
	for i, s := range ss {
		out[i] = satisfyJSON{Need: s.Need.String(), Atom: string(s.Atom)}
	}
	return out
}

func refsToJSON(rs []model.NodeRef) []string {
	if len(rs) == 0 {
		return nil
	}
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.String()
	}
	return out
}

// buildAtoms splits a Need's Atoms into FRs and NFRs (matching the wire
// schema's two separate arrays).
func buildAtoms(atoms []model.Atom) (frs, nfrs []atomJSON) {
	for _, a := range atoms {
		aj := atomJSON{ID: string(a.ID), Kind: string(a.Kind), Text: a.Text}
		switch a.Kind {
		case model.KindFR:
			frs = append(frs, aj)
		case model.KindNFR:
			nfrs = append(nfrs, aj)
		}
	}
	return
}

func idStrings(ids []model.NodeID) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	slices.Sort(out)
	return out
}

func refStr(r model.NodeRef) string {
	if r == (model.NodeRef{}) {
		return ""
	}
	return r.String()
}
