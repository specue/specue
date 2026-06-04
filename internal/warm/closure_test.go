package warm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/modules"
	"github.com/specue/specue/internal/source"
)

// writeFile lays down a file under dir with content, creating parents.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

// fakeClosure builds a Closure of N small modules under base; each module has a
// cue.mod and one node file with the given marker as content. Returns the
// closure and the per-module dirs (so a stub resolve can find them).
func fakeClosure(t *testing.T, base string, mods ...string) (modules.Closure, []string) {
	t.Helper()
	var c modules.Closure
	var dirs []string
	for _, name := range mods {
		dir := filepath.Join(base, name)
		dirs = append(dirs, dir)
		writeFile(t, dir, "cue.mod/module.cue",
			"module: \"x.test/"+name+"@v0\"\nlanguage: version: \"v0.16.0\"\n")
		writeFile(t, dir, "x.cue", "package "+name+"\nmarker: \""+name+"\"\n")
		c.Modules = append(c.Modules, modules.ResolvedModule{
			Path:    model.ModulePath("x.test/" + name + "@v0"),
			Version: source.Version("v0.1.0"),
			Dir:     dir,
		})
	}
	return c, dirs
}

// stubClosureResolve emulates cue's resolve step: it creates the extract dir for
// every module in the closure (read-only, as cue does) so the warmer's freshness
// check has something to find. It records the number of times it ran so a test
// can assert idempotency.
func stubClosureResolve(w *Warmer, closure modules.Closure) (ClosureResolveFunc, *int) {
	calls := 0
	return func(addr string, roots []string) error {
		calls++
		for _, m := range closure.Modules {
			spec := ModuleSpec{Path: string(m.Path), Version: string(m.Version), Dir: m.Dir}
			ext := w.moduleExtractDir(spec)
			// Real cue does not re-extract an existing module: if the dir is there,
			// it serves the cached content as-is. Mirror that — the warmer's clear
			// step is what makes a re-warm possible. (Without this the stub would
			// fail to write into a read-only extract left from a previous call.)
			if _, err := os.Stat(ext); err == nil {
				continue
			}
			if err := os.MkdirAll(ext, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(ext, "x.cue"), []byte("package "+filepath.Base(m.Dir)+"\n"), 0o444); err != nil {
				return err
			}
			if err := os.Chmod(ext, 0o555); err != nil {
				return err
			}
		}
		return nil
	}, &calls
}

func TestEnsureClosureWarm_ColdThenNoOp(t *testing.T) {
	cache := t.TempDir()
	w, err := New(cache, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = chmodTreeWritable(cache) })

	closure, dirs := fakeClosure(t, t.TempDir(), "alpha", "beta")
	resolve, calls := stubClosureResolve(w, closure)

	rewarmed, err := w.EnsureClosureWarm(closure, dirs, resolve)
	require.NoError(t, err)
	assert.True(t, rewarmed, "cold cache should warm")
	assert.Equal(t, 1, *calls)

	// Same content → no-op, no second resolve.
	rewarmed, err = w.EnsureClosureWarm(closure, dirs, resolve)
	require.NoError(t, err)
	assert.False(t, rewarmed, "fresh closure should be a no-op")
	assert.Equal(t, 1, *calls, "fresh modules should not re-warm")
}

func TestEnsureClosureWarm_RewarmsOnContentChange(t *testing.T) {
	cache := t.TempDir()
	w, err := New(cache, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = chmodTreeWritable(cache) })

	closure, dirs := fakeClosure(t, t.TempDir(), "alpha", "beta")
	resolve, calls := stubClosureResolve(w, closure)

	_, err = w.EnsureClosureWarm(closure, dirs, resolve)
	require.NoError(t, err)
	assert.Equal(t, 1, *calls)

	// Change alpha's content; beta is untouched. Both still warm because the warm
	// step republishes the whole closure into a fresh ephemeral registry — but
	// only alpha's extract needed to be cleared first.
	writeFile(t, closure.Modules[0].Dir, "x.cue", "package alpha\nmarker: \"alpha-v2\"\n")

	rewarmed, err := w.EnsureClosureWarm(closure, dirs, resolve)
	require.NoError(t, err)
	assert.True(t, rewarmed, "content change in one module should re-warm")
	assert.Equal(t, 2, *calls)
}

func TestEnsureClosureWarm_ClearsReadOnlyExtract(t *testing.T) {
	cache := t.TempDir()
	w, err := New(cache, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = chmodTreeWritable(cache) })

	closure, dirs := fakeClosure(t, t.TempDir(), "alpha")
	resolve, _ := stubClosureResolve(w, closure)

	// Pre-seed a read-only extract with no stamp → must be cleared and re-warmed.
	spec := ModuleSpec{Path: string(closure.Modules[0].Path), Version: string(closure.Modules[0].Version), Dir: closure.Modules[0].Dir}
	require.NoError(t, os.MkdirAll(w.moduleExtractDir(spec), 0o755))
	require.NoError(t, os.Chmod(w.moduleExtractDir(spec), 0o555))

	_, err = w.EnsureClosureWarm(closure, dirs, resolve)
	require.NoError(t, err, "must cope with a read-only pre-existing extract")
}

// injectSelfSource is the textual append injection: a fresh module.cue gets the
// source line; an existing one with `source:` is left alone.
func TestInjectSelfSource(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "module.cue")

	// Fresh: gets the line appended.
	require.NoError(t, os.WriteFile(p, []byte("module: \"x@v0\"\nlanguage: version: \"v0.16.0\"\n"), 0o644))
	require.NoError(t, injectSelfSource(p))
	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Contains(t, string(raw), "source: kind: \"self\"")

	// Already declares a source: left alone.
	original := "module: \"x@v0\"\nsource: kind: \"git\"\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))
	require.NoError(t, injectSelfSource(p))
	raw, err = os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, original, string(raw), "an existing source field is left untouched")
}
