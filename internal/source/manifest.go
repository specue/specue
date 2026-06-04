package source

import (
	"fmt"

	"cuelang.org/go/cue"

	"github.com/specue/specue/internal/model"
)

// ModuleKind is a module's role; it gates which node types the module may hold
// (role-gate, enforced by the compiler).
type ModuleKind string

const (
	KindService    ModuleKind = "service"
	KindDomain     ModuleKind = "domain"
	KindGovernance ModuleKind = "governance"
	KindTopology   ModuleKind = "topology"
	KindCode       ModuleKind = "code"
)

// Version is a module's semver tag (vMAJOR.MINOR.PATCH); shape enforced by the
// #semver schema, so it is carried as a typed string here.
type Version string

// Manifest is a parsed spec.mod: a module's identity and its dependencies. Only
// what the source layer needs to place nodes (Path, Kind) plus the raw requires,
// which the module-manager layer resolves later.
type Manifest struct {
	Path     model.ModulePath
	Version  Version
	Kind     ModuleKind
	Requires []ModuleRequire
	// Ignore holds gitignore-style globs the code scanner skips (kind:code only).
	Ignore []string
	// CodeRoot is where the code scan begins, relative to spec.mod.cue's own
	// directory. Empty (the default) means "." — scan the manifest's directory.
	// kind:code modules in a spec.d/code/ subfolder set "../.." so the scan
	// reaches the repo's source tree from a manifest that does not claim
	// nested sibling modules. Honoured only for kind:code.
	CodeRoot string
}

// ModuleRequire is one dependency declaration; the module-manager locates it and
// adds it to the closure. Import naming/scoping is CUE-native (cue.mod + the
// import statement), so a require carries only identity, version, and the local
// replace path.
type ModuleRequire struct {
	Module  model.ModulePath
	Version Version
	Replace string // a filesystem path, not an identifier
}

// ParseManifest validates a spec.mod against #Module and decodes it.
func (p *cueParser) ParseManifest(filename string, src []byte) (Manifest, error) {
	data := p.ctx.CompileBytes(src, cue.Filename(filename))
	if data.Err() != nil {
		return Manifest{}, fmt.Errorf("%s: parse: %w", filename, data.Err())
	}
	unified := p.module.Unify(data)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return Manifest{}, fmt.Errorf("%s: schema: %w", filename, err)
	}
	var d manifestDTO
	if err := unified.Decode(&d); err != nil {
		return Manifest{}, fmt.Errorf("%s: decode: %w", filename, err)
	}
	return toManifest(d), nil
}

type manifestDTO struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Require []struct {
		Module  string `json:"module"`
		Version string `json:"version"`
		Replace string `json:"replace"`
	} `json:"require"`
	Ignore   []string `json:"ignore"`
	CodeRoot string   `json:"code_root"`
}

func toManifest(d manifestDTO) Manifest {
	m := Manifest{
		Path:     model.ModulePath(d.Module),
		Version:  Version(d.Version),
		Kind:     ModuleKind(d.Kind),
		Ignore:   d.Ignore,
		CodeRoot: d.CodeRoot,
	}
	for _, r := range d.Require {
		m.Requires = append(m.Requires, ModuleRequire{
			Module:  model.ModulePath(r.Module),
			Version: Version(r.Version),
			Replace: r.Replace,
		})
	}
	return m
}
