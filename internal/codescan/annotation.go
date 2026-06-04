// Package codescan gathers code facts — the //specue: annotations bound in
// source — for the compiler. It is the second fact source (the first is source's
// parser facts); the two feed the compiler, which collides them into statuses. It
// reads from fs.FS (mockable) and produces compiler.CodeFact: it resolves the
// annotation's lexical ref (alias/slug/element/rev) but does not resolve it
// against the graph — that is the compiler's job.
package codescan

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

// annRe matches a namespaced, verb-tagged specue annotation in a // (code) or
// # (manifest) comment. The ref is captured raw; parseTarget splits it.
var annRe = regexp.MustCompile(`(?://|#)\s*specue:([a-z]+):([A-Za-z0-9_/.#:@-]+)`)

// rawAnnotation is one matched annotation before its ref is parsed.
type rawAnnotation struct {
	verb   string
	ref    string
	file   model.FilePath
	line   int
	isTest bool
}

// parse turns a raw annotation into a CodeFact target (verb + parsed ref). The
// caller stamps the module/candidates. An unknown verb yields ok=false.
func (r rawAnnotation) parse() (compiler.AnnotationVerb, compiler.AnnotationTarget, bool) {
	verb := compiler.AnnotationVerb(r.verb)
	if !knownVerb(verb) {
		return "", compiler.AnnotationTarget{}, false
	}
	return verb, parseTarget(r.ref), true
}

// parseTarget splits a ref into alias / slug / element / rev. Lexical only:
//
//	example:validate-graph#idempotent@3
//	→ alias=example slug=validate-graph element=single-verdict rev=3
func parseTarget(ref string) compiler.AnnotationTarget {
	ref, rev := splitRev(ref)
	var alias string
	if a, rest, ok := strings.Cut(ref, ":"); ok {
		alias, ref = a, rest
	}
	slug, elem, _ := strings.Cut(ref, "#")
	return compiler.AnnotationTarget{
		Alias:   model.Alias(alias),
		Slug:    model.Slug(slug),
		Element: model.ElementID(elem),
		Rev:     rev,
	}
}

// splitRev separates an optional @N revision pin (no pin → rev 0).
func splitRev(ref string) (string, int) {
	i := strings.LastIndexByte(ref, '@')
	if i < 0 {
		return ref, 0
	}
	n, err := strconv.Atoi(ref[i+1:])
	if err != nil {
		return ref, 0
	}
	return ref[:i], n
}

func knownVerb(v compiler.AnnotationVerb) bool {
	switch v {
	case compiler.VerbReq, compiler.VerbTest,
		compiler.VerbProduces, compiler.VerbConsumes, compiler.VerbPublishes,
		compiler.VerbSubscribes, compiler.VerbServes, compiler.VerbCalls,
		compiler.VerbReads, compiler.VerbWrites, compiler.VerbGrants:
		return true
	}
	return false
}
