package source

import (
	"fmt"

	"cuelang.org/go/cue"

	"github.com/specue/specue/internal/model"
)

// WorkFile is the workspace manifest's filename — the landscape entry point, CUE
// content like the module manifest.
const WorkFile = "spec.work.cue"

// Workspace is a parsed spec.work: the landscape's module set and where each lives.
// It is the entry point the engine resolves from — the whole graph, not a single
// module's require closure. A module's git repository is derived from its dir.
type Workspace struct {
	// Root is the directory module dirs are relative to; empty = the work file's
	// own directory (filled by the caller that knows the file's location).
	Root string
	// PlanBase is the branch a plan forks from and diffs against; empty = the
	// repo's current branch.
	PlanBase string
	Modules  []WorkModule
}

// WorkModule is one landscape module: its canonical path and its directory
// (relative to the workspace root).
type WorkModule struct {
	Path model.ModulePath
	Dir  string
}

// ParseWork validates a spec.work against #Work and decodes it.
func (p *cueParser) ParseWork(filename string, src []byte) (Workspace, error) {
	data := p.ctx.CompileBytes(src, cue.Filename(filename))
	if data.Err() != nil {
		return Workspace{}, fmt.Errorf("%s: parse: %w", filename, data.Err())
	}
	unified := p.work.Unify(data)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return Workspace{}, fmt.Errorf("%s: schema: %w", filename, err)
	}
	var d workDTO
	if err := unified.Decode(&d); err != nil {
		return Workspace{}, fmt.Errorf("%s: decode: %w", filename, err)
	}
	return toWorkspace(d), nil
}

type workDTO struct {
	Root     string `json:"root"`
	PlanBase string `json:"plan_base"`
	Modules  []struct {
		Path string `json:"path"`
		Dir  string `json:"dir"`
	} `json:"modules"`
}

func toWorkspace(d workDTO) Workspace {
	w := Workspace{Root: d.Root, PlanBase: d.PlanBase}
	for _, m := range d.Modules {
		w.Modules = append(w.Modules, WorkModule{Path: model.ModulePath(m.Path), Dir: m.Dir})
	}
	return w
}
