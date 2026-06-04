// Package markdown is the default set of renderers: one file per node under
// <module-dir>/<slug>.md plus a README at the root listing modules. Every file
// opens with YAML frontmatter for the machine-readable shape; the body is the
// describe-style narrative with relative markdown links between nodes.
package markdown

import (
	"path"
	"strings"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// Layout is the default (flat) layout, kept for backward compatibility with
// callers that previously instantiated `markdown.Layout{}` directly.
type Layout = flatLayout

// flatLayout implements the historical single-directory-per-module shape:
// every module collapses to one dir, slashes and `@` become dashes, so the
// tree stays two-deep regardless of how many slashes a module path carries.
type flatLayout struct {
	stripPrefix string
}

// NodePath returns the canonical path of a node's file inside the tree.
func (l flatLayout) NodePath(id model.NodeID) render.RelPath {
	return render.RelPath(l.dir(id.Module) + "/" + string(id.Slug) + ".md")
}

func (l flatLayout) dir(m model.ModulePath) string {
	s := stripModulePrefix(string(m), l.stripPrefix)
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "@", "-")
	return s
}

// treeLayout splits the module path on `/` into nested directories, dropping
// the trailing `@vN` (CUE module version) from the leaf directory. The slug
// file sits at the deepest level.
type treeLayout struct {
	stripPrefix string
}

func (l treeLayout) NodePath(id model.NodeID) render.RelPath {
	return render.RelPath(l.dir(id.Module) + "/" + string(id.Slug) + ".md")
}

//specue:req:render-doc#tree-layout-mirrors-module-path
//specue:req:render-doc#strip-prefix-shortens-paths
func (l treeLayout) dir(m model.ModulePath) string {
	s := stripModulePrefix(string(m), l.stripPrefix)
	// strip @vN from final segment
	if at := strings.LastIndex(s, "@"); at >= 0 {
		if slash := strings.LastIndex(s, "/"); slash < at {
			s = s[:at]
		}
	}
	return s
}

// stripModulePrefix removes prefix from s when it matches and returns the
// remainder without a leading slash, so a strip of "specue.io" against
// "specue.io/governance@v0" yields "governance@v0" — not "/governance@v0"
// which would later flatten to "-governance-v0". A trailing slash on the
// prefix is tolerated, with or without one in the input.
func stripModulePrefix(s, prefix string) string {
	if prefix == "" {
		return s
	}
	p := strings.TrimSuffix(prefix, "/")
	if p == "" {
		return s
	}
	if strings.HasPrefix(s, p+"/") {
		return s[len(p)+1:]
	}
	if s == p {
		return ""
	}
	return s
}

// linkText is the human-visible label of a cross-link in the body — the slug
// (with #element for an atom ref). The URL still carries the full module path
// to keep links unambiguous; the visible text drops the module prefix so a
// reader is not asked to parse `gitlab.example.com/gp/...:add-task` to find
// "add-task". For frontmatter and the index — where machine consumers want the
// full id — the renderer still uses .String().
//
//specue:req:render-doc#link-text-is-slug-only
func linkText(id model.NodeID) string {
	return string(id.Slug)
}

// linkAtomText is linkText for an atom: slug#fr-NN.
//
//specue:req:render-doc#link-text-is-slug-only
func linkAtomText(ref model.AtomRef) string {
	return string(ref.Need.Slug) + "#" + string(ref.Atom)
}

// linkPath computes the relative-link from one node file to another, using the
// layout's actual paths (so flat and tree layouts each produce correct ../).
func linkPath(from, to model.NodeID, layout render.Layout) string {
	if from.Module == to.Module {
		// same module → same directory → just slug.md
		return string(to.Slug) + ".md"
	}
	fromPath := string(layout.NodePath(from))
	toPath := string(layout.NodePath(to))
	rel, err := relPath(path.Dir(fromPath), toPath)
	if err != nil {
		return "../" + toPath
	}
	return rel
}

// relPath computes a forward-slash relative path from base dir to target.
func relPath(base, target string) (string, error) {
	bp := strings.Split(strings.Trim(base, "/"), "/")
	tp := strings.Split(strings.Trim(target, "/"), "/")
	// drop common prefix
	i := 0
	for i < len(bp) && i < len(tp)-1 && bp[i] == tp[i] {
		i++
	}
	var parts []string
	for j := i; j < len(bp); j++ {
		parts = append(parts, "..")
	}
	parts = append(parts, tp[i:]...)
	return strings.Join(parts, "/"), nil
}
