package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// walletSpec is a known-good single-module spec used across the engine tests; the
// CLI resolves it in module mode (-C).
const walletSpec = "../compiler/testdata/example/spec"

// TestMain isolates the context registry for the whole package: SPECUE_HOME
// points at a throwaway dir so no test ever reads the developer's real ~/.specue
// (which would make "no active context" assertions flaky). A test that needs its own
// registry overrides SPECUE_HOME via t.Setenv.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "specue-cli-test-home-")
	if err != nil {
		panic(err)
	}
	os.Setenv("SPECUE_HOME", home)
	// Keep the suite hermetic: the implicit schema-cache warm shells out to `cue`
	// and writes the real user cache. Off for tests; the warm path has its own.
	os.Setenv("SPECUE_NO_AUTOWARM", "1")
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}

func run(args ...string) (stdout, stderr string, code int) {
	var out, errb bytes.Buffer
	code = Execute(args, &out, &errb)
	return out.String(), errb.String(), code
}

//specue:test:validate-graph
//specue:test:validate-graph#single-verdict
func TestValidateCleanModuleHuman(t *testing.T) {
	out, _, code := run("validate", "-C", walletSpec)
	assert.Equal(t, exitOK, code)
	assert.Contains(t, out, "node(s) valid")
}

func TestValidateCleanModuleJSON(t *testing.T) {
	out, _, code := run("validate", "-C", walletSpec, "--json")
	require.Equal(t, exitOK, code)

	var got struct {
		OK    bool `json:"ok"`
		Nodes int  `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.True(t, got.OK)
	assert.Positive(t, got.Nodes)
}

// TestValidateNoSpecTreeIsActionable is the core invariant: a resolution failure
// exits 2 AND tells the caller what to do (a non-empty `try:` line). No error
// leaves the reader guessing the remedy.
func TestValidateNoSpecTreeIsActionable(t *testing.T) {
	_, errOut, code := run("validate", "-C", t.TempDir())
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "error:")
	assert.Contains(t, errOut, "try:", "every error must carry an actionable next step")
	// the fix names a concrete lever: pick a context, cd into a module, or scaffold one
	assert.True(t, strings.Contains(errOut, "context") || strings.Contains(errOut, "module"),
		"the fix points at a way to get a landscape")
}

func TestValidateErrorJSONCarriesFix(t *testing.T) {
	out, _, code := run("validate", "-C", t.TempDir(), "--json")
	require.Equal(t, exitUsage, code)

	var got struct {
		Error string `json:"error"`
		Fix   string `json:"fix"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.NotEmpty(t, got.Error)
	assert.NotEmpty(t, got.Fix, "JSON errors must carry the fix too")
}

func TestUnknownCommandIsUsageError(t *testing.T) {
	_, errOut, code := run("frobnicate")
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
}

// TestStaleModuleDirIsActionable pins the preflight: a context whose module dir was
// deleted fails with a message naming the dir and the remedy, NOT the opaque
// "stat .: no such file or directory" that surfaced from CUE before the check.
func TestStaleModuleDirIsActionable(t *testing.T) {
	dir := gitTempDir(t) // sets SPECUE_GIT; a repo so init/module add are allowed
	mod := filepath.Join(dir, "svc")
	_, _, code := run("init", mod, "x.test/svc@v0")
	require.Equal(t, exitOK, code)
	_, _, code = run("context", "create", "stale")
	require.Equal(t, exitOK, code)
	_, _, code = run("context", "module", "add", mod, "--workspace", "stale")
	require.Equal(t, exitOK, code)
	_, _, code = run("context", "use", "stale")
	require.Equal(t, exitOK, code)

	// The module dir vanishes after registration.
	require.NoError(t, os.RemoveAll(mod))

	_, errOut, code := run("validate")
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "does not exist", "names the vanished dir")
	assert.Contains(t, errOut, "module remove", "the fix drops the stale entry")
	assert.NotContains(t, errOut, "stat .", "no opaque CUE error leaks through")
}
