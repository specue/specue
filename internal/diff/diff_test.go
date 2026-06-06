package diff_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/diff"
	"github.com/specue/specue/internal/model"
)

const mod = model.ModulePath("svc")

func uc(slug model.Slug, title string, els ...model.Element) model.PlacedNode {
	return model.PlacedNode{Module: mod, Node: model.Node{
		Slug: slug, Type: model.TypeContract, Title: title,
		Body: &model.Body{Contract: &model.ContractBody{Elements: els}},
	}}
}

func inv(id model.ElementID, text string, deps ...model.Dep) model.Element {
	return model.Element{ID: id, Text: text, Deps: deps}
}

func dep(to model.Slug, role model.Role) model.Dep {
	return model.Dep{To: model.NodeID{Module: mod, Slug: to}, Role: role}
}

func nodeOf(t *testing.T, d diff.Delta, slug model.Slug) diff.NodeDelta {
	t.Helper()
	for _, n := range d.Nodes {
		if n.ID.Slug == slug {
			return n
		}
	}
	t.Fatalf("no delta for %s", slug)
	return diff.NodeDelta{}
}

//specue:test:diff-refs#typed-over-the-spec-graph
//specue:test:diff-refs#every-change-named
func TestAddedRemovedNodes(t *testing.T) {
	a := []model.PlacedNode{uc("keep", "Keep"), uc("gone", "Gone")}
	b := []model.PlacedNode{uc("keep", "Keep"), uc("fresh", "Fresh")}

	d := diff.Compute(a, b)
	require.Len(t, d.Nodes, 2)
	assert.Equal(t, diff.Added, nodeOf(t, d, "fresh").Change)
	assert.Equal(t, diff.Removed, nodeOf(t, d, "gone").Change)
}

func TestIdenticalIsEmpty(t *testing.T) {
	a := []model.PlacedNode{uc("x", "X", inv("i1", "holds"))}
	b := []model.PlacedNode{uc("x", "X", inv("i1", "holds"))}
	assert.True(t, diff.Compute(a, b).Empty(), "identical snapshots → no delta")
}

func TestModifiedTitle(t *testing.T) {
	a := []model.PlacedNode{uc("x", "Old title")}
	b := []model.PlacedNode{uc("x", "New title")}
	nd := nodeOf(t, diff.Compute(a, b), "x")
	assert.Equal(t, diff.Modified, nd.Change)
	assert.Empty(t, nd.Elements)
	assert.Empty(t, nd.Edges)
}

func TestElementAddedRemovedModified(t *testing.T) {
	a := []model.PlacedNode{uc("x", "X", inv("keep", "k"), inv("drop", "d"), inv("edit", "before"))}
	b := []model.PlacedNode{uc("x", "X", inv("keep", "k"), inv("edit", "after"), inv("new", "n"))}

	nd := nodeOf(t, diff.Compute(a, b), "x")
	require.Equal(t, diff.Modified, nd.Change)
	got := map[model.ElementID]diff.Change{}
	for _, e := range nd.Elements {
		got[e.ID] = e.Change
	}
	assert.Equal(t, diff.Added, got["new"])
	assert.Equal(t, diff.Removed, got["drop"])
	assert.Equal(t, diff.Modified, got["edit"])
	assert.NotContains(t, got, model.ElementID("keep"), "unchanged element absent from delta")
}

// TestElementKindOrWhenChangeIsModified proves the element signature tracks the
// invariant fields (ADR-14): changing only an invariant's kind (nature) or its when guard —
// id and text unchanged — still marks the element Modified.
func TestElementKindOrWhenChangeIsModified(t *testing.T) {
	plain := model.Element{ID: "i", Text: "t"}
	typed := model.Element{ID: "i", Text: "t", Kind: model.KindRejects, When: "bad input"}

	a := []model.PlacedNode{uc("x", "X", plain)}
	b := []model.PlacedNode{uc("x", "X", typed)}

	nd := nodeOf(t, diff.Compute(a, b), "x")
	require.Equal(t, diff.Modified, nd.Change)
	require.Len(t, nd.Elements, 1)
	assert.Equal(t, model.ElementID("i"), nd.Elements[0].ID)
	assert.Equal(t, diff.Modified, nd.Elements[0].Change,
		"kind/when are part of the element signature")
}

func TestEdgeRewire(t *testing.T) {
	a := []model.PlacedNode{uc("x", "X", inv("i", "t", dep("old", "call")))}
	b := []model.PlacedNode{uc("x", "X", inv("i", "t", dep("new", "call")))}

	nd := nodeOf(t, diff.Compute(a, b), "x")
	require.Equal(t, diff.Modified, nd.Change)
	// A retargeted edge is a remove of the old + add of the new.
	var added, removed model.Slug
	for _, e := range nd.Edges {
		switch e.Change {
		case diff.Added:
			added = e.To.Slug
		case diff.Removed:
			removed = e.To.Slug
		}
	}
	assert.Equal(t, model.Slug("new"), added)
	assert.Equal(t, model.Slug("old"), removed)
}
