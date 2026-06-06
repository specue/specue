package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

func TestDeriveBranchExclusion(t *testing.T) {
	// caller core-depends on core-dep and branch-depends on branch-dep.
	caller := model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: "caller", Type: model.TypeContract,
		Body: &model.Body{Contract: &model.ContractBody{Elements: []model.Element{
			{Text: "core", Deps: []model.Dep{{To: model.NodeRef{Module: "svc", Slug: "core-dep"}}}},
			{ID: "v", When: "w", Deps: []model.Dep{{To: model.NodeRef{Module: "svc", Slug: "branch-dep"}, Branch: true}}},
		}}},
	}}
	coreDep := uc("svc", "core-dep", model.Public)
	branchDep := uc("svc", "branch-dep", model.Public)

	g, _ := New().Compile(Input{Modules: []source.LoadedModule{
		loadedMod("svc", source.KindService, []model.PlacedNode{caller, coreDep, branchDep}),
	}})

	n, _ := g.Node(model.NodeID{Module: "svc", Slug: "caller"})
	core := model.NodeID{Module: "svc", Slug: "core-dep"}
	branch := model.NodeID{Module: "svc", Slug: "branch-dep"}

	assert.ElementsMatch(t, []model.NodeID{core, branch}, n.Uses, "Uses includes branch deps")
	assert.Equal(t, []model.NodeID{core}, n.CoreUses, "CoreUses excludes branch deps")
}

func TestDeriveSatisfiesAndRealizes(t *testing.T) {
	story := model.PlacedNode{Module: "m", Node: model.Node{
		Slug: "cashout", Type: model.TypeNeed,
		Body: &model.Body{Need: &model.NeedBody{
			Atoms: []model.Atom{{Kind: model.KindFR, ID: "fr-01", Text: "x"}},
		}},
	}}
	impl := model.PlacedNode{Module: "m", Node: model.Node{
		Slug: "do-cashout", Type: model.TypeContract,
		Body: &model.Body{Contract: &model.ContractBody{Elements: []model.Element{
			{Text: "done", Satisfies: []model.AtomRef{{Need: model.NodeID{Module: "m", Slug: "cashout"}, Atom: "fr-01"}}},
		}}},
	}}
	g, _ := New().Compile(Input{Modules: []source.LoadedModule{
		{Manifest: source.Manifest{Path: "m"}, Nodes: []model.PlacedNode{story, impl}},
	}})

	n, _ := g.Node(model.NodeID{Module: "m", Slug: "do-cashout"})
	storyID := model.NodeID{Module: "m", Slug: "cashout"}
	require.Len(t, n.Satisfies, 1)
	assert.Equal(t, AtomAddr{Need: storyID, Atom: "fr-01"}, n.Satisfies[0])
	assert.Equal(t, []model.NodeID{storyID}, n.Realizes, "realizes the story it satisfies an atom of")
}

// TestDeriveSatisfiesAcrossKinds pins that a satisfies edge discharges its atom
// regardless of the invariant's kind or guard — retyping a guarantee
// returns/rejects (and a rejects being conditional, with a When) must not change
// which atoms a Contract covers. This is the coverage-survival guarantee the
// self-spec re-typing relied on (ADR-14).
func TestDeriveSatisfiesAcrossKinds(t *testing.T) {
	need := model.PlacedNode{Module: "m", Node: model.Node{
		Slug: "need", Type: model.TypeNeed,
		Body: &model.Body{Need: &model.NeedBody{
			Atoms: []model.Atom{
				{Kind: model.KindFR, ID: "fr-01", Text: "a"},
				{Kind: model.KindFR, ID: "fr-02", Text: "b"},
			},
		}},
	}}
	needID := model.NodeID{Module: "m", Slug: "need"}
	impl := model.PlacedNode{Module: "m", Node: model.Node{
		Slug: "impl", Type: model.TypeContract,
		Body: &model.Body{Contract: &model.ContractBody{Elements: []model.Element{
			{ID: "r", Kind: model.KindReturns, Text: "returned",
				Satisfies: []model.AtomRef{{Need: needID, Atom: "fr-01"}}},
			{ID: "x", Kind: model.KindRejects, When: "bad input", Text: "refused",
				Satisfies: []model.AtomRef{{Need: needID, Atom: "fr-02"}}},
		}}},
	}}
	g, _ := New().Compile(Input{Modules: []source.LoadedModule{
		{Manifest: source.Manifest{Path: "m"}, Nodes: []model.PlacedNode{need, impl}},
	}})

	n, _ := g.Node(model.NodeID{Module: "m", Slug: "impl"})
	assert.ElementsMatch(t, []AtomAddr{
		{Need: needID, Atom: "fr-01"},
		{Need: needID, Atom: "fr-02"},
	}, n.Satisfies, "both a returns and a guarded rejects invariant discharge their atoms")
	assert.Equal(t, []model.NodeID{needID}, n.Realizes)
}

func TestDeriveTopology(t *testing.T) {
	port := model.PlacedNode{Module: "topo", Node: model.Node{
		Slug: "report-channel", Type: model.TypePort,
		Body: &model.Body{Port: &model.PortBody{Kind: model.PortChannel, Transport: "kafka"}},
	}}
	producer := model.PlacedNode{Module: "topo", Node: model.Node{
		Slug: "describe-node", Type: model.TypeContract,
		Body: &model.Body{Contract: &model.ContractBody{Elements: []model.Element{
			{Text: "queued", Deps: []model.Dep{
				{To: model.NodeRef{Module: "topo", Slug: "report-channel"}, Role: model.RoleProduce}}},
		}}},
	}}
	g, _ := New().Compile(Input{Modules: []source.LoadedModule{
		{Manifest: source.Manifest{Path: "topo"}, Nodes: []model.PlacedNode{port, producer}},
	}})

	p, _ := g.Node(model.NodeID{Module: "topo", Slug: "report-channel"})
	assert.Equal(t, []model.NodeID{{Module: "topo", Slug: "describe-node"}}, p.Topology.ProducedBy)
	assert.Empty(t, p.Topology.ConsumedBy)
}
