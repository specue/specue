package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlacedNodeID(t *testing.T) {
	p := PlacedNode{
		Module: "specue/example",
		File:   "spec/validate-graph.cue",
		Node:   Node{Slug: "validate-graph", Type: TypeUseCase},
	}
	id := p.ID()
	assert.Equal(t, ModulePath("specue/example"), id.Module)
	assert.Equal(t, Slug("validate-graph"), id.Slug)
	assert.Equal(t, "specue/example:validate-graph", id.String())
}

func TestNodeIDIdentity(t *testing.T) {
	// Same slug in different modules are distinct identities.
	a := NodeID{Module: "example", Slug: "apply"}
	b := NodeID{Module: "consumer", Slug: "apply"}
	assert.NotEqual(t, a, b)
	assert.Equal(t, a, NodeID{Module: "example", Slug: "apply"})
}
