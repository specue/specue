package scc

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// normalize sorts each component and the set of components for stable comparison,
// since Find's output order is not guaranteed.
func normalize(sccs [][]string) [][]string {
	for _, c := range sccs {
		sort.Strings(c)
	}
	sort.Slice(sccs, func(i, j int) bool { return sccs[i][0] < sccs[j][0] })
	return sccs
}

func TestFindCycle(t *testing.T) {
	// a→b→c→a is one component; d→b is its own.
	got := Find(map[string][]string{
		"a": {"b"}, "b": {"c"}, "c": {"a"}, "d": {"b"},
	})
	assert.ElementsMatch(t, [][]string{{"a", "b", "c"}, {"d"}}, normalize(got))
}

func TestFindAcyclic(t *testing.T) {
	// No cycle → every node is its own singleton component.
	got := Find(map[string][]string{"a": {"b"}, "b": {"c"}, "c": nil})
	require.Len(t, got, 3)
	for _, c := range got {
		assert.Len(t, c, 1)
	}
}

func TestFindSelfEdge(t *testing.T) {
	got := Find(map[string][]string{"a": {"a"}})
	assert.Equal(t, [][]string{{"a"}}, got)
}

func TestFindEdgeToAbsentNode(t *testing.T) {
	// Edge to a key not in adj is ignored, doesn't panic.
	got := Find(map[string][]string{"a": {"ghost"}})
	assert.Equal(t, [][]string{{"a"}}, got)
}

// TestFindThreeComponents is the worked example: three cycles linked by
// one-way edges. SCCs: {0,1}, {2,3}, {4,5,6}.
//
//	6→4, 6→5 · 4→1, 4→3, 4→5 · 5→6, 5→3 · 3→2 · 2→3, 2→1, 2→0 · 1→0 · 0→1
func TestFindThreeComponents(t *testing.T) {
	got := Find(map[string][]string{
		"6": {"4", "5"},
		"4": {"1", "3", "5"},
		"5": {"6", "3"},
		"3": {"2"},
		"2": {"3", "1", "0"},
		"1": {"0"},
		"0": {"1"},
	})
	assert.ElementsMatch(t, [][]string{
		{"0", "1"},
		{"2", "3"},
		{"4", "5", "6"},
	}, normalize(got))
}

// FuzzFind asserts the SCC invariants hold on random graphs regardless of the
// (non-deterministic) traversal order: every node appears in exactly one
// component, and two nodes share a component iff they are mutually reachable.
func FuzzFind(f *testing.F) {
	f.Add(int64(1), 6, 10)
	f.Add(int64(42), 12, 30)
	f.Fuzz(func(t *testing.T, seed int64, n, edges int) {
		if n <= 0 || n > 50 || edges < 0 || edges > 200 {
			t.Skip()
		}
		adj := randomGraph(seed, n, edges)
		sccs := Find(adj)

		assertPartition(t, adj, sccs)
		assertMutualReachability(t, adj, sccs)
	})
}

func randomGraph(seed int64, n, edges int) map[string][]string {
	rng := rand.New(rand.NewSource(seed))
	adj := map[string][]string{}
	for i := 0; i < n; i++ {
		adj[fmt.Sprint(i)] = nil
	}
	for i := 0; i < edges; i++ {
		from := fmt.Sprint(rng.Intn(n))
		to := fmt.Sprint(rng.Intn(n))
		adj[from] = append(adj[from], to)
	}
	return adj
}

// assertPartition: every node is in exactly one component.
func assertPartition(t *testing.T, adj map[string][]string, sccs [][]string) {
	t.Helper()
	seen := map[string]int{}
	for _, c := range sccs {
		for _, node := range c {
			seen[node]++
		}
	}
	for node := range adj {
		assert.Equal(t, 1, seen[node], "node %s in exactly one component", node)
	}
}

// assertMutualReachability: two nodes share a component iff each reaches the other.
func assertMutualReachability(t *testing.T, adj map[string][]string, sccs [][]string) {
	t.Helper()
	comp := map[string]int{}
	for i, c := range sccs {
		for _, node := range c {
			comp[node] = i
		}
	}
	nodes := make([]string, 0, len(adj))
	for k := range adj {
		nodes = append(nodes, k)
	}
	for _, a := range nodes {
		for _, b := range nodes {
			mutual := reaches(adj, a, b) && reaches(adj, b, a)
			same := comp[a] == comp[b]
			assert.Equalf(t, mutual, same, "nodes %s,%s: mutual=%v same-component=%v", a, b, mutual, same)
		}
	}
}

// reaches is a plain DFS reachability check (independent of Tarjan).
func reaches(adj map[string][]string, from, to string) bool {
	if from == to {
		return true
	}
	seen := map[string]bool{}
	var dfs func(string) bool
	dfs = func(v string) bool {
		for _, w := range adj[v] {
			if _, known := adj[w]; !known {
				continue
			}
			if w == to {
				return true
			}
			if !seen[w] {
				seen[w] = true
				if dfs(w) {
					return true
				}
			}
		}
		return false
	}
	return dfs(from)
}
