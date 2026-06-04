package engine_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/engine"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// TestScanTargetsForOnlyCodeModules pins that the builder emits a target for each
// code-kind module and skips the rest, using the injected lister/FS (no git, no
// real reads of the listed files) — so any caller reuses the same selection.
func TestScanTargetsForOnlyCodeModules(t *testing.T) {
	root := t.TempDir()
	writeModule(t, filepath.Join(root, "example"), "x.test/example@v0", "service")
	writeModule(t, filepath.Join(root, "example-code"), "x.test/example-code@v0", "code")

	work := source.Workspace{Modules: []source.WorkModule{
		{Path: "x.test/example@v0", Dir: filepath.Join(root, "example")},
		{Path: "x.test/example-code@v0", Dir: filepath.Join(root, "example-code")},
	}}
	dirs := map[model.ModulePath]string{
		"x.test/example@v0":      filepath.Join(root, "example"),
		"x.test/example-code@v0": filepath.Join(root, "example-code"),
	}

	// Stub lister: every code module "tracks" one file; record which dirs were asked.
	var listed []string
	lister := func(dir string) ([]string, error) {
		listed = append(listed, dir)
		return []string{"apply.go"}, nil
	}
	fsFor := func(string) fs.FS { return fstest.MapFS{"apply.go": {Data: []byte("package x")}} }

	targets, err := engine.ScanTargetsFor(work, dirs, lister, fsFor)
	require.NoError(t, err)

	require.Len(t, targets, 1, "only the code module yields a target")
	assert.Equal(t, model.ModulePath("x.test/example-code@v0"), targets[0].Module)
	assert.Equal(t, []string{"apply.go"}, targets[0].Files)
	assert.Equal(t, []string{filepath.Join(root, "example-code")}, listed,
		"the lister is asked only for the code module's dir")
}

func writeModule(t *testing.T, dir, modulePath, kind string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	src := "module: \"" + modulePath + "\"\nversion: \"v0.1.0\"\nkind:    \"" + kind + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, source.ManifestFile), []byte(src), 0o644))
}

// TestScanTargetsForHonorsIgnore pins that a code module's ignore globs drop files
// from the scan set — testdata/fixtures/generated code that carry foreign
// annotations stay out, real code stays in.
func TestScanTargetsForHonorsIgnore(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "code")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	src := "module: \"x.test/code@v0\"\nversion: \"v0.1.0\"\nkind:    \"code\"\n" +
		"ignore: [\"testdata/\", \"*.gen.go\"]\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, source.ManifestFile), []byte(src), 0o644))

	work := source.Workspace{Modules: []source.WorkModule{{Path: "x.test/code@v0", Dir: dir}}}
	dirs := map[model.ModulePath]string{"x.test/code@v0": dir}

	lister := func(string) ([]string, error) {
		return []string{"internal/app.go", "testdata/fixture.go", "api/types.gen.go"}, nil
	}
	fsFor := func(string) fs.FS { return fstest.MapFS{} }

	targets, err := engine.ScanTargetsFor(work, dirs, lister, fsFor)
	require.NoError(t, err)
	require.Len(t, targets, 1)
	assert.Equal(t, []string{"internal/app.go"}, targets[0].Files,
		"testdata/ and *.gen.go are excluded; real code kept")
}
