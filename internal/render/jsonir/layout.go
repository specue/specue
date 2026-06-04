// Package jsonir is the JSON Intermediate Representation renderer: one .json
// file per node under <module-dir>/<slug>.json plus an index.json at the root
// listing every module and node. The file layout mirrors the markdown
// renderer's so a caller switching --format finds the same shape on disk; only
// the file extension and the body format change.
package jsonir

import (
	"strings"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/render"
)

// IndexPath is where the index file lands inside the tree.
const IndexPath render.RelPath = "index.json"

// NodePath returns the canonical path of a node's JSON file inside the tree.
// Same flattening as the markdown layout — only the extension differs — so a
// caller that already knows the markdown layout finds the JSON in the same
// shape.
func NodePath(id model.NodeID) render.RelPath {
	return render.RelPath(moduleDir(id.Module) + "/" + string(id.Slug) + ".json")
}

// moduleDir turns a module path into a single safe directory name.
func moduleDir(m model.ModulePath) string {
	s := string(m)
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "@", "-")
	return s
}
