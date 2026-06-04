package model

// ModulePath is a module's canonical path (its spec.mod `module` line), e.g.
// "specue/example". It is the namespace within which slugs are unique.
type ModulePath string

// FilePath is the path of a source file (relative to its FS root), carried for
// diagnostics and the file:line a binding renders at — a typed string so it is
// never confused with an identifier or free text.
type FilePath string

// NodeID is a node's full identity: a slug is unique only within its module, so
// the module is part of the identity. The engine keys its graph by NodeID
// (nodeAt[Module][Slug]); display uses NodeID.String().
type NodeID struct {
	Module ModulePath
	Slug   Slug
}

func (id NodeID) String() string {
	return string(id.Module) + ":" + string(id.Slug)
}

// PlacedNode is an authored Node together with where it was found — like an AST
// node paired with its FileSet position. The Node itself stays purely authored
// (it carries only what the text says); the module and file are stamped by the
// source layer from the node's location, never written by the author. Identity
// is (Module, Node.Slug); the compiler resolves refs against these.
type PlacedNode struct {
	Module ModulePath
	File   FilePath // path of the source file, for diagnostics
	Node   Node
}

// ID returns the node's full identity.
func (p PlacedNode) ID() NodeID {
	return NodeID{Module: p.Module, Slug: p.Node.Slug}
}
