package codescan

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

const implGo = `package example

//specue:req:validate-graph
func Apply() {}

//specue:produces:report-channel
func publish() {}

//specue:req:validate-graph#single-verdict@3
func reApply() {}
`

const testGo = `package example

//specue:test:validate-graph
func TestApply(t *testing.T) {}
`

func TestScanGoSource(t *testing.T) {
	fsys := fstest.MapFS{
		"apply.go":      {Data: []byte(implGo)},
		"apply_test.go": {Data: []byte(testGo)},
	}
	facts, err := NewScanner().Scan([]ScanTarget{{
		FS:     fsys,
		Root:   ".",
		Module: "example",
	}})
	require.NoError(t, err)
	require.Len(t, facts, 4)

	byVerb := map[compiler.AnnotationVerb][]compiler.CodeFact{}
	for _, f := range facts {
		assert.Equal(t, model.ModulePath("example"), f.Module)
		byVerb[f.Verb] = append(byVerb[f.Verb], f)
	}

	// Two req (one whole-contract, one element-scoped with rev), one produces, one covers.
	require.Len(t, byVerb[compiler.VerbReq], 2)
	require.Len(t, byVerb[compiler.VerbProduces], 1)
	require.Len(t, byVerb[compiler.VerbTest], 1)

	// The covers fact is a test, the req in apply.go is not.
	cov := byVerb[compiler.VerbTest][0]
	assert.True(t, cov.IsTest, "covers in _test.go is a test context")
	assert.Equal(t, model.Slug("validate-graph"), cov.Target.Slug)

	// The element-scoped req parsed element + rev.
	var scoped compiler.CodeFact
	for _, f := range byVerb[compiler.VerbReq] {
		if f.Target.Element != "" {
			scoped = f
		}
	}
	assert.Equal(t, model.ElementID("single-verdict"), scoped.Target.Element)
	assert.Equal(t, 3, scoped.Target.Rev)
}

func TestScanInfraVerbRoleMapping(t *testing.T) {
	fsys := fstest.MapFS{"x.go": {Data: []byte("//specue:consumes:example:validate-graph\n")}}
	facts, err := NewScanner().Scan([]ScanTarget{{
		FS: fsys, Root: ".", Module: "consumer",
	}})
	require.NoError(t, err)
	require.Len(t, facts, 1)

	f := facts[0]
	assert.Equal(t, compiler.VerbConsumes, f.Verb)
	assert.Equal(t, model.RoleConsume, f.Verb.Role(), "consumes maps to the consume role")
	assert.Equal(t, model.Alias("example"), f.Target.Alias, "qualified ref split")
	assert.Equal(t, model.Slug("validate-graph"), f.Target.Slug)
}

func TestScanIgnoresUnknownVerb(t *testing.T) {
	fsys := fstest.MapFS{"x.go": {Data: []byte("//specue:bogus:validate-graph\n//notspecue:req:x\n")}}
	facts, err := NewScanner().Scan([]ScanTarget{{
		FS: fsys, Root: ".", Module: "m",
	}})
	require.NoError(t, err)
	assert.Empty(t, facts, "unknown verb and non-specue comments are ignored")
}

// A doc comment that mentions an annotation as syntax (the marker sits inside an
// already-open comment) must not be scanned as a real binding — otherwise the
// prose orphans against a slug that does not exist.
func TestScanIgnoresAnnotationQuotedInDocComment(t *testing.T) {
	src := "// A CODE annotation still does:\n" +
		"// `//specue:req:alias:slug` is a lexical string, not a real binding.\n" +
		"//specue:req:validate-graph\n" + // the real one, on its own line
		"x := 1 //specue:req:trailing-ok\n" // trailing after code is real too
	fsys := fstest.MapFS{"x.go": {Data: []byte(src)}}
	facts, err := NewScanner().Scan([]ScanTarget{{FS: fsys, Root: ".", Module: "m"}})
	require.NoError(t, err)

	var slugs []string
	for _, f := range facts {
		slugs = append(slugs, string(f.Target.Slug))
	}
	assert.ElementsMatch(t, []string{"validate-graph", "trailing-ok"}, slugs,
		"only the standalone and trailing annotations bind; the doc-comment mention is prose")
}

// A // inside a string literal on a code line must not be mistaken for the start
// of a comment: the trailing annotation after it is still real. This is the
// counterpart to the doc-comment case — there the line IS a comment; here it is
// code with an incidental // in a string.
func TestScanBindsTrailingAnnotationAfterURLInString(t *testing.T) {
	src := "url := \"http://example.com\" //specue:req:fetch-rates\n"
	fsys := fstest.MapFS{"x.go": {Data: []byte(src)}}
	facts, err := NewScanner().Scan([]ScanTarget{{FS: fsys, Root: ".", Module: "m"}})
	require.NoError(t, err)
	require.Len(t, facts, 1)
	assert.Equal(t, model.Slug("fetch-rates"), facts[0].Target.Slug)
}
