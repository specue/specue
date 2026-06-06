package modules_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"

	"github.com/specue/specue/internal/modules"
)

// TestSchemaImportTypeChecks proves an authored node file can import the schema
// module and unify against #Contract — the foundation of CUE-native authoring
// (type-checking + editor autocomplete come from this import).
func TestSchemaImportTypeChecks(t *testing.T) {
	schema, err := modules.NewSchemaModule()
	require.NoError(t, err)
	defer schema.Cleanup()

	closure := modules.Closure{Modules: []modules.ResolvedModule{schema.ResolvedModule}}

	overlay := map[string]load.Source{
		"/work/example/cue.mod/module.cue": load.FromString(`
module: "specue.test/example@v0"
language: version: "v0.16.0"
deps: "specue.io/schema@v0": v: "v0.0.1"
`),
		"/work/example/apply.cue": load.FromString(`
package example
import s "specue.io/schema@v0:spec"

example: s.#Container & {
	type: "Container", slug: "example", title: "Wallet", confidence: "CONFIRMED", kind: "service"
}
validateGraph: s.#Contract & {
	type:       "Contract"
	slug:       "validate-graph"
	title:      "Apply"
	confidence: "CONFIRMED"
	service:    example
	invariants: [{id: "post", text: "done"}]
}
`),
	}

	ctx := cuecontext.New()
	insts := load.Instances([]string{"."}, &load.Config{
		Dir:      "/work/example",
		Overlay:  overlay,
		Registry: closure.Registry(),
	})
	require.NoError(t, insts[0].Err)
	v := ctx.BuildInstance(insts[0])
	require.NoError(t, v.Err())
	require.NoError(t, v.Validate(cue.Concrete(true)))

	slug, _ := v.LookupPath(cue.ParsePath("validateGraph.slug")).String()
	assert.Equal(t, "validate-graph", slug, "node type-checked against imported #Contract")
}

// TestSchemaImportRejectsBadField proves the imported schema actually constrains:
// a bad confidence enum fails.
func TestSchemaImportRejectsBadField(t *testing.T) {
	schema, err := modules.NewSchemaModule()
	require.NoError(t, err)
	defer schema.Cleanup()
	closure := modules.Closure{Modules: []modules.ResolvedModule{schema.ResolvedModule}}

	overlay := map[string]load.Source{
		"/work/example/cue.mod/module.cue": load.FromString(`
module: "specue.test/example@v0"
language: version: "v0.16.0"
deps: "specue.io/schema@v0": v: "v0.0.1"
`),
		"/work/example/bad.cue": load.FromString(`
package example
import s "specue.io/schema@v0:spec"
svc: s.#Container & {type: "Container", slug: "w", title: "W", confidence: "CONFIRMED", kind: "service"}
bad: s.#Contract & {type: "Contract", slug: "x", title: "t", confidence: "BOGUS", service: svc, invariants: [{id: "post", text: "y"}]}
`),
	}
	ctx := cuecontext.New()
	insts := load.Instances([]string{"."}, &load.Config{Dir: "/work/example", Overlay: overlay, Registry: closure.Registry()})
	v := ctx.BuildInstance(insts[0])
	err = firstErr(insts[0].Err, v.Err(), v.Validate(cue.Concrete(true)))
	require.Error(t, err, "bad confidence must be rejected by the imported schema")
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
