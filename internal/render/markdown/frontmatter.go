package markdown

import (
	"fmt"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
	"gopkg.in/yaml.v3"
)

// frontmatter is the YAML preamble every node file opens with: the
// machine-readable shape a tool consumes (jq, mkdocs, custom dashboards) without
// re-parsing the markdown body.
//
//specue:req:render-doc#machine-readable-frontmatter
type frontmatter struct {
	ID           string   `yaml:"id"`
	Type         string   `yaml:"type"`
	Module       string   `yaml:"module"`
	Status       string   `yaml:"status,omitempty"`
	Confidence   string   `yaml:"confidence,omitempty"`
	Title        string   `yaml:"title,omitempty"`
	Service      string   `yaml:"service,omitempty"`
	Domain       string   `yaml:"domain,omitempty"`
	Satisfies    []string `yaml:"satisfies,omitempty"`
	DecidedBy    []string `yaml:"decided_by,omitempty"`
	Realizes     []string `yaml:"realizes,omitempty"`
	RenderedFrom string   `yaml:"rendered_from,omitempty"`
}

// writeFrontmatter dispatches on the configured shape. Each shape is a small
// builder that picks fields from the populated `frontmatter` struct (which the
// node renderer fills as if for the full shape — shape selection is pure
// projection).
func writeFrontmatter(fm frontmatter, cfg Config) (string, error) {
	shape := cfg.Frontmatter
	if shape == "" {
		shape = FrontmatterFull
	}
	switch shape {
	case FrontmatterNone:
		return "", nil
	case FrontmatterFull:
		return fenced(fm)
	case FrontmatterMinimal:
		return fenced(minimalShape(fm))
	case FrontmatterMark:
		return fenced(markShape(fm, cfg.Space))
	case FrontmatterMkDocs:
		return fenced(mkdocsShape(fm))
	default:
		return "", fmt.Errorf("unknown frontmatter shape %q", shape)
	}
}

// fenced marshals any YAML value and wraps it in `---` fences.
func fenced(v any) (string, error) {
	body, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(body)
	b.WriteString("---\n\n")
	return b.String(), nil
}

// minimalShape: title, type, status only.
type minimalFM struct {
	Title  string `yaml:"title,omitempty"`
	Type   string `yaml:"type"`
	Status string `yaml:"status,omitempty"`
}

func minimalShape(fm frontmatter) minimalFM {
	return minimalFM{Title: fm.Title, Type: fm.Type, Status: fm.Status}
}

// markShape: kovetskiy/mark (Confluence) PascalCase keys.
type markFM struct {
	Title  string   `yaml:"Title"`
	Space  string   `yaml:"Space,omitempty"`
	Parent string   `yaml:"Parent,omitempty"`
	Labels []string `yaml:"Labels,omitempty"`
}

// markShape extracts Title/Space/Parent/Labels. Parent is the module's last
// path segment (with @vN stripped) — the natural Confluence parent for nodes
// of the same module.
func markShape(fm frontmatter, space string) markFM {
	out := markFM{Title: fm.Title}
	if out.Title == "" {
		out.Title = fm.ID
	}
	if space != "" {
		out.Space = space
	}
	out.Parent = moduleLeaf(fm.Module)
	if fm.Type != "" {
		out.Labels = append(out.Labels, strings.ToLower(fm.Type))
	}
	if fm.Status != "" {
		out.Labels = append(out.Labels, fm.Status)
	}
	return out
}

// mkdocsShape: MkDocs Material — title + tags. Icon is added per-type when
// the type maps to a sensible Material icon.
type mkdocsFM struct {
	Title string   `yaml:"title,omitempty"`
	Icon  string   `yaml:"icon,omitempty"`
	Tags  []string `yaml:"tags,omitempty"`
}

// materialIcons is a deliberately small map — only the icons the Material
// theme already ships with no extra packs needed.
var materialIcons = map[string]string{
	"UseCase":   "material/play-circle-outline",
	"Need":      "material/clipboard-text-outline",
	"Domain":    "material/shape-outline",
	"Port":      "material/connection",
	"Container": "material/cube-outline",
	"ADR":       "material/gavel",
	"Plan":      "material/notebook-edit-outline",
}

func mkdocsShape(fm frontmatter) mkdocsFM {
	out := mkdocsFM{Title: fm.Title}
	if icon, ok := materialIcons[fm.Type]; ok {
		out.Icon = icon
	}
	if fm.Type != "" {
		out.Tags = append(out.Tags, strings.ToLower(fm.Type))
	}
	if fm.Status != "" {
		out.Tags = append(out.Tags, fm.Status)
	}
	return out
}

// moduleLeaf returns the last `/`-separated segment of a module path with the
// trailing @vN stripped — the natural display name for a module.
func moduleLeaf(m string) string {
	s := m
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	if i := strings.LastIndex(s, "@"); i >= 0 {
		s = s[:i]
	}
	return s
}

// baseFrontmatter fills the fields every node carries: id, type, module,
// confidence, title, rendered_from. Type-specific fields (status, satisfies,
// decided_by, realizes, service, domain) are layered on by the node renderer.
func baseFrontmatter(n *compiler.ResolvedNode, ctx render.Context) frontmatter {
	id := n.ID()
	nd := n.Node()
	return frontmatter{
		ID:           id.String(),
		Type:         string(nd.Type),
		Module:       string(id.Module),
		Confidence:   string(nd.Confidence),
		Title:        nd.Title,
		RenderedFrom: ctx.Revisions[id.Module],
	}
}

// linkTo returns a relative markdown link from one file to another, computed
// from the layout. Same-module link is `<slug>.md`; cross-module link is the
// shortest forward-slash relative path between the two file locations.
func linkTo(from, to model.NodeID, layout render.Layout) string {
	return linkPath(from, to, layout)
}

// atomLink is the in-file anchor on a Need FR — Need file plus anchor on the
// atom id (the FR's element id within the Need).
func atomLink(from model.NodeID, ref model.AtomRef, layout render.Layout) string {
	base := linkTo(from, ref.Need, layout)
	return base + "#" + string(ref.Atom)
}

// strList renders []string with empty omitted, so an empty slice is dropped
// from frontmatter rather than serialised as `[]`.
func strList(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}
