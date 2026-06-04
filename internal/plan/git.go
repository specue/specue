// Package plan is the planning layer: a plan is a named change-set whose content
// lives on plan/<id> branches across the affected repositories, anchored by a Plan
// record in the governance module. This file is the git wrapper the layer drives —
// every git invocation goes through one place so the binary is injectable (tests
// point at their own git, no assumption it is installed on the host) and the rest
// of the layer depends on the capability, not on exec.
package plan

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Git runs git commands in a working tree. The binary path is a field, not a
// hardcoded "git": production resolves it from PATH, a test injects its own. Only
// the operations the planning layer needs are exposed.
type Git interface {
	// RepoRoot returns the git toplevel containing dir (a module's repository).
	RepoRoot(dir string) (string, error)
	// CurrentBranch returns the checked-out branch of the repo at root.
	CurrentBranch(root string) (string, error)
	// BranchExists reports whether branch exists in the repo at root.
	BranchExists(root, branch string) (bool, error)
	// CreateBranch creates branch at the tip of from (without switching to it).
	CreateBranch(root, branch, from string) error
	// Checkout switches the repo at root to branch.
	Checkout(root, branch string) error
	// Commit stages everything and commits with msg; a no-op (nothing to commit)
	// is not an error. Use only where committing the whole tree is intended (the
	// post-merge flip in accept, on base); for a single authored file use
	// CommitPaths so unrelated working-tree changes are not swept in.
	Commit(root, msg string) error
	// CommitPaths stages exactly paths (relative to root) and commits them with msg;
	// nothing staged is a no-op. Unlike Commit it does not `add -A`, so it never
	// captures unrelated changes — register uses it for the plan record alone.
	CommitPaths(root, msg string, paths ...string) error
	// IsClean reports whether the repo at root has no uncommitted changes (tracked or
	// untracked). The planning verbs check this before any checkout: a checkout in a
	// dirty tree silently overwrites untracked files and moves tracked edits onto the
	// wrong branch.
	IsClean(root string) (bool, error)
	// DeleteBranch deletes branch; force allows deleting an unmerged branch.
	DeleteBranch(root, branch string, force bool) error
	// ArchiveSubtree extracts the subtree at subdir of ref (a branch/commit) in the
	// repo at root into dest — a read of a revision without touching the working
	// tree (the pattern Go uses to read a module from VCS). subdir is relative to
	// the repo root; dest is created. A leading/trailing slash on subdir is fine.
	ArchiveSubtree(root, ref, subdir, dest string) error
	// Head returns the current commit hash of the repo at root — a marker to reset
	// back to if a later step fails.
	Head(root string) (string, error)
	// Tag creates a lightweight annotated tag at commit (or at HEAD if commit is
	// empty) with the given name and message. Used by accept to mark each
	// landed plan so a release index can find them by `git tag --list plan/*`.
	Tag(root, name, commit, msg string) error
	// Merge merges branch into the current branch with --no-ff. A merge conflict is
	// reported as conflicted=true (and the merge is aborted) rather than an error;
	// a genuine git failure is the error.
	Merge(root, branch, msg string) (conflicted bool, err error)
	// ResetHard resets the current branch hard to commit (undo a merge).
	ResetHard(root, commit string) error
	// SubtreeChanged reports whether subdir differs between baseRef and ref — i.e.
	// whether this branch actually touched that module. Used to overlay a module
	// only from the plan that changed it.
	SubtreeChanged(root, baseRef, ref, subdir string) (bool, error)
	// ListBranches returns the local branches in root whose name starts with prefix,
	// the prefix stripped (so ListBranches(root, "plan/") yields plan ids). Empty
	// when none match.
	ListBranches(root, prefix string) ([]string, error)
	// ListFiles returns the git-tracked files under dir, as paths relative to dir.
	// This is how the scanner enumerates a code module's source: git filters out
	// node_modules / vendor / built artifacts via .gitignore, so the scanner sees
	// exactly what the repository tracks (MANIFESTO P20). Untracked files are not
	// returned until added.
	ListFiles(dir string) ([]string, error)
}

// gitCLI drives the git binary at Bin.
type gitCLI struct {
	Bin string
}

// NewGit returns a Git driving the binary at path (e.g. resolved from PATH for
// production, or a test's own git). An empty path defaults to "git".
func NewGit(path string) Git {
	if path == "" {
		path = "git"
	}
	return &gitCLI{Bin: path}
}

func (g *gitCLI) run(dir string, args ...string) (string, error) {
	cmd := exec.Command(g.Bin, args...)
	cmd.Dir = dir
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errBuf.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func (g *gitCLI) RepoRoot(dir string) (string, error) {
	return g.run(dir, "rev-parse", "--show-toplevel")
}

func (g *gitCLI) CurrentBranch(root string) (string, error) {
	// branch --show-current reports the checked-out branch name even on an unborn
	// branch (a fresh `git init` with no commits yet) — unlike `rev-parse HEAD`,
	// which errors there. Empty in detached-HEAD, which callers treat as "no branch".
	return g.run(root, "branch", "--show-current")
}

func (g *gitCLI) BranchExists(root, branch string) (bool, error) {
	// rev-parse --verify exits non-zero when the ref is absent; a clean exit means
	// it exists. A genuine git failure (bad repo) also exits non-zero, but the
	// caller treats "absent" benignly and a later op surfaces a real fault.
	_, err := g.run(root, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil, nil
}

func (g *gitCLI) CreateBranch(root, branch, from string) error {
	_, err := g.run(root, "branch", branch, from)
	return err
}

func (g *gitCLI) Checkout(root, branch string) error {
	_, err := g.run(root, "checkout", branch)
	return err
}

func (g *gitCLI) Commit(root, msg string) error {
	if _, err := g.run(root, "add", "-A"); err != nil {
		return err
	}
	// `git commit` exits non-zero when there is nothing staged; treat that as a
	// no-op so register/use are idempotent.
	clean, err := g.nothingToCommit(root)
	if err != nil {
		return err
	}
	if clean {
		return nil
	}
	_, err = g.run(root, "commit", "-m", msg)
	return err
}

func (g *gitCLI) nothingToCommit(root string) (bool, error) {
	out, err := g.run(root, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out == "", nil
}

func (g *gitCLI) CommitPaths(root, msg string, paths ...string) error {
	args := append([]string{"add", "--"}, paths...)
	if _, err := g.run(root, args...); err != nil {
		return err
	}
	// Commit only the staged paths; --only ignores anything else in the index/tree.
	staged, err := g.run(root, "diff", "--cached", "--name-only")
	if err != nil {
		return err
	}
	if staged == "" {
		return nil // nothing changed in those paths — idempotent no-op
	}
	cargs := append([]string{"commit", "-m", msg, "--only", "--"}, paths...)
	_, err = g.run(root, cargs...)
	return err
}

func (g *gitCLI) IsClean(root string) (bool, error) {
	return g.nothingToCommit(root)
}

func (g *gitCLI) Head(root string) (string, error) {
	return g.run(root, "rev-parse", "HEAD")
}

func (g *gitCLI) Tag(root, name, commit, msg string) error {
	args := []string{"tag", "-a", name, "-m", msg}
	if commit != "" {
		args = append(args, commit)
	}
	_, err := g.run(root, args...)
	return err
}

func (g *gitCLI) ListBranches(root, prefix string) ([]string, error) {
	// for-each-ref over the local branches; %(refname:short) gives the bare name.
	out, err := g.run(root, "for-each-ref", "--format=%(refname:short)", "refs/heads/"+prefix)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var ids []string
	for _, line := range strings.Split(out, "\n") {
		ids = append(ids, strings.TrimPrefix(line, prefix))
	}
	return ids, nil
}

func (g *gitCLI) ListFiles(dir string) ([]string, error) {
	// Run in dir so the returned paths are relative to it; --cached lists tracked
	// files. -z would be safer for exotic names, but the scanner's path model is
	// line-based, so newline-split matches the rest of the layer.
	out, err := g.run(dir, "ls-files")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

func (g *gitCLI) Merge(root, branch, msg string) (bool, error) {
	_, err := g.run(root, "merge", "--no-ff", "-m", msg, branch)
	if err == nil {
		return false, nil
	}
	// A merge conflict leaves the merge in progress; abort it and report the
	// conflict (not a hard error — the caller surfaces it to the user).
	if g.inMergeConflict(root) {
		_, _ = g.run(root, "merge", "--abort")
		return true, nil
	}
	return false, err
}

// inMergeConflict reports whether the repo is mid-merge with conflicts (an
// unmerged path in status).
func (g *gitCLI) inMergeConflict(root string) bool {
	out, err := g.run(root, "status", "--porcelain")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		// porcelain conflict markers: UU, AA, DD, AU, UA, DU, UD.
		if len(line) >= 2 && (line[0] == 'U' || line[1] == 'U' || line[:2] == "AA" || line[:2] == "DD") {
			return true
		}
	}
	return false
}

func (g *gitCLI) ResetHard(root, commit string) error {
	_, err := g.run(root, "reset", "--hard", commit)
	return err
}

func (g *gitCLI) SubtreeChanged(root, baseRef, ref, subdir string) (bool, error) {
	subdir = strings.Trim(subdir, "/")
	args := []string{"diff", "--quiet", baseRef, ref}
	if subdir != "" {
		args = append(args, "--", subdir)
	}
	// git diff --quiet exits 1 when there is a difference, 0 when none.
	cmd := exec.Command(g.Bin, args...)
	cmd.Dir = root
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		return true, nil
	}
	return false, fmt.Errorf("git diff %s..%s -- %s: %w", baseRef, ref, subdir, err)
}

func (g *gitCLI) DeleteBranch(root, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := g.run(root, "branch", flag, branch)
	return err
}

func (g *gitCLI) ArchiveSubtree(root, ref, subdir, dest string) error {
	subdir = strings.Trim(subdir, "/")
	args := []string{"archive", "--format=tar", ref}
	if subdir != "" {
		args = append(args, subdir)
	}
	cmd := exec.Command(g.Bin, args...)
	cmd.Dir = root
	var tarBuf, errBuf bytes.Buffer
	cmd.Stdout = &tarBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git archive %s %s: %w: %s", ref, subdir, err, strings.TrimSpace(errBuf.String()))
	}
	// git archive of a subdir emits paths relative to the repo root (subdir/...),
	// so strip the subdir prefix to land files at dest root.
	return extractTar(&tarBuf, dest, subdir)
}

// extractTar unpacks a tar stream into dest, stripping prefix (and its slash) from
// each entry path. It is deliberately minimal — git archive emits regular files
// and directories only.
func extractTar(r *bytes.Buffer, dest, prefix string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(hdr.Name, prefix)
		name = strings.TrimPrefix(name, "/")
		if name == "" {
			continue
		}
		target := filepath.Join(dest, filepath.FromSlash(name))
		if hdr.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
}
