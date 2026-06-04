package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// planLandscape builds a single git repo holding a governance module and one service
// module, registers them as the active context "test", and returns the repo root and
// git binary. $SPECUE_GIT and $SPECUE_HOME are set by the caller's t.Setenv.
// Plan tests then run against the active context (no --workspace needed). Skips if
// git is unavailable.
func planLandscape(t *testing.T) (root, gitPath string) {
	t.Helper()
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available")
	}
	// A per-test specue home so the context registry is isolated.
	t.Setenv("SPECUE_HOME", filepath.Join(t.TempDir(), ".specue"))
	t.Setenv("SPECUE_GIT", bin)

	root = t.TempDir()
	git := func(args ...string) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		out, e := cmd.CombinedOutput()
		require.NoErrorf(t, e, "git %v: %s", args, out)
	}
	write := func(rel, content string) {
		full := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}

	git("init", "-b", "base")
	git("config", "user.email", "t@t")
	git("config", "user.name", "t")

	write("governance/spec.mod.cue", `module: "x.test/governance@v0"
version: "v0.0.1"
kind:    "governance"
`)
	write("governance/cue.mod/module.cue", `module: "x.test/governance@v0"
language: version: "v0.16.0"
deps: "specue.io/schema@v0": v: "v0.0.1"
`)
	write("example/spec.mod.cue", `module: "x.test/example@v0"
version: "v0.0.1"
kind:    "service"
`)
	write("example/cue.mod/module.cue", `module: "x.test/example@v0"
language: version: "v0.16.0"
deps: "specue.io/schema@v0": v: "v0.0.1"
`)
	git("add", "-A")
	git("commit", "-m", "base")

	// Register the landscape as the active context "test" via the CLI.
	mustRun := func(args ...string) {
		_, errOut, code := run(args...)
		require.Equalf(t, exitOK, code, "%v: %s", args, errOut)
	}
	mustRun("context", "create", "test")
	mustRun("context", "module", "add", filepath.Join(root, "governance"), "--workspace", "test")
	mustRun("context", "module", "add", filepath.Join(root, "example"), "--workspace", "test")
	mustRun("context", "use", "test")

	return root, bin
}

func TestPlanRegisterListShow(t *testing.T) {
	planLandscape(t)

	// register
	_, errOut, code := run("plan", "register", "gp-1076", "--title", "Drop virtual id")
	require.Equalf(t, exitOK, code, "stderr: %s", errOut)

	// list shows it
	out, _, code := run("plan", "list", "--json")
	require.Equal(t, exitOK, code)
	var listed struct {
		Plans []struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"plans"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &listed))
	require.Len(t, listed.Plans, 1)
	assert.Equal(t, "gp-1076", listed.Plans[0].ID)
	assert.Equal(t, "Drop virtual id", listed.Plans[0].Title)
	assert.Equal(t, "proposed", listed.Plans[0].Status)

	// show one
	out, _, code = run("plan", "show", "gp-1076")
	require.Equal(t, exitOK, code)
	assert.Contains(t, out, "gp-1076")
	assert.Contains(t, out, "proposed")
}

func TestPlanListEmpty(t *testing.T) {
	planLandscape(t)

	out, _, code := run("plan", "list")
	require.Equal(t, exitOK, code)
	assert.Contains(t, out, "no open plans")
}

func TestPlanShowMissingIsActionable(t *testing.T) {
	planLandscape(t)

	_, errOut, code := run("plan", "show", "nope")
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
	assert.Contains(t, errOut, "plan list", "the fix points at how to find valid ids")
}

func TestPlanAcceptNeedsConfirmationOrForce(t *testing.T) {
	planLandscape(t)
	_, _, code := run("plan", "register", "gp-1076")
	require.Equal(t, exitOK, code)

	// --no-input without --force must refuse to prompt and exit 2 with the fix.
	_, errOut, code := run("plan", "accept", "gp-1076", "--no-input")
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "--force", "the fix names how to proceed unattended")
}

// TestPlanUseThenAcceptLifecycle is the full working cycle: register, use (checkout
// the plan branch), then accept. It guards the chain of bugs where accept, run after
// use, merged into the plan branch instead of base and could not free the branch —
// requiring plan_base to be set and accept to checkout base first.
func TestPlanUseThenAcceptLifecycle(t *testing.T) {
	planLandscape(t)

	_, _, code := run("plan", "register", "np")
	require.Equal(t, exitOK, code)
	_, _, code = run("plan", "use", "np")
	require.Equal(t, exitOK, code)

	_, errOut, code := run("plan", "accept", "np", "--force")
	require.Equalf(t, exitOK, code, "accept after use must land cleanly; stderr: %s", errOut)

	out, _, code := run("plan", "list")
	require.Equal(t, exitOK, code)
	assert.Contains(t, out, "no open plans", "accepted plan is consumed even when reached via use")
}

func TestPlanAcceptConsumesPlan(t *testing.T) {
	planLandscape(t)
	_, _, code := run("plan", "register", "gp-1076")
	require.Equal(t, exitOK, code)

	// accept --force lands the (empty) plan; it is then consumed.
	_, errOut, code := run("plan", "accept", "gp-1076", "--force")
	require.Equalf(t, exitOK, code, "stderr: %s", errOut)

	// list no longer shows it — the accepted plan's branch is gone.
	out, _, code := run("plan", "list")
	require.Equal(t, exitOK, code)
	assert.Contains(t, out, "no open plans", "accepted plan is consumed, not listed as open")
}

// TestPlanInModuleModeNeedsWorkspace pins that a plan verb run against a single
// module (module mode) points at the workspace bootstrap — a plan needs a workspace
// with a governance module, not a lone module.
func TestPlanInModuleModeNeedsWorkspace(t *testing.T) {
	// walletSpec is a single service module; -C resolves it in module mode.
	_, errOut, code := run("plan", "list", "-C", walletSpec)
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
	assert.Contains(t, errOut, "workspace", "the fix points at creating a workspace")
	assert.Contains(t, errOut, "--kind governance", "and the governance module a plan needs")
}

// TestPlanNoLandscapeGivesGovernanceBootstrap pins that a plan verb with nothing
// resolved points straight at the workspace+governance bootstrap — NOT the generic
// "scaffold a module" hint.
func TestPlanNoLandscapeGivesGovernanceBootstrap(t *testing.T) {
	t.Setenv("SPECUE_HOME", filepath.Join(t.TempDir(), ".specue")) // no active context
	_, errOut, code := run("plan", "register", "foo", "-C", t.TempDir())
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "workspace", "the fix points at creating a workspace")
	assert.Contains(t, errOut, "--kind governance", "and the command that creates the governance module")
	assert.NotContains(t, errOut, "scaffold a module", "must not give the generic module hint")
}

// TestPlanWorkspaceWithoutGovernance pins that a workspace lacking a governance
// module gets "add one to this workspace", not the full bootstrap.
func TestPlanWorkspaceWithoutGovernance(t *testing.T) {
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available")
	}
	t.Setenv("SPECUE_HOME", filepath.Join(t.TempDir(), ".specue"))
	t.Setenv("SPECUE_GIT", bin)
	dir := t.TempDir()
	// git-native (P20): the module must live in a repo for init / module add.
	out, e := exec.Command(bin, "init", dir).CombinedOutput()
	require.NoErrorf(t, e, "git init: %s", out)
	_, _, code := run("init", filepath.Join(dir, "svc"), "x.test/svc@v0")
	require.Equal(t, exitOK, code)
	_, _, code = run("context", "create", "nogov")
	require.Equal(t, exitOK, code)
	_, _, code = run("context", "module", "add", filepath.Join(dir, "svc"), "--workspace", "nogov")
	require.Equal(t, exitOK, code)

	_, errOut, code := run("plan", "list", "--workspace", "nogov")
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "no governance module in this workspace")
	assert.Contains(t, errOut, "module add", "the fix adds a governance module to the workspace")
}
