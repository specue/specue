package compiler_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/codescan"
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/modules"
	"github.com/specue/specue/internal/source"
	"github.com/specue/specue/internal/specload"
)

const walletModule = model.ModulePath("specue.test/example@v0")

// TestIntegration wires the layers end to end on a real testdata module: the
// module-manager resolves the require closure, specload loads the CUE spec as one
// resolved value tree, codescan gathers the //specue: facts from the code tree,
// and the compiler collides them into statuses. This proves the layers compose
// (each is unit-tested in isolation elsewhere).
func TestIntegration(t *testing.T) {
	specDir, err := filepath.Abs("testdata/example/spec")
	require.NoError(t, err)
	codeFS := os.DirFS("testdata/example/code")

	schema, err := modules.NewSchemaModule()
	require.NoError(t, err)
	defer schema.Cleanup()

	resolver := modules.NewResolver(mustParser(t), modules.NewReplaceLocator())
	closure, err := resolver.Resolve([]modules.RootModule{{Path: walletModule, Dir: specDir}})
	require.NoError(t, err)
	closure.Modules = append(closure.Modules, schema.ResolvedModule)

	mods, err := specload.New().Load(closure)
	require.NoError(t, err)

	facts, err := codescan.NewScanner().Scan([]codescan.ScanTarget{{
		FS:     codeFS,
		Root:   ".",
		Module: walletModule,
	}})
	require.NoError(t, err)

	g, diags := compiler.New().Compile(compiler.Input{Modules: mods, Facts: facts})

	assertNoGates(t, diags)

	// validate-graph: req + covering test → proven; describe-node: no code → asserted.
	assert.Equal(t, compiler.StatusProven, statusOf(t, g, "validate-graph"))
	assert.Equal(t, compiler.StatusAsserted, statusOf(t, g, "describe-node"))

	// The Port's topology was aggregated from validate-graph's produce edge.
	port := nodeOf(t, g, "report-channel")
	assert.Equal(t, []model.NodeID{{Module: walletModule, Slug: "validate-graph"}},
		port.Topology.ProducedBy)
}

func mustParser(t *testing.T) source.Parser {
	t.Helper()
	p, err := source.NewCUEParser()
	require.NoError(t, err)
	return p
}

func nodeOf(t *testing.T, g *compiler.ResolvedGraph, slug model.Slug) *compiler.ResolvedNode {
	t.Helper()
	n, ok := g.Node(model.NodeID{Module: walletModule, Slug: slug})
	require.True(t, ok, "node %s present", slug)
	return n
}

func statusOf(t *testing.T, g *compiler.ResolvedGraph, slug model.Slug) compiler.ResolvedNodeStatus {
	return nodeOf(t, g, slug).Status
}

func assertNoGates(t *testing.T, diags []compiler.Diagnostic) {
	t.Helper()
	for _, d := range diags {
		assert.NotEqualf(t, compiler.Gate, d.Severity(), "unexpected gate: %s %s", d.Code, d.Message)
	}
}
