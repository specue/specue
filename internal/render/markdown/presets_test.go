package markdown_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
	"github.com/specue/specue/internal/render/markdown"
	"github.com/specue/specue/internal/source"
)

// fixture builds the same small graph as render_test, but in-package so the
// preset tests don't have to expose helpers.
func fixture(t *testing.T) (*compiler.ResolvedGraph, map[model.ModulePath]string) {
	t.Helper()
	svc := model.ModulePath("ex.test/gp/svc@v0")
	prod := model.ModulePath("ex.test/gp/dom@v0")
	gov := model.ModulePath("ex.test/gp/gov@v0")

	storyRef := model.NodeRef{Module: prod, Slug: "as-user"}
	adrRef := model.NodeRef{Module: gov, Slug: "adr-01"}
	svcRef := model.NodeRef{Module: svc, Slug: "service"}

	contract := model.PlacedNode{Module: svc, Node: model.Node{
		Slug: "do-thing", Type: model.TypeContract, Title: "Do the thing",
		Confidence: model.Confirmed,
		Body: &model.Body{Contract: &model.ContractBody{
			Service: svcRef, Trigger: "caller asks",
			Elements: []model.Element{
				{ID: "atomic", Text: "Each call is atomic.",
					Satisfies: []model.AtomRef{{Need: storyRef, Atom: "fr-01"}},
					DecidedBy: []model.NodeRef{adrRef}},
			},
		}},
	}}
	container := model.PlacedNode{Module: svc, Node: model.Node{
		Slug: "service", Type: model.TypeContainer, Title: "The service",
		Confidence: model.Confirmed,
		Body:       &model.Body{Container: &model.ContainerBody{Kind: model.ContainerService}},
	}}
	story := model.PlacedNode{Module: prod, Node: model.Node{
		Slug: "as-user", Type: model.TypeNeed, Title: "As a user",
		Confidence: model.Confirmed,
		Body: &model.Body{Need: &model.NeedBody{
			Domain: model.NodeRef{Module: prod, Slug: "domain"}, Consumer: "user",
			Atoms: []model.Atom{{ID: "fr-01", Kind: model.KindFR, Text: "Atomic."}},
		}},
	}}
	product := model.PlacedNode{Module: prod, Node: model.Node{
		Slug: "domain", Type: model.TypeDomain, Title: "Example domain",
		Confidence: model.Confirmed, Body: &model.Body{},
	}}
	adr := model.PlacedNode{Module: gov, Node: model.Node{
		Slug: "adr-01", Type: model.TypeADR, Title: "Atomic is required",
		Confidence: model.Confirmed,
		Body: &model.Body{Prose: "We require atomicity.",
			Gov: &model.GovBody{Lifecycle: model.LifecycleAccepted}},
	}}

	c := compiler.New()
	g, _ := c.Compile(compiler.Input{Modules: []source.LoadedModule{
		{Manifest: source.Manifest{Path: svc, Kind: source.KindService}, Nodes: []model.PlacedNode{contract, container}},
		{Manifest: source.Manifest{Path: prod, Kind: source.KindDomain}, Nodes: []model.PlacedNode{story, product}},
		{Manifest: source.Manifest{Path: gov, Kind: source.KindGovernance}, Nodes: []model.PlacedNode{adr}},
	}})
	require.NotNil(t, g)
	return g, map[model.ModulePath]string{svc: "abc", prod: "abc", gov: "abc"}
}

// TestFlatVsTreeLayout: same graph, two layouts → different paths.
func TestFlatVsTreeLayout(t *testing.T) {
	g, revs := fixture(t)

	flat, err := markdown.New(markdown.Config{Layout: markdown.LayoutFlat}).
		Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	tree, err := markdown.New(markdown.Config{Layout: markdown.LayoutTree}).
		Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	assert.Contains(t, flat, render.RelPath("ex.test-gp-svc-v0/do-thing.md"))
	assert.Contains(t, tree, render.RelPath("ex.test/gp/svc/do-thing.md"))
}

// TestStripPrefixShortens: strip-prefix removes the prefix from directory.
func TestStripPrefixShortens(t *testing.T) {
	g, revs := fixture(t)
	out, err := markdown.New(markdown.Config{
		Layout:      markdown.LayoutTree,
		StripPrefix: "ex.test/gp/",
	}).Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	assert.Contains(t, out, render.RelPath("svc/do-thing.md"))
	assert.Contains(t, out, render.RelPath("dom/as-user.md"))
}

// TestCrossLinkResolveInTree: tree layout, link from svc/do-thing to dom/as-user
// should be ../dom/as-user.md#fr-01 (one ../).
func TestCrossLinkResolveInTree(t *testing.T) {
	g, revs := fixture(t)
	out, err := markdown.New(markdown.Config{
		Layout:      markdown.LayoutTree,
		StripPrefix: "ex.test/gp/",
	}).Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	uc := string(out["svc/do-thing.md"])
	require.NotEmpty(t, uc)
	assert.Contains(t, uc, "(../dom/as-user.md#fr-01)")
	assert.Contains(t, uc, "(../gov/adr-01.md)")
	assert.Contains(t, uc, "(service.md)", "same-module link stays bare")
}

// TestFrontmatterMinimal: only title, type, status keys.
func TestFrontmatterMinimal(t *testing.T) {
	g, revs := fixture(t)
	out, err := markdown.New(markdown.Config{Frontmatter: markdown.FrontmatterMinimal}).
		Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	uc := string(out["ex.test-gp-svc-v0/do-thing.md"])
	assert.Contains(t, uc, "title: Do the thing")
	assert.Contains(t, uc, "type: Contract")
	assert.NotContains(t, uc, "module: ex.test")
	assert.NotContains(t, uc, "rendered_from:")
	assert.NotContains(t, uc, "satisfies:")
}

// TestFrontmatterMark: PascalCase keys, Labels, Space optional.
func TestFrontmatterMark(t *testing.T) {
	g, revs := fixture(t)
	out, err := markdown.New(markdown.Config{
		Frontmatter: markdown.FrontmatterMark, Space: "ENG"}).
		Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	uc := string(out["ex.test-gp-svc-v0/do-thing.md"])
	assert.Contains(t, uc, "Title: Do the thing")
	assert.Contains(t, uc, "Space: ENG")
	assert.Contains(t, uc, "Parent: svc")
	assert.Contains(t, uc, "- contract")
}

// TestFrontmatterMkdocs: lowercase, tags, optional icon.
func TestFrontmatterMkdocs(t *testing.T) {
	g, revs := fixture(t)
	out, err := markdown.New(markdown.Config{Frontmatter: markdown.FrontmatterMkDocs}).
		Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	uc := string(out["ex.test-gp-svc-v0/do-thing.md"])
	assert.Contains(t, uc, "title: Do the thing")
	assert.Contains(t, uc, "tags:")
	assert.Contains(t, uc, "- contract")
	assert.Contains(t, uc, "icon:")
}

// TestFrontmatterNone: no fence at all.
func TestFrontmatterNone(t *testing.T) {
	g, revs := fixture(t)
	out, err := markdown.New(markdown.Config{Frontmatter: markdown.FrontmatterNone}).
		Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	uc := string(out["ex.test-gp-svc-v0/do-thing.md"])
	assert.False(t, strings.HasPrefix(uc, "---\n"), "no frontmatter fence")
	assert.True(t, strings.HasPrefix(uc, "# "), "starts with heading")
}

// TestNavSnippet: nav.yml entry is produced via EmitNavSnippet.
func TestNavSnippet(t *testing.T) {
	g, revs := fixture(t)
	r := markdown.New(markdown.Config{
		Layout: markdown.LayoutTree, StripPrefix: "ex.test/gp/",
		NavSnippetPath: "nav.yml",
	})
	out, err := r.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	path, content, ok := markdown.EmitNavSnippet(r, out)
	require.True(t, ok)
	assert.Equal(t, render.RelPath("nav.yml"), path)
	s := string(content)
	assert.Contains(t, s, "nav:")
	assert.Contains(t, s, "svc:")
	assert.Contains(t, s, "do-thing: svc/do-thing.md")
}

// TestRenderInvariantKindAndWhen: a rejects invariant renders as one "Rejects
// when <cond>" sentence; a plain invariant shows no nature/when adornment.
// (ADR-14: one invariant kind.)
func TestRenderInvariantKindAndWhen(t *testing.T) {
	svc := model.ModulePath("ex.test/svc@v0")
	svcRef := model.NodeRef{Module: svc, Slug: "service"}
	contract := model.PlacedNode{Module: svc, Node: model.Node{
		Slug: "do-thing", Type: model.TypeContract, Title: "Do the thing",
		Confidence: model.Confirmed,
		Body: &model.Body{Contract: &model.ContractBody{
			Service: svcRef,
			Elements: []model.Element{
				{ID: "plain-inv", Text: "Always holds."},
				{ID: "refuses", Kind: model.KindRejects, When: "input is bad", Text: "the call is refused."},
			},
		}},
	}}
	container := model.PlacedNode{Module: svc, Node: model.Node{
		Slug: "service", Type: model.TypeContainer, Title: "The service",
		Confidence: model.Confirmed,
		Body:       &model.Body{Container: &model.ContainerBody{Kind: model.ContainerService}},
	}}
	g, _ := compiler.New().Compile(compiler.Input{Modules: []source.LoadedModule{
		{Manifest: source.Manifest{Path: svc, Kind: source.KindService}, Nodes: []model.PlacedNode{contract, container}},
	}})
	require.NotNil(t, g)

	out, err := markdown.New(markdown.Config{Layout: markdown.LayoutFlat}).
		Render(render.Input{Graph: g, Revisions: map[model.ModulePath]string{svc: "abc"}})
	require.NoError(t, err)
	body := string(out["ex.test-svc-v0/do-thing.md"])
	require.NotEmpty(t, body)

	plainIdx := strings.Index(body, "plain-inv")
	rejIdx := strings.Index(body, "refuses")
	require.Greater(t, plainIdx, 0)
	require.Greater(t, rejIdx, plainIdx)

	// The rejects section reads as one "Rejects when …" sentence.
	rejSection := body[rejIdx:]
	assert.Contains(t, rejSection, "**Rejects** when input is bad.")
	// The plain invariant (before the rejects section) shows no adornment.
	plainSection := body[plainIdx:rejIdx]
	assert.NotContains(t, plainSection, "*(")
	assert.NotContains(t, plainSection, "Rejects")
	assert.NotContains(t, plainSection, "*When*")
}

// TestDefaultUnchanged: zero-config New(Config{}) is byte-identical to Default().
func TestDefaultUnchanged(t *testing.T) {
	g, revs := fixture(t)
	a, err := markdown.Default().Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	b, err := markdown.New(markdown.Config{}).Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	require.Equal(t, len(a), len(b))
	for k, va := range a {
		assert.Equal(t, va, b[k], "file %s differs", k)
	}
}
