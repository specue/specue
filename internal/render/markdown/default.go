package markdown

import "github.com/specue/specue/internal/render"

// Default returns the full default markdown Renderer: one NodeRenderer per
// authored node type, the Index for the root, and the standard Layout. The
// output is byte-identical to the historical renderer — every preset is opt-in
// via New(Config{...}).
//
//specue:req:render-doc#defaults-need-no-flags
func Default() *render.Renderer { return New(Config{}) }

// New builds a markdown Renderer configured for one of the supported presets.
// The zero Config matches Default() exactly.
//
//specue:req:render-doc#layout-is-chosen-per-run
//specue:req:render-doc#frontmatter-shape-is-chosen
func New(cfg Config) *render.Renderer {
	if cfg.Layout == "" {
		cfg.Layout = LayoutFlat
	}
	if cfg.Frontmatter == "" {
		cfg.Frontmatter = FrontmatterFull
	}
	var layout render.Layout
	switch cfg.Layout {
	case LayoutTree:
		layout = treeLayout{stripPrefix: cfg.StripPrefix}
	default:
		layout = flatLayout{stripPrefix: cfg.StripPrefix}
	}
	nodes := []render.NodeRenderer{
		Contract{cfg: cfg}, Need{cfg: cfg}, ADR{cfg: cfg}, Plan{cfg: cfg},
		Port{cfg: cfg}, Container{cfg: cfg}, Domain{cfg: cfg},
	}
	var idx render.IndexRenderer = Index{cfg: cfg}
	if cfg.WithIndexPages {
		// With per-folder index.md pages, the root README.md is replaced by
		// an index.md at the root — the per-folder pass owns every index.
		idx = nopIndex{}
	}
	r := render.New(nodes, idx, layout)
	if cfg.NavSnippetPath != "" {
		// Wrap the renderer so the nav snippet is added as one extra Tree entry.
		r = wrapWithNavSnippet(r, cfg.NavSnippetPath, layout)
	}
	if cfg.WithIndexPages {
		r = wrapWithIndexPages(r, cfg)
	}
	if cfg.WithTagsPage {
		r = wrapWithTagsPage(r, cfg)
	}
	return r
}
