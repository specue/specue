package markdown

import (
	"fmt"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Need renders a Need: frontmatter with domain/status, consumer + description,
// FR/NFR list with anchors so cross-files (UseCases satisfying them) can link
// directly to the atom.
type Need struct{ cfg Config }

func (Need) Type() model.NodeType { return model.TypeNeed }

func (nr Need) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	nd := n.Node().Body.Need
	fm := baseFrontmatter(n, ctx)
	fm.Status = string(n.Status)
	fm.Domain = refStr(nd.Domain)

	head, err := writeFrontmatter(fm, nr.cfg)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(head)
	fmt.Fprintf(&b, "# %s\n\n", titleOrSlug(n))
	if nr.cfg.WithStatusAdmonitions {
		b.WriteString(statusAdmonition(n, ctx))
	}
	fmt.Fprintf(&b, "Domain: %s\n\n", linkRef(n.ID(), nd.Domain, ctx.Layout))
	if nd.Consumer != "" {
		fmt.Fprintf(&b, "**Consumer:** %s\n\n", nd.Consumer)
	}
	if nd.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", nd.Description)
	}
	if len(nd.Atoms) > 0 {
		b.WriteString("## Requirements\n\n")
		var lookup atomSatisfierLookup
		if nr.cfg.WithStatusAdmonitions {
			lookup = buildAtomLookup(ctx, n.ID())
		}
		for _, a := range nd.Atoms {
			fmt.Fprintf(&b, "### <a id=%q></a>%s\n\n%s\n\n", string(a.ID), string(a.ID), a.Text)
			if nr.cfg.WithStatusAdmonitions {
				b.WriteString(atomInlineStatus(n.ID(), a.ID, lookup, ctx.Layout))
			}
		}
	}
	return render.FileContent(b.String()), nil
}
