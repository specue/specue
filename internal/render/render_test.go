package render_test

import (
	"sort"
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

// fixtureGraph builds a small graph with one service module (a Container, a
// Contract with two named invariants, one of which satisfies a Need FR),
// one domain module (the Need + Domain), and one governance module (one ADR
// cited from the Contract). Enough to exercise every renderer-side invariant.
func fixtureGraph(t *testing.T) (*compiler.ResolvedGraph, map[model.ModulePath]string) {
	t.Helper()

	svc := model.ModulePath("ex.test/svc@v0")
	prod := model.ModulePath("ex.test/dom@v0")
	gov := model.ModulePath("ex.test/gov@v0")

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
				{ID: "single-verdict", Text: "A repeat is a no-op."},
				{Text: "The thing is done."},
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
			Domain:      model.NodeRef{Module: prod, Slug: "domain"},
			Consumer:    "user",
			Description: "X, so that Y",
			Atoms:       []model.Atom{{ID: "fr-01", Kind: model.KindFR, Text: "The thing is atomic."}},
		}},
	}}
	product := model.PlacedNode{Module: prod, Node: model.Node{
		Slug: "domain", Type: model.TypeDomain, Title: "Example domain",
		Confidence: model.Confirmed,
		Body:       &model.Body{},
	}}

	adr := model.PlacedNode{Module: gov, Node: model.Node{
		Slug: "adr-01", Type: model.TypeADR, Title: "Atomic is required",
		Confidence: model.Confirmed,
		Body: &model.Body{
			Prose: "We require atomicity to keep callers honest.",
			Gov:   &model.GovBody{Lifecycle: model.LifecycleAccepted},
		},
	}}

	c := compiler.New()
	g, _ := c.Compile(compiler.Input{Modules: []source.LoadedModule{
		{Manifest: source.Manifest{Path: svc, Kind: source.KindService}, Nodes: []model.PlacedNode{contract, container}},
		{Manifest: source.Manifest{Path: prod, Kind: source.KindDomain}, Nodes: []model.PlacedNode{story, product}},
		{Manifest: source.Manifest{Path: gov, Kind: source.KindGovernance}, Nodes: []model.PlacedNode{adr}},
	}})
	require.NotNil(t, g)

	return g, map[model.ModulePath]string{
		svc:  "abc123def456",
		prod: "abc123def456",
		gov:  "abc123def456",
	}
}

// TestRenderDerivedFromResolvedGraph asserts the renderer reads from the same
// resolved graph the rest of the tool sees: every authored node ends up with
// exactly one file, no extra files leak in, and the only non-node file is the
// configured index. (render-doc#derived-from-resolved-graph + #one-file-per-node)
//
//specue:test:render-doc#derived-from-resolved-graph
//specue:test:render-doc#one-file-per-node
func TestRenderDerivedFromResolvedGraph(t *testing.T) {
	g, revs := fixtureGraph(t)
	tree, err := markdown.Default().Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	var nodeIDs []string
	for n := range g.Nodes() {
		nodeIDs = append(nodeIDs, n.ID().String())
	}
	sort.Strings(nodeIDs)

	gotNonIndex := 0
	for path := range tree {
		if path == "README.md" {
			continue
		}
		gotNonIndex++
		assert.True(t, strings.HasSuffix(string(path), ".md"), "%s is .md", path)
	}
	assert.Equal(t, len(nodeIDs), gotNonIndex, "one file per node, no extras")
	assert.Contains(t, tree, render.RelPath("README.md"))
}

// TestRenderMachineReadableFrontmatter asserts every node file opens with
// `---` YAML frontmatter carrying the type, the module, the rendered_from
// revision, and (for elements with satisfies/decided_by) the flattened refs.
//
//specue:test:render-doc#machine-readable-frontmatter
func TestRenderMachineReadableFrontmatter(t *testing.T) {
	g, revs := fixtureGraph(t)
	tree, err := markdown.Default().Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	uc := string(tree["ex.test-svc-v0/do-thing.md"])
	require.NotEmpty(t, uc, "UC file present")

	assert.True(t, strings.HasPrefix(uc, "---\n"), "frontmatter fence")
	assert.Contains(t, uc, "id: ex.test/svc@v0:do-thing")
	assert.Contains(t, uc, "type: Contract")
	assert.Contains(t, uc, "module: ex.test/svc@v0")
	assert.Contains(t, uc, "rendered_from: abc123def456")
	assert.Contains(t, uc, "ex.test/dom@v0:as-user#fr-01", "satisfies flattened into frontmatter")
	assert.Contains(t, uc, "ex.test/gov@v0:adr-01", "decided_by flattened into frontmatter")

	// Frontmatter is delimited by a second `---`; everything after is body.
	assert.True(t, strings.Index(uc, "\n---\n") > 0, "closing fence present")
}

// TestRenderCrossLinksResolveAsMarkdown asserts cross-references between nodes
// are written as relative markdown links to the target's actual file (so a
// reader on a forge clicks through), with an anchor when an element is
// addressed (so a satisfies link lands on the FR, not the file top).
//
//specue:test:render-doc#cross-links-resolve-as-markdown
func TestRenderCrossLinksResolveAsMarkdown(t *testing.T) {
	g, revs := fixtureGraph(t)
	tree, err := markdown.Default().Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	uc := string(tree["ex.test-svc-v0/do-thing.md"])
	require.NotEmpty(t, uc)

	// satisfies points cross-module with the atom anchor
	assert.Contains(t, uc, "(../ex.test-dom-v0/as-user.md#fr-01)",
		"satisfies link is relative + anchored on the atom")
	// decided_by points cross-module to the ADR file (no anchor — whole-node)
	assert.Contains(t, uc, "(../ex.test-gov-v0/adr-01.md)",
		"decided_by link is relative to the ADR file")
	// same-module service link is just <slug>.md
	assert.Contains(t, uc, "(service.md)", "same-module link is a bare filename")

	story := string(tree["ex.test-dom-v0/as-user.md"])
	assert.Contains(t, story, "<a id=\"fr-01\"></a>", "story carries the anchor satisfies links land on")
}

// TestRenderDeterministic asserts two runs over the same graph produce
// byte-identical output (no map iteration leaks into ordering of the body).
// This is what makes the tree diffable across renders.
func TestRenderDeterministic(t *testing.T) {
	g, revs := fixtureGraph(t)
	a, err := markdown.Default().Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	b, err := markdown.Default().Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	require.Equal(t, len(a), len(b))
	for k, va := range a {
		assert.Equal(t, va, b[k], "file %s differs between runs", k)
	}
}
