// Package render projects a resolved graph onto a documentation tree — one file
// per node plus a root index — produced by a set of per-node-type renderers.
//
// The transform is pure: it takes a resolved graph and per-module source
// revisions and returns a map of file paths to contents. Writing to disk is
// the caller's job; this makes the render trivially testable.
//
// Renderers are pluggable: NodeRenderer formats one node, IndexRenderer the
// root listing. The markdown subpackage carries the default set. Swapping in
// another format (asciidoc, html) is a different set of renderers, not a
// rewrite of this layer.
package render

import (
	"fmt"
	"sort"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

// RelPath is a forward-slash, no-leading-slash path relative to the tree root —
// the form the caller resolves against its destination directory. A typed
// string so it is never confused with arbitrary text or an absolute path.
type RelPath string

// FileContent is the textual body of one file in the tree. A typed string so
// the Tree's key/value sides cannot be swapped at a callsite.
type FileContent string

// Tree is the renderer's output: every file the caller writes, keyed by its
// path relative to the destination root.
type Tree map[RelPath]FileContent

// Context carries everything a renderer needs beyond the node itself: the full
// resolved graph (for cross-references), the source revision of each module
// (frontmatter `rendered_from`), and the path layout (so renderers compute
// relative links the same way the tree is laid out).
type Context struct {
	Graph     *compiler.ResolvedGraph
	Revisions map[model.ModulePath]string // module path → git sha
	Layout    Layout
}

// NodeRenderer formats one node into the textual content of its file. One
// renderer per NodeType; Render is called on a node of its declared type only.
type NodeRenderer interface {
	Type() model.NodeType
	Render(n *compiler.ResolvedNode, ctx Context) (FileContent, error)
}

// IndexRenderer formats the root index — the entry file a reader opens first.
// It sees every node grouped by module, in deterministic order.
type IndexRenderer interface {
	// Path is the index's path inside the tree (e.g. "README.md").
	Path() RelPath
	Render(byModule map[model.ModulePath][]*compiler.ResolvedNode, ctx Context) (FileContent, error)
}

// Layout decides where a node's file lives. Renderers consult it through
// Context to compute cross-links — both sides must agree on the path.
type Layout interface {
	// NodePath is the canonical relative path of a node's file inside the tree.
	NodePath(id model.NodeID) RelPath
}

// Renderer is the orchestrator. It dispatches each node to the renderer that
// matches its type, optionally builds the index, and returns the tree.
type Renderer struct {
	nodes  map[model.NodeType]NodeRenderer
	index  IndexRenderer
	layout Layout
}

// New builds a Renderer from a set of node renderers, the index renderer and
// the layout. A second renderer for an already-claimed type replaces the first.
func New(nodes []NodeRenderer, index IndexRenderer, layout Layout) *Renderer {
	m := make(map[model.NodeType]NodeRenderer, len(nodes))
	for _, r := range nodes {
		m[r.Type()] = r
	}
	return &Renderer{nodes: m, index: index, layout: layout}
}

// Layout exposes the Renderer's path layout — post-processors that need to
// build a per-directory view of the tree consult it to map a NodeID back to
// the same RelPath the Renderer used.
func (r *Renderer) Layout() Layout { return r.layout }

// Input is what Render needs: the resolved graph (the same one validate and
// describe see — render is a derived view, never a second source) and the
// per-module revision map that frontmatter records as `rendered_from`.
//
//specue:req:render-doc#derived-from-resolved-graph
type Input struct {
	Graph     *compiler.ResolvedGraph
	Revisions map[model.ModulePath]string
}

// Render projects the graph onto a Tree. Each node lands at layout.NodePath(id)
// with content from its type's renderer; the index lands at index.Path().
// A node whose type has no registered renderer is skipped (an unknown future
// type is not fatal — older renderers still produce something useful).
//
//specue:req:render-doc#one-file-per-node
func (r *Renderer) Render(in Input) (Tree, error) {
	if in.Graph == nil {
		return nil, fmt.Errorf("render: nil graph")
	}
	ctx := Context{Graph: in.Graph, Revisions: in.Revisions, Layout: r.layout}

	byModule := map[model.ModulePath][]*compiler.ResolvedNode{}
	for n := range in.Graph.Nodes() {
		byModule[n.ID().Module] = append(byModule[n.ID().Module], n)
	}
	for _, ns := range byModule {
		sortByID(ns)
	}

	out := Tree{}
	for _, ns := range byModule {
		for _, n := range ns {
			nr, ok := r.nodes[n.Node().Type]
			if !ok {
				continue
			}
			content, err := nr.Render(n, ctx)
			if err != nil {
				return nil, fmt.Errorf("render %s: %w", n.ID(), err)
			}
			out[r.layout.NodePath(n.ID())] = content
		}
	}

	if r.index != nil {
		content, err := r.index.Render(byModule, ctx)
		if err != nil {
			return nil, fmt.Errorf("render index: %w", err)
		}
		// An IndexRenderer that returns an empty Path() is a deliberate no-op:
		// a post-processor (e.g. per-folder index.md pages) owns the index
		// slot and the orchestrator must not write a phantom file at "".
		if p := r.index.Path(); p != "" {
			out[p] = content
		}
	}
	return out, nil
}

// sortByID orders a slice of resolved nodes by their string identity, so the
// output is deterministic (same input → byte-identical tree).
func sortByID(ns []*compiler.ResolvedNode) {
	sort.Slice(ns, func(i, j int) bool { return ns[i].ID().String() < ns[j].ID().String() })
}
