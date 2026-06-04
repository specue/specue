package compiler

import (
	"fmt"

	"github.com/specue/specue/internal/model"
)

// bindFacts collides scanner facts onto resolved nodes: it sets each node's
// ReqElems / CoverElems / InfraProof from the code annotations that bind it. A
// fact that resolves to no node is an orphan (a gate). Binding is module-scoped:
// a bare ref resolves in the carrying module first, then a candidate module that
// owns the slug (the deploy-repo-serves-many case).
//
//specue:req:validate-graph#dangling-binding
func bindFacts(g *ResolvedGraph, facts []CodeFact) []Diagnostic {
	var diags []Diagnostic
	for _, f := range facts {
		n := resolveFact(g, f)
		if n == nil {
			d := newDiag(OrphanBinding, model.NodeID{Module: f.Module, Slug: f.Target.Slug},
				fmt.Sprintf("//specue:%s:%s in module %s resolves to no node", f.Verb, f.Target.Slug, f.Module))
			d.Location = Location{File: f.File, Line: f.Line}
			diags = append(diags, d)
			continue
		}
		if d := bindOne(n, f); d != nil {
			diags = append(diags, *d)
		}
	}
	return diags
}

// resolveFact finds the node a code fact binds. A code annotation is a lexical
// string, not a CUE reference, so the slug is resolved here against the carrying
// module first, then any candidate module that owns the slug (a deploy repo
// serving several modules). An aliased annotation names its module directly.
func resolveFact(g *ResolvedGraph, f CodeFact) *ResolvedNode {
	if f.Target.Alias != "" {
		// An aliased code annotation can only mean a module the carrying module
		// imports; without the require table here we resolve it by the slug across
		// loaded modules, preferring an exact carrying-module hit.
		if n := g.bySlug[f.Module][f.Target.Slug]; n != nil {
			return n
		}
		return nil
	}
	if n := g.bySlug[f.Module][f.Target.Slug]; n != nil {
		return n
	}
	for _, cand := range f.Candidates {
		if cand == f.Module {
			continue
		}
		if n := g.bySlug[cand][f.Target.Slug]; n != nil {
			return n
		}
	}
	// A code module holds no nodes of its own — it binds contracts in the modules it
	// requires (the import closure declares what it may implement). So a bare slug
	// that missed the carrying module resolves across its required modules.
	if n := g.resolveInRequires(f.Module, f.Target.Slug); n != nil {
		return n
	}
	return nil
}

// resolveInRequires resolves a bare slug against the modules that `mod` requires —
// the path a code module's annotation takes, since the contract it binds lives in a
// required service module, not in the code module itself.
func (g *ResolvedGraph) resolveInRequires(mod model.ModulePath, slug model.Slug) *ResolvedNode {
	info, ok := g.mods[mod]
	if !ok {
		return nil
	}
	for _, req := range info.Requires {
		if n := g.bySlug[req.Module][slug]; n != nil {
			return n
		}
	}
	return nil
}

// bindOne records one fact on its node by verb. req marks implementation,
// test (in a test file) marks proof, infra verbs record an edge proof. A mismatch
// (infra anchor on a non-UC, or an edge the spec doesn't declare) is advisory.
//
//specue:req:validate-graph#unbindable-target
func bindOne(n *ResolvedNode, f CodeFact) *Diagnostic {
	switch f.Verb {
	case VerbReq, VerbTest:
		// req/test implement and prove a logical contract — only a UseCase carries
		// code-bindable elements. A code module may import UseCase/Port (the role-gate
		// on its require), but binding a Need/Domain/Plan/ADR with code is
		// meaningless: those hold no implementable elements. Gate it with the remedy.
		if n.Node().Type != model.TypeUseCase {
			d := newDiag(UnbindableTarget, n.ID(), fmt.Sprintf(
				"//specue:%s binds %s %q, which holds no code — only a UseCase is implementable; point the annotation at the UseCase that realizes it",
				f.Verb, n.Node().Type, n.Node().Slug))
			return &d
		}
		if f.Verb == VerbReq {
			ensureReq(n)
			n.ReqElems[f.Target.Element] = append(n.ReqElems[f.Target.Element], bindingOf(f))
			return nil
		}
		if f.IsTest {
			ensureCover(n)
			n.CoverElems[f.Target.Element] = append(n.CoverElems[f.Target.Element], bindingOf(f))
		}
		return nil
	default:
		return bindInfra(n, f)
	}
}

// bindInfra records an infra-edge proof on a use case, scoped to the element.
func bindInfra(n *ResolvedNode, f CodeFact) *Diagnostic {
	role := f.Verb.Role()
	if role == "" {
		return nil // not an infra verb
	}
	if n.Node().Type != model.TypeUseCase {
		d := newDiag(OrphanBinding, n.ID(),
			fmt.Sprintf("infra anchor %s:%s but %s is not a use case", f.Verb, f.Target.Slug, n.ID().Slug))
		return &d
	}
	if n.InfraProof == nil {
		n.InfraProof = map[InfraKey][]Binding{}
	}
	key := InfraKey{Role: role, Element: f.Target.Element}
	n.InfraProof[key] = append(n.InfraProof[key], bindingOf(f))
	return nil
}

func ensureReq(n *ResolvedNode) {
	if n.ReqElems == nil {
		n.ReqElems = map[model.ElementID][]Binding{}
	}
}

func ensureCover(n *ResolvedNode) {
	if n.CoverElems == nil {
		n.CoverElems = map[model.ElementID][]Binding{}
	}
}

func bindingOf(f CodeFact) Binding {
	return Binding{SourceModule: f.Module, File: f.File, Line: f.Line, IsTest: f.IsTest}
}
