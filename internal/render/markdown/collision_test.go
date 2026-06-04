package markdown_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/render"
	"github.com/specue/specue/internal/render/markdown"
)

// TestResolveNodeFolderCollisions: a node `settings.md` lives next to a
// `settings/` directory that holds sibling node files. The collision
// resolver moves the node to `settings/index.md` and rewrites in-file
// relative links by prepending `../`.
func TestResolveNodeFolderCollisions(t *testing.T) {
	body := strings.Join([]string{
		"# settings",
		"",
		"See [neighbour](neighbour.md) and [external](https://example.com/x.md).",
		"Anchor: [self](#top). Sub: [child](settings/create.md).",
		"",
	}, "\n")

	tree := render.Tree{
		render.RelPath("be/game-service/spec/settings-sor/settings.md"):        render.FileContent(body),
		render.RelPath("be/game-service/spec/settings-sor/neighbour.md"):       render.FileContent("# neighbour\n"),
		render.RelPath("be/game-service/spec/settings-sor/settings/create.md"): render.FileContent("# create\n"),
		render.RelPath("be/game-service/spec/settings-sor/settings/update.md"): render.FileContent("# update\n"),
	}

	markdown.DetectCollisions(tree).Apply(tree)

	// Old path is gone.
	_, hasOld := tree[render.RelPath("be/game-service/spec/settings-sor/settings.md")]
	assert.False(t, hasOld, "settings.md must be removed")

	// New path holds the body.
	moved, hasNew := tree[render.RelPath("be/game-service/spec/settings-sor/settings/index.md")]
	require.True(t, hasNew, "settings/index.md must exist")
	movedStr := string(moved)

	// Relative sibling link prefixed.
	assert.Contains(t, movedStr, "[neighbour](../neighbour.md)")
	// Absolute URL untouched.
	assert.Contains(t, movedStr, "[external](https://example.com/x.md)")
	// Anchor-only untouched.
	assert.Contains(t, movedStr, "[self](#top)")
	// Relative subpath also prefixed (becomes ../settings/create.md, which
	// from <dir>/settings/index.md resolves to <dir>/settings/create.md).
	assert.Contains(t, movedStr, "[child](../settings/create.md)")

	// Untouched sibling files remain.
	_, ok := tree[render.RelPath("be/game-service/spec/settings-sor/neighbour.md")]
	assert.True(t, ok)
	_, ok = tree[render.RelPath("be/game-service/spec/settings-sor/settings/create.md")]
	assert.True(t, ok)
}

// TestResolveNodeFolderCollisionsRewritesCrossFileLinks: a link FROM another
// file pointing AT the moved node must be rewritten to point at the new
// `<dir>/<slug>/index.md` location, preserving any anchor. Absolute URLs are
// untouched.
func TestResolveNodeFolderCollisionsRewritesCrossFileLinks(t *testing.T) {
	parentBody := strings.Join([]string{
		"# parent",
		"",
		"See [link](../b/target.md#anchor).",
		"Skip [external](https://example.com/x.md).",
		"Plain [no-anchor](../b/target.md).",
		"",
	}, "\n")

	tree := render.Tree{
		render.RelPath("a/parent.md"):       render.FileContent(parentBody),
		render.RelPath("b/target.md"):       render.FileContent("# target\n"),
		render.RelPath("b/target/child.md"): render.FileContent("# child\n"),
	}

	markdown.DetectCollisions(tree).Apply(tree)

	// Node moved.
	_, hasOld := tree[render.RelPath("b/target.md")]
	assert.False(t, hasOld)
	_, hasNew := tree[render.RelPath("b/target/index.md")]
	assert.True(t, hasNew)

	// Parent's link rewritten. From a/parent.md, b/target/index.md is
	// ../b/target/index.md. Anchor preserved.
	parent := string(tree[render.RelPath("a/parent.md")])
	assert.Contains(t, parent, "[link](../b/target/index.md#anchor)")
	assert.Contains(t, parent, "[no-anchor](../b/target/index.md)")
	// Absolute URL untouched.
	assert.Contains(t, parent, "[external](https://example.com/x.md)")
}

// TestResolveNodeFolderCollisionsSuppressesAutoIndex: with --with-index-pages
// the auto-generated index.md for the collided directory must be suppressed
// — the node now IS that directory's index.
func TestResolveNodeFolderCollisionsSuppressesAutoIndex(t *testing.T) {
	// Use the shared fixture but craft a tree by hand to model the collision.
	g, revs := fixture(t)
	r := markdown.New(markdown.Config{
		Layout:         markdown.LayoutTree,
		StripPrefix:    "ex.test/gp/",
		WithIndexPages: true,
	})
	tree, err := r.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	// Synthesise a collision: rename the leaf so its parent slug clashes
	// with the slug of a synthetic UseCase placed at svc/do-thing.md beside
	// a svc/do-thing/ subdir.
	tree[render.RelPath("svc/do-thing/child.md")] = render.FileContent("# child\n")

	// Mirror the real pipeline: detect first (so EmitIndexPages can skip the
	// collided dir), then Apply the moves at the end.
	collisions := markdown.DetectCollisions(tree)

	// Run EmitIndexPages with the skipper: the auto-index for svc/do-thing
	// must NOT replace the node-as-index.
	ctx := render.Context{Graph: g, Revisions: revs, Layout: r.Layout()}
	entries, ok := markdown.EmitIndexPages(r, tree, ctx, collisions)
	require.True(t, ok)
	_, autoGenerated := entries[render.RelPath("svc/do-thing/index.md")]
	assert.False(t, autoGenerated, "auto-index must be suppressed for collided dir")

	collisions.Apply(tree)

	// The UseCase moved.
	_, hasOld := tree[render.RelPath("svc/do-thing.md")]
	assert.False(t, hasOld)
	_, hasNew := tree[render.RelPath("svc/do-thing/index.md")]
	assert.True(t, hasNew)

	// Sanity: other dirs still get their auto-index.
	_, hasSvcIdx := entries[render.RelPath("svc/index.md")]
	assert.True(t, hasSvcIdx)
}
