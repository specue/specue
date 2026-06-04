package markdown_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/render"
	"github.com/specue/specue/internal/render/markdown"
)

// TestWithIndexPages: tree layout + WithIndexPages emits an index.md at every
// directory level (root + each intermediate dir + leaf-module dir), the root
// README.md is gone, and bodies carry the status summary and child links.
func TestWithIndexPages(t *testing.T) {
	g, revs := fixture(t)
	r := markdown.New(markdown.Config{
		Layout:         markdown.LayoutTree,
		StripPrefix:    "ex.test/gp/",
		WithIndexPages: true,
	})
	out, err := r.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	ctx := render.Context{Graph: g, Revisions: revs, Layout: r.Layout()}
	entries, ok := markdown.EmitIndexPages(r, out, ctx, nil)
	require.True(t, ok)
	for k, v := range entries {
		out[k] = v
	}

	// README.md must NOT be emitted when WithIndexPages is on.
	_, hasReadme := out[render.RelPath("README.md")]
	assert.False(t, hasReadme, "README.md must be replaced by root index.md")

	// Root index.md exists.
	rootIdx, ok := out[render.RelPath("index.md")]
	require.True(t, ok, "root index.md must exist")
	rootStr := string(rootIdx)
	assert.Contains(t, rootStr, "# Spec")
	assert.Contains(t, rootStr, "## Areas")
	assert.Contains(t, rootStr, "[svc](svc/index.md)")
	assert.Contains(t, rootStr, "[dom](dom/index.md)")
	assert.Contains(t, rootStr, "[gov](gov/index.md)")

	// Leaf-module index.md (svc) exists and lists contracts with status.
	svcIdx, ok := out[render.RelPath("svc/index.md")]
	require.True(t, ok)
	svcStr := string(svcIdx)
	assert.Contains(t, svcStr, "# svc")
	assert.Contains(t, svcStr, "## Contracts")
	assert.Contains(t, svcStr, "[do-thing](do-thing.md)")
	assert.Contains(t, svcStr, "Do the thing")
	// Status counts on the header line.
	assert.True(t,
		strings.Contains(svcStr, "UseCase") || strings.Contains(svcStr, "Container"),
		"leaf module index should carry type counts")

	// Every leaf-module dir produced an index.md.
	for _, dir := range []string{"svc", "dom", "gov"} {
		_, ok := out[render.RelPath(dir+"/index.md")]
		assert.True(t, ok, "%s/index.md must exist", dir)
	}
}

// TestWithIndexPagesOff: default tree layout still emits README.md and no
// per-folder index.md.
func TestWithIndexPagesOff(t *testing.T) {
	g, revs := fixture(t)
	r := markdown.New(markdown.Config{
		Layout:      markdown.LayoutTree,
		StripPrefix: "ex.test/gp/",
	})
	out, err := r.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	ctx := render.Context{Graph: g, Revisions: revs, Layout: r.Layout()}
	_, ok := markdown.EmitIndexPages(r, out, ctx, nil)
	assert.False(t, ok, "no index pages without --with-index-pages")
	_, hasReadme := out[render.RelPath("README.md")]
	assert.True(t, hasReadme, "default tree layout keeps README.md")
	_, hasRootIdx := out[render.RelPath("index.md")]
	assert.False(t, hasRootIdx, "no root index.md without flag")
}
