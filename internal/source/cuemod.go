package source

import (
	"fmt"

	"cuelang.org/go/cue"

	"github.com/specue/specue/internal/model"
)

// CUEModFile is the CUE module file's path within a module (its own version +
// dependency versions, which CUE uses to resolve imports). It is distinct from
// spec.mod.cue (our manifest) — both carry versions, and a consistency check gates
// them being out of sync.
const CUEModFile = "cue.mod/module.cue"

// CUEModule is the parsed cue.mod/module.cue: the module's own path (with @vN
// suffix) and the version it pins each dependency at. This is CUE's view of
// versions, compared against spec.mod's in the consistency gate.
type CUEModule struct {
	Module string                       // e.g. "x.test/example@v0"
	Deps   map[model.ModulePath]Version // dep base path → pinned version (v field)
}

// ParseCUEMod decodes a cue.mod/module.cue. It is CUE's own format, not our schema,
// so it is compiled and decoded directly (no #Module unification).
func (p *cueParser) ParseCUEMod(filename string, src []byte) (CUEModule, error) {
	v := p.ctx.CompileBytes(src, cue.Filename(filename))
	if v.Err() != nil {
		return CUEModule{}, fmt.Errorf("%s: parse: %w", filename, v.Err())
	}
	var d cueModDTO
	if err := v.Decode(&d); err != nil {
		return CUEModule{}, fmt.Errorf("%s: decode: %w", filename, err)
	}
	out := CUEModule{Module: d.Module, Deps: map[model.ModulePath]Version{}}
	for path, dep := range d.Deps {
		out.Deps[model.ModulePath(path)] = Version(dep.V)
	}
	return out, nil
}

type cueModDTO struct {
	Module string `json:"module"`
	Deps   map[string]struct {
		V string `json:"v"`
	} `json:"deps"`
}
