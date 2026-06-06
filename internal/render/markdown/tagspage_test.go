package markdown_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/render"
	"github.com/specue/specue/internal/render/markdown"
)

// TestWithTagsPage: WithTagsPage emits a tags.md grouping nodes by type AND
// by status, each entry carrying the module-qualified id, link, title and a
// status badge. Empty sections are skipped.
func TestWithTagsPage(t *testing.T) {
	g, revs := fixture(t)
	r := markdown.New(markdown.Config{
		Layout:       markdown.LayoutTree,
		StripPrefix:  "ex.test/gp/",
		WithTagsPage: true,
	})
	out, err := r.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	ctx := render.Context{Graph: g, Revisions: revs, Layout: r.Layout()}
	path, content, ok := markdown.EmitTagsPage(r, out, ctx, nil)
	require.True(t, ok)
	assert.Equal(t, render.RelPath("tags.md"), path)
	out[path] = content
	body := string(content)

	// Page header + intro.
	assert.Contains(t, body, "# Tags")
	assert.Contains(t, body, "Every node carries tags")

	// Type sections present for every node type in the fixture.
	assert.Contains(t, body, "## contract")
	assert.Contains(t, body, "## need")
	assert.Contains(t, body, "## domain")
	assert.Contains(t, body, "## adr")
	assert.Contains(t, body, "## container")

	// Plan/port sections must NOT appear — fixture has none of those.
	assert.NotContains(t, body, "## plan\n")
	assert.NotContains(t, body, "## port\n")

	// Status sections derived from the fixture: ADR is accepted; Contract is
	// proven (has both req binding and a covering atom in fixture? actually
	// the fixture has no scan binding so the Contract is asserted). Need is
	// uncovered. Container has no scan so it ends up asserted too. Assert at
	// least one of the present statuses appears.
	assert.Contains(t, body, "## accepted")

	// Each entry carries the stripped module path + slug, the URL, title,
	// and a status badge in *italics*.
	assert.Contains(t, body, "[`svc:do-thing`](svc/do-thing.md)")
	assert.Contains(t, body, "Do the thing")
	assert.Contains(t, body, "[`gov:adr-01`](gov/adr-01.md)")
	assert.Contains(t, body, "*accepted*")

	// Frontmatter for default (full) shape.
	assert.True(t, strings.HasPrefix(body, "---\n"), "default frontmatter shape fences tags.md")
	assert.Contains(t, body, "title: Tags")
	assert.Contains(t, body, "type: Tags")
}

// TestWithTagsPageOff: no flag → no tags.md and EmitTagsPage returns false.
func TestWithTagsPageOff(t *testing.T) {
	g, revs := fixture(t)
	r := markdown.New(markdown.Config{Layout: markdown.LayoutTree, StripPrefix: "ex.test/gp/"})
	out, err := r.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	ctx := render.Context{Graph: g, Revisions: revs, Layout: r.Layout()}
	_, _, ok := markdown.EmitTagsPage(r, out, ctx, nil)
	assert.False(t, ok)
	_, hasTags := out[render.RelPath("tags.md")]
	assert.False(t, hasTags)
}

// TestWithTagsPageMkdocs: mkdocs frontmatter hides the auto-tag plugin pills
// on tags.md itself (the page IS the index).
func TestWithTagsPageMkdocs(t *testing.T) {
	g, revs := fixture(t)
	r := markdown.New(markdown.Config{
		Layout:       markdown.LayoutTree,
		StripPrefix:  "ex.test/gp/",
		Frontmatter:  markdown.FrontmatterMkDocs,
		WithTagsPage: true,
	})
	_, err := r.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	ctx := render.Context{Graph: g, Revisions: revs, Layout: r.Layout()}
	_, content, ok := markdown.EmitTagsPage(r, render.Tree{}, ctx, nil)
	require.True(t, ok)
	body := string(content)
	assert.Contains(t, body, "title: Tags")
	assert.Contains(t, body, "hide:")
	assert.Contains(t, body, "- tags")
}
