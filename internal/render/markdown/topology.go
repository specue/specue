package markdown

import (
	"fmt"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Port renders a Port (L2 transport surface): frontmatter with kind/transport,
// schema ref if any, and the derived topology (produced/consumed/served/called).
type Port struct{ cfg Config }

func (Port) Type() model.NodeType { return model.TypePort }

func (pr Port) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	p := n.Node().Body.Port
	fm := baseFrontmatter(n, ctx)
	head, err := writeFrontmatter(fm, pr.cfg)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(head)
	fmt.Fprintf(&b, "# %s\n\n", titleOrSlug(n))
	fmt.Fprintf(&b, "Kind: `%s`  •  transport: `%s`\n\n", p.Kind, p.Transport)
	if p.Schema != (model.NodeRef{}) {
		fmt.Fprintf(&b, "Schema: %s\n\n", linkRef(n.ID(), p.Schema, ctx.Layout))
	}
	writeRole(&b, n.ID(), "Produced by", n.Topology.ProducedBy, ctx)
	writeRole(&b, n.ID(), "Consumed by", n.Topology.ConsumedBy, ctx)
	writeRole(&b, n.ID(), "Served by", n.Topology.ServedBy, ctx)
	writeRole(&b, n.ID(), "Called by", n.Topology.CalledBy, ctx)
	return render.FileContent(b.String()), nil
}

func writeRole(b *strings.Builder, from model.NodeID, label string, ids []model.NodeID, ctx render.Context) {
	if len(ids) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n", label)
	for _, id := range ids {
		fmt.Fprintf(b, "- [%s](%s)\n", linkText(id), linkTo(from, id, ctx.Layout))
	}
	b.WriteString("\n")
}

// Container renders a Container (boundary box / external actor): frontmatter
// with kind/boundary, the prose body if any.
type Container struct{ cfg Config }

func (Container) Type() model.NodeType { return model.TypeContainer }

func (cr Container) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	c := n.Node().Body.Container
	fm := baseFrontmatter(n, ctx)
	head, err := writeFrontmatter(fm, cr.cfg)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(head)
	fmt.Fprintf(&b, "# %s\n\n", titleOrSlug(n))
	fmt.Fprintf(&b, "Kind: `%s`  •  boundary: %t\n\n", c.Kind, c.Boundary)
	if prose := n.Node().Body.Prose; prose != "" {
		b.WriteString(prose)
		if !strings.HasSuffix(prose, "\n") {
			b.WriteString("\n")
		}
	}
	return render.FileContent(b.String()), nil
}
