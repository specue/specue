package plan_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/plan"
)

// gitRepo inits an empty repo on branch "base" with one commit, returning its root
// and a Git wrapper driving the injected binary.
func gitRepo(t *testing.T) (root string, g plan.Git) {
	t.Helper()
	bin := gitBin(t)
	root = t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, out)
	}
	run("init", "-b", "base")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")
	require.NoError(t, os.WriteFile(filepath.Join(root, "f.txt"), []byte("v1\n"), 0o644))
	run("add", "-A")
	run("commit", "-m", "init")
	return root, plan.NewGit(bin)
}

func TestGitRepoRootAndCurrentBranch(t *testing.T) {
	root, g := gitRepo(t)

	got, err := g.RepoRoot(root)
	require.NoError(t, err)
	// macOS /var → /private/var symlink: compare resolved paths.
	wantResolved, _ := filepath.EvalSymlinks(root)
	gotResolved, _ := filepath.EvalSymlinks(got)
	assert.Equal(t, wantResolved, gotResolved)

	branch, err := g.CurrentBranch(root)
	require.NoError(t, err)
	assert.Equal(t, "base", branch)
}

func TestGitListBranches(t *testing.T) {
	root, g := gitRepo(t)
	require.NoError(t, g.CreateBranch(root, "plan/gp-1", "base"))
	require.NoError(t, g.CreateBranch(root, "plan/gp-2", "base"))
	require.NoError(t, g.CreateBranch(root, "feature/x", "base"))

	ids, err := g.ListBranches(root, "plan/")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"gp-1", "gp-2"}, ids, "prefix is stripped; non-matching branches excluded")
}

func TestGitListBranchesNoneMatch(t *testing.T) {
	root, g := gitRepo(t)
	ids, err := g.ListBranches(root, "plan/")
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestGitBranchLifecycle(t *testing.T) {
	root, g := gitRepo(t)

	exists, err := g.BranchExists(root, "plan/x")
	require.NoError(t, err)
	assert.False(t, exists, "absent branch reported absent, not an error")

	require.NoError(t, g.CreateBranch(root, "plan/x", "base"))
	exists, err = g.BranchExists(root, "plan/x")
	require.NoError(t, err)
	assert.True(t, exists)

	require.NoError(t, g.Checkout(root, "plan/x"))
	cur, _ := g.CurrentBranch(root)
	assert.Equal(t, "plan/x", cur)

	// Delete needs another branch checked out first.
	require.NoError(t, g.Checkout(root, "base"))
	require.NoError(t, g.DeleteBranch(root, "plan/x", false))
	exists, _ = g.BranchExists(root, "plan/x")
	assert.False(t, exists)
}

func TestGitCommitNoOpIsNotError(t *testing.T) {
	root, g := gitRepo(t)
	// Nothing changed since init → Commit is a no-op, not an error.
	require.NoError(t, g.Commit(root, "empty"))
}

func TestGitCommitStagesAndCommits(t *testing.T) {
	root, g := gitRepo(t)
	bin := gitBin(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "g.txt"), []byte("new\n"), 0o644))

	require.NoError(t, g.Commit(root, "add g"))

	// The file is now tracked at HEAD (committed, not just staged).
	cmd := exec.Command(bin, "ls-files", "g.txt")
	cmd.Dir = root
	out, _ := cmd.Output()
	assert.Equal(t, "g.txt\n", string(out))
}

func TestGitDeleteUnmergedNeedsForce(t *testing.T) {
	root, g := gitRepo(t)
	require.NoError(t, g.CreateBranch(root, "plan/x", "base"))
	require.NoError(t, g.Checkout(root, "plan/x"))
	require.NoError(t, os.WriteFile(filepath.Join(root, "h.txt"), []byte("x\n"), 0o644))
	require.NoError(t, g.Commit(root, "diverge"))
	require.NoError(t, g.Checkout(root, "base"))

	// -d refuses an unmerged branch; -D (force) deletes it.
	assert.Error(t, g.DeleteBranch(root, "plan/x", false), "unmerged branch needs force")
	assert.NoError(t, g.DeleteBranch(root, "plan/x", true))
}
