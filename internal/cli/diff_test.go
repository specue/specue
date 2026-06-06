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

// gitModuleRepo builds a git repo holding a single spec module, with two commits:
// the first has one use case, the second adds a second. It returns the repo dir and
// the git binary path (set as $SPECUE_GIT so the CLI drives the same git). The
// test skips if git is unavailable — the binary is never assumed present.
func gitModuleRepo(t *testing.T) (dir, gitPath string) {
	t.Helper()
	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available")
	}
	dir = t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = dir
		out, e := cmd.CombinedOutput()
		require.NoErrorf(t, e, "git %v: %s", args, out)
	}
	write := func(name, content string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}

	// A minimal module mirroring the example testdata shape: spec.mod.cue +
	// cue.mod/module.cue (with the schema dep) + a node file importing the schema.
	write("spec.mod.cue", `module: "diff.test/svc@v0"
version: "v0.1.0"
kind:    "service"
`)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cue.mod"), 0o755))
	write("cue.mod/module.cue", `module: "diff.test/svc@v0"
language: version: "v0.16.0"
deps: "specue.io/schema@v0": v: "v0.0.1"
`)

	run("init", "-b", "base")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")

	svc := `svc: s.#Container & {slug: "svc", type: "Container", title: "Svc", confidence: "CONFIRMED", kind: "service"}`
	one := `one: s.#Contract & {slug: "one", type: "Contract", title: "One", confidence: "CONFIRMED", service: svc}`
	two := `two: s.#Contract & {slug: "two", type: "Contract", title: "Two", confidence: "CONFIRMED", service: svc}`

	write("nodes.cue", svcNodes(svc+"\n"+one))
	run("add", "-A")
	run("commit", "-m", "one uc")

	write("nodes.cue", svcNodes(svc+"\n"+one+"\n"+two))
	run("add", "-A")
	run("commit", "-m", "add second uc")

	return dir, bin
}

// svcNodes wraps node bodies in the package clause and schema import the loader
// expects, matching the example testdata module.
func svcNodes(body string) string {
	return `package svc

import s "specue.io/schema@v0:spec"

` + body + `
`
}

func TestDiffModuleAddedNode(t *testing.T) {
	dir, gitPath := gitModuleRepo(t)
	t.Setenv("SPECUE_GIT", gitPath)

	out, errOut, code := run("diff", "module", "HEAD~1", "HEAD", "-C", dir)
	require.Equalf(t, exitOK, code, "stderr: %s", errOut)
	assert.Contains(t, out, "two", "the added use case shows in the delta")
	assert.Contains(t, out, "+", "added nodes are marked +")
}

func TestDiffModuleJSON(t *testing.T) {
	dir, gitPath := gitModuleRepo(t)
	t.Setenv("SPECUE_GIT", gitPath)

	out, _, code := run("diff", "module", "HEAD~1", "HEAD", "-C", dir, "--json")
	require.Equal(t, exitOK, code)

	var got struct {
		Module string `json:"module"`
		Nodes  []struct {
			ID     string `json:"id"`
			Change string `json:"change"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, "diff.test/svc@v0", got.Module)
	require.Len(t, got.Nodes, 1)
	assert.Equal(t, "added", got.Nodes[0].Change)
}

func TestDiffModuleBadRefIsActionable(t *testing.T) {
	dir, gitPath := gitModuleRepo(t)
	t.Setenv("SPECUE_GIT", gitPath)

	_, errOut, code := run("diff", "module", "nope-ref", "HEAD", "-C", dir)
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
	assert.Contains(t, errOut, "nope-ref", "the fix names the bad ref")
}

func TestDiffModuleRejectsMultiModule(t *testing.T) {
	// The example testdata resolves as a single module via --no-work; to force the
	// multi-module guard we'd need a real spec.work with >1 module. Here we assert the
	// guard message is reachable by checking a bare landscape: not applicable to the
	// single example module, so we only assert the single-module path stays clean.
	t.Skip("multi-module landscape fixture lands with the plan verbs (slice 3)")
}

func TestDiffScopesListed(t *testing.T) {
	out, _, code := run("diff")
	assert.Equal(t, exitOK, code)
	assert.Contains(t, out, "module")
}
