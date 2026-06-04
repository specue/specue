package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// codeModWithReqs assembles a service module holding one UC and a code module that
// requires it, plus the given facts, and returns the resolved graph + diags.
func codeModWithReqs(t *testing.T, facts ...CodeFact) (*ResolvedGraph, []Diagnostic) {
	t.Helper()
	svc := uc("x.test/example@v0", "validate-graph", model.Public)
	codeMod := source.LoadedModule{
		Manifest: source.Manifest{
			Path: "x.test/wallet-code@v0", Kind: source.KindCode,
			Requires: []source.ModuleRequire{{Module: "x.test/example@v0", Version: "v0.1.0"}},
		},
	}
	return New().Compile(Input{
		Modules: []source.LoadedModule{
			loadedMod("x.test/example@v0", source.KindService, []model.PlacedNode{svc}),
			codeMod,
		},
		Facts: facts,
	})
}

func reqFrom(mod model.ModulePath, slug model.Slug, file string) CodeFact {
	return CodeFact{Module: mod, Verb: VerbReq, Target: AnnotationTarget{Slug: slug}, File: model.FilePath(file), Line: 1}
}
func testFrom(mod model.ModulePath, slug model.Slug, file string) CodeFact {
	return CodeFact{Module: mod, Verb: VerbTest, Target: AnnotationTarget{Slug: slug}, File: model.FilePath(file), Line: 1, IsTest: true}
}

func stateOf(v BindingsView, target model.NodeID) BindState {
	for _, r := range v.Rows {
		if r.Target == target && r.Element == "" {
			return r.State
		}
	}
	return ""
}

//specue:test:report-bindings#scoped-to-code-module
func TestBindingsForRejectsNonCodeModule(t *testing.T) {
	g, diags := codeModWithReqs(t)
	_, ok := g.BindingsFor("x.test/example@v0", diags) // a service module
	assert.False(t, ok, "bindings is only for a code module")
}

//specue:test:report-bindings#per-element-state
func TestBindingsStatesAcrossRequires(t *testing.T) {
	target := model.NodeID{Module: "x.test/example@v0", Slug: "validate-graph"}

	t.Run("unbound when no req", func(t *testing.T) {
		g, diags := codeModWithReqs(t)
		v, ok := g.BindingsFor("x.test/wallet-code@v0", diags)
		require.True(t, ok)
		assert.Equal(t, BindUnbound, stateOf(v, target))
	})

	t.Run("bound with req only", func(t *testing.T) {
		g, diags := codeModWithReqs(t, reqFrom("x.test/wallet-code@v0", "validate-graph", "a.go"))
		v, _ := g.BindingsFor("x.test/wallet-code@v0", diags)
		assert.Equal(t, BindBound, stateOf(v, target))
	})

	t.Run("proven with req and test", func(t *testing.T) {
		g, diags := codeModWithReqs(t,
			reqFrom("x.test/wallet-code@v0", "validate-graph", "a.go"),
			testFrom("x.test/wallet-code@v0", "validate-graph", "a_test.go"))
		v, _ := g.BindingsFor("x.test/wallet-code@v0", diags)
		assert.Equal(t, BindProven, stateOf(v, target))
	})

	t.Run("duplicate with two reqs", func(t *testing.T) {
		g, diags := codeModWithReqs(t,
			reqFrom("x.test/wallet-code@v0", "validate-graph", "a.go"),
			reqFrom("x.test/wallet-code@v0", "validate-graph", "b.go"))
		v, _ := g.BindingsFor("x.test/wallet-code@v0", diags)
		assert.Equal(t, BindDuplicate, stateOf(v, target))
	})
}

// kindStateOf returns the state of the whole-contract row of a given kind.
func kindStateOf(v BindingsView, target model.NodeID, kind BindKind) BindState {
	for _, r := range v.Rows {
		if r.Target == target && r.Element == "" && r.Kind == kind {
			return r.State
		}
	}
	return ""
}

// TestBindingsInfraFactAxis pins the fact axis: an infra edge a contract declares
// shows as its own kind row (produce), unbound until the anchor exists, then bound —
// and never proven (a fact needs no test, unlike a req).
func TestBindingsInfraFactAxis(t *testing.T) {
	target := model.NodeID{Module: "x.test/example@v0", Slug: "validate-graph"}
	// A UC that declares a produce edge.
	svc := uc("x.test/example@v0", "validate-graph", model.Public,
		model.Dep{To: model.NodeRef{Module: "x.test/topo@v0", Slug: "report-channel"}, Role: model.RoleProduce})
	codeMod := source.LoadedModule{Manifest: source.Manifest{
		Path: "x.test/wallet-code@v0", Kind: source.KindCode,
		Requires: []source.ModuleRequire{{Module: "x.test/example@v0", Version: "v0.1.0"}},
	}}
	build := func(facts ...CodeFact) BindingsView {
		g, diags := New().Compile(Input{
			Modules: []source.LoadedModule{
				loadedMod("x.test/example@v0", source.KindService, []model.PlacedNode{svc}),
				codeMod,
			},
			Facts: facts,
		})
		v, ok := g.BindingsFor("x.test/wallet-code@v0", diags)
		require.True(t, ok)
		return v
	}

	t.Run("unbound with no anchor", func(t *testing.T) {
		assert.Equal(t, BindUnbound, kindStateOf(build(), target, "produce"))
	})
	t.Run("bound with a produces anchor, never proven", func(t *testing.T) {
		f := CodeFact{Module: "x.test/wallet-code@v0", Verb: VerbProduces,
			Target: AnnotationTarget{Slug: "validate-graph"}, File: "p.go", Line: 1}
		assert.Equal(t, BindBound, kindStateOf(build(f), target, "produce"))
	})
}

// TestBindingsOrphanDedup pins that several annotations to the same dead slug
// collapse into ONE orphan row carrying every location, not a row per annotation.
func TestBindingsOrphanDedup(t *testing.T) {
	g, diags := codeModWithReqs(t,
		reqFrom("x.test/wallet-code@v0", "ghost", "g1.go"),
		reqFrom("x.test/wallet-code@v0", "ghost", "g2.go"))
	v, _ := g.BindingsFor("x.test/wallet-code@v0", diags)

	var orphans []BindingRow
	for _, r := range v.Rows {
		if r.State == BindOrphan {
			orphans = append(orphans, r)
		}
	}
	require.Len(t, orphans, 1, "two annotations to one dead slug = one orphan row")
	assert.Len(t, orphans[0].Locations, 2, "the row carries both locations")
}
