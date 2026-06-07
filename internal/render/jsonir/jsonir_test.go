package jsonir_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
	"github.com/specue/specue/internal/render/jsonir"
	"github.com/specue/specue/internal/source"
)

// TestJSONIRRendersTreeAndIndex builds a minimal graph (one Need with a single
// FR, one Contract that satisfies it, one whole-contract req binding) and
// asserts the JSON IR renderer lays the tree out as index.json plus one .json
// file per node, with the expected shape on each file.
func TestJSONIRRendersTreeAndIndex(t *testing.T) {
	svc := model.ModulePath("ex.test/svc@v0")
	prod := model.ModulePath("ex.test/dom@v0")

	storyRef := model.NodeRef{Module: prod, Slug: "as-user"}
	svcRef := model.NodeRef{Module: svc, Slug: "service"}

	contract := model.PlacedNode{
		Module: svc,
		File:   "uc.cue",
		Node: model.Node{
			Slug: "do-thing", Type: model.TypeContract, Title: "Do the thing",
			Confidence: model.Confirmed,
			Body: &model.Body{Contract: &model.ContractBody{
				Service:     svcRef,
				Interaction: model.InteractionSync,
				Trigger:     "caller asks",
				Elements: []model.Element{
					{ID: "single-verdict",
						Text:      "A repeat is a no-op.",
						Satisfies: []model.AtomRef{{Need: storyRef, Atom: "fr-01"}}},
				},
			}},
		},
	}
	story := model.PlacedNode{
		Module: prod,
		Node: model.Node{
			Slug: "as-user", Type: model.TypeNeed, Title: "As a user",
			Confidence: model.Confirmed,
			Body: &model.Body{Need: &model.NeedBody{
				Domain:   model.NodeRef{Module: prod, Slug: "domain"},
				Consumer: "user",
				Atoms:    []model.Atom{{ID: "fr-01", Kind: model.KindFR, Text: "Atomic."}},
			}},
		},
	}
	domain := model.PlacedNode{
		Module: prod,
		Node: model.Node{
			Slug: "domain", Type: model.TypeDomain, Title: "Example domain",
			Confidence: model.Confirmed, Body: &model.Body{},
		},
	}

	// One whole-contract //req binding on the Contract, so the renderer emits
	// a bindings.req entry with no `element` key.
	fact := compiler.CodeFact{
		Module: svc,
		File:   "do.go",
		Line:   42,
		Verb:   compiler.VerbReq,
		Target: compiler.AnnotationTarget{Slug: "do-thing"},
	}

	c := compiler.New()
	g, _ := c.Compile(compiler.Input{
		Modules: []source.LoadedModule{
			{Manifest: source.Manifest{Path: svc, Kind: source.KindService}, Nodes: []model.PlacedNode{contract}},
			{Manifest: source.Manifest{Path: prod, Kind: source.KindDomain}, Nodes: []model.PlacedNode{story, domain}},
		},
		Facts: []compiler.CodeFact{fact},
	})
	require.NotNil(t, g)

	revs := map[model.ModulePath]string{svc: "abc123def456", prod: "abc123def456"}
	tree, err := jsonir.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	// Shape: index.json + one file per authored node.
	require.Contains(t, tree, jsonir.IndexPath, "index.json present")
	ucPath := render.RelPath("ex.test-svc-v0/do-thing.json")
	needPath := render.RelPath("ex.test-dom-v0/as-user.json")
	domainPath := render.RelPath("ex.test-dom-v0/domain.json")
	require.Contains(t, tree, ucPath)
	require.Contains(t, tree, needPath)
	require.Contains(t, tree, domainPath)
	assert.Len(t, tree, 4, "index + 3 nodes")

	// Each file parses as JSON.
	parse := func(path render.RelPath) map[string]any {
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(tree[path]), &m), "parse %s", path)
		return m
	}

	idx := parse(jsonir.IndexPath)
	assert.Equal(t, "abc123def456", idx["rendered_from"])
	mods, _ := idx["modules"].([]any)
	assert.Len(t, mods, 2)
	nodes, _ := idx["nodes"].([]any)
	assert.Len(t, nodes, 3)

	uc := parse(ucPath)
	assert.Equal(t, "ex.test/svc@v0:do-thing", uc["id"])
	assert.Equal(t, "Contract", uc["type"])
	assert.Equal(t, "ex.test/svc@v0:service", uc["service"])
	assert.Equal(t, "sync", uc["interaction"])
	assert.Equal(t, "caller asks", uc["trigger"])
	invs, ok := uc["invariants"].([]any)
	require.True(t, ok, "invariants array present")
	require.Len(t, invs, 1)
	inv0, _ := invs[0].(map[string]any)
	assert.Equal(t, "single-verdict", inv0["id"])
	_, hasKind := inv0["kind"]
	assert.False(t, hasKind, "a plain invariant omits kind")
	sat, _ := inv0["satisfies"].([]any)
	require.Len(t, sat, 1)
	sat0, _ := sat[0].(map[string]any)
	assert.Equal(t, "ex.test/dom@v0:as-user", sat0["need"])
	assert.Equal(t, "fr-01", sat0["atom"])

	// derived.realizes points at the satisfied Need.
	derived, _ := uc["derived"].(map[string]any)
	require.NotNil(t, derived, "derived present")
	real, _ := derived["realizes"].([]any)
	assert.Contains(t, real, "ex.test/dom@v0:as-user")

	// bindings.req carries the whole-contract code binding (no `element`).
	bindings, _ := uc["bindings"].(map[string]any)
	require.NotNil(t, bindings, "bindings present")
	req, _ := bindings["req"].([]any)
	require.Len(t, req, 1)
	req0, _ := req[0].(map[string]any)
	_, hasElem := req0["element"]
	assert.False(t, hasElem, "whole-contract binding omits element")
	assert.Equal(t, "do.go:42", req0["loc"])
	assert.Equal(t, "ex.test/svc@v0", req0["source_module"])

	need := parse(needPath)
	assert.Equal(t, "Need", need["type"])
	frs, _ := need["frs"].([]any)
	require.Len(t, frs, 1)
	fr0, _ := frs[0].(map[string]any)
	assert.Equal(t, "fr-01", fr0["id"])
	assert.Equal(t, "fr", fr0["kind"])
}
