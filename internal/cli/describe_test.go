package cli

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const applyOp = "specue.test/example@v0:validate-graph"

//specue:test:describe-node#shown-in-full
func TestDescribeFullNode(t *testing.T) {
	out, _, code := run("describe", applyOp, "-C", walletSpec)
	require.Equal(t, exitOK, code)
	assert.Contains(t, out, "Contract")
	assert.Contains(t, out, "service")
	assert.Contains(t, out, "single-verdict", "named invariant is shown")
}

func TestDescribeJSONCarriesTypeAndStatus(t *testing.T) {
	out, _, code := run("describe", applyOp, "-C", walletSpec, "--json")
	require.Equal(t, exitOK, code)

	var got struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, applyOp, got.ID)
	assert.Equal(t, "Contract", got.Type)
	assert.NotEmpty(t, got.Status)
}

//specue:test:describe-node#identity-is-module-qualified
func TestDescribeBareSlugRejected(t *testing.T) {
	_, errOut, code := run("describe", "validate-graph", "-C", walletSpec)
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "module:slug", "the fix names the required form")
	assert.Contains(t, errOut, "try:")
}

func TestDescribeMissingNodeIsActionable(t *testing.T) {
	_, errOut, code := run("describe", "specue.test/example@v0:nope", "-C", walletSpec)
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
}

// TestDescribeElementScoped covers describe-node#element-scoped: a #element
// suffix narrows the output to one named invariant, dropping the node-level
// service/trigger header and the other elements.
//
//specue:test:describe-node#element-scoped
func TestDescribeElementScoped(t *testing.T) {
	out, _, code := run("describe", applyOp+"#single-verdict", "-C", walletSpec)
	require.Equal(t, exitOK, code)
	assert.Contains(t, out, "single-verdict")
	assert.NotContains(t, out, "trigger:", "node-level header is dropped for an element view")
}

// TestDescribeUnknownElementIsActionable covers the fix-suggestion path: the
// node is real but the suffix names no element, so the error tells the caller
// to drop the suffix to see every element.
//
//specue:test:describe-node#element-scoped
func TestDescribeUnknownElementIsActionable(t *testing.T) {
	_, errOut, code := run("describe", applyOp+"#nope", "-C", walletSpec)
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "drop the `#nope`")
	assert.Contains(t, errOut, "try:")
}
