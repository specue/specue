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

// gitTempDir returns a fresh temp dir that is a git repository — the git-native
// invariant (MANIFESTO P20) means init / module add only operate inside one.
// SPECUE_GIT is pointed at the located git so the gate uses the same binary.
func gitTempDir(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available")
	}
	t.Setenv("SPECUE_GIT", bin)
	dir := t.TempDir()
	cmd := exec.Command(bin, "init", dir)
	out, e := cmd.CombinedOutput()
	require.NoErrorf(t, e, "git init: %s", out)
	return dir
}

func TestInitScaffoldsModule(t *testing.T) {
	dir := gitTempDir(t)
	out, errOut, code := run("init", dir, "x.test/svc@v0")
	require.Equalf(t, exitOK, code, "stderr: %s", errOut)
	assert.Contains(t, out, "x.test/svc@v0")

	mf, err := os.ReadFile(filepath.Join(dir, "spec.mod.cue"))
	require.NoError(t, err)
	assert.Contains(t, string(mf), `module: "x.test/svc@v0"`)
	assert.Contains(t, string(mf), `kind:    "service"`)

	cm, err := os.ReadFile(filepath.Join(dir, "cue.mod", "module.cue"))
	require.NoError(t, err)
	assert.Contains(t, string(cm), "specue.io/schema@v0")
}

func TestInitGovernanceThenValidate(t *testing.T) {
	dir := gitTempDir(t)
	_, _, code := run("init", dir, "x.test/gov@v0", "--kind", "governance")
	require.Equal(t, exitOK, code)

	// A freshly initialized module validates clean (no nodes yet).
	_, _, code = run("validate", "-C", dir)
	assert.Equal(t, exitOK, code, "a fresh module is a valid empty graph")
}

func TestInitJSON(t *testing.T) {
	dir := gitTempDir(t)
	out, _, code := run("init", dir, "x.test/svc@v0", "--json")
	require.Equal(t, exitOK, code)
	var got struct {
		Module string   `json:"module"`
		Kind   string   `json:"kind"`
		Files  []string `json:"files"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, "x.test/svc@v0", got.Module)
	assert.Equal(t, "service", got.Kind)
	assert.Len(t, got.Files, 2)
}

//specue:test:init-module#no-overwrite
func TestInitRefusesOverwrite(t *testing.T) {
	dir := gitTempDir(t)
	_, _, code := run("init", dir, "x.test/svc@v0")
	require.Equal(t, exitOK, code)

	_, errOut, code := run("init", dir, "x.test/other@v0")
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
	assert.Contains(t, errOut, "already", "the error names the conflict")
	assert.Contains(t, errOut, dir, "the error names which directory is already a module")
}

func TestInitBadModulePathIsActionable(t *testing.T) {
	// A path with no dot in the first element is what CUE rejects on load; init must
	// catch it at creation, not let it through to break later commands.
	_, errOut, code := run("init", t.TempDir(), "governance", "--kind", "governance")
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
	assert.Contains(t, errOut, "@v0", "the fix shows the required path shape")
}

// TestInitOutsideGitIsActionable pins the git-native gate (MANIFESTO P20): init in
// a directory not inside any git repository is refused, with `git init` as the fix.
//
//specue:test:init-module#git-repository-required
func TestInitOutsideGitIsActionable(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	// A bare temp dir with no enclosing repo. Point SPECUE_GIT at real git so the
	// gate's failure is "not a repo", not "git missing".
	bin, _ := exec.LookPath("git")
	t.Setenv("SPECUE_GIT", bin)
	_, errOut, code := run("init", t.TempDir(), "x.test/svc@v0")
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
	assert.Contains(t, errOut, "git init", "the fix names how to make it a repository")
}

// TestBindingsIsolatedModuleIsActionable pins that bindings on an isolated (module
// mode) code module refuses with the workspace remedy — its required contracts are
// not loaded, so the view would be all-orphan noise. -C makes the isolation
// explicit (an implicit cwd module would defer to an active context instead).
func TestBindingsIsolatedModuleIsActionable(t *testing.T) {
	dir := gitTempDir(t)
	_, _, code := run("init", dir, "x.test/code@v0", "--kind", "code")
	require.Equal(t, exitOK, code)

	_, errOut, code := run("bindings", "-C", dir)
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
	assert.Contains(t, errOut, "workspace", "the fix points at using a workspace")
}

// TestActiveContextBeatsImplicitCwdModule pins the precedence rule: an active
// context stays in force even when the cwd happens to be a module dir — only an
// explicit -C downgrades to module mode. This is what lets you work in a workspace
// while standing in a module's directory (e.g. a dogfooded code module at the repo
// root). Regression for the silent isolation that ignored the active context.
func TestActiveContextBeatsImplicitCwdModule(t *testing.T) {
	dir := gitTempDir(t) // a git repo; sets SPECUE_GIT
	t.Setenv("SPECUE_HOME", filepath.Join(dir, ".sghome"))
	svc := filepath.Join(dir, "svc")

	mustOK := func(args ...string) {
		_, errOut, code := run(args...)
		require.Equalf(t, exitOK, code, "%v: %s", args, errOut)
	}
	mustOK("init", svc, "x.test/svc@v0")
	mustOK("context", "create", "ws")
	mustOK("context", "module", "add", svc, "--workspace", "ws")
	mustOK("context", "use", "ws")

	// cwd at the module dir (no -C): the cwd is a module, but the active context wins.
	t.Chdir(svc)

	// get goes through dispatch, which prints the run-mode banner on stderr.
	_, errOut, code := run("get", "contract")
	require.Equalf(t, exitOK, code, "stderr: %s", errOut)
	assert.Contains(t, errOut, "workspace: ws", "the active context wins over the implicit cwd module")
	assert.NotContains(t, errOut, "isolated", "an implicit cwd module must not silently isolate")
}

func TestInitBadKindIsActionable(t *testing.T) {
	_, errOut, code := run("init", t.TempDir(), "x.test/svc@v0", "--kind", "frob")
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
	assert.Contains(t, errOut, "governance", "the fix lists the valid kinds")
}
