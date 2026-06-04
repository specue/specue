package markdown

import (
	"fmt"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Domain renders a Domain node (audience-facing root). Carries no specific
// body beyond title and prose; Needs link to it.
type Domain struct{ cfg Config }

func (Domain) Type() model.NodeType { return model.TypeDomain }

func (d Domain) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	fm := baseFrontmatter(n, ctx)
	head, err := writeFrontmatter(fm, d.cfg)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(head)
	fmt.Fprintf(&b, "# %s\n\n", titleOrSlug(n))
	if prose := n.Node().Body.Prose; prose != "" {
		b.WriteString(prose)
		if !strings.HasSuffix(prose, "\n") {
			b.WriteString("\n")
		}
	}
	return render.FileContent(b.String()), nil
}
