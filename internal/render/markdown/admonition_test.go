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

// admonitionFixture builds a richer graph than the shared fixture: a Need
// with two atoms — fr-01 satisfied by a proven UC, fr-02 unclaimed — plus
// an accepted ADR. The UC is therefore proven, the Need is partial.
func admonitionFixture(t *testing.T) (*compiler.ResolvedGraph, map[model.ModulePath]string) {
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
			Atoms: []model.Atom{
				{ID: "fr-01", Kind: model.KindFR, Text: "Atomic."},
				{ID: "fr-02", Kind: model.KindFR, Text: "Idempotent."},
			},
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

	// Bind do-thing as proven: a req + a covering test.
	facts := []compiler.CodeFact{
		{Module: svc, Verb: compiler.VerbReq,
			Target: compiler.AnnotationTarget{Slug: "do-thing"}, File: "x.go", Line: 1},
		{Module: svc, Verb: compiler.VerbTest,
			Target: compiler.AnnotationTarget{Slug: "do-thing"}, File: "x_test.go", Line: 1, IsTest: true},
	}
	g, _ := compiler.New().Compile(compiler.Input{
		Modules: []source.LoadedModule{
			{Manifest: source.Manifest{Path: svc, Kind: source.KindService}, Nodes: []model.PlacedNode{contract, container}},
			{Manifest: source.Manifest{Path: prod, Kind: source.KindDomain}, Nodes: []model.PlacedNode{story, product}},
			{Manifest: source.Manifest{Path: gov, Kind: source.KindGovernance}, Nodes: []model.PlacedNode{adr}},
		},
		Facts: facts,
	})
	require.NotNil(t, g)
	return g, map[model.ModulePath]string{svc: "abc", prod: "abc", gov: "abc"}
}

// TestStatusAdmonitions_On: every node-type page opens with a status
// admonition block after the H1, and elements/atoms carry an inline marker.
func TestStatusAdmonitions_On(t *testing.T) {
	g, revs := admonitionFixture(t)
	r := markdown.New(markdown.Config{
		Layout:                markdown.LayoutTree,
		StripPrefix:           "ex.test/gp/",
		WithStatusAdmonitions: true,
	})
	out, err := r.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)

	// Contract: proven → success admonition + inline *Proven.* on the
	// invariant.
	uc := string(out["svc/do-thing.md"])
	require.NotEmpty(t, uc)
	assertAdmonitionAfterH1(t, uc, `!!! success "Proven"`)
	assert.Contains(t, uc, "All invariants have an implementation and a test.")
	assert.Contains(t, uc, "*Proven.*")

	// Need: partial → warning Partial — 1/2 covered. fr-01 covered, fr-02
	// uncovered.
	need := string(out["dom/as-user.md"])
	require.NotEmpty(t, need)
	assertAdmonitionAfterH1(t, need, `!!! warning "Partial — 1/2 covered"`)
	assert.Contains(t, need, "Some requirements have no proven contract.")
	// fr-01 covered by do-thing#atomic.
	assert.Contains(t, need, "*Covered by [do-thing#atomic]")
	// fr-02 uncovered.
	assert.Contains(t, need, "**Uncovered.**")

	// ADR: accepted → note Accepted.
	adr := string(out["gov/adr-01.md"])
	require.NotEmpty(t, adr)
	assertAdmonitionAfterH1(t, adr, `!!! note "Accepted"`)
	assert.Contains(t, adr, "This decision is in effect.")
}

// TestStatusAdmonitions_Off: flag off → no `!!!` blocks and no inline
// markers anywhere.
func TestStatusAdmonitions_Off(t *testing.T) {
	g, revs := admonitionFixture(t)
	r := markdown.New(markdown.Config{
		Layout:      markdown.LayoutTree,
		StripPrefix: "ex.test/gp/",
	})
	out, err := r.Render(render.Input{Graph: g, Revisions: revs})
	require.NoError(t, err)
	for path, content := range out {
		body := string(content)
		assert.NotContains(t, body, "!!!", "no admonition blocks in %s", path)
		assert.NotContains(t, body, "*Proven.*", "no inline element status in %s", path)
		assert.NotContains(t, body, "**Uncovered.**", "no inline atom status in %s", path)
	}
}

// assertAdmonitionAfterH1 checks the admonition line appears between the H1
// and the first non-blank content line after it.
func assertAdmonitionAfterH1(t *testing.T, body, want string) {
	t.Helper()
	lines := strings.Split(body, "\n")
	hIdx := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "# ") {
			hIdx = i
			break
		}
	}
	require.GreaterOrEqual(t, hIdx, 0, "no H1 found")
	// Find first non-blank line after H1.
	for i := hIdx + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		assert.Equal(t, want, lines[i],
			"first non-blank line after H1 should be the admonition opener")
		return
	}
	t.Fatalf("no content after H1")
}
