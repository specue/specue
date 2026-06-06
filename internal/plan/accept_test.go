package plan_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/source"
)

// TestAcceptMergesAndFlips runs the full accept: register → use → add a valid node
// on the branch → return to base → accept. The plan's content lands on base and
// the Plan record flips proposed→accepted.
//specue:test:accept-plan#branches-merged-everywhere
//specue:test:accept-plan#plan-record-closes
func TestAcceptMergesAndFlips(t *testing.T) {
	_, dirs, mgr, bin := projectRepo(t)
	exampleDir := dirs["x.test/example@v0"]
	root := filepath.Dir(exampleDir)
	govDir := dirs["x.test/governance@v0"]

	require.NoError(t, mgr.Register("gp-1", "add reverse"))
	require.NoError(t, mgr.Use("gp-1"))
	require.NoError(t, os.WriteFile(filepath.Join(exampleDir, "nodes.cue"), []byte(planNodes), 0o644))
	commit(t, bin, root, "plan: add reverse")
	require.NoError(t, mgr.Base())

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	proj, err := plan.NewProjector(parser, plan.NewGit(bin))
	require.NoError(t, err)
	t.Cleanup(func() { _ = proj.Close() })

	res, err := mgr.Accept(proj, dirs, "gp-1")
	require.NoError(t, err)
	assert.True(t, res.OK(), "clean plan accepts: %+v", res)

	// On base now: the reverse node is merged in, and the plan record reads
	// accepted (committed on base).
	assert.Equal(t, "base", currentBranch(t, bin, root))
	wt, _ := os.ReadFile(filepath.Join(exampleDir, "nodes.cue"))
	assert.Equal(t, planNodes, string(wt), "plan content merged onto base")
	rec, _ := os.ReadFile(filepath.Join(govDir, "plan-gp-1.cue"))
	assert.Contains(t, string(rec), `status:     "accepted"`, "Plan flipped to accepted")

	// The spent branch is deleted: an accepted plan is consumed (its record lives on
	// base), so `plan list` (which enumerates plan/<id> branches) no longer shows it.
	assert.False(t, branchExists(t, bin, root, "plan/gp-1"), "accepted plan's branch is removed")
}

// TestAcceptRollsBackOnBrokenSpec proves that a plan whose merged result is invalid
// (a reference to a node that does not exist — CUE rejects it) does not land: the
// merge is rolled back and base stays on its pre-merge commit.
//specue:test:accept-plan#merge-only-if-valid
func TestAcceptRollsBackOnBrokenSpec(t *testing.T) {
	_, dirs, mgr, bin := projectRepo(t)
	exampleDir := dirs["x.test/example@v0"]
	root := filepath.Dir(exampleDir)

	headBefore := revParse(t, bin, root, "base")

	require.NoError(t, mgr.Register("gp-bad", ""))
	require.NoError(t, mgr.Use("gp-bad"))
	// A use case whose dependency targets a non-existent node — CUE fails to resolve
	// the cross reference, so loading the merged landscape errors.
	broken := `package example
import s "specue.io/schema@v0:spec"
example: s.#Container & {type:"Container", slug:"example", title:"Wallet", confidence:"CONFIRMED", kind:"service"}
apply: s.#Contract & {type:"Contract", slug:"apply", title:"Apply", confidence:"CONFIRMED", service:example, postconditions:[{text:"done"}]}
ghostref: s.#Contract & {type:"Contract", slug:"ghostref", title:"G", confidence:"CONFIRMED", service:example, postconditions:[{text:"x", depends_on:[{to: missing}]}]}
`
	require.NoError(t, os.WriteFile(filepath.Join(exampleDir, "nodes.cue"), []byte(broken), 0o644))
	commit(t, bin, root, "plan: broken")
	require.NoError(t, mgr.Base())

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	proj, err := plan.NewProjector(parser, plan.NewGit(bin))
	require.NoError(t, err)
	t.Cleanup(func() { _ = proj.Close() })

	_, err = mgr.Accept(proj, dirs, "gp-bad")
	assert.Error(t, err, "a plan whose merged spec is invalid does not accept")

	// Base is back to its pre-merge commit; the broken content is not on base.
	assert.Equal(t, headBefore, revParse(t, bin, root, "base"), "merge rolled back")
	wt, _ := os.ReadFile(filepath.Join(exampleDir, "nodes.cue"))
	assert.Equal(t, baseNodes, string(wt), "working tree example back to base")
}

// TestAcceptFromPlanBranch is the regression for the bug where a caller still on
// the Plan branch would silently fail accept (baseBranch fell back to the
// current branch — the Plan branch — and merging a branch into itself broke).
// The fix: baseBranch never returns a plan/<id> branch and falls back to main
// or master instead, so accept lands cleanly regardless of where the caller is.
//
//specue:test:accept-plan#works-from-anywhere
func TestAcceptFromPlanBranch(t *testing.T) {
	_, dirs, mgr, bin := projectRepo(t)
	exampleDir := dirs["x.test/example@v0"]
	root := filepath.Dir(exampleDir)
	govDir := dirs["x.test/governance@v0"]

	// Rename "base" → "main" so the fallback path is exercised (projectRepo sets
	// up "base" as the default branch).
	for _, args := range [][]string{{"checkout", "-b", "main"}, {"branch", "-D", "base"}} {
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, out)
	}

	require.NoError(t, mgr.Register("gp-anywhere", "add reverse"))
	require.NoError(t, mgr.Use("gp-anywhere"))
	require.NoError(t, os.WriteFile(filepath.Join(exampleDir, "nodes.cue"), []byte(planNodes), 0o644))
	commit(t, bin, root, "plan: add reverse")

	// Stay on the plan branch — do not return to base before accept.
	assert.Equal(t, "plan/gp-anywhere", currentBranch(t, bin, root),
		"precondition: caller is still on the plan branch")

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	proj, err := plan.NewProjector(parser, plan.NewGit(bin))
	require.NoError(t, err)
	t.Cleanup(func() { _ = proj.Close() })

	res, err := mgr.Accept(proj, dirs, "gp-anywhere")
	require.NoError(t, err)
	assert.True(t, res.OK(), "accept succeeds from the plan branch: %+v", res)
	assert.Equal(t, "main", currentBranch(t, bin, root),
		"after accept, the worktree is on the resolved base branch")
	wt, _ := os.ReadFile(filepath.Join(exampleDir, "nodes.cue"))
	assert.Equal(t, planNodes, string(wt), "plan content landed on main")
	rec, _ := os.ReadFile(filepath.Join(govDir, "plan-gp-anywhere.cue"))
	assert.Contains(t, string(rec), `status:     "accepted"`, "Plan flipped to accepted")
}

// TestRegisterRefusesFromPlanBranch covers register-plan#from-base-only: a
// register run while the caller is on another plan branch is refused with an
// actionable error pointing at the base.
//
//specue:test:register-plan#from-base-only
func TestRegisterRefusesFromPlanBranch(t *testing.T) {
	_, dirs, mgr, bin := projectRepo(t)
	root := filepath.Dir(dirs["x.test/example@v0"])

	// Rename base → main so the fallback resolves a real "main" base.
	for _, args := range [][]string{{"checkout", "-b", "main"}, {"branch", "-D", "base"}} {
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, out)
	}

	require.NoError(t, mgr.Register("gp-first", "first plan"))
	require.NoError(t, mgr.Use("gp-first"))
	assert.Equal(t, "plan/gp-first", currentBranch(t, bin, root))

	// Try to register a second plan while still on the first plan's branch.
	err := mgr.Register("gp-second", "should fail")
	require.Error(t, err, "register from a plan branch is refused")
	assert.Contains(t, err.Error(), "main", "the error names the expected base")
}

// TestAcceptTagsTheLanding covers accept-plan#tags-the-landing: a successful
// accept marks the merge commit of every affected repo with `plan/<id>`, so
// `git tag --list` enumerates landed plans.
//
//specue:test:accept-plan#tags-the-landing
func TestAcceptTagsTheLanding(t *testing.T) {
	_, dirs, mgr, bin := projectRepo(t)
	exampleDir := dirs["x.test/example@v0"]
	root := filepath.Dir(exampleDir)

	require.NoError(t, mgr.Register("gp-tag", "add reverse"))
	require.NoError(t, mgr.Use("gp-tag"))
	require.NoError(t, os.WriteFile(filepath.Join(exampleDir, "nodes.cue"), []byte(planNodes), 0o644))
	commit(t, bin, root, "plan: add reverse")
	require.NoError(t, mgr.Base())

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	proj, err := plan.NewProjector(parser, plan.NewGit(bin))
	require.NoError(t, err)
	t.Cleanup(func() { _ = proj.Close() })

	res, err := mgr.Accept(proj, dirs, "gp-tag")
	require.NoError(t, err)
	require.True(t, res.OK(), "clean plan accepts")

	// `git tag --list plan/gp-tag` returns exactly the tag and it points at HEAD
	// of base (the merge commit).
	tagsCmd := exec.Command(bin, "tag", "--list", "plan/gp-tag")
	tagsCmd.Dir = root
	tags, err := tagsCmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "plan/gp-tag\n", string(tags), "the plan's tag lands at accept")

	tagSha := revParse(t, bin, root, "plan/gp-tag^{commit}")
	headSha := revParse(t, bin, root, "HEAD")
	assert.Equal(t, headSha, tagSha, "tag points at the merge commit (current base HEAD)")
}

func revParse(t *testing.T, bin, root, ref string) string {
	t.Helper()
	cmd := exec.Command(bin, "rev-parse", ref)
	cmd.Dir = root
	out, err := cmd.Output()
	require.NoError(t, err)
	return trim(string(out))
}
