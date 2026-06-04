package engine

import (
	"io/fs"
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"

	"github.com/specue/specue/internal/codescan"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// FileLister yields the files to scan under a module directory, relative to it.
// Production passes git ls-files (tracked files only — MANIFESTO P20); a test can
// pass any lister. An empty result means "scan the whole tree" is not implied here:
// the builder always lists explicitly, so a module with no listed files is scanned
// as empty.
type FileLister func(dir string) ([]string, error)

// FSFor opens the filesystem rooted at a module directory. Production passes
// os.DirFS; a test passes an in-memory FS. Kept separate from FileLister so the
// listing source (git) and the read source (disk) are injected independently.
type FSFor func(dir string) fs.FS

// ScanTargetsFor builds one codescan.ScanTarget per code-kind module in the
// landscape, so any caller (the CLI, the server, a future bindings query) derives
// the same scan set from a workspace without duplicating the walk. It is pure but
// for the injected lister/fsFor: which modules are code is read from their
// manifests on disk, then each code module's files come from the lister and its FS
// from fsFor.
//
// dirs maps each module path to its absolute directory (what the resolver already
// computes). A module whose manifest can't be read is skipped — it cannot be a code
// module — and a non-code module contributes no target. A lister error aborts, so a
// caller that wants best-effort can swallow it.
func ScanTargetsFor(work source.Workspace, dirs map[model.ModulePath]string, lister FileLister, fsFor FSFor) ([]codescan.ScanTarget, error) {
	var targets []codescan.ScanTarget
	for _, wm := range work.Modules {
		dir := dirs[wm.Path]
		mf, err := source.ReadManifest(filepath.Join(dir, source.ManifestFile))
		if err != nil || mf.Kind != source.KindCode {
			continue
		}
		// code_root lets a code module live in a subfolder (spec.d/code/) while its
		// scan starts higher up (e.g. "../.." for the repo root), so the manifest
		// does not claim sibling spec modules as its own subpackages. Default "."
		// is backward-compat: scan the manifest's own directory.
		scanDir := dir
		if mf.CodeRoot != "" {
			scanDir = filepath.Join(dir, mf.CodeRoot)
		}
		files, err := lister(scanDir)
		if err != nil {
			return nil, err
		}
		// Drop files the module's ignore globs exclude (testdata/fixtures/generated
		// code that carry foreign annotations) — gitignore semantics, over the
		// git-tracked set the lister returned.
		files = applyIgnore(files, mf.Ignore)
		targets = append(targets, codescan.ScanTarget{
			FS:     fsFor(scanDir),
			Root:   ".",
			Files:  files,
			Module: wm.Path,
		})
	}
	return targets, nil
}

// applyIgnore drops files matching any gitignore-style glob in patterns; an empty
// patterns list keeps everything. Compiled per module (cheap — patterns are few).
func applyIgnore(files, patterns []string) []string {
	if len(patterns) == 0 {
		return files
	}
	ig := ignore.CompileIgnoreLines(patterns...)
	kept := files[:0:0]
	for _, f := range files {
		if !ig.MatchesPath(f) {
			kept = append(kept, f)
		}
	}
	return kept
}
