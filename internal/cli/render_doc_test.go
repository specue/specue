package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenderWritesTreeUnderDestination is the happy path: run render against a
// known-good module, point at an empty dir, see files appear under it. Covers
// the destination-is-explicit and one-file-per-node invariants end-to-end.
//
//specue:test:render-doc
//specue:test:render-doc#destination-is-explicit
//specue:test:render-doc#one-file-per-node
func TestRenderWritesTreeUnderDestination(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "doc")
	out, _, code := run("render", dest, "-C", walletSpec)
	require.Equal(t, exitOK, code, "render exits ok: %s", out)
	assert.Contains(t, out, "rendered")

	// At least the index + one node file landed under dest.
	entries, err := os.ReadDir(dest)
	require.NoError(t, err)
	assert.NotEmpty(t, entries, "tree written")

	readme, err := os.ReadFile(filepath.Join(dest, "README.md"))
	require.NoError(t, err)
	assert.Contains(t, string(readme), "Modules", "README lists modules")
}

// TestRenderRefusesNonEmptyDestination guards #refuses-non-empty-destination:
// pre-create a sibling file, ensure the run exits with an actionable error,
// and confirm the sibling is untouched.
//
//specue:test:render-doc#refuses-non-empty-destination
func TestRenderRefusesNonEmptyDestination(t *testing.T) {
	dest := t.TempDir() // TempDir returns an EXISTING empty dir
	sibling := filepath.Join(dest, "keep.txt")
	require.NoError(t, os.WriteFile(sibling, []byte("hands off"), 0o644))

	_, errOut, code := run("render", dest, "-C", walletSpec)
	assert.Equal(t, exitUsage, code, "render refuses to write into a non-empty dir")
	assert.Contains(t, errOut, "not empty")
	assert.Contains(t, errOut, "try:", "actionable fix is shown")

	got, err := os.ReadFile(sibling)
	require.NoError(t, err)
	assert.Equal(t, "hands off", string(got), "the sibling is untouched")
}

// TestRenderDestinationIsExplicit asserts a missing destination argument is a
// usage error, not a default-to-cwd surprise.
//
//specue:test:render-doc#destination-is-explicit
func TestRenderDestinationIsExplicit(t *testing.T) {
	_, errOut, code := run("render", "-C", walletSpec)
	assert.NotEqual(t, exitOK, code, "render needs an explicit destination")
	// cobra reports the missing arg before our verb runs — its message names the
	// command, which is what tells the user where to look.
	assert.True(t, strings.Contains(errOut, "render") || strings.Contains(errOut, "arg"),
		"the error points at the verb's missing argument: %s", errOut)
}
