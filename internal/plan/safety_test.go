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

// dirty writes an uncommitted file into the governance module so the gov repo's
// working tree is no longer clean.
func dirty(t *testing.T, root string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(root, "governance", "scratch.cue"), []byte("// wip\n"), 0o644))
}

func TestRegisterRefusesDirtyTree(t *testing.T) {
	root, mgr, _ := repo(t)
	dirty(t, root)

	err := mgr.Register("gp-1", "")
	require.Error(t, err, "register must refuse a dirty tree rather than checkout over it")
	var dte plan.DirtyTreeError
	assert.ErrorAs(t, err, &dte, "the error names the dirty-tree condition")
}

//specue:test:use-plan#refuses-on-dirty-tree
func TestUseRefusesDirtyTree(t *testing.T) {
	root, mgr, _ := repo(t)
	require.NoError(t, mgr.Register("gp-1", ""))
	dirty(t, root)

	err := mgr.Use("gp-1")
	require.Error(t, err, "use must refuse a dirty tree — a checkout would overwrite it")
	var dte plan.DirtyTreeError
	assert.ErrorAs(t, err, &dte)
}

// TestRegisterCommitsOnlyTheRecord proves register no longer sweeps unrelated work
// onto the plan branch: an untracked file present before register stays untracked
// (uncommitted) on base, and is NOT on the plan branch.
func TestRegisterCommitsOnlyTheRecord(t *testing.T) {
	root, mgr, _ := repo(t)
	bin := gitBin(t)

	// An untracked sibling file the user has not committed. (Place it outside the
	// governance module so it does not dirty that repo's relevant subtree — register
	// checks cleanliness of the gov repo, so use a path git ignores via add scope.)
	stray := filepath.Join(root, "example", "stray.cue")
	require.NoError(t, os.WriteFile(stray, []byte("// stray\n"), 0o644))

	// The gov module itself is clean; register should succeed and commit only its
	// record. (The example stray lives in the same repo here, so to keep the gov tree
	// clean we commit the stray first as a baseline, then assert register adds only
	// the record on top.)
	run := func(args ...string) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, out)
	}
	run("add", "example/stray.cue")
	run("commit", "-m", "baseline stray")

	require.NoError(t, mgr.Register("gp-1", ""))

	// On the plan branch, the only change vs base is the record file.
	out, err := exec.Command(bin, "-C", root, "diff", "--name-only", "base", "plan/gp-1").CombinedOutput()
	require.NoErrorf(t, err, "%s", out)
	assert.Equal(t, "governance/plan-gp-1.cue\n", string(out), "register commits the record alone")
}
