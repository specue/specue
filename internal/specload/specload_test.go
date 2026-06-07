package specload_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/modules"
	"github.com/specue/specue/internal/source"
	"github.com/specue/specue/internal/specload"
)

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

// TestLoadResolvesCrossModuleRef is the core promise: a reference authored in
// module consumer pointing at a node in module example is loaded ALREADY RESOLVED to
// example's full address — CUE resolved it, and specload recovered the target
// module from where the node lives, not from any string the author wrote.
func TestLoadResolvesCrossModuleRef(t *testing.T) {
	base := t.TempDir()
	exampleDir := filepath.Join(base, "example")
	consumerDir := filepath.Join(base, "consumer")

	// schema module on disk.
	schema, err := modules.NewSchemaModule()
	require.NoError(t, err)
	defer schema.Cleanup()

	// example: owns validate-graph.
	write(t, exampleDir, "cue.mod/module.cue", "module: \"specue.test/example@v0\"\nlanguage: version: \"v0.16.0\"\ndeps: \"specue.io/schema@v0\": v: \"v0.0.1\"\n")
	write(t, exampleDir, source.ManifestFile, "module: \"specue.test/example@v0\"\nversion: \"v0.1.0\"\nkind: \"service\"\n")
	write(t, exampleDir, "nodes.cue", `package example
import s "specue.io/schema@v0:spec"
svc: s.#Container & {type:"Container", slug:"example", title:"W", confidence:"CONFIRMED", kind:"service"}
validateGraph: s.#Contract & {type:"Contract", slug:"validate-graph", title:"Apply", confidence:"CONFIRMED", service:svc, invariants:[{id:"post", text:"done"}]}
`)

	// consumer: depends on example, references validate-graph cue-natively.
	write(t, consumerDir, "cue.mod/module.cue", "module: \"specue.test/consumer@v0\"\nlanguage: version: \"v0.16.0\"\ndeps: \"specue.io/schema@v0\": v: \"v0.0.1\"\ndeps: \"specue.test/example@v0\": v: \"v0.1.0\"\n")
	write(t, consumerDir, source.ManifestFile, "module: \"specue.test/consumer@v0\"\nversion: \"v0.1.0\"\nkind: \"service\"\nrequire: [{module: \"specue.test/example@v0\", version: \"v0.1.0\", replace: \"../example\"}]\n")
	write(t, consumerDir, "nodes.cue", `package consumer
import (
	s "specue.io/schema@v0:spec"
	w "specue.test/example@v0:example"
)
svc: s.#Container & {type:"Container", slug:"consumer", title:"C", confidence:"CONFIRMED", kind:"service"}
grant: s.#Contract & {
	type:"Contract", slug:"grant", title:"Grant", confidence:"CONFIRMED", service:svc
	invariants:[{id:"post", text:"granted", depends_on:[{to: w.validateGraph}]}]
}
`)

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	resolver := modules.NewResolver(parser, modules.NewReplaceLocator())
	closure, err := resolver.Resolve([]modules.RootModule{{Path: "specue.test/consumer@v0", Dir: consumerDir}})
	require.NoError(t, err)
	closure.Modules = append(closure.Modules, schema.ResolvedModule)

	mods, err := specload.New().Load(closure)
	require.NoError(t, err)

	// consumer is the only root → the only loaded module.
	require.Len(t, mods, 1)
	var grant model.PlacedNode
	for _, n := range mods[0].Nodes {
		if n.Node.Slug == "grant" {
			grant = n
		}
	}
	require.Equal(t, model.Slug("grant"), grant.Node.Slug, "grant loaded")

	dep := grant.Node.Body.Contract.Elements[0].Deps[0]
	assert.Equal(t, model.NodeID{Module: "specue.test/example@v0", Slug: "validate-graph"}, dep.To,
		"cross-module ref resolved to example's full address")
}

// TestLoadMultiPackageModule proves a module may organize nodes across sub-folders
// (each a CUE sub-package) and they all load as one module, with a cross-folder
// reference resolved CUE-natively: a Contract in navigation/ points at the service
// Container in the root package via a module-path import.
//specue:test:build-graph#multi-folder-modules
func TestLoadMultiPackageModule(t *testing.T) {
	base := t.TempDir()
	svcDir := filepath.Join(base, "svc")

	schema, err := modules.NewSchemaModule()
	require.NoError(t, err)
	defer schema.Cleanup()

	write(t, svcDir, "cue.mod/module.cue", "module: \"x.test/svc@v0\"\nlanguage: version: \"v0.16.0\"\ndeps: \"specue.io/schema@v0\": v: \"v0.0.1\"\n")
	write(t, svcDir, source.ManifestFile, "module: \"x.test/svc@v0\"\nversion: \"v0.1.0\"\nkind: \"service\"\n")
	// Root package: the service Container every Contract points at.
	write(t, svcDir, "root.cue", `package svc
import s "specue.io/schema@v0:spec"
svc: s.#Container & {type:"Container", slug:"specue", title:"the tool", confidence:"CONFIRMED", kind:"service"}
`)
	// Sub-folder package: a Contract that references the root package's service
	// node by importing the module path — the cross-folder, CUE-native edge.
	write(t, svcDir, "navigation/nav.cue", `package navigation
import (
	s "specue.io/schema@v0:spec"
	root "x.test/svc@v0:svc"
)
listResources: s.#Contract & {type:"Contract", slug:"list-resources", title:"List", confidence:"CONFIRMED", service: root.svc, invariants:[{id:"post", text:"listed"}]}
`)

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	resolver := modules.NewResolver(parser, modules.NewReplaceLocator())
	closure, err := resolver.Resolve([]modules.RootModule{{Path: "x.test/svc@v0", Dir: svcDir}})
	require.NoError(t, err)
	closure.Modules = append(closure.Modules, schema.ResolvedModule)

	mods, err := specload.New().Load(closure)
	require.NoError(t, err)
	require.Len(t, mods, 1)

	var slugs []string
	for _, n := range mods[0].Nodes {
		slugs = append(slugs, string(n.Node.Slug))
	}
	assert.ElementsMatch(t, []string{"specue", "list-resources"}, slugs,
		"nodes from both the root package and the navigation/ sub-package load as one module")
}

// TestLoadNodelessModuleIsEmpty proves a module with only a manifest and no node
// files (e.g. a fresh governance module before any Plan/ADR) loads as zero nodes,
// not a build error — CUE's "no package files" is a legitimate empty module.
func TestLoadNodelessModuleIsEmpty(t *testing.T) {
	base := t.TempDir()
	govDir := filepath.Join(base, "governance")
	write(t, govDir, "cue.mod/module.cue", "module: \"x.test/governance@v0\"\nlanguage: version: \"v0.16.0\"\ndeps: \"specue.io/schema@v0\": v: \"v0.0.1\"\n")
	write(t, govDir, source.ManifestFile, "module: \"x.test/governance@v0\"\nversion: \"v0.0.1\"\nkind: \"governance\"\n")

	schema, err := modules.NewSchemaModule()
	require.NoError(t, err)
	defer schema.Cleanup()

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	resolver := modules.NewResolver(parser, modules.NewReplaceLocator())
	closure, err := resolver.Resolve([]modules.RootModule{{Path: "x.test/governance@v0", Dir: govDir}})
	require.NoError(t, err)
	closure.Modules = append(closure.Modules, schema.ResolvedModule)

	mods, err := specload.New().Load(closure)
	require.NoError(t, err, "a node-less module is not a load error")
	require.Len(t, mods, 1)
	assert.Empty(t, mods[0].Nodes, "no node files → zero nodes")
}

// TestLoadInvariantKindWhenAndElementDep proves the invariant loader (ADR-14):
//   - `kind: "rejects" | "returns"` maps to Element.Kind,
//   - `when` maps to Element.When,
//   - a guarded invariant (When set) has branch deps (Dep.Branch), while an
//     unguarded one does not — the rule that replaced "variations section ⇒
//     branch",
//   - G2: a dep whose `to` targets an invariant (not a whole node) resolves to
//     the owning Contract, never an empty NodeRef.
func TestLoadInvariantKindWhenAndElementDep(t *testing.T) {
	base := t.TempDir()
	svcDir := filepath.Join(base, "svc")

	schema, err := modules.NewSchemaModule()
	require.NoError(t, err)
	defer schema.Cleanup()

	write(t, svcDir, "cue.mod/module.cue", "module: \"x.test/svc@v0\"\nlanguage: version: \"v0.16.0\"\ndeps: \"specue.io/schema@v0\": v: \"v0.0.1\"\n")
	write(t, svcDir, source.ManifestFile, "module: \"x.test/svc@v0\"\nversion: \"v0.1.0\"\nkind: \"service\"\n")
	write(t, svcDir, "nodes.cue", `package svc
import s "specue.io/schema@v0:spec"
svc: s.#Container & {type:"Container", slug:"svc", title:"S", confidence:"CONFIRMED", kind:"service"}
dep: s.#Contract & {type:"Contract", slug:"dep", title:"Dep", confidence:"CONFIRMED", service:svc, invariants:[{id:"g", text:"guarantee"}]}
uc: s.#Contract & {
	type:"Contract", slug:"uc", title:"UC", confidence:"CONFIRMED", service:svc
	invariants:[
		{id:"plain-inv", text:"always"},
		{id:"ret", kind:"returns", text:"result"},
		// guarded rejects whose dep targets dep's invariant (element-grained, G2):
		{id:"rej", kind:"rejects", when:"bad input", text:"refused", depends_on:[{to: dep.invariants[0]}]},
	]
}
`)

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	resolver := modules.NewResolver(parser, modules.NewReplaceLocator())
	closure, err := resolver.Resolve([]modules.RootModule{{Path: "x.test/svc@v0", Dir: svcDir}})
	require.NoError(t, err)
	closure.Modules = append(closure.Modules, schema.ResolvedModule)

	mods, err := specload.New().Load(closure)
	require.NoError(t, err)
	require.Len(t, mods, 1)

	var uc model.PlacedNode
	for _, n := range mods[0].Nodes {
		if n.Node.Slug == "uc" {
			uc = n
		}
	}
	els := uc.Node.Body.Contract.Elements
	require.Len(t, els, 3)

	// element 0: plain, unguarded.
	assert.Equal(t, model.KindPlain, els[0].Kind)
	assert.Empty(t, els[0].When)

	// element 1: returns.
	assert.Equal(t, model.KindReturns, els[1].Kind)

	// element 2: guarded rejects with an element-target dep.
	rej := els[2]
	assert.Equal(t, model.KindRejects, rej.Kind)
	assert.Equal(t, "bad input", rej.When)
	require.Len(t, rej.Deps, 1)
	assert.True(t, rej.Deps[0].Branch, "a guarded (When) invariant's deps are branch deps")
	assert.Equal(t, model.NodeID{Module: "x.test/svc@v0", Slug: "dep"}, rej.Deps[0].To,
		"a dep targeting an invariant resolves to the owning Contract (G2)")
}

// TestLoadRejectsMisTypedEdge proves the schema gate (ADR-15): CUE type-checks
// every edge at load, so a `service` aimed at a node that is not a Container
// makes the build fail at resolution — a mis-aimed relationship never reaches the
// graph. (The same gate rejects a depends_on whose `to` does not match its role;
// the service case is the simplest to exercise.)
//specue:test:build-graph#edges-are-type-checked
func TestLoadRejectsMisTypedEdge(t *testing.T) {
	base := t.TempDir()
	svcDir := filepath.Join(base, "svc")

	schema, err := modules.NewSchemaModule()
	require.NoError(t, err)
	defer schema.Cleanup()

	write(t, svcDir, "cue.mod/module.cue", "module: \"x.test/svc@v0\"\nlanguage: version: \"v0.16.0\"\ndeps: \"specue.io/schema@v0\": v: \"v0.0.1\"\n")
	write(t, svcDir, source.ManifestFile, "module: \"x.test/svc@v0\"\nversion: \"v0.1.0\"\nkind: \"service\"\n")
	// adr is an #ADR, not a #Container — service must be a #Container, so this
	// edge is ill-typed and CUE must refuse it at resolution.
	write(t, svcDir, "nodes.cue", `package svc
import s "specue.io/schema@v0:spec"
adr: s.#ADR & {type:"ADR", slug:"adr-x", title:"A", confidence:"CONFIRMED", status:"accepted"}
uc: s.#Contract & {type:"Contract", slug:"uc", title:"UC", confidence:"CONFIRMED", service: adr, invariants:[{id:"g", text:"guarantee"}]}
`)

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	resolver := modules.NewResolver(parser, modules.NewReplaceLocator())
	closure, err := resolver.Resolve([]modules.RootModule{{Path: "x.test/svc@v0", Dir: svcDir}})
	require.NoError(t, err)
	closure.Modules = append(closure.Modules, schema.ResolvedModule)

	_, err = specload.New().Load(closure)
	require.Error(t, err, "a service aimed at a non-Container is refused at load")
}
