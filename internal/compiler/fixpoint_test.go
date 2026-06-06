package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// ucWith builds a Contract node with a chosen interaction and core deps.
func ucWith(slug model.Slug, interaction model.Interaction, deps ...model.Slug) model.PlacedNode {
	var ds []model.Dep
	for _, d := range deps {
		ds = append(ds, model.Dep{To: model.NodeRef{Module: "svc", Slug: d}})
	}
	return model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: slug, Type: model.TypeContract,
		Body: &model.Body{Contract: &model.ContractBody{
			Interaction: interaction,
			Elements:    []model.Element{{Text: "x", Deps: ds}},
		}},
	}}
}

//specue:test:validate-graph#sync-cycle
func TestSyncCycleIsGate(t *testing.T) {
	// a→b→a, b is sync → the cycle is a gate.
	a := ucWith("a", model.InteractionAsync, "b")
	b := ucWith("b", model.InteractionSync, "a")
	_, diags := New().Compile(Input{Modules: []source.LoadedModule{
		loadedMod("svc", source.KindService, []model.PlacedNode{a, b}),
	}})
	assert.Contains(t, codesOf(diags), SyncCycle)
	assert.NotContains(t, codesOf(diags), AsyncCycle)
}

func TestAsyncCycleIsAdvisory(t *testing.T) {
	// a→b→a, both async → tolerated choreography.
	a := ucWith("a", model.InteractionAsync, "b")
	b := ucWith("b", model.InteractionAsync, "a")
	_, diags := New().Compile(Input{Modules: []source.LoadedModule{
		loadedMod("svc", source.KindService, []model.PlacedNode{a, b}),
	}})
	codes := codesOf(diags)
	assert.Contains(t, codes, AsyncCycle)
	assert.NotContains(t, codes, SyncCycle)
	// And an advisory never makes the graph red.
	for _, d := range diags {
		if d.Code == AsyncCycle {
			assert.Equal(t, Advisory, d.Severity())
		}
	}
}

func TestNoCycleNoDiagnostic(t *testing.T) {
	a := ucWith("a", model.InteractionAsync, "b")
	b := ucWith("b", model.InteractionAsync)
	_, diags := New().Compile(Input{Modules: []source.LoadedModule{
		loadedMod("svc", source.KindService, []model.PlacedNode{a, b}),
	}})
	codes := codesOf(diags)
	assert.NotContains(t, codes, SyncCycle)
	assert.NotContains(t, codes, AsyncCycle)
}

// blockedSetup compiles a graph, then stubs statuses and runs propagateBlocked
// directly (the status pass isn't wired yet).
func TestBlockedPropagation(t *testing.T) {
	// ready→gap: ready implemented, depends on an asserted (unbuilt) node.
	readyN := ucWith("ready", model.InteractionAsync, "gap")
	gap := ucWith("gap", model.InteractionAsync)
	g, _ := New().Compile(Input{Modules: []source.LoadedModule{
		loadedMod("svc", source.KindService, []model.PlacedNode{readyN, gap}),
	}})

	mustNode(t, g, "ready").Status = StatusImplemented
	mustNode(t, g, "gap").Status = StatusAsserted
	propagateBlocked(g)

	assert.Equal(t, StatusBlocked, mustNode(t, g, "ready").Status, "ready blocks on an asserted core dep")
	assert.True(t, mustNode(t, g, "ready").Blocked)
}

func TestBlockedAbstractDoesNotBlock(t *testing.T) {
	// A dep with an abstract binding is deliverable by design — no block.
	readyN := ucWith("ready", model.InteractionAsync, "concept")
	concept := ucWith("concept", model.InteractionAsync)
	concept.Node.Body.Contract.Binding = model.BindingAbstract
	g, _ := New().Compile(Input{Modules: []source.LoadedModule{
		loadedMod("svc", source.KindService, []model.PlacedNode{readyN, concept}),
	}})

	mustNode(t, g, "ready").Status = StatusImplemented
	mustNode(t, g, "concept").Status = StatusAsserted
	propagateBlocked(g)

	assert.NotEqual(t, StatusBlocked, mustNode(t, g, "ready").Status, "abstract dep never blocks")
}

func TestBranchDepDoesNotBlock(t *testing.T) {
	// A branch dep on an asserted node must NOT block the parent (CoreUses excludes it).
	ready := model.PlacedNode{Module: "svc", Node: model.Node{
		Slug: "ready", Type: model.TypeContract,
		Body: &model.Body{Contract: &model.ContractBody{Elements: []model.Element{
			{ID: "v", When: "w", Deps: []model.Dep{{To: model.NodeRef{Module: "svc", Slug: "gap"}, Branch: true}}},
		}}},
	}}
	gap := ucWith("gap", model.InteractionAsync)
	g, _ := New().Compile(Input{Modules: []source.LoadedModule{
		loadedMod("svc", source.KindService, []model.PlacedNode{ready, gap}),
	}})

	mustNode(t, g, "ready").Status = StatusImplemented
	mustNode(t, g, "gap").Status = StatusAsserted
	propagateBlocked(g)

	assert.NotEqual(t, StatusBlocked, mustNode(t, g, "ready").Status, "branch dep is excluded from CoreUses")
}

func mustNode(t *testing.T, g *ResolvedGraph, slug model.Slug) *ResolvedNode {
	t.Helper()
	n, ok := g.Node(model.NodeID{Module: "svc", Slug: slug})
	require.True(t, ok, "node %s present", slug)
	return n
}
