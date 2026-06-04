package modules_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/modules"
	"github.com/specue/specue/internal/source"
)

// writeFile writes content under dir, creating parents.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

func mustParser(t *testing.T) source.Parser {
	t.Helper()
	p, err := source.NewCUEParser()
	require.NoError(t, err)
	return p
}

// twoModules lays out a consumer root that requires example (local replace ../example).
func twoModules(t *testing.T) (root string) {
	t.Helper()
	base := t.TempDir()
	writeFile(t, filepath.Join(base, "example"), source.ManifestFile, `
module: "specue.test/example@v0"
version: "v0.1.0"
kind:    "service"
`)
	writeFile(t, filepath.Join(base, "consumer"), source.ManifestFile, `
module: "specue.test/consumer@v0"
version: "v0.1.0"
kind:    "service"
require: [{
	module:  "specue.test/example@v0"
	version: "v0.1.0"
	replace: "../example"
}]
`)
	return filepath.Join(base, "consumer")
}

func TestResolveClosure(t *testing.T) {
	root := twoModules(t)
	r := modules.NewResolver(mustParser(t), modules.NewReplaceLocator())

	closure, err := r.Resolve([]modules.RootModule{{Path: "specue.test/consumer@v0", Dir: root}})
	require.NoError(t, err)

	byPath := map[model.ModulePath]modules.ResolvedModule{}
	for _, m := range closure.Modules {
		byPath[m.Path] = m
	}
	require.Len(t, byPath, 2, "root + transitive dep")

	consumer := byPath["specue.test/consumer@v0"]
	assert.True(t, consumer.IsRoot)
	example := byPath["specue.test/example@v0"]
	assert.False(t, example.IsRoot)
	assert.Equal(t, source.KindService, example.Manifest.Kind)
	assert.DirExists(t, example.Dir, "example located at its replace dir")
}

// TestRegistryResolvesImport proves the closure's registry lets CUE resolve a
// cue-native cross-module import end to end — the whole point of the layer.
func TestRegistryResolvesImport(t *testing.T) {
	base := t.TempDir()
	// example dependency (real dir, has a cue.mod for cue's module system).
	exampleDir := filepath.Join(base, "example")
	writeFile(t, exampleDir, "cue.mod/module.cue", "module: \"specue.test/example@v0\"\nlanguage: version: \"v0.16.0\"\n")
	writeFile(t, exampleDir, "example.cue", "package example\n#Validate: {slug: \"validate-graph\"}\n")
	writeFile(t, exampleDir, source.ManifestFile, `
module: "specue.test/example@v0"
version: "v0.1.0"
kind:    "service"
`)

	closure := modules.Closure{Modules: []modules.ResolvedModule{
		{Path: "specue.test/example@v0", Version: "v0.1.0", Dir: exampleDir},
	}}

	overlay := map[string]load.Source{
		"/work/consumer/cue.mod/module.cue": load.FromString(`
module: "specue.test/consumer@v0"
language: version: "v0.16.0"
deps: "specue.test/example@v0": v: "v0.0.1"
`),
		"/work/consumer/consumer.cue": load.FromString(`
package consumer
import w "specue.test/example@v0:example"
grant: onTopOf: w.#Validate.slug
`),
	}
	ctx := cuecontext.New()
	insts := load.Instances([]string{"."}, &load.Config{
		Dir:      "/work/consumer",
		Overlay:  overlay,
		Registry: closure.Registry(),
	})
	require.NoError(t, insts[0].Err)
	v := ctx.BuildInstance(insts[0])
	require.NoError(t, v.Err())

	got, _ := v.LookupPath(cue.ParsePath("grant.onTopOf")).String()
	assert.Equal(t, "validate-graph", got, "closure registry resolved the cross-module import")
}
