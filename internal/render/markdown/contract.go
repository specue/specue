package markdown

import (
	"fmt"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Contract renders a Contract node as one .md file: frontmatter carrying the
// machine-readable shape, then a body identical in content to `describe`
// (trigger, invariants with satisfies/decided_by, postconditions, realizes).
type Contract struct{ cfg Config }

func (Contract) Type() model.NodeType { return model.TypeContract }

//specue:req:render-doc#cross-links-resolve-as-markdown
func (u Contract) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	uc := n.Node().Body.Contract
	fm := baseFrontmatter(n, ctx)
	fm.Status = string(n.Status)
	fm.Service = refStr(uc.Service)
	fm.Satisfies = collectUCSatisfies(uc)
	fm.DecidedBy = collectUCDecidedBy(uc)
	fm.Realizes = idStrings(n.Realizes)

	head, err := writeFrontmatter(fm, u.cfg)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(head)
	fmt.Fprintf(&b, "# %s\n\n", titleOrSlug(n))
	if u.cfg.WithStatusAdmonitions {
		b.WriteString(statusAdmonition(n, ctx))
	}
	fmt.Fprintf(&b, "Service: %s  •  binding: %s  •  interaction: %s\n\n",
		linkRef(n.ID(), uc.Service, ctx.Layout), uc.Binding, uc.Interaction)
	if uc.Trigger != "" {
		fmt.Fprintf(&b, "**Trigger.** %s\n\n", uc.Trigger)
	}

	writeElements(&b, n, uc.Elements, model.KindInvariant, "Invariants", ctx, u.cfg)
	writeElements(&b, n, uc.Elements, model.KindVariation, "Variations", ctx, u.cfg)
	writeElements(&b, n, uc.Elements, model.KindPost, "Postconditions", ctx, u.cfg)
	writeElements(&b, n, uc.Elements, model.KindPre, "Preconditions", ctx, u.cfg)

	if len(n.Realizes) > 0 {
		b.WriteString("## Realizes\n\n")
		for _, sid := range n.Realizes {
			fmt.Fprintf(&b, "- [%s](%s)\n", linkText(sid), linkTo(n.ID(), sid, ctx.Layout))
		}
		b.WriteString("\n")
	}

	return render.FileContent(b.String()), nil
}

// writeElements emits one section per element kind, with anchors on named
// elements (so cross-files can link `#<element-id>`). Empty kinds are skipped.
func writeElements(b *strings.Builder, n *compiler.ResolvedNode, els []model.Element, kind model.ElementKind, title string, ctx render.Context, cfg Config) {
	first := true
	for _, e := range els {
		if e.Kind != kind {
			continue
		}
		if first {
			fmt.Fprintf(b, "## %s\n\n", title)
			first = false
		}
		writeElement(b, n.ID(), e, ctx)
		if cfg.WithStatusAdmonitions {
			b.WriteString(elementInlineStatus(n, e))
		}
	}
	if !first {
		b.WriteString("\n")
	}
}

// writeElement renders one element: an `<a id>` anchor on named ones, the text,
// when/then for variations, and any satisfies/decided_by as markdown links.
func writeElement(b *strings.Builder, from model.NodeID, e model.Element, ctx render.Context) {
	if e.ID != "" {
		fmt.Fprintf(b, "### <a id=%q></a>%s\n\n", string(e.ID), string(e.ID))
	} else {
		b.WriteString("### —\n\n")
	}
	if e.Text != "" {
		fmt.Fprintf(b, "%s\n\n", e.Text)
	}
	if e.When != "" || e.Then != "" {
		fmt.Fprintf(b, "*When* %s → *then* %s\n\n", e.When, e.Then)
	}
	if len(e.Satisfies) > 0 {
		b.WriteString("Satisfies: ")
		parts := make([]string, len(e.Satisfies))
		for i, s := range e.Satisfies {
			parts[i] = fmt.Sprintf("[%s](%s)", linkAtomText(s), atomLink(from, s, ctx.Layout))
		}
		fmt.Fprintf(b, "%s\n\n", strings.Join(parts, ", "))
	}
	if len(e.DecidedBy) > 0 {
		b.WriteString("Decided by: ")
		parts := make([]string, len(e.DecidedBy))
		for i, d := range e.DecidedBy {
			parts[i] = fmt.Sprintf("[%s](%s)", linkText(d), linkTo(from, d, ctx.Layout))
		}
		fmt.Fprintf(b, "%s\n\n", strings.Join(parts, ", "))
	}
}

// collectUCSatisfies flattens every element's satisfies into the node-level
// frontmatter list — a deduped, sorted view for tooling that wants the
// node-level summary without walking elements.
func collectUCSatisfies(uc *model.ContractBody) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, e := range uc.Elements {
		for _, s := range e.Satisfies {
			key := s.String()
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	return strList(out)
}

// collectUCDecidedBy flattens every element's decided_by into the node-level
// frontmatter list, deduped.
func collectUCDecidedBy(uc *model.ContractBody) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, e := range uc.Elements {
		for _, d := range e.DecidedBy {
			key := d.String()
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	return strList(out)
}

// linkRef renders a NodeRef as a markdown link, or as plain text if the ref
// is the zero value (an unset optional field).
func linkRef(from model.NodeID, ref model.NodeRef, layout render.Layout) string {
	if ref == (model.NodeRef{}) {
		return "—"
	}
	return fmt.Sprintf("[%s](%s)", linkText(ref), linkTo(from, ref, layout))
}

// refStr is the bare module:slug form for frontmatter (or empty for the zero
// ref).
func refStr(ref model.NodeRef) string {
	if ref == (model.NodeRef{}) {
		return ""
	}
	return ref.String()
}

// idStrings flattens a NodeID slice into string form for frontmatter.
func idStrings(ids []model.NodeID) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

// titleOrSlug falls back to the slug when the node carries no title (rare for
// authored nodes, but the renderer stays robust).
func titleOrSlug(n *compiler.ResolvedNode) string {
	if t := n.Node().Title; t != "" {
		return t
	}
	return string(n.ID().Slug)
}
