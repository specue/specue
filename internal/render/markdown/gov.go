package markdown

import (
	"fmt"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// ADR renders an ADR: frontmatter with lifecycle, the prose body as the markdown
// content. ADRs are leaves in the link graph (other nodes link TO an ADR via
// decided_by, but the ADR itself carries no outbound spec refs).
type ADR struct{ cfg Config }

func (ADR) Type() model.NodeType { return model.TypeADR }

func (a ADR) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	fm := baseFrontmatter(n, ctx)
	if gov := n.Node().Body.Gov; gov != nil {
		fm.Status = string(gov.Lifecycle)
	}
	head, err := writeFrontmatter(fm, a.cfg)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(head)
	fmt.Fprintf(&b, "# %s\n\n", titleOrSlug(n))
	if a.cfg.WithStatusAdmonitions {
		b.WriteString(statusAdmonition(n, ctx))
	}
	if prose := n.Node().Body.Prose; prose != "" {
		b.WriteString(prose)
		if !strings.HasSuffix(prose, "\n") {
			b.WriteString("\n")
		}
	}
	return render.FileContent(b.String()), nil
}

// Plan renders a Plan record: frontmatter with lifecycle and branch, the prose
// body as the markdown content.
type Plan struct{ cfg Config }

func (Plan) Type() model.NodeType { return model.TypePlan }

func (p Plan) Render(n *compiler.ResolvedNode, ctx render.Context) (render.FileContent, error) {
	fm := baseFrontmatter(n, ctx)
	if gov := n.Node().Body.Gov; gov != nil {
		fm.Status = string(gov.Lifecycle)
	}
	head, err := writeFrontmatter(fm, p.cfg)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(head)
	fmt.Fprintf(&b, "# %s\n\n", titleOrSlug(n))
	if p.cfg.WithStatusAdmonitions {
		b.WriteString(statusAdmonition(n, ctx))
	}
	if gov := n.Node().Body.Gov; gov != nil && gov.Branch != "" {
		fmt.Fprintf(&b, "Branch: `%s`\n\n", gov.Branch)
	}
	if prose := n.Node().Body.Prose; prose != "" {
		b.WriteString(prose)
		if !strings.HasSuffix(prose, "\n") {
			b.WriteString("\n")
		}
	}
	return render.FileContent(b.String()), nil
}
