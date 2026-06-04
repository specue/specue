package engine

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/specue/specue/internal/codescan"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// Digest is a content hash (sha256). A fixed-size array so it is comparable by
// value — inputKey compares with ==.
type Digest [32]byte

// inputKey is the content identity of a build's two inputs. Same key ⇒ same graph
// (the pipeline is a pure function of these); different ⇒ rebuild.
type inputKey struct {
	spec Digest
	code Digest
}

// SourceKeyer computes a content key for one input. Injectable because the live
// tree and a git ref key differently (a stat-walk vs a tree SHA), and a content
// hash is needed where mtime is unreliable.
type SourceKeyer interface {
	Key() (Digest, error)
}

// statKeyer hashes (path, size, modtime) of the in-scope files — cheap, for a
// real filesystem with monotonic mtime (the live tree). Unsound where mtime is
// zero or stable, so not for an in-memory test FS.
type statKeyer struct {
	enum fileEnumerator
}

func (k statKeyer) Key() (Digest, error) {
	return hashFiles(k.enum, func(h hash.Hash, info fs.FileInfo) {
		var buf [16]byte
		binary.LittleEndian.PutUint64(buf[0:8], uint64(info.Size()))
		binary.LittleEndian.PutUint64(buf[8:16], uint64(info.ModTime().UnixNano()))
		h.Write(buf[:])
	})
}

// contentKeyer hashes (path, content) of the in-scope files — correct over any
// fs.FS (including an in-memory test FS), at the cost of reading each file.
type contentKeyer struct {
	enum fileEnumerator
}

func (k contentKeyer) Key() (Digest, error) {
	return hashFiles(k.enum, func(h hash.Hash, info fs.FileInfo) {})
}

// scopedFile is one file to hash, with the FS it lives in.
type scopedFile struct {
	fsys fs.FS
	path string
}

// fileEnumerator yields, in any order, the files in scope for one input.
type fileEnumerator func() ([]scopedFile, error)

// hashFiles walks the enumerated files in sorted order, mixing path + per-file
// meta (size/mtime, or nothing) + content into one digest.
func hashFiles(enum fileEnumerator, meta func(hash.Hash, fs.FileInfo)) (Digest, error) {
	files, err := enum()
	if err != nil {
		return Digest{}, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })
	h := sha256.New()
	for _, f := range files {
		h.Write([]byte(f.path))
		h.Write([]byte{0})
		info, err := fs.Stat(f.fsys, f.path)
		if err != nil {
			return Digest{}, err
		}
		meta(h, info)
		raw, err := fs.ReadFile(f.fsys, f.path)
		if err != nil {
			return Digest{}, err
		}
		h.Write(raw)
	}
	return Digest(h.Sum(nil)), nil
}

// enumerate walks one tree, collecting the files a predicate accepts. Both inputs
// reduce to this: the spec input accepts every file under a module dir, the code
// input accepts the files a scanner kind matches.
func enumerate(fsys fs.FS, root string, accept func(model.FilePath) bool) ([]scopedFile, error) {
	var out []scopedFile
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && accept(model.FilePath(p)) {
			out = append(out, scopedFile{fsys: fsys, path: p})
		}
		return nil
	})
	return out, err
}

// specEnumerator lists the work file (when one exists on disk) plus every .cue file
// under each module dir the landscape lists. The work file is included so
// adding/removing a module re-keys the build; an in-memory workspace has no file to
// hash, so only its module dirs are walked. Dirs are real OS paths (CUE requires
// them), walked via os.DirFS.
func specEnumerator(cfg Config) fileEnumerator {
	return func() ([]scopedFile, error) {
		dirs, workFile, err := workModuleDirs(cfg)
		if err != nil {
			return nil, err
		}
		var out []scopedFile
		if workFile != "" {
			out = append(out, scopedFile{fsys: os.DirFS(filepath.Dir(workFile)), path: filepath.Base(workFile)})
		}
		for _, dir := range dirs {
			files, err := enumerate(os.DirFS(dir), ".", isCUE)
			if err != nil {
				return nil, err
			}
			out = append(out, files...)
		}
		return out, nil
	}
}

func isCUE(p model.FilePath) bool { return strings.HasSuffix(string(p), ".cue") }

// workModuleDirs resolves each module's absolute dir for the keyer, mirroring the
// engine's own resolution. It returns the dirs and the work file path (empty when
// the landscape was supplied in memory — nothing on disk to hash). An in-memory
// Workspace wins over WorkFile, exactly as the engine's workspace() does.
func workModuleDirs(cfg Config) (dirs []string, workFile string, err error) {
	work, base, workFile, err := keyWorkspace(cfg)
	if err != nil {
		return nil, "", err
	}
	root := work.Root
	if root == "" {
		root = base
	} else if !filepath.IsAbs(root) {
		root = filepath.Join(base, root)
	}
	for _, m := range work.Modules {
		dirs = append(dirs, resolveDir(root, m.Dir))
	}
	return dirs, workFile, nil
}

// keyWorkspace returns the workspace, the base dir relative roots resolve against,
// and the work file path (empty for an in-memory workspace). It mirrors the
// engine's workspace() so the key tracks exactly the inputs the build reads.
func keyWorkspace(cfg Config) (work source.Workspace, base, workFile string, err error) {
	if cfg.Workspace != nil {
		return *cfg.Workspace, "", "", nil
	}
	raw, err := os.ReadFile(cfg.WorkFile)
	if err != nil {
		return source.Workspace{}, "", "", err
	}
	parser, err := source.NewCUEParser()
	if err != nil {
		return source.Workspace{}, "", "", err
	}
	work, err = parser.ParseWork(cfg.WorkFile, raw)
	if err != nil {
		return source.Workspace{}, "", "", err
	}
	return work, filepath.Dir(cfg.WorkFile), cfg.WorkFile, nil
}

// codeEnumerator lists the files each scan target's scanner would read, so the key
// tracks exactly the scanned set (it changes when a rescan is due, not on files the
// scanner ignores). It mirrors the scanner's selection: an explicit Files list
// (git ls-files) when present, else the scannable files under Root.
func codeEnumerator(cfg Config) fileEnumerator {
	return func() ([]scopedFile, error) {
		var out []scopedFile
		for _, t := range cfg.ScanTargets {
			if len(t.Files) > 0 {
				for _, rel := range t.Files {
					if codescan.IsScannable(model.FilePath(rel)) {
						out = append(out, scopedFile{fsys: t.FS, path: joinFSPath(t.Root, rel)})
					}
				}
				continue
			}
			files, err := enumerate(t.FS, t.Root, codescan.IsScannable)
			if err != nil {
				return nil, err
			}
			out = append(out, files...)
		}
		return out, nil
	}
}

// joinFSPath joins an fs.FS path, treating "." / "" root as the rel as-is.
func joinFSPath(root, rel string) string {
	if root == "" || root == "." {
		return rel
	}
	return root + "/" + rel
}

// newContentKeyers builds content-hash keyers for both inputs (correct over any
// FS). The default for tests and any FS without reliable mtime.
func newContentKeyers(cfg Config) (spec, code SourceKeyer) {
	return contentKeyer{enum: specEnumerator(cfg)}, contentKeyer{enum: codeEnumerator(cfg)}
}

// newStatKeyers builds stat-based keyers for both inputs (cheap, for a live
// filesystem with monotonic mtime).
func newStatKeyers(cfg Config) (spec, code SourceKeyer) {
	return statKeyer{enum: specEnumerator(cfg)}, statKeyer{enum: codeEnumerator(cfg)}
}
