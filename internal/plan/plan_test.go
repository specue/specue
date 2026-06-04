package plan_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/source"
)

// gitBin resolves the git binary the test drives. It is looked up once; a missing
// git skips the test (planning needs a real git, injected — see the git wrapper).
func gitBin(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available")
	}
	return bin
}

// repo lays out a single-repo landscape: a governance module and one service
// module, committed on the base branch. Returns the repo root and a Manager.
func repo(t *testing.T) (root string, mgr *plan.Manager, govModule model.ModulePath) {
	t.Helper()
	bin := gitBin(t)
	root = t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, out)
	}
	write := func(rel, content string) {
		full := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}

	run("init", "-b", "base")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")
	write("governance/spec.mod.cue", "module: \"x.test/governance@v0\"\nversion: \"v0.0.1\"\nkind: \"governance\"\n")
	write("example/spec.mod.cue", "module: \"x.test/example@v0\"\nversion: \"v0.0.1\"\nkind: \"service\"\n")
	run("add", "-A")
	run("commit", "-m", "base")

	govModule = "x.test/governance@v0"
	work := source.Workspace{PlanBase: "base", Modules: []source.WorkModule{
		{Path: govModule, Dir: filepath.Join(root, "governance")},
		{Path: "x.test/example@v0", Dir: filepath.Join(root, "example")},
	}}
	dirs := map[model.ModulePath]string{
		govModule:         filepath.Join(root, "governance"),
		"x.test/example@v0": filepath.Join(root, "example"),
	}
	mgr, err := plan.NewManager(work, dirs, plan.NewGit(bin), govModule)
	require.NoError(t, err)
	return root, mgr, govModule
}

//specue:test:register-plan#plan-is-a-branch-set
func TestRegisterAnchorsPlanOnBranch(t *testing.T) {
	root, mgr, _ := repo(t)
	bin := gitBin(t)

	require.NoError(t, mgr.Register("gp-1076", "Drop virtual id"))

	// The plan branch exists, base is still checked out, and the record lives ON
	// the branch — not on base.
	assert.Equal(t, "base", currentBranch(t, bin, root))
	assert.True(t, branchExists(t, bin, root, "plan/gp-1076"))
	assert.NoFileExists(t, filepath.Join(root, "governance/plan-gp-1076.cue"), "record not on base")

	recordOnBranch := show(t, bin, root, "plan/gp-1076", "governance/plan-gp-1076.cue")
	assert.Contains(t, recordOnBranch, `slug:       "plan-gp-1076"`)
	assert.Contains(t, recordOnBranch, `branch:     "plan/gp-1076"`)
	assert.Contains(t, recordOnBranch, `status:     "proposed"`)
}

//specue:test:use-plan#checks-out-every-branch
//specue:test:return-to-base#every-module-returns
func TestUseAndBaseSwitchBranches(t *testing.T) {
	root, mgr, _ := repo(t)
	bin := gitBin(t)
	require.NoError(t, mgr.Register("gp-1076", ""))

	require.NoError(t, mgr.Use("gp-1076"))
	assert.Equal(t, "plan/gp-1076", currentBranch(t, bin, root), "use switches to the plan branch")
	assert.FileExists(t, filepath.Join(root, "governance/plan-gp-1076.cue"), "record visible in working tree on the plan branch")

	require.NoError(t, mgr.Base())
	assert.Equal(t, "base", currentBranch(t, bin, root), "base returns to plan_base")
}

//specue:test:drop-plan#branches-and-record-removed
func TestDropRemovesBranchAndRecord(t *testing.T) {
	root, mgr, _ := repo(t)
	bin := gitBin(t)
	require.NoError(t, mgr.Register("gp-1076", ""))
	require.NoError(t, mgr.Use("gp-1076"))

	// Drop from the plan branch: it returns to base, deletes the branch (force,
	// since it diverged from base), and removes the record.
	require.NoError(t, mgr.Drop("gp-1076", true))
	assert.Equal(t, "base", currentBranch(t, bin, root))
	assert.False(t, branchExists(t, bin, root, "plan/gp-1076"), "plan branch gone")
}

// --- git helpers (verify state independently of the wrapper under test) -------

func gitOut(t *testing.T, bin, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func currentBranch(t *testing.T, bin, root string) string {
	return trim(gitOut(t, bin, root, "rev-parse", "--abbrev-ref", "HEAD"))
}

func branchExists(t *testing.T, bin, root, b string) bool {
	cmd := exec.Command(bin, "rev-parse", "--verify", "--quiet", "refs/heads/"+b)
	cmd.Dir = root
	return cmd.Run() == nil
}

func show(t *testing.T, bin, root, ref, path string) string {
	return gitOut(t, bin, root, "show", ref+":"+path)
}

func trim(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r' || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	return s
}
