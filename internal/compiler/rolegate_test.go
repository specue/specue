package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// TestRoleGateCodeModuleHoldsNoNodes pins the code-kind invariant: a code module is
// manifest + require only, so any spec node authored in it is a gate violation, and
// the message says "no spec nodes" rather than naming the type.
//
//specue:test:validate-graph#role-gate
func TestRoleGateCodeModuleHoldsNoNodes(t *testing.T) {
	mod := loadedMod("x.test/wallet-code@v0", source.KindCode,
		[]model.PlacedNode{uc("x.test/wallet-code@v0", "validate-graph", model.Public)})

	g, diags := New().Compile(Input{Modules: []source.LoadedModule{mod}})

	assert.Contains(t, codesOf(diags), RoleGateViolation)
	var msg string
	for _, d := range diags {
		if d.Code == RoleGateViolation {
			msg = d.Message
		}
	}
	assert.Contains(t, msg, "no spec nodes", "the code-module message states the real rule")

	n, ok := g.Node(model.NodeID{Module: "x.test/wallet-code@v0", Slug: "validate-graph"})
	require.True(t, ok)
	assert.Equal(t, StatusBroken, n.Status, "a node in a code module is broken")
}
