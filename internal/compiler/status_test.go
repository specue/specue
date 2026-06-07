package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// req/cover build code facts for a slug in module svc.
func req(slug model.Slug) CodeFact {
	return CodeFact{Module: "svc", Verb: VerbReq, Target: AnnotationTarget{Slug: slug}, File: "x.go", Line: 1}
}
func cover(slug model.Slug) CodeFact {
	return CodeFact{Module: "svc", Verb: VerbTest, Target: AnnotationTarget{Slug: slug}, File: "x_test.go", Line: 1, IsTest: true}
}

func statusOf(g *ResolvedGraph, slug model.Slug) ResolvedNodeStatus {
	n, _ := g.Node(model.NodeID{Module: "svc", Slug: slug})
	return n.Status
}

// reqElem/coverElem build element-scoped code facts (//req:slug#element).
func reqElem(slug model.Slug, elem model.ElementID) CodeFact {
	return CodeFact{Module: "svc", Verb: VerbReq, Target: AnnotationTarget{Slug: slug, Element: elem}, File: "x.go", Line: 1}
}
func coverElem(slug model.Slug, elem model.ElementID) CodeFact {
	return CodeFact{Module: "svc", Verb: VerbTest, Target: AnnotationTarget{Slug: slug, Element: elem}, File: "x_test.go", Line: 1, IsTest: true}
}

// TestGuardedElementNeedsScopedBinding pins the isGuarded rule (ADR-14): a
// guarded invariant (one with a When — e.g. a rejects) is a conditional branch,
// so a whole-contract //req does NOT auto-cover it; it needs its own scoped
// binding. The atom such an element satisfies stays uncovered until the scoped
// bind exists. (The rule keys on When != "" — formerly a separate variation kind.)
func TestGuardedElementNeedsScopedBinding(t *testing.T) {
	need := model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: "need", Type: model.TypeNeed,
		Body: &model.Body{Need: &model.NeedBody{Atoms: []model.Atom{
			{Kind: model.KindFR, ID: "fr-01", Text: "a"},
		}}},
	}}
	needID := model.NodeID{Module: "svc", Slug: "need"}
	// The contract's only satisfier of fr-01 is a guarded rejects invariant.
	c := model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: "uc", Type: model.TypeContract,
		Body: &model.Body{Contract: &model.ContractBody{Elements: []model.Element{
			{ID: "guard", Kind: model.KindRejects, When: "bad input", Text: "refused",
				Satisfies: []model.AtomRef{{Need: needID, Atom: "fr-01"}}},
		}}},
	}}
	mods := []source.LoadedModule{
		{Manifest: source.Manifest{Path: "svc"}, Nodes: []model.PlacedNode{need, c}},
	}

	// Whole-contract req+cover only: the guarded element is NOT auto-covered.
	g, _ := New().Compile(Input{Modules: mods, Facts: []CodeFact{req("uc"), cover("uc")}})
	nd, _ := g.Node(needID)
	assert.Equal(t, StatusUncovered, nd.Status,
		"a whole-contract bind does not cover a guarded invariant's atom")

	// Scoped bind on the guarded element: now the atom is proven.
	g2, _ := New().Compile(Input{Modules: mods,
		Facts: []CodeFact{reqElem("uc", "guard"), coverElem("uc", "guard")}})
	nd2, _ := g2.Node(needID)
	assert.Equal(t, StatusCovered, nd2.Status,
		"a scoped bind on the guarded element covers its atom")
}

func TestContractStatusTiers(t *testing.T) {
	asserted := contract("svc", "asserted", model.Public)
	implemented := contract("svc", "implemented", model.Public)
	proven := contract("svc", "proven", model.Public)

	g, _ := New().Compile(Input{
		Modules: []source.LoadedModule{loadedMod("svc", source.KindService,
			[]model.PlacedNode{asserted, implemented, proven})},
		Facts: []CodeFact{
			req("implemented"),
			req("proven"), cover("proven"),
		},
	})

	assert.Equal(t, StatusAsserted, statusOf(g, "asserted"), "no code → asserted (GAP)")
	assert.Equal(t, StatusImplemented, statusOf(g, "implemented"), "req only → implemented")
	assert.Equal(t, StatusProven, statusOf(g, "proven"), "req + covering test → proven")
}

func TestStoryCoverageTiers(t *testing.T) {
	story := model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: "tale", Type: model.TypeNeed,
		Body: &model.Body{Need: &model.NeedBody{Atoms: []model.Atom{
			{Kind: model.KindFR, ID: "fr-01", Text: "a"},
			{Kind: model.KindFR, ID: "fr-02", Text: "b"},
		}}},
	}}
	// uc1 satisfies+proves fr-01; uc2 satisfies fr-02 but only implemented.
	uc1 := ucSat("uc1", "tale", "fr-01")
	uc2 := ucSat("uc2", "tale", "fr-02")

	g, _ := New().Compile(Input{
		Modules: []source.LoadedModule{
			{Manifest: source.Manifest{Path: "svc"}, Nodes: []model.PlacedNode{story, uc1, uc2}},
		},
		Facts: []CodeFact{req("uc1"), cover("uc1"), req("uc2")},
	})

	// fr-01 proven, fr-02 only covered → partial (not all proven).
	tale, _ := g.Node(model.NodeID{Module: "svc", Slug: "tale"})
	assert.Equal(t, StatusPartial, tale.Status)
}

func TestStoryDeliveredWhenAllProven(t *testing.T) {
	story := model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: "tale", Type: model.TypeNeed,
		Body: &model.Body{Need: &model.NeedBody{Atoms: []model.Atom{
			{Kind: model.KindFR, ID: "fr-01", Text: "a"},
		}}},
	}}
	impl := ucSat("uc1", "tale", "fr-01")
	g, _ := New().Compile(Input{
		Modules: []source.LoadedModule{
			{Manifest: source.Manifest{Path: "svc"}, Nodes: []model.PlacedNode{story, impl}},
		},
		Facts: []CodeFact{req("uc1"), cover("uc1")},
	})
	tale, _ := g.Node(model.NodeID{Module: "svc", Slug: "tale"})
	assert.Equal(t, StatusCovered, tale.Status, "all atoms proven → delivered")
}

func TestStoryOrphanWhenNoCoverage(t *testing.T) {
	story := model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: "tale", Type: model.TypeNeed,
		Body: &model.Body{Need: &model.NeedBody{Atoms: []model.Atom{
			{Kind: model.KindFR, ID: "fr-01", Text: "a"},
		}}},
	}}
	impl := ucSat("uc1", "tale", "fr-01") // satisfies but no code
	g, _ := New().Compile(Input{
		Modules: []source.LoadedModule{
			{Manifest: source.Manifest{Path: "svc"}, Nodes: []model.PlacedNode{story, impl}},
		},
	})
	tale, _ := g.Node(model.NodeID{Module: "svc", Slug: "tale"})
	assert.Equal(t, StatusUncovered, tale.Status, "asserted satisfier doesn't cover")
}

func TestBlockedSatisfierDoesNotDeliver(t *testing.T) {
	story := model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: "tale", Type: model.TypeNeed,
		Body: &model.Body{Need: &model.NeedBody{Atoms: []model.Atom{
			{Kind: model.KindFR, ID: "fr-01", Text: "a"},
		}}},
	}}
	// uc1 satisfies fr-01, is implemented+tested, but core-depends on an asserted gap.
	uc1 := model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: "uc1", Type: model.TypeContract,
		Body: &model.Body{Contract: &model.ContractBody{Elements: []model.Element{
			{Text: "done",
				Satisfies: []model.AtomRef{{Need: model.NodeID{Module: "svc", Slug: "tale"}, Atom: "fr-01"}},
				Deps:      []model.Dep{{To: model.NodeRef{Module: "svc", Slug: "gap"}}}},
		}}},
	}}
	gap := contract("svc", "gap", model.Public)

	g, _ := New().Compile(Input{
		Modules: []source.LoadedModule{
			{Manifest: source.Manifest{Path: "svc"}, Nodes: []model.PlacedNode{story, uc1, gap}},
		},
		Facts: []CodeFact{req("uc1"), cover("uc1")},
	})

	assert.Equal(t, StatusBlocked, statusOf(g, "uc1"), "uc1 blocks on asserted gap")
	tale, _ := g.Node(model.NodeID{Module: "svc", Slug: "tale"})
	assert.Equal(t, StatusUncovered, tale.Status, "a blocked satisfier does not deliver the atom")
}

// ucSat builds a Contract whose postcondition satisfies story#atom.
func ucSat(slug model.Slug, story model.Slug, atom model.AtomID) model.PlacedNode {
	return model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: slug, Type: model.TypeContract,
		Body: &model.Body{Contract: &model.ContractBody{Elements: []model.Element{
			{Text: "x", Satisfies: []model.AtomRef{{Need: model.NodeID{Module: "svc", Slug: story}, Atom: atom}}},
		}}},
	}}
}
