package query_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/query"
	"github.com/specue/specue/internal/source"
)

// graph builds a tiny landscape: a story with one FR, and two Contracts — one
// satisfies the FR and depends on the other (a contract dep), the second produces
// to a port. Enough to exercise nodes, dep_edges, infra_edges, satisfies, atoms,
// and FTS.
func graph(t *testing.T) (*compiler.ResolvedGraph, []compiler.Diagnostic) {
	t.Helper()
	story := model.PlacedNode{Module: "prod", Node: model.Node{
		Slug: "describe-node", Type: model.TypeNeed, Visibility: model.Public,
		Title: "Describe a node",
		Body: &model.Body{Need: &model.NeedBody{
			Atoms: []model.Atom{{Kind: model.KindFR, ID: "fr-01", Text: "the node is described idempotently"}},
		}},
	}}
	apply := model.PlacedNode{Module: "example", Node: model.Node{
		Slug: "validate-graph", Type: model.TypeContract, Visibility: model.Public,
		Title: "Validate the graph",
		Body: &model.Body{Contract: &model.ContractBody{
			Service: model.NodeID{Module: "example", Slug: "example"},
			Elements: []model.Element{
				{Kind: model.KindPost, Text: "verdict emitted", Deps: []model.Dep{
					{To: model.NodeID{Module: "example", Slug: "validate"}},
					{To: model.NodeID{Module: "topo", Slug: "report-channel"}, Role: model.RoleProduce},
				}},
				{Kind: model.KindInvariant, ID: "single-verdict", Text: "a run emits a single verdict",
					Satisfies: []model.AtomRef{{Need: model.NodeID{Module: "prod", Slug: "describe-node"}, Atom: "fr-01"}}},
			},
		}},
	}}
	validate := model.PlacedNode{Module: "example", Node: model.Node{
		Slug: "validate", Type: model.TypeContract, Visibility: model.Public, Title: "Validate input",
		Body: &model.Body{Contract: &model.ContractBody{Service: model.NodeID{Module: "example", Slug: "example"}}},
	}}

	g, diags := compiler.New().Compile(compiler.Input{Modules: []source.LoadedModule{
		{Manifest: source.Manifest{Path: "prod", Kind: source.KindDomain}, Nodes: []model.PlacedNode{story}},
		{Manifest: source.Manifest{Path: "example", Kind: source.KindService}, Nodes: []model.PlacedNode{apply, validate}},
	}})
	return g, diags
}

//specue:test:query-graph#runs-against-projection
func TestProjectionAndQuery(t *testing.T) {
	g, diags := graph(t)
	db, err := query.Build(g, diags)
	require.NoError(t, err)
	defer db.Close()

	t.Run("nodes projected", func(t *testing.T) {
		rows, err := db.Query(`SELECT id, type FROM nodes ORDER BY id`)
		require.NoError(t, err)
		require.Len(t, rows, 3)
		ids := []string{rows[0]["id"].(string), rows[1]["id"].(string), rows[2]["id"].(string)}
		assert.Contains(t, ids, "prod:describe-node")
		assert.Contains(t, ids, "example:validate-graph")
		assert.Contains(t, ids, "example:validate")
	})

	t.Run("dep edge", func(t *testing.T) {
		rows, err := db.Query(`SELECT to_id FROM dep_edges WHERE from_id = 'example:validate-graph'`)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "example:validate", rows[0]["to_id"])
	})

	t.Run("infra edge with role", func(t *testing.T) {
		rows, err := db.Query(`SELECT to_id, role FROM infra_edges WHERE role = 'produce'`)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "topo:report-channel", rows[0]["to_id"])
	})

	t.Run("satisfies an atom", func(t *testing.T) {
		rows, err := db.Query(`SELECT uc_id FROM satisfies WHERE need_id = 'prod:describe-node' AND atom = 'fr-01'`)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "example:validate-graph", rows[0]["uc_id"])
	})

	t.Run("recursive CTE blast radius", func(t *testing.T) {
		// who transitively depends on validate? validate-graph does.
		rows, err := db.Query(`WITH RECURSIVE up(id) AS (
			SELECT 'example:validate'
			UNION SELECT from_id FROM dep_edges JOIN up ON to_id = up.id
		) SELECT id FROM up WHERE id != 'example:validate'`)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "example:validate-graph", rows[0]["id"])
	})

	t.Run("full-text search", func(t *testing.T) {
		rows, err := db.Query(`SELECT id FROM nodes_fts WHERE nodes_fts MATCH 'idempotently'`)
		require.NoError(t, err)
		// the story atom mentions "idempotently"; FTS sees it.
		require.NotEmpty(t, rows)
	})
}

// TestOrphansProjected pins that an annotation binding nothing lands in the orphans
// table, so query sees the same dangling code the bindings view does (they would
// otherwise diverge: nodes empty, orphans not).
func TestOrphansProjected(t *testing.T) {
	// A code module annotating a slug that resolves to no node.
	g, diags := compiler.New().Compile(compiler.Input{
		Modules: []source.LoadedModule{{Manifest: source.Manifest{Path: "code", Kind: source.KindCode}}},
		Facts: []compiler.CodeFact{{
			Module: "code", Verb: compiler.VerbReq,
			Target: compiler.AnnotationTarget{Slug: "ghost"}, File: "x.go", Line: 7,
		}},
	})
	db, err := query.Build(g, diags)
	require.NoError(t, err)
	defer db.Close()

	rows, err := db.Query(`SELECT slug, reason, loc FROM orphans`)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "ghost", rows[0]["slug"])
	assert.Equal(t, "orphan-binding", rows[0]["reason"])
	assert.Equal(t, "x.go:7", rows[0]["loc"])
}

// TestPreJoinedViews covers query-graph#pre-joined-views: node_describe yields
// one row per element of a Contract, and fr_coverage yields one row per
// satisfying UC for a given story atom (with its status), so the common reads
// are one statement instead of a chain of joins.
//
//specue:test:query-graph#pre-joined-views
func TestPreJoinedViews(t *testing.T) {
	g, diags := graph(t)
	db, err := query.Build(g, diags)
	require.NoError(t, err)
	defer db.Close()

	// node_describe surfaces validate-graph's named invariant (single-verdict).
	rows, err := db.Query(`SELECT element, element_kind FROM node_describe
		WHERE id = 'example:validate-graph' AND element != ''`)
	require.NoError(t, err)
	require.NotEmpty(t, rows, "node_describe lists the UC's elements")
	found := false
	for _, r := range rows {
		if r["element"] == "single-verdict" {
			assert.Equal(t, "invariant", r["element_kind"])
			found = true
		}
	}
	assert.True(t, found, "the named invariant appears as an element row")

	// fr_coverage maps atom → UCs that satisfy it (with their status).
	rows, err = db.Query(`SELECT uc_id, uc_status FROM fr_coverage
		WHERE need_id = 'prod:describe-node' AND atom = 'fr-01'`)
	require.NoError(t, err)
	require.NotEmpty(t, rows, "an atom satisfied by a UC has a coverage row")
	assert.Equal(t, "example:validate-graph", rows[0]["uc_id"])
}

//specue:test:query-graph#cannot-mutate
func TestQueryIsReadOnly(t *testing.T) {
	g, diags := graph(t)
	db, err := query.Build(g, diags)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Query(`DELETE FROM nodes`)
	assert.Error(t, err, "writes are rejected on a read-only projection")
}
