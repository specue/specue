package markdown

import (
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/specue/specue/internal/render"
)

// Collisions describes node-vs-folder name collisions in a rendered tree: a
// node whose file `<dir>/<slug>.md` shares its name with a sibling directory
// `<dir>/<slug>/` that already holds OTHER node files. With MkDocs Material's
// use_directory_urls=true both paths route to the same URL, so an auto-index
// at `<dir>/<slug>/index.md` (from --with-index-pages) would mask the node's
// own content.
//
// A Collisions is built once via DetectCollisions, then:
//   - passed to EmitIndexPages so it skips dirs that a node will fill, and
//   - Apply'd at the end of the pipeline to move each colliding file to
//     `<dir>/<slug>/index.md` and rewrite cross-file references.
//
// A nil *Collisions is valid everywhere: ShouldSkip returns false, Apply is
// a no-op. Callers that don't care about collisions pass nil.
type Collisions struct {
	moves      []collisionMove
	suppressed map[string]bool
}

type collisionMove struct{ from, to, dir string }

// IndexSkipper is the predicate EmitIndexPages consults to decide whether to
// emit an auto-index for a directory. *Collisions implements it; a caller may
// pass nil to mean "never skip".
type IndexSkipper interface {
	ShouldSkip(dir string) bool
}

// ShouldSkip reports whether dir's index.md slot is already claimed by a
// node-as-index. nil receiver returns false so a caller can pass a nil
// *Collisions when they have no collisions to apply.
func (c *Collisions) ShouldSkip(dir string) bool {
	if c == nil {
		return false
	}
	return c.suppressed[dir]
}

// PathResolver returns the post-Apply path of a tree-relative target. If the
// path will be moved by Apply, returns the moved location; otherwise returns
// path unchanged. Emitters that produce file references via Layout.NodePath
// (which doesn't know about collisions) wrap each path through this so their
// output matches the eventual tree shape.
type PathResolver interface {
	ResolvePath(path string) string
}

// ResolvePath returns the post-Apply path for a tree-relative target. nil
// receiver and unknown paths return path unchanged.
func (c *Collisions) ResolvePath(path string) string {
	if c == nil {
		return path
	}
	for _, mv := range c.moves {
		if mv.from == path {
			return mv.to
		}
	}
	return path
}

// DetectCollisions scans the tree and returns the move list and the set of
// suppressed directories. Does NOT mutate the tree. Returns a non-nil zero
// value when no collisions exist, so the caller can always pass it on.
//
//specue:req:render-doc#node-name-collisions-resolved
func DetectCollisions(tree render.Tree) *Collisions {
	if tree == nil {
		return &Collisions{}
	}

	// Every directory that holds at least one .md file in the tree.
	dirHasFiles := map[string]bool{}
	for rel := range tree {
		s := string(rel)
		if !strings.HasSuffix(s, ".md") {
			continue
		}
		if i := strings.LastIndex(s, "/"); i >= 0 {
			dirHasFiles[s[:i]] = true
		}
	}

	// Sorted iteration keeps the result deterministic across runs.
	var paths []string
	for rel := range tree {
		paths = append(paths, string(rel))
	}
	slices.Sort(paths)

	c := &Collisions{suppressed: map[string]bool{}}
	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			continue
		}
		dir, base := splitPath(p)
		if base == "index.md" || base == "README.md" {
			continue
		}
		slug := strings.TrimSuffix(base, ".md")
		collideDir := slug
		if dir != "" {
			collideDir = dir + "/" + slug
		}
		// A collision exists when some other .md file lives at or under
		// collideDir.
		if !anyFileUnder(dirHasFiles, collideDir) {
			continue
		}
		newPath := collideDir + "/index.md"
		// Skip if the new slot is already occupied — would overwrite.
		if _, exists := tree[render.RelPath(newPath)]; exists {
			continue
		}
		c.moves = append(c.moves, collisionMove{from: p, to: newPath, dir: collideDir})
		c.suppressed[collideDir] = true
	}
	return c
}

// Apply moves each colliding file to its new path, rewrites in-file relative
// links by prepending "../" so they still resolve from the deeper location,
// and walks every other file rewriting markdown links whose absolute target
// landed on a moved file. nil receiver and empty Collisions are no-ops.
//
//specue:req:render-doc#node-name-collisions-resolved
func (c *Collisions) Apply(tree render.Tree) {
	if c == nil || len(c.moves) == 0 || tree == nil {
		return
	}
	for _, mv := range c.moves {
		body := string(tree[render.RelPath(mv.from)])
		body = prefixRelativeLinks(body)
		tree[render.RelPath(mv.to)] = render.FileContent(body)
		delete(tree, render.RelPath(mv.from))
	}
	movedTargets := make(map[string]string, len(c.moves))
	for _, mv := range c.moves {
		movedTargets[mv.from] = mv.to
	}
	rewriteCrossLinks(tree, movedTargets)
}

// splitPath splits a forward-slash tree path into (dir, base). dir is "" for
// root-level files.
func splitPath(p string) (dir, base string) {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[:i], p[i+1:]
	}
	return "", p
}

// anyFileUnder reports whether any directory in dirs equals collideDir or
// sits beneath it. The set value is irrelevant; we only check membership of
// the prefix.
func anyFileUnder(dirs map[string]bool, collideDir string) bool {
	for d := range dirs {
		if d == collideDir || strings.HasPrefix(d, collideDir+"/") {
			return true
		}
	}
	return false
}

// linkPattern matches `](path)` where path is non-absolute. Captures the path
// so prefixRelativeLinks can prepend "../".
var linkPattern = regexp.MustCompile(`\]\((?P<p>[^)#][^)]*)\)`)

// crossLinkPattern matches `](path)` and `](path#anchor)`. Captures the path
// and the optional anchor separately so the anchor survives a rewrite.
var crossLinkPattern = regexp.MustCompile(`\]\((?P<p>[^)#][^)#]*)(?P<a>#[^)]*)?\)`)

// prefixRelativeLinks rewrites every relative markdown link in body by
// prepending "../" so it still resolves after the file moves one dir deeper.
// Anchor-only links (#foo) and absolute URLs (http://, https://, leading /)
// are untouched.
func prefixRelativeLinks(body string) string {
	return linkPattern.ReplaceAllStringFunc(body, func(m string) string {
		sub := linkPattern.FindStringSubmatch(m)
		p := sub[1]
		if isAbsoluteOrScheme(p) {
			return m
		}
		return "](../" + p + ")"
	})
}

// rewriteCrossLinks walks every file in the tree and rewrites markdown links
// whose resolved absolute target lands on a moved file. Anchors survive the
// rewrite. Files that were themselves moved have already had prefixRelative-
// Links run on them — their links are walked again, harmlessly, because the
// re-resolved paths still land on the original targets.
func rewriteCrossLinks(tree render.Tree, movedTargets map[string]string) {
	for rel, content := range tree {
		fileDir, _ := splitPath(string(rel))
		body := string(content)
		rewritten := crossLinkPattern.ReplaceAllStringFunc(body, func(m string) string {
			sub := crossLinkPattern.FindStringSubmatch(m)
			p := sub[1]
			anchor := ""
			if len(sub) > 2 {
				anchor = sub[2]
			}
			if isAbsoluteOrScheme(p) {
				return m
			}
			target := resolveAgainst(fileDir, p)
			newTarget, ok := movedTargets[target]
			if !ok {
				return m
			}
			return "](" + relPathFromDir(fileDir, newTarget) + anchor + ")"
		})
		if rewritten != body {
			tree[rel] = render.FileContent(rewritten)
		}
	}
}

// isAbsoluteOrScheme reports whether a link path is an absolute URL, root-
// relative, or a non-path scheme like mailto:. Such links are never rewritten.
func isAbsoluteOrScheme(p string) bool {
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") || strings.HasPrefix(p, "/") {
		return true
	}
	if i := strings.Index(p, ":"); i > 0 {
		pre := p[:i]
		if !strings.ContainsAny(pre, "/.#") {
			return true
		}
	}
	return false
}

// resolveAgainst returns the absolute (tree-relative) path of link `p` as
// resolved from `fileDir`. Both use forward slashes.
func resolveAgainst(fileDir, p string) string {
	if fileDir == "" {
		return path.Clean(p)
	}
	return path.Clean(path.Join(fileDir, p))
}

// relPathFromDir returns `to` expressed relative to `fromDir`. Both are
// forward-slash tree paths. If fromDir is empty (root), returns `to` as-is.
func relPathFromDir(fromDir, to string) string {
	if fromDir == "" {
		return to
	}
	fromParts := strings.Split(fromDir, "/")
	toParts := strings.Split(to, "/")
	i := 0
	for i < len(fromParts) && i < len(toParts) && fromParts[i] == toParts[i] {
		i++
	}
	var out []string
	for j := i; j < len(fromParts); j++ {
		out = append(out, "..")
	}
	out = append(out, toParts[i:]...)
	return strings.Join(out, "/")
}
