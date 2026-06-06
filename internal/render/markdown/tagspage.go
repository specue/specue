package markdown

import (
	"fmt"
	"sort"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// wrapWithTagsPage stashes the cfg on a registry keyed by the renderer pointer;
// the CLI looks it up after Render via EmitTagsPage. Same pattern as
// wrapWithIndexPages / wrapWithNavSnippet.
//
//specue:req:render-doc#tags-page-on-request
func wrapWithTagsPage(inner *render.Renderer, cfg Config) *render.Renderer {
	tagsHook.store(inner, cfg)
	return inner
}

type tagsHookRegistry struct {
	m map[*render.Renderer]Config
}

func (r *tagsHookRegistry) store(rr *render.Renderer, cfg Config) {
	if r.m == nil {
		r.m = map[*render.Renderer]Config{}
	}
	r.m[rr] = cfg
}

func (r *tagsHookRegistry) load(rr *render.Renderer) (Config, bool) {
	c, ok := r.m[rr]
	return c, ok
}

var tagsHook = &tagsHookRegistry{}

// EmitTagsPage builds the tags.md page for r. Returns (relpath, content, true)
// when WithTagsPage is on for the given renderer, ("", "", false) otherwise.
// resolve maps a node's layout-computed path to its post-collision location
// (pass nil if no collisions are in play).
// The CLI adds the entry to the Tree after index-pages and nav-snippet.
func EmitTagsPage(r *render.Renderer, _ render.Tree, ctx render.Context, resolve PathResolver) (render.RelPath, render.FileContent, bool) {
	cfg, ok := tagsHook.load(r)
	if !ok {
		return "", "", false
	}
	return render.RelPath("tags.md"), render.FileContent(buildTagsPage(ctx, cfg, resolve)), true
}

// tagsEntry is one bullet under a section.
type tagsEntry struct {
	fullID string // module-stripped + ":" + slug
	url    string // relative URL to the node's .md
	title  string
	status string
}

// type sections, in the order they appear in tags.md.
var tagsTypeOrder = []model.NodeType{
	model.TypeContract, model.TypeNeed, model.TypeDomain,
	model.TypeADR, model.TypePlan,
	model.TypePort, model.TypeContainer,
}

// status sections, in the order they appear in tags.md.
var tagsStatusOrder = []string{
	"proven", "implemented", "asserted", "broken",
	"covered", "partial", "uncovered",
	"accepted", "proposed", "superseded",
}

// buildTagsPage walks the graph, groups every node by lowercase type AND by
// status, sorts each bucket by full id, and emits the markdown body.
func buildTagsPage(ctx render.Context, cfg Config, resolve PathResolver) string {
	byType := map[string][]tagsEntry{}
	byStatus := map[string][]tagsEntry{}

	if ctx.Graph != nil {
		for n := range ctx.Graph.Nodes() {
			id := n.ID()
			typ := strings.ToLower(string(n.Node().Type))
			status := nodeStatus(n)
			title := titleOrSlug(n)
			// The label mirrors the on-disk directory: strip the configured
			// prefix and the trailing CUE @vN — the same shape the tree
			// layout uses for its directory names — so the reader sees
			// `be/auth-service/spec/tokens:slug`, not `…@v0:slug`.
			moduleLabel := stripModulePrefix(string(id.Module), cfg.StripPrefix)
			if at := strings.LastIndex(moduleLabel, "@"); at >= 0 {
				if slash := strings.LastIndex(moduleLabel, "/"); slash < at {
					moduleLabel = moduleLabel[:at]
				}
			}
			fullID := moduleLabel + ":" + string(id.Slug)
			// tags.md lives at the root, so the relative URL is just the
			// layout-relative path of the node file (no `../`). Run it through
			// the resolver so a collided node points at its post-move location
			// (`<dir>/<slug>/index.md`), not the dangling pre-move `<slug>.md`.
			url := string(ctx.Layout.NodePath(id))
			if resolve != nil {
				url = resolve.ResolvePath(url)
			}

			e := tagsEntry{fullID: fullID, url: url, title: title, status: status}
			byType[typ] = append(byType[typ], e)
			if status != "" {
				byStatus[status] = append(byStatus[status], e)
			}
		}
	}

	var b strings.Builder
	b.WriteString(tagsFrontmatter(cfg))
	b.WriteString("# Tags\n\n")
	b.WriteString("Every node carries tags for its type and status. Click a heading or use the table of contents.\n\n")

	for _, t := range tagsTypeOrder {
		key := strings.ToLower(string(t))
		writeTagSection(&b, key, byType[key])
	}
	for _, s := range tagsStatusOrder {
		writeTagSection(&b, s, byStatus[s])
	}
	return b.String()
}

// nodeStatus returns the lowercase status string the renderer would put in
// frontmatter. For ADR/Plan that is Lifecycle; for everything else it is the
// compiler-assigned ResolvedNodeStatus.
func nodeStatus(n *compiler.ResolvedNode) string {
	nd := n.Node()
	if nd.Type == model.TypeADR || nd.Type == model.TypePlan {
		if gov := nd.Body.Gov; gov != nil {
			return string(gov.Lifecycle)
		}
		return ""
	}
	return string(n.Status)
}

func writeTagSection(b *strings.Builder, name string, entries []tagsEntry) {
	if len(entries) == 0 {
		return
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].fullID < entries[j].fullID })
	// The {#tag:name} explicit id matches Material's pill-link target
	// (href="…/tags/#tag:<name>"), so a click on a status or type pill on a
	// node page lands on the right heading here. Requires the attr_list
	// markdown extension to honour the {…} attribute syntax.
	fmt.Fprintf(b, "## %s { #tag:%s }\n\n", name, name)
	for _, e := range entries {
		if e.status != "" {
			fmt.Fprintf(b, "- [`%s`](%s) — %s · *%s*\n", e.fullID, e.url, e.title, e.status)
		} else {
			fmt.Fprintf(b, "- [`%s`](%s) — %s\n", e.fullID, e.url, e.title)
		}
	}
	b.WriteString("\n")
}

// tagsFrontmatter renders the YAML preamble for tags.md, dispatching on the
// active shape. For mkdocs we hide the auto-tag plugin's pills on this page
// itself (the page IS the tags index).
func tagsFrontmatter(cfg Config) string {
	shape := cfg.Frontmatter
	if shape == "" {
		shape = FrontmatterFull
	}
	switch shape {
	case FrontmatterNone:
		return ""
	case FrontmatterMkDocs:
		return fencedAny(map[string]any{
			"title": "Tags",
			"hide":  []string{"tags"},
		})
	case FrontmatterMark:
		m := map[string]any{"Title": "Tags"}
		if cfg.Space != "" {
			m["Space"] = cfg.Space
		}
		// Parent is the root; we surface the strip-prefix tail when present,
		// otherwise leave it off — tags.md is a top-level page.
		return fencedAny(m)
	case FrontmatterMinimal:
		return fencedAny(map[string]any{"title": "Tags", "type": "Tags"})
	case FrontmatterFull:
		fallthrough
	default:
		return fencedAny(map[string]any{"title": "Tags", "type": "Tags"})
	}
}
