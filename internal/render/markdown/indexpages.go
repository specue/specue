package markdown

import (
	"fmt"
	"slices"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
	"gopkg.in/yaml.v3"
)

// nopIndex is the IndexRenderer used when WithIndexPages is on — the per-folder
// index pass owns every index.md, including the root, so the standard
// README.md must NOT be produced.
type nopIndex struct{}

func (nopIndex) Path() render.RelPath { return "" }
func (nopIndex) Render(_ map[model.ModulePath][]*compiler.ResolvedNode, _ render.Context) (render.FileContent, error) {
	return "", nil
}

// wrapWithIndexPages stashes the cfg on a registry keyed by the renderer
// pointer; the CLI looks it up after Render via EmitIndexPages. Same pattern
// as wrapWithNavSnippet.
//
//specue:req:render-doc#index-pages-on-request
func wrapWithIndexPages(inner *render.Renderer, cfg Config) *render.Renderer {
	indexHook.store(inner, cfg)
	return inner
}

type indexHookRegistry struct {
	m map[*render.Renderer]Config
}

func (r *indexHookRegistry) store(rr *render.Renderer, cfg Config) {
	if r.m == nil {
		r.m = map[*render.Renderer]Config{}
	}
	r.m[rr] = cfg
}

func (r *indexHookRegistry) load(rr *render.Renderer) (Config, bool) {
	c, ok := r.m[rr]
	return c, ok
}

var indexHook = &indexHookRegistry{}

// EmitIndexPages walks the tree and produces an index.md for every directory
// level (root → leaf). Returns (entries, true) when WithIndexPages is on for
// the given renderer, (nil, false) otherwise. skip names directories whose
// index slot is already taken by a node (folder-vs-slug collision); pass nil
// to emit one for every directory. The CLI merges the entries into the Tree
// AFTER nav-snippet so a future index never clobbers nav.
func EmitIndexPages(r *render.Renderer, tree render.Tree, ctx render.Context, skip IndexSkipper) (map[render.RelPath]render.FileContent, bool) {
	cfg, ok := indexHook.load(r)
	if !ok {
		return nil, false
	}
	return buildIndexPages(tree, ctx, cfg, skip), true
}

// dirInfo accumulates summary data for one directory in the rendered tree.
type dirInfo struct {
	path        string // forward-slash, "" = root
	subdirs     map[string]*dirInfo
	leafFiles   []string                 // file paths under this dir (every .md file, transitively, for counting)
	directLeafs []string                 // .md files directly in this dir (not in subdirs)
	directSubs  []string                 // subdir names (one level down)
	nodes       []*compiler.ResolvedNode // every node whose file lives under this dir
	nodeByPath  map[string]*compiler.ResolvedNode
}

func newDirInfo(path string) *dirInfo {
	return &dirInfo{
		path:       path,
		subdirs:    map[string]*dirInfo{},
		nodeByPath: map[string]*compiler.ResolvedNode{},
	}
}

// buildIndexPages constructs the directory tree from the rendered Tree and the
// graph, then emits one index.md per directory.
func buildIndexPages(tree render.Tree, ctx render.Context, cfg Config, skip IndexSkipper) map[render.RelPath]render.FileContent {
	// 1) Map file paths to their resolved nodes (so we can read type/status/title).
	pathToNode := map[string]*compiler.ResolvedNode{}
	if ctx.Graph != nil {
		for n := range ctx.Graph.Nodes() {
			p := string(ctx.Layout.NodePath(n.ID()))
			pathToNode[p] = n
		}
	}

	// 2) Build the dir tree from every .md file in the tree (excluding any
	//    pre-existing index.md and README.md — we own them).
	root := newDirInfo("")
	for rel := range tree {
		s := string(rel)
		if !strings.HasSuffix(s, ".md") {
			continue
		}
		base := s
		if i := strings.LastIndex(s, "/"); i >= 0 {
			base = s[i+1:]
		}
		if base == "index.md" || base == "README.md" {
			continue
		}
		insertFile(root, s, pathToNode[s])
	}

	// 3) Walk the dir tree and emit one index.md per directory.
	out := map[render.RelPath]render.FileContent{}
	emitDir(root, cfg, ctx, out, skip)
	return out
}

func insertFile(root *dirInfo, fullPath string, n *compiler.ResolvedNode) {
	parts := strings.Split(fullPath, "/")
	cur := root
	cur.leafFiles = append(cur.leafFiles, fullPath)
	if n != nil {
		cur.nodes = append(cur.nodes, n)
		cur.nodeByPath[fullPath] = n
	}
	for i := 0; i < len(parts)-1; i++ {
		name := parts[i]
		next, ok := cur.subdirs[name]
		if !ok {
			subPath := name
			if cur.path != "" {
				subPath = cur.path + "/" + name
			}
			next = newDirInfo(subPath)
			cur.subdirs[name] = next
			cur.directSubs = append(cur.directSubs, name)
		}
		cur = next
		cur.leafFiles = append(cur.leafFiles, fullPath)
		if n != nil {
			cur.nodes = append(cur.nodes, n)
			cur.nodeByPath[fullPath] = n
		}
	}
	// register the leaf file as a direct child of its immediate parent
	cur.directLeafs = append(cur.directLeafs, fullPath)
}

func emitDir(d *dirInfo, cfg Config, ctx render.Context, out map[render.RelPath]render.FileContent, skip IndexSkipper) {
	// Recurse first so children exist before we link to them.
	subNames := append([]string(nil), d.directSubs...)
	slices.Sort(subNames)
	for _, name := range subNames {
		emitDir(d.subdirs[name], cfg, ctx, out, skip)
	}

	// If a node has taken this directory's index.md slot (folder/slug
	// collision resolved earlier), skip the auto-index — the node IS the
	// index.
	if skip != nil && skip.ShouldSkip(d.path) {
		return
	}

	content := renderIndexPage(d, cfg, ctx)
	rel := "index.md"
	if d.path != "" {
		rel = d.path + "/index.md"
	}
	out[render.RelPath(rel)] = render.FileContent(content)
}

// renderIndexPage produces the markdown body for one directory's index.md.
func renderIndexPage(d *dirInfo, cfg Config, _ render.Context) string {
	dirName := "Spec"
	parentName := ""
	if d.path != "" {
		if i := strings.LastIndex(d.path, "/"); i >= 0 {
			dirName = d.path[i+1:]
			parentName = d.path[:i]
			if j := strings.LastIndex(parentName, "/"); j >= 0 {
				parentName = parentName[j+1:]
			}
		} else {
			dirName = d.path
		}
	}

	var b strings.Builder
	b.WriteString(indexFrontmatter(dirName, parentName, cfg))
	fmt.Fprintf(&b, "# %s\n\n", dirName)

	totalNodes := len(d.nodes)
	statusCounts := countByStatus(d.nodes)
	typeCounts := countByType(d.nodes)

	isRoot := d.path == ""
	isLeafModule := len(d.directSubs) == 0 && len(d.directLeafs) > 0

	switch {
	case isRoot:
		b.WriteString("This documentation is rendered from a Specue spec.\n\n")
		if totalNodes > 0 {
			fmt.Fprintf(&b, "%d nodes across %d modules", totalNodes, countLeafModules(d))
			if len(statusCounts) > 0 {
				fmt.Fprintf(&b, " · %s", strings.Join(statusCounts, " · "))
			}
			b.WriteString("\n\n")
		}
	case isLeafModule:
		// Header line: type breakdown.
		if len(typeCounts) > 0 {
			fmt.Fprintf(&b, "%s\n\n", strings.Join(typeCounts, " · "))
		}
		if len(statusCounts) > 0 {
			fmt.Fprintf(&b, "**Status:** %s\n\n", strings.Join(statusCounts, " · "))
		}
	default:
		// Intermediate dir summary.
		if totalNodes > 0 {
			parts := []string{fmt.Sprintf("%d nodes", totalNodes)}
			parts = append(parts, statusCounts...)
			fmt.Fprintf(&b, "%s\n\n", strings.Join(parts, " · "))
		}
	}

	// Modules / Areas section: child subdirectories.
	if len(d.directSubs) > 0 {
		heading := "Modules"
		if isRoot {
			heading = "Areas"
		}
		fmt.Fprintf(&b, "## %s\n\n", heading)
		subs := append([]string(nil), d.directSubs...)
		slices.Sort(subs)
		for _, name := range subs {
			sub := d.subdirs[name]
			fmt.Fprintf(&b, "- [%s](%s) — %s\n", name, name+"/index.md", subSummary(sub))
		}
		b.WriteString("\n")
	}

	// Contracts section: direct leaf .md files.
	if len(d.directLeafs) > 0 {
		b.WriteString("## Contracts\n\n")
		leafs := append([]string(nil), d.directLeafs...)
		slices.Sort(leafs)
		for _, fp := range leafs {
			base := fp
			if i := strings.LastIndex(fp, "/"); i >= 0 {
				base = fp[i+1:]
			}
			slug := strings.TrimSuffix(base, ".md")
			n := d.nodeByPath[fp]
			if n != nil {
				title := titleOrSlug(n)
				status := string(n.Status)
				if status != "" {
					fmt.Fprintf(&b, "- [%s](%s) — %s · *%s*\n", slug, base, title, status)
				} else {
					fmt.Fprintf(&b, "- [%s](%s) — %s\n", slug, base, title)
				}
			} else {
				fmt.Fprintf(&b, "- [%s](%s)\n", slug, base)
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// subSummary is the one-line description shown next to a child link.
func subSummary(d *dirInfo) string {
	parts := []string{}
	if len(d.directSubs) > 0 {
		parts = append(parts, fmt.Sprintf("%d module%s", len(d.directSubs), plural(len(d.directSubs))))
	}
	if len(d.nodes) > 0 {
		parts = append(parts, fmt.Sprintf("%d node%s", len(d.nodes), plural(len(d.nodes))))
	}
	sc := countByStatus(d.nodes)
	parts = append(parts, sc...)
	if len(parts) == 0 {
		return "(empty)"
	}
	return strings.Join(parts, " · ")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// countByStatus counts ResolvedNode.Status across the slice, returning
// "proven: 5", "implemented: 4", … in a stable order.
func countByStatus(nodes []*compiler.ResolvedNode) []string {
	if len(nodes) == 0 {
		return nil
	}
	c := map[compiler.ResolvedNodeStatus]int{}
	for _, n := range nodes {
		if n.Status != "" {
			c[n.Status]++
		}
	}
	order := []compiler.ResolvedNodeStatus{
		compiler.StatusProven, compiler.StatusImplemented, compiler.StatusAsserted,
		compiler.StatusBlocked, compiler.StatusBroken,
		compiler.StatusCovered, compiler.StatusPartial, compiler.StatusUncovered,
	}
	var out []string
	for _, s := range order {
		if c[s] > 0 {
			out = append(out, fmt.Sprintf("%d %s", c[s], string(s)))
		}
	}
	return out
}

// countLeafModules counts directories that hold node files directly (no further
// subdirs of their own). Used in the root summary "N nodes across M modules".
func countLeafModules(d *dirInfo) int {
	if len(d.directSubs) == 0 {
		if len(d.directLeafs) > 0 {
			return 1
		}
		return 0
	}
	total := 0
	if len(d.directLeafs) > 0 {
		total = 1
	}
	for _, sub := range d.subdirs {
		total += countLeafModules(sub)
	}
	return total
}

// indexFrontmatter renders the YAML preamble for an index.md, dispatching on
// the active shape and mapping every shape to a generic "Index" kind.
func indexFrontmatter(dirName, parentName string, cfg Config) string {
	shape := cfg.Frontmatter
	if shape == "" {
		shape = FrontmatterFull
	}
	switch shape {
	case FrontmatterNone:
		return ""
	case FrontmatterMkDocs:
		return fencedAny(map[string]any{
			"title": dirName,
			"tags":  []string{"index"},
		})
	case FrontmatterMark:
		m := map[string]any{
			"Title": dirName,
		}
		if cfg.Space != "" {
			m["Space"] = cfg.Space
		}
		if parentName != "" {
			m["Parent"] = parentName
		}
		return fencedAny(m)
	case FrontmatterMinimal:
		return fencedAny(map[string]any{
			"title": dirName,
			"type":  "Index",
		})
	case FrontmatterFull:
		fallthrough
	default:
		return fencedAny(map[string]any{
			"type":          "Index",
			"title":         dirName,
			"rendered_from": "",
		})
	}
}

func fencedAny(v any) string {
	body, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(body)
	b.WriteString("---\n\n")
	return b.String()
}
