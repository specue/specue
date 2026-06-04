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

func TestUseCaseStatusTiers(t *testing.T) {
	asserted := uc("svc", "asserted", model.Public)
	implemented := uc("svc", "implemented", model.Public)
	proven := uc("svc", "proven", model.Public)

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
		Slug: "uc1", Type: model.TypeUseCase,
		Body: &model.Body{UseCase: &model.UseCaseBody{Elements: []model.Element{
			{Kind: model.KindPost, Text: "done",
				Satisfies: []model.AtomRef{{Need: model.NodeID{Module: "svc", Slug: "tale"}, Atom: "fr-01"}},
				Deps:      []model.Dep{{To: model.NodeRef{Module: "svc", Slug: "gap"}}}},
		}}},
	}}
	gap := uc("svc", "gap", model.Public)

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

// ucSat builds a UseCase whose postcondition satisfies story#atom.
func ucSat(slug model.Slug, story model.Slug, atom model.AtomID) model.PlacedNode {
	return model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: slug, Type: model.TypeUseCase,
		Body: &model.Body{UseCase: &model.UseCaseBody{Elements: []model.Element{
			{Kind: model.KindPost, Text: "x", Satisfies: []model.AtomRef{{Need: model.NodeID{Module: "svc", Slug: story}, Atom: atom}}},
		}}},
	}}
}
