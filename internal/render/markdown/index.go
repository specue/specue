package markdown

import (
	"fmt"
	"sort"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Index renders the README at the tree root: every module with a count by node
// type and a flat list linking each node's .md file. This is the entry point a
// reader opens first on a markdown forge.
type Index struct{ cfg Config }

func (Index) Path() render.RelPath { return "README.md" }

func (i Index) Render(byModule map[model.ModulePath][]*compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	mods := make([]model.ModulePath, 0, len(byModule))
	for m := range byModule {
		mods = append(mods, m)
	}
	sort.Slice(mods, func(i, j int) bool { return string(mods[i]) < string(mods[j]) })

	var b strings.Builder
	b.WriteString("# Spec\n\n")
	b.WriteString("This documentation is rendered from a Specue spec. Each node is one file; cross-references are markdown links.\n\n")
	b.WriteString("## Modules\n\n")
	for _, m := range mods {
		nodes := byModule[m]
		counts := countByType(nodes)
		rev := ctx.Revisions[m]
		fmt.Fprintf(&b, "### `%s`", stripModulePrefix(string(m), i.cfg.StripPrefix))
		if rev != "" {
			fmt.Fprintf(&b, " — `%s`", short(rev))
		}
		b.WriteString("\n\n")
		if len(counts) > 0 {
			fmt.Fprintf(&b, "%s\n\n", strings.Join(counts, " • "))
		}
		for _, n := range nodes {
			path := ctx.Layout.NodePath(n.ID())
			title := titleOrSlug(n)
			label := stripModulePrefix(string(n.ID().Module), i.cfg.StripPrefix) + ":" + string(n.ID().Slug)
			fmt.Fprintf(&b, "- [`%s`](%s) — %s\n", label, string(path), title)
		}
		b.WriteString("\n")
	}
	return render.FileContent(b.String()), nil
}

// countByType returns "UseCase: 12", "Need: 8", … in a stable order.
func countByType(nodes []*compiler.ResolvedNode) []string {
	c := map[model.NodeType]int{}
	for _, n := range nodes {
		c[n.Node().Type]++
	}
	order := []model.NodeType{
		model.TypeDomain, model.TypeNeed, model.TypeUseCase,
		model.TypeContainer, model.TypePort, model.TypeADR, model.TypePlan,
	}
	var out []string
	for _, t := range order {
		if c[t] > 0 {
			out = append(out, fmt.Sprintf("%s: %d", t, c[t]))
		}
	}
	return out
}

// short truncates a sha to its first 12 chars (the convention git uses for
// human-display refs).
func short(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
