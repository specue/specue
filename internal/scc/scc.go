// Package scc finds strongly connected components of a directed graph (Tarjan's
// algorithm), generic over any comparable node key. It is isolated from the
// compiler so the algorithm is testable on its own and reusable; callers pass an
// adjacency map and get back the SCCs.
package scc

// Find returns the strongly connected components of the directed graph given by
// adj (node → its out-neighbours). Each returned slice is one component; a
// component of one node with no self-edge is just that node. Edges to keys absent
// from adj are ignored.
//
// The order of components and of nodes within a component follows Go's map
// iteration and is therefore not stable across runs — Find guarantees a correct
// partition, not a stable order. Callers that need reproducible output (reports,
// golden) sort the result themselves.
func Find[N comparable](adj map[N][]N) [][]N {
	t := &tarjan[N]{
		adj:     adj,
		index:   map[N]int{},
		low:     map[N]int{},
		onStack: map[N]bool{},
	}
	for v := range adj {
		if _, seen := t.index[v]; !seen {
			t.connect(v)
		}
	}
	return t.sccs
}

type tarjan[N comparable] struct {
	adj     map[N][]N
	index   map[N]int
	low     map[N]int
	onStack map[N]bool
	stack   []N
	counter int
	sccs    [][]N
}

func (t *tarjan[N]) connect(v N) {
	t.index[v] = t.counter
	t.low[v] = t.counter
	t.counter++
	t.stack = append(t.stack, v)
	t.onStack[v] = true

	for _, w := range t.adj[v] {
		if _, known := t.adj[w]; !known {
			continue // edge to a node absent from the graph — ignore
		}
		if _, seen := t.index[w]; !seen {
			t.connect(w)
			t.low[v] = min(t.low[v], t.low[w])
		} else if t.onStack[w] {
			t.low[v] = min(t.low[v], t.index[w])
		}
	}

	if t.low[v] == t.index[v] {
		t.sccs = append(t.sccs, t.popComponent(v))
	}
}

// popComponent pops the stack down to v, yielding one component.
func (t *tarjan[N]) popComponent(v N) []N {
	var comp []N
	for {
		w := t.stack[len(t.stack)-1]
		t.stack = t.stack[:len(t.stack)-1]
		t.onStack[w] = false
		comp = append(comp, w)
		if w == v {
			return comp
		}
	}
}
