package plan_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/diff"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/source"
)

const walletCueMod = `module: "x.test/example@v0"
language: version: "v0.16.0"
deps: "specue.io/schema@v0": v: "v0.0.1"
`
const walletSpecMod = "module: \"x.test/example@v0\"\nversion: \"v0.0.1\"\nkind: \"service\"\n"

// baseNodes is the example module on base: a service container + one use case.
const baseNodes = `package example
import s "specue.io/schema@v0:spec"
example: s.#Container & {type:"Container", slug:"example", title:"Wallet", confidence:"CONFIRMED", kind:"service"}
apply: s.#Contract & {type:"Contract", slug:"apply", title:"Apply", confidence:"CONFIRMED", service:example, invariants:[{id:"post", text:"done"}]}
`

// planNodes adds a second use case on the plan branch.
const planNodes = `package example
import s "specue.io/schema@v0:spec"
example: s.#Container & {type:"Container", slug:"example", title:"Wallet", confidence:"CONFIRMED", kind:"service"}
apply: s.#Contract & {type:"Contract", slug:"apply", title:"Apply", confidence:"CONFIRMED", service:example, invariants:[{id:"post", text:"done"}]}
reverse: s.#Contract & {type:"Contract", slug:"reverse", title:"Reverse", confidence:"CONFIRMED", service:example, invariants:[{id:"post", text:"compensated"}]}
`

// projectRepo lays out a single-repo workspace (governance + example), commits base,
// and returns the workspace, dirs, manager, and git binary.
func projectRepo(t *testing.T) (source.Workspace, map[model.ModulePath]string, *plan.Manager, string) {
	t.Helper()
	bin := gitBin(t)
	root := t.TempDir()
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
	write("governance/cue.mod/module.cue", "module: \"x.test/governance@v0\"\nlanguage: version: \"v0.16.0\"\ndeps: \"specue.io/schema@v0\": v: \"v0.0.1\"\n")
	write("example/spec.mod.cue", walletSpecMod)
	write("example/cue.mod/module.cue", walletCueMod)
	write("example/nodes.cue", baseNodes)
	run("add", "-A")
	run("commit", "-m", "base")

	gov := model.ModulePath("x.test/governance@v0")
	work := source.Workspace{PlanBase: "base", Modules: []source.WorkModule{
		{Path: gov, Dir: filepath.Join(root, "governance")},
		{Path: "x.test/example@v0", Dir: filepath.Join(root, "example")},
	}}
	dirs := map[model.ModulePath]string{
		gov:               filepath.Join(root, "governance"),
		"x.test/example@v0": filepath.Join(root, "example"),
	}
	mgr, err := plan.NewManager(work, dirs, plan.NewGit(bin), gov)
	require.NoError(t, err)
	return work, dirs, mgr, bin
}

// TestPlanDiffProjectsAddedNode registers a plan, edits the example module on its
// branch (adds a use case), returns to base, and asserts plan diff projects the
// added node from base — without the working tree leaving base.
//specue:test:pending-overlay#viewed-without-checkout
func TestPlanDiffProjectsAddedNode(t *testing.T) {
	_, dirs, mgr, bin := projectRepo(t)
	root := filepath.Dir(dirs["x.test/example@v0"])

	require.NoError(t, mgr.Register("gp-1", "add reverse"))
	require.NoError(t, mgr.Use("gp-1"))

	// On the plan branch, add the reverse use case and commit.
	require.NoError(t, os.WriteFile(filepath.Join(root, "example/nodes.cue"), []byte(planNodes), 0o644))
	commit(t, bin, root, "plan: add reverse")
	require.NoError(t, mgr.Base())

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	proj, err := plan.NewProjector(parser, plan.NewGit(bin))
	require.NoError(t, err)
	t.Cleanup(func() { _ = proj.Close() })

	delta, err := mgr.Diff(proj, dirs, "gp-1")
	require.NoError(t, err)

	// The projection sees the added reverse node. The governance #Plan record the
	// plan adds on its branch is bookkeeping, not a spec change, so it is filtered
	// out of the overlay. Working tree stayed on base.
	added := map[model.Slug]bool{}
	for _, nd := range delta.Nodes {
		if nd.Change == diff.Added {
			added[nd.ID.Slug] = true
		}
		assert.NotEqual(t, model.TypePlan, nd.Type, "the plan's own #Plan record is filtered from its overlay")
	}
	assert.True(t, added["reverse"], "plan diff projects the added use case from base")
	assert.Equal(t, "base", currentBranch(t, bin, root), "working tree stayed on base")

	wt, _ := os.ReadFile(filepath.Join(root, "example/nodes.cue"))
	assert.Equal(t, baseNodes, string(wt), "working tree example unchanged")
}

// TestPlanDiffFromPlanBranchReadsBaseFromGit edits on the plan branch and asks for
// the diff WITHOUT returning to base. The worktree carries plan content; if the
// base side were read from the worktree the diff would be empty. The invariant
// pending-overlay#base-side-read-through-git requires base to come from the base
// branch via git-fs — so the delta surfaces the same `reverse` add as when run
// from base.
//
//specue:test:pending-overlay#base-side-read-through-git
func TestPlanDiffFromPlanBranchReadsBaseFromGit(t *testing.T) {
	_, dirs, mgr, bin := projectRepo(t)
	root := filepath.Dir(dirs["x.test/example@v0"])

	require.NoError(t, mgr.Register("gp-2", "add reverse"))
	require.NoError(t, mgr.Use("gp-2"))

	// Stay on the plan branch: edit, commit, and ask for the diff right here.
	require.NoError(t, os.WriteFile(filepath.Join(root, "example/nodes.cue"), []byte(planNodes), 0o644))
	commit(t, bin, root, "plan: add reverse")
	assert.Equal(t, "plan/gp-2", currentBranch(t, bin, root), "test stays on the plan branch")

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	proj, err := plan.NewProjector(parser, plan.NewGit(bin))
	require.NoError(t, err)
	t.Cleanup(func() { _ = proj.Close() })

	delta, err := mgr.Diff(proj, dirs, "gp-2")
	require.NoError(t, err)

	added := map[model.Slug]bool{}
	for _, nd := range delta.Nodes {
		if nd.Change == diff.Added {
			added[nd.ID.Slug] = true
		}
	}
	assert.True(t, added["reverse"], "base side read from git → reverse surfaces as added even when worktree is the plan")
}

func commit(t *testing.T, bin, root, msg string) {
	t.Helper()
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-m", msg}} {
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, out)
	}
}
