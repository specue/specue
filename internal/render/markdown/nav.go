package markdown

import (
	"fmt"
	"sort"
	"strings"

	"github.com/specue/specue/internal/render"
)

// wrapWithNavSnippet returns a Renderer-like value that delegates to inner and
// adds one extra Tree entry — an MkDocs-style `nav:` snippet covering every
// rendered node, grouped by module-as-tree. Used when Config.NavSnippetPath
// is set.
//
// Implementation: we cannot replace *render.Renderer's behaviour from outside
// the package, so wrapWithNavSnippet builds a NEW Renderer with the same set
// of node renderers PLUS a navIndex stand-in. The standard IndexRenderer slot
// is already taken by Index{}, so we instead piggy-back on Index by stashing
// the nav path on Index{} and writing the snippet alongside the README — but
// that mixes responsibilities. Cleaner: emit the nav by running Render once
// then post-processing the Tree. We do that in a thin wrapper struct that
// satisfies the Renderer.Render shape via a new Renderer (we just re-call).
//specue:req:render-doc#nav-snippet-on-request
func wrapWithNavSnippet(inner *render.Renderer, navPath string, layout render.Layout) *render.Renderer {
	// Stash the nav config on a package-level map keyed by the renderer pointer
	// — we look it up after Render returns. Cleaner than threading a new type
	// through render.Renderer because the orchestrator deliberately does not
	// know about post-processors.
	navHook.store(inner, navConfig{Path: navPath, Layout: layout})
	return inner
}

type navConfig struct {
	Path   string
	Layout render.Layout
}

// navHookRegistry maps a Renderer to its nav config. Lookups happen in the
// CLI wrapper after Render — see runRender's post-processing.
type navHookRegistry struct {
	m map[*render.Renderer]navConfig
}

func (r *navHookRegistry) store(rr *render.Renderer, cfg navConfig) {
	if r.m == nil {
		r.m = map[*render.Renderer]navConfig{}
	}
	r.m[rr] = cfg
}

func (r *navHookRegistry) load(rr *render.Renderer) (navConfig, bool) {
	c, ok := r.m[rr]
	return c, ok
}

var navHook = &navHookRegistry{}

// EmitNavSnippet returns the (relpath, content) for the nav snippet attached
// to r, or ("", "", false) if none. The CLI calls this after Render to add
// the entry to the Tree.
func EmitNavSnippet(r *render.Renderer, tree render.Tree) (render.RelPath, render.FileContent, bool) {
	cfg, ok := navHook.load(r)
	if !ok {
		return "", "", false
	}
	return render.RelPath(cfg.Path), render.FileContent(buildNavSnippet(tree, cfg.Layout)), true
}

// buildNavSnippet walks the tree's .md entries (excluding the nav file itself
// and README.md) and emits an MkDocs-compatible nav:. The tree is built from
// each file's path components — flat layouts produce a one-level nav (one
// entry per module), tree layouts produce nested entries that mirror the dir
// structure.
func buildNavSnippet(tree render.Tree, _ render.Layout) string {
	type entry struct{ path, title string }
	var leaves []entry
	indexes := map[string]string{} // dir → path of its index.md
	for rel := range tree {
		s := string(rel)
		if s == "README.md" || !strings.HasSuffix(s, ".md") {
			continue
		}
		base := s
		if i := strings.LastIndex(s, "/"); i >= 0 {
			base = s[i+1:]
		}
		if base == "index.md" {
			dir := strings.TrimSuffix(s, base)
			dir = strings.TrimSuffix(dir, "/") // "" for root
			indexes[dir] = s
			continue
		}
		leaves = append(leaves, entry{path: s, title: titleFromPath(s)})
	}
	sort.Slice(leaves, func(i, j int) bool { return leaves[i].path < leaves[j].path })

	root := &navNode{}
	for _, e := range leaves {
		parts := strings.Split(e.path, "/")
		root.insert(parts, e.path, e.title)
	}
	root.collapseSingleChildChains()
	root.attachIndexes(indexes, "")
	var b strings.Builder
	b.WriteString("nav:\n")
	if rootIdx, ok := indexes[""]; ok {
		// The root index.md leads the nav: when navigation.indexes is on,
		// Material shows it as the site landing page.
		fmt.Fprintf(&b, "  - %s\n", rootIdx)
	}
	root.write(&b, 1)
	return b.String()
}

// attachIndexes binds each directory's index.md to its corresponding navNode
// so the emitter can write `- name: dir/index.md` (mkdocs Material's
// navigation.indexes landing-page convention) instead of an empty group.
// Called AFTER collapseSingleChildChains so the collapsed name still maps
// back to a directory path that the index map knows.
func (n *navNode) attachIndexes(indexes map[string]string, prefix string) {
	for _, c := range n.children {
		if c.path != "" {
			continue // leaf, no index
		}
		dir := prefix
		if dir != "" {
			dir += "/"
		}
		dir += c.name
		if idx, ok := indexes[dir]; ok {
			c.indexPath = idx
		}
		c.attachIndexes(indexes, dir)
	}
}

type navNode struct {
	name      string // dir name OR leaf title
	path      string // file path (leaf only)
	indexPath string // index.md for a group, when navigation.indexes is in play
	children  []*navNode
	byName    map[string]*navNode
}

func (n *navNode) insert(parts []string, fullPath, title string) {
	if len(parts) == 1 {
		n.children = append(n.children, &navNode{name: title, path: fullPath})
		return
	}
	head, rest := parts[0], parts[1:]
	if n.byName == nil {
		n.byName = map[string]*navNode{}
	}
	child, ok := n.byName[head]
	if !ok {
		child = &navNode{name: head}
		n.byName[head] = child
		n.children = append(n.children, child)
	}
	child.insert(rest, fullPath, title)
}

// collapseSingleChildChains walks the tree and merges any group whose only
// child is itself a group (no leaves of its own) into one entry with a
// slash-joined name. Done bottom-up so a chain like a/b/c/leaf.md collapses
// to "a/b/c" in a single pass. The root is never collapsed — it carries the
// top-level entries the nav: list shows.
//
//specue:req:render-doc#nav-snippet-on-request
func (n *navNode) collapseSingleChildChains() {
	for _, c := range n.children {
		c.collapseSingleChildChains()
	}
	merged := make([]*navNode, 0, len(n.children))
	for _, c := range n.children {
		// A group is a non-leaf (path == "") with exactly one child; the child
		// must itself be a group (a single leaf stays as a leaf under its
		// parent so a one-node module is `parent: parent/leaf.md`, not a
		// title-less wrapper).
		for c.path == "" && len(c.children) == 1 && c.children[0].path == "" {
			child := c.children[0]
			c.name = c.name + "/" + child.name
			c.children = child.children
		}
		merged = append(merged, c)
	}
	n.children = merged
	// byName is no longer consulted after collapse; clear it so nothing stale
	// outlives the rebuild.
	n.byName = nil
}

func (n *navNode) write(b *strings.Builder, depth int) {
	// stable order: sort by name
	sort.Slice(n.children, func(i, j int) bool { return n.children[i].name < n.children[j].name })
	indent := strings.Repeat("  ", depth)
	for _, c := range n.children {
		if c.path != "" {
			fmt.Fprintf(b, "%s- %s: %s\n", indent, yamlQuote(c.name), c.path)
			continue
		}
		fmt.Fprintf(b, "%s- %s:\n", indent, yamlQuote(c.name))
		// A group's own index.md leads its children when navigation.indexes
		// is on — Material shows it as the section landing page.
		if c.indexPath != "" {
			fmt.Fprintf(b, "%s- %s\n", strings.Repeat("  ", depth+1), c.indexPath)
		}
		c.write(b, depth+1)
	}
}

// titleFromPath strips the .md and returns the basename — the leaf slug is
// the natural display title without re-parsing the file.
func titleFromPath(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		p = p[i+1:]
	}
	return strings.TrimSuffix(p, ".md")
}

// yamlQuote wraps a name in quotes when it contains characters that would
// otherwise break YAML scalar parsing (colon, space). Conservative: only quote
// when needed.
func yamlQuote(s string) string {
	if strings.ContainsAny(s, ":#") {
		return "\"" + strings.ReplaceAll(s, "\"", "\\\"") + "\""
	}
	return s
}
