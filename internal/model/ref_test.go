package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeRef(t *testing.T) {
	// A resolved ref carries its target's full address.
	ref := NodeRef{Module: "specue/example", Slug: "validate-graph"}
	assert.Equal(t, "specue/example:validate-graph", ref.String())
}

func TestAtomRef(t *testing.T) {
	ref := AtomRef{Need: NodeID{Module: "specue/example", Slug: "cashout"}, Atom: "fr-01"}
	assert.Equal(t, "specue/example:cashout#fr-01", ref.String())
}

func TestDepIsInfra(t *testing.T) {
	assert.False(t, Dep{To: NodeRef{Slug: "x"}}.IsInfra(), "dep without role is not infra")
	assert.True(t, Dep{To: NodeRef{Slug: "report-channel"}, Role: RoleProduce}.IsInfra(), "dep with role is infra")
}

func TestElementNamed(t *testing.T) {
	assert.False(t, Element{}.Named(), "unnamed element is not Named")
	assert.True(t, Element{ID: "win-limit"}.Named(), "element with ID is Named")
}
