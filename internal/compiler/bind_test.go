package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// TestReqOnNeedIsUnbindable pins the import-role half of the code-module
// contract: a code annotation may resolve to a Contract (implementable) but binding
// a Need — which holds no code — is a gate, with the remedy pointing at the
// realizing Contract. This is what stops a code module from "implementing" a product
// story directly.
//
//specue:test:validate-graph#unbindable-target
func TestReqOnNeedIsUnbindable(t *testing.T) {
	story := model.PlacedNode{
		Module: "x.test/domain@v0",
		Node: model.Node{
			Slug: "describe-node", Type: model.TypeNeed, Visibility: model.Public,
			Body: &model.Body{Need: &model.NeedBody{}},
		},
	}
	fact := CodeFact{Module: "x.test/domain@v0", Verb: VerbReq,
		Target: AnnotationTarget{Slug: "describe-node"}, File: "x.go", Line: 1}

	_, diags := New().Compile(Input{
		Modules: []source.LoadedModule{loadedMod("x.test/domain@v0", source.KindDomain, []model.PlacedNode{story})},
		Facts:   []CodeFact{fact},
	})

	assert.Contains(t, codesOf(diags), UnbindableTarget)
	var msg string
	for _, d := range diags {
		if d.Code == UnbindableTarget {
			msg = d.Message
		}
	}
	assert.Contains(t, msg, "Contract", "the fix names the implementable type")
}

// TestReqAcrossRequiresResolves pins that a code module's bare annotation resolves
// against a module it requires — the contract it implements lives in the required
// service module, not in the (node-free) code module.
func TestReqAcrossRequiresResolves(t *testing.T) {
	svc := contract("x.test/example@v0", "validate-graph", model.Public)
	codeMod := source.LoadedModule{
		Manifest: source.Manifest{
			Path: "x.test/wallet-code@v0", Kind: source.KindCode,
			Requires: []source.ModuleRequire{{Module: "x.test/example@v0", Version: "v0.1.0"}},
		},
	}
	fact := CodeFact{Module: "x.test/wallet-code@v0", Verb: VerbReq,
		Target: AnnotationTarget{Slug: "validate-graph"}, File: "apply.go", Line: 1}

	g, diags := New().Compile(Input{
		Modules: []source.LoadedModule{
			loadedMod("x.test/example@v0", source.KindService, []model.PlacedNode{svc}),
			codeMod,
		},
		Facts: []CodeFact{fact},
	})

	assert.NotContains(t, codesOf(diags), OrphanBinding, "a bare req resolves via the code module's requires")
	n, ok := g.Node(model.NodeID{Module: "x.test/example@v0", Slug: "validate-graph"})
	assert.True(t, ok)
	assert.Equal(t, StatusImplemented, n.Status, "the required module's Contract is implemented by the code module")
}
