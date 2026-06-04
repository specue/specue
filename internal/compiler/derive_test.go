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
		Slug: "caller", Type: model.TypeUseCase,
		Body: &model.Body{UseCase: &model.UseCaseBody{Elements: []model.Element{
			{Kind: model.KindPost, Text: "core", Deps: []model.Dep{{To: model.NodeRef{Module: "svc", Slug: "core-dep"}}}},
			{Kind: model.KindVariation, ID: "v", When: "w", Then: "t",
				Deps: []model.Dep{{To: model.NodeRef{Module: "svc", Slug: "branch-dep"}, Branch: true}}},
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
		Slug: "do-cashout", Type: model.TypeUseCase,
		Body: &model.Body{UseCase: &model.UseCaseBody{Elements: []model.Element{
			{Kind: model.KindPost, Text: "done", Satisfies: []model.AtomRef{{Need: model.NodeID{Module: "m", Slug: "cashout"}, Atom: "fr-01"}}},
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

func TestDeriveTopology(t *testing.T) {
	port := model.PlacedNode{Module: "topo", Node: model.Node{
		Slug: "report-channel", Type: model.TypePort,
		Body: &model.Body{Port: &model.PortBody{Kind: model.PortChannel, Transport: "kafka"}},
	}}
	producer := model.PlacedNode{Module: "topo", Node: model.Node{
		Slug: "describe-node", Type: model.TypeUseCase,
		Body: &model.Body{UseCase: &model.UseCaseBody{Elements: []model.Element{
			{Kind: model.KindPost, Text: "queued", Deps: []model.Dep{
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
