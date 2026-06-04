package compiler

import "github.com/specue/specue/internal/model"

// CodeFact is one resolved code annotation — a binding the scanner found in
// source. The compiler owns this type (it is the compiler's input contract); the
// scanner depends on the compiler to produce it, not the reverse, keeping the
// compiler a pure downstream transform.
//
// The annotation's ref is already parsed (alias/slug/element/rev split out by the
// scanner — a lexical concern), so the compiler only resolves it against the
// graph and collides it onto the node.
type CodeFact struct {
	// Module is the module whose source carried the annotation; a bare slug
	// resolves here first.
	Module model.ModulePath
	// Candidates are sibling modules the carrying source also implements (a
	// deploy repo serving several modules); a bare slug falls back to whichever
	// owns it. Empty in the common single-module case.
	Candidates []model.ModulePath

	Verb   AnnotationVerb
	Target AnnotationTarget
	File   model.FilePath
	Line   int
	IsTest bool
}

// AnnotationTarget is the parsed target of an annotation: the node it binds,
// optionally scoped to an element, optionally pinned to a rev.
type AnnotationTarget struct {
	Alias   model.Alias
	Slug    model.Slug
	Element model.ElementID // "" = whole-contract
	Rev     int             // 0 = unpinned
}

// AnnotationVerb is the closed set of code annotation verbs. req/test bind
// implementation and proof; the rest are infra verbs that anchor an infra edge's
// role.
type AnnotationVerb string

const (
	VerbReq  AnnotationVerb = "req"  // implementation of a contract
	VerbTest AnnotationVerb = "test" // a test proving a contract

	VerbProduces   AnnotationVerb = "produces"
	VerbConsumes   AnnotationVerb = "consumes"
	VerbPublishes  AnnotationVerb = "publishes"
	VerbSubscribes AnnotationVerb = "subscribes"
	VerbServes     AnnotationVerb = "serves"
	VerbCalls      AnnotationVerb = "calls"
	VerbReads      AnnotationVerb = "reads"
	VerbWrites     AnnotationVerb = "writes"
	VerbGrants     AnnotationVerb = "grants"
)

// Role maps an infra verb to the dep Role it anchors; empty for req/test (not
// infra verbs) or an unknown verb.
func (v AnnotationVerb) Role() model.Role {
	switch v {
	case VerbProduces:
		return model.RoleProduce
	case VerbConsumes:
		return model.RoleConsume
	case VerbPublishes:
		return model.RolePublish
	case VerbSubscribes:
		return model.RoleSubscribe
	case VerbServes:
		return model.RoleServe
	case VerbCalls:
		return model.RoleCall
	case VerbReads:
		return model.RoleRead
	case VerbWrites:
		return model.RoleWrite
	case VerbGrants:
		return model.RoleGrant
	}
	return ""
}
