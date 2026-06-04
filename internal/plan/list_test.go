package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/source"
)

// projector builds a Projector for the test's git binary.
func projector(t *testing.T, bin string) *plan.Projector {
	t.Helper()
	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	proj, err := plan.NewProjector(parser, plan.NewGit(bin))
	require.NoError(t, err)
	t.Cleanup(func() { _ = proj.Close() })
	return proj
}

func TestListEnumeratesPlansFromBranches(t *testing.T) {
	_, _, mgr, bin := projectRepo(t)
	proj := projector(t, bin)

	require.NoError(t, mgr.Register("gp-1", "First"))
	require.NoError(t, mgr.Register("gp-2", "Second"))

	plans, err := mgr.List(proj)
	require.NoError(t, err)
	require.Len(t, plans, 2, "both registered plans are discovered from their branches")
	// sorted by id
	assert.Equal(t, "gp-1", plans[0].ID)
	assert.Equal(t, "First", plans[0].Title)
	assert.Equal(t, "proposed", plans[0].Status)
	assert.Equal(t, "plan/gp-1", plans[0].Branch)
	assert.Equal(t, "gp-2", plans[1].ID)
}

func TestListEmptyWhenNoPlans(t *testing.T) {
	_, _, mgr, bin := projectRepo(t)
	proj := projector(t, bin)

	plans, err := mgr.List(proj)
	require.NoError(t, err)
	assert.Empty(t, plans)
}

func TestShowReadsOneRecord(t *testing.T) {
	_, _, mgr, bin := projectRepo(t)
	proj := projector(t, bin)
	require.NoError(t, mgr.Register("gp-1", "First"))

	info, err := mgr.Show(proj, "gp-1")
	require.NoError(t, err)
	assert.Equal(t, "gp-1", info.ID)
	assert.Equal(t, "First", info.Title)
	assert.Equal(t, "proposed", info.Status)
}

func TestShowMissingPlanErrors(t *testing.T) {
	_, _, mgr, bin := projectRepo(t)
	proj := projector(t, bin)

	_, err := mgr.Show(proj, "nope")
	require.Error(t, err, "a plan with no branch is an error, not an empty result")
}

func TestListDropReflectsRemoval(t *testing.T) {
	_, _, mgr, bin := projectRepo(t)
	proj := projector(t, bin)
	require.NoError(t, mgr.Register("gp-1", ""))
	require.NoError(t, mgr.Register("gp-2", ""))

	require.NoError(t, mgr.Drop("gp-1", true))
	plans, err := mgr.List(proj)
	require.NoError(t, err)
	require.Len(t, plans, 1, "a dropped plan no longer lists")
	assert.Equal(t, "gp-2", plans[0].ID)
}
