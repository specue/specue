package warm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/source"
)

// newWarmer builds a Warmer over a temp cache and registers cleanup that restores
// write permission — the stub leaves a read-only extract (as cue does), which
// would otherwise defeat t.TempDir's RemoveAll.
func newWarmer(t *testing.T) *Warmer {
	t.Helper()
	cache := t.TempDir()
	w, err := New(cache, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = chmodTreeWritable(cache) })
	return w
}

// stubResolve stands in for `cue vet` against the live registry: the real resolve
// materializes the schema extract, so the stub creates the extract dir (read-only,
// as cue does) to exercise the freshness/clear logic without invoking cue.
func stubResolve(w *Warmer) (ResolveFunc, *int) {
	calls := 0
	return func(addr string) error {
		calls++
		dir := w.extractDir()
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "spec.cue"), []byte("package spec\n"), 0o444); err != nil {
			return err
		}
		// cue makes the extract read-only; clearExtract must cope with that.
		return os.Chmod(dir, 0o555)
	}, &calls
}

//specue:test:warm-schema#no-op-when-current
func TestEnsureWarm_ColdThenNoOp(t *testing.T) {
	w := newWarmer(t)
	resolve, calls := stubResolve(w)
	w.cueRun = resolve

	rewarmed, err := w.EnsureWarm()
	require.NoError(t, err)
	assert.True(t, rewarmed, "cold cache should warm")
	assert.Equal(t, 1, *calls)

	// Content unchanged → no-op, no second resolve.
	rewarmed, err = w.EnsureWarm()
	require.NoError(t, err)
	assert.False(t, rewarmed, "fresh cache should be a no-op")
	assert.Equal(t, 1, *calls, "fresh should not resolve again")
}

func TestEnsureWarm_RewarmsOnContentChange(t *testing.T) {
	w := newWarmer(t)
	resolve, calls := stubResolve(w)
	w.cueRun = resolve

	_, err := w.EnsureWarm()
	require.NoError(t, err)

	// Simulate a schema content change: overwrite the recorded key with a stale one.
	require.NoError(t, os.WriteFile(w.keyStamp(), []byte("stale-key"), 0o644))

	rewarmed, err := w.EnsureWarm()
	require.NoError(t, err)
	assert.True(t, rewarmed, "stale content key should re-warm")
	assert.Equal(t, 2, *calls)

	// Stamp now holds the real content key.
	got, err := os.ReadFile(w.keyStamp())
	require.NoError(t, err)
	want, err := source.SchemaContentKey()
	require.NoError(t, err)
	assert.Equal(t, want, string(got))
}

func TestEnsureWarm_ClearsReadOnlyExtract(t *testing.T) {
	w := newWarmer(t)
	resolve, _ := stubResolve(w)
	w.cueRun = resolve

	// Pre-seed a read-only extract with no stamp → must be cleared and re-warmed.
	dir := w.extractDir()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.Chmod(dir, 0o555))

	_, err := w.EnsureWarm()
	require.NoError(t, err, "must cope with a read-only pre-existing extract")
}
