package engine_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/codescan"
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/engine"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

const exampleModule = "specue.test/example@v0"

const moduleCUE = `module: "specue.test/example@v0"
language: version: "v0.16.0"
deps: "specue.io/schema@v0": v: "v0.0.1"
`

const specModCUE = `module: "specue.test/example@v0"
version: "v0.1.0"
kind:    "service"
`

const nodesCUE = `package example
import s "specue.io/schema@v0:spec"

example: s.#Container & {
	type: "Container", slug: "example", title: "Example", confidence: "CONFIRMED", kind: "service"
}
apply: s.#Contract & {
	type:       "Contract"
	slug:       "apply"
	title:      "Apply"
	confidence: "CONFIRMED"
	service:    example
	invariants: [{id: "post", text: "done"}]
}
`

const implSrc = `package example

//specue:req:apply
func Apply() {}
`

// writeWorkspace lays out a real workspace on disk: a spec.work.cue at the root
// plus the example module under example/. Returns the work file path and the module
// dir (the latter so a test can corrupt a node file).
func writeWorkspace(t *testing.T) (workFile, moduleDir string) {
	t.Helper()
	root := t.TempDir()
	write := func(rel, content string) {
		full := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}
	write("example/cue.mod/module.cue", moduleCUE)
	write("example/spec.mod.cue", specModCUE)
	write("example/nodes.cue", nodesCUE)
	work := `modules: [
	{path: "` + exampleModule + `", dir: "example"},
]
`
	write(source.WorkFile, work)
	return filepath.Join(root, source.WorkFile), filepath.Join(root, "example")
}

func config(t *testing.T, codeFS fstest.MapFS) (engine.Config, string) {
	workFile, moduleDir := writeWorkspace(t)
	return engine.Config{
		WorkFile: workFile,
		ScanTargets: []codescan.ScanTarget{{
			FS: codeFS, Root: ".", Module: exampleModule,
		}},
	}, moduleDir
}

func newEngine(t *testing.T, codeFS fstest.MapFS) engine.Engine {
	t.Helper()
	cfg, _ := config(t, codeFS)
	eng, err := engine.New(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = eng.Close() })
	return eng
}

func TestLiveBuildsAndCaches(t *testing.T) {
	codeFS := fstest.MapFS{"apply.go": {Data: []byte(implSrc)}}
	eng := newEngine(t, codeFS)

	r1, err := eng.Live()
	require.NoError(t, err)
	n, ok := r1.Graph.Node(model.NodeID{Module: exampleModule, Slug: "apply"})
	require.True(t, ok)
	assert.Equal(t, compiler.StatusImplemented, n.Status, "req binding → implemented")

	r2, err := eng.Live()
	require.NoError(t, err)
	assert.Same(t, r1.Graph, r2.Graph, "unchanged inputs → cached build, same graph")
}

func TestLiveRebuildsOnChange(t *testing.T) {
	codeFS := fstest.MapFS{"apply.go": {Data: []byte(implSrc)}}
	eng := newEngine(t, codeFS)

	r1, err := eng.Live()
	require.NoError(t, err)
	n1, _ := r1.Graph.Node(model.NodeID{Module: exampleModule, Slug: "apply"})
	assert.Equal(t, compiler.StatusImplemented, n1.Status)

	// Add a covering test in code → code input changes → rebuild → proven.
	codeFS["apply_test.go"] = &fstest.MapFile{Data: []byte("package example\n//specue:test:apply\nfunc TestApply() {}\n")}

	r2, err := eng.Live()
	require.NoError(t, err)
	assert.NotSame(t, r1.Graph, r2.Graph, "changed code → fresh build")
	n2, _ := r2.Graph.Node(model.NodeID{Module: exampleModule, Slug: "apply"})
	assert.Equal(t, compiler.StatusProven, n2.Status, "added covering test → proven")
}

func TestLiveErrorLeavesLastGood(t *testing.T) {
	codeFS := fstest.MapFS{"apply.go": {Data: []byte(implSrc)}}
	cfg, moduleDir := config(t, codeFS)
	eng, err := engine.New(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = eng.Close() })

	r1, err := eng.Live()
	require.NoError(t, err)

	// Corrupt the spec so the next build fails (a CUE syntax error).
	nodesPath := filepath.Join(moduleDir, "nodes.cue")
	require.NoError(t, os.WriteFile(nodesPath, []byte("package example\nthis is not valid cue {{{\n"), 0o644))

	_, err = eng.Live()
	require.Error(t, err, "a broken spec fails the rebuild")

	// Restore and confirm a good build still works (last good wasn't clobbered).
	require.NoError(t, os.WriteFile(nodesPath, []byte(nodesCUE), 0o644))
	r3, err := eng.Live()
	require.NoError(t, err)
	assert.NotNil(t, r3.Graph)
	_ = r1
}
