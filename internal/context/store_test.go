package context_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/context"
)

//specue:test:create-context#name-is-unique
func TestCreateIsSortedAndRejectsDuplicate(t *testing.T) {
	var s context.Store
	require.NoError(t, s.Create("example"))
	require.NoError(t, s.Create("consumer"))

	require.Len(t, s.Contexts, 2)
	assert.Equal(t, "consumer", s.Contexts[0].Name, "sorted by name")

	var dup *context.DuplicateContextError
	assert.ErrorAs(t, s.Create("example"), &dup, "create never clobbers an existing context")
}

func TestAddRemoveModule(t *testing.T) {
	var s context.Store
	require.NoError(t, s.Create("example"))

	e, ok := s.EntryPtr("example")
	require.True(t, ok)
	assert.True(t, e.AddModule("x.test/gov@v0", "/abs/gov"))
	assert.True(t, e.AddModule("x.test/svc@v0", "/abs/svc"))
	assert.False(t, e.AddModule("x.test/gov@v0", "/abs/gov2"), "re-add updates, reports not-new")

	require.Len(t, e.Modules, 2)
	assert.Equal(t, "x.test/gov@v0", e.Modules[0].Path, "modules sorted by path")
	assert.Equal(t, "/abs/gov2", e.Modules[0].Dir, "AddModule updates the dir")

	assert.True(t, e.RemoveModule("x.test/gov@v0"))
	assert.Len(t, e.Modules, 1)
	assert.False(t, e.RemoveModule("x.test/gov@v0"), "removing a gone module reports false")
}

//specue:test:use-context#context-must-exist
func TestUseRequiresExisting(t *testing.T) {
	var s context.Store
	require.NoError(t, s.Create("example"))

	require.NoError(t, s.Use("example"))
	e, ok := s.ActiveEntry()
	require.True(t, ok)
	assert.Equal(t, "example", e.Name)

	var unknown *context.UnknownContextError
	assert.ErrorAs(t, s.Use("nope"), &unknown, "Use of a missing context errors; active never dangles")
}

func TestRemoveClearsActive(t *testing.T) {
	var s context.Store
	require.NoError(t, s.Create("example"))
	require.NoError(t, s.Use("example"))

	assert.True(t, s.Remove("example"))
	_, ok := s.ActiveEntry()
	assert.False(t, ok, "removing the active context clears the active selection")
	assert.False(t, s.Remove("example"), "removing a gone context reports false")
}

//specue:test:create-context#survives-across-invocations
func TestFileRepositoryRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".specue", "contexts.json")
	repo := context.NewFileRepository(path)

	// first load: empty (nothing persisted)
	s, err := repo.Load()
	require.NoError(t, err)
	assert.Empty(t, s.Contexts)

	require.NoError(t, s.Create("example"))
	e, _ := s.EntryPtr("example")
	e.PlanBase = "main"
	e.AddModule("x.test/gov@v0", "/abs/gov")
	require.NoError(t, s.Use("example"))
	require.NoError(t, repo.Save(s)) // creates the dir

	got, err := repo.Load()
	require.NoError(t, err)
	assert.Equal(t, "example", got.Active)
	require.Len(t, got.Contexts, 1)
	assert.Equal(t, "main", got.Contexts[0].PlanBase)
	require.Len(t, got.Contexts[0].Modules, 1)
	assert.Equal(t, "/abs/gov", got.Contexts[0].Modules[0].Dir)
}
