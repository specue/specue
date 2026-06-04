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

// TestMaterializeSubtreeReadsBranchWithoutCheckout proves a module's tree is
// extracted from a branch ref into a temp dir (subdir prefix stripped) while the
// working tree stays on base — the read-only projection a plan diff needs.
func TestMaterializeSubtreeReadsBranchWithoutCheckout(t *testing.T) {
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
	write("example/spec.mod.cue", "module: \"x.test/example@v0\"\n")
	write("example/nodes.cue", "// base\n")
	run("add", "-A")
	run("commit", "-m", "base")

	// Diverge example/nodes.cue on a plan branch.
	run("checkout", "-b", "plan/x")
	write("example/nodes.cue", "// plan version\n")
	run("add", "-A")
	run("commit", "-m", "plan edit")
	run("checkout", "base")

	mz := plan.NewMaterializer(plan.NewGit(bin))
	out, err := mz.Subtree(root, "plan/x", "example")
	require.NoError(t, err)
	t.Cleanup(func() { _ = out.Cleanup() })

	// Files landed at dest root (example/ prefix stripped), carrying the plan-branch
	// content — not what is in the working tree (which is on base).
	got, err := os.ReadFile(filepath.Join(out.Dir, "nodes.cue"))
	require.NoError(t, err)
	assert.Equal(t, "// plan version\n", string(got))
	assert.FileExists(t, filepath.Join(out.Dir, "spec.mod.cue"))

	// The working tree was untouched (still base content).
	wt, _ := os.ReadFile(filepath.Join(root, "example/nodes.cue"))
	assert.Equal(t, "// base\n", string(wt), "working tree stays on base")
}
