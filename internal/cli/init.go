package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"cuelang.org/go/mod/module"

	"github.com/specue/specue/internal/source"
)

// schema dep the scaffolded cue.mod pins, matching what the migrator emits so a
// fresh module resolves the shared schema the same way every other module does.
const (
	scaffoldLangVersion = "v0.16.0"
	scaffoldSchemaVer   = "v0.0.1"
)

// validKinds are the module roles `init` accepts, mirroring source.ModuleKind.
var validKinds = []string{
	string(source.KindService), string(source.KindDomain),
	string(source.KindGovernance), string(source.KindTopology),
	string(source.KindCode),
}

// InitReport is the typed result of `init`: the module created and the files
// written, so the caller sees exactly what landed.
type InitReport struct {
	Module string   `json:"module"`
	Kind   string   `json:"kind"`
	Dir    string   `json:"dir"` // absolute, canonical
	Files  []string `json:"files"`

	// inputDir is the directory exactly as the user typed it, used for the next-step
	// hint so it reads `work add spec`, not an absolute path.
	inputDir string
}

// runInit scaffolds a new spec module at dir: spec.mod.cue (with the given kind) and
// cue.mod/module.cue (pinning the shared schema). It refuses to overwrite an
// existing spec.mod.cue — initializing over a module is almost always a mistake, so
// the fix points at editing it instead. The module path is the canonical
// `path/name@vMAJOR` the manifest declares.
//
//specue:req:init-module
func runInit(dir, modulePath, kind, name string, useLayout bool) (InitReport, *Problem) {
	if !validKind(kind) {
		p := Errorf(fmt.Sprintf("pass one of: %v", validKinds), "unknown module kind %q", kind)
		return InitReport{}, &p
	}
	if name != "" && !useLayout {
		p := Errorf("add --layout spec.d, or drop --name and use a plain dir",
			"--name only applies with --layout spec.d")
		return InitReport{}, &p
	}
	// Validate the module path the same way CUE will on load, so a bad path fails
	// here with a clear fix — not later when `plan list` / resolution chokes on a
	// module that was written but can never load. CUE requires a dot in the first
	// path element and an @vMAJOR suffix (e.g. x.test/governance@v0).
	if err := module.CheckPath(modulePath); err != nil {
		p := Errorf("use a path like example.com/name@v0 — the first element needs a dot and an @vMAJOR suffix",
			"invalid module path %q: %v", modulePath, err)
		return InitReport{}, &p
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		p := Errorf("pass a valid directory with -C", "cannot resolve %q: %v", dir, err)
		return InitReport{}, &p
	}
	// With --layout spec.d the user names the repo root; init places the module at
	// repo/<LayoutDir>/<kind>/[<name>/]. The on-disk module dir extends from there;
	// the git check still runs against the deepest existing ancestor, so a fresh
	// layout subfolder does not need to exist yet.
	codeRoot := "" // empty = "." (manifest's own dir); only set when --layout shifts the dir
	if useLayout {
		sub := kind
		if name != "" {
			sub = filepath.Join(sub, name)
		}
		abs = filepath.Join(abs, source.LayoutDir, sub)
		if kind == string(source.KindCode) {
			// <LayoutDir>/code/spec.mod.cue → repo root is two levels up.
			codeRoot = "../.."
		}
	}
	// Specue is git-native (MANIFESTO P20): a module lives in git. Refuse to
	// scaffold outside a repository here, with the remedy, rather than let it fail
	// later when a plan or the scanner first runs git.
	if p := requireGitRepo(abs); p != nil {
		return InitReport{}, p
	}
	manifest := filepath.Join(abs, source.ManifestFile)
	if fileExists(manifest) {
		p := Errorf(fmt.Sprintf("edit %s, or init in a fresh directory with -C <dir>", manifest),
			"%s is already a module (%s exists)", abs, source.ManifestFile)
		return InitReport{}, &p
	}
	if err := os.MkdirAll(filepath.Join(abs, "cue.mod"), 0o755); err != nil {
		p := Errorf("check the directory is writable", "cannot create cue.mod: %v", err)
		return InitReport{}, &p
	}

	manifestSrc := fmt.Sprintf("module: %q\nversion: \"v0.1.0\"\nkind:    %q\n", modulePath, kind)
	if codeRoot != "" {
		manifestSrc += fmt.Sprintf("code_root: %q\n", codeRoot)
	}
	cueModSrc := fmt.Sprintf("module: %q\nlanguage: version: %q\ndeps: %q: v: %q\n",
		modulePath, scaffoldLangVersion, source.SchemaModulePath, scaffoldSchemaVer)

	if err := os.WriteFile(manifest, []byte(manifestSrc), 0o644); err != nil {
		p := Errorf("check the directory is writable", "cannot write %s: %v", source.ManifestFile, err)
		return InitReport{}, &p
	}
	cueMod := filepath.Join(abs, source.CUEModFile)
	if err := os.WriteFile(cueMod, []byte(cueModSrc), 0o644); err != nil {
		p := Errorf("check the directory is writable", "cannot write %s: %v", source.CUEModFile, err)
		return InitReport{}, &p
	}

	return InitReport{
		Module: modulePath, Kind: kind, Dir: abs, inputDir: dir,
		Files: []string{source.ManifestFile, source.CUEModFile},
	}, nil
}

func validKind(k string) bool {
	return slices.Contains(validKinds, k)
}

func (r InitReport) renderHuman(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "initialized %s module %s in %s\n", r.Kind, r.Module, r.Dir); err != nil {
		return err
	}
	for _, f := range r.Files {
		if _, err := fmt.Fprintf(w, "  + %s\n", f); err != nil {
			return err
		}
	}
	hintDir := r.inputDir
	if hintDir == "" {
		hintDir = r.Dir
	}
	// A possibility, not an imperative — init may be all the user wanted.
	_, err := fmt.Fprintf(w, "\nto include it in a workspace: `%s %s`\n", cmdPath(cmdContext, subModule, subAdd), hintDir)
	return err
}

func (r InitReport) jsonValue() any { return r }
