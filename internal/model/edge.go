package model

// The three edge classes are not a tagged union but three typed lists on every
// WHAT-element (see Element): Deps, Satisfies, DecidedBy. This keeps each class's
// own shape — only a dep carries Role/Carries/Branch, only satisfies targets an
// atom — instead of one struct with mostly-empty fields. An edge attached to no
// element does not exist.

// Role is the direction of an infrastructure dep, from the perspective of the
// service that owns the element. Empty Role means a plain contract dep, not infra.
type Role string

const (
	RoleProduce   Role = "produce"
	RolePublish   Role = "publish"
	RoleConsume   Role = "consume"
	RoleSubscribe Role = "subscribe"
	RoleServe     Role = "serve"
	RoleCall      Role = "call"
	RoleRead      Role = "read"
	RoleWrite     Role = "write"
	RoleGrant     Role = "grant"
)

// Dep is a dependency on another contract or piece of infrastructure (v1's
// depends_on / invokes / infra, collapsed). Role tells the two apart.
//
// Two attributes are load-bearing for status derivation and must survive any
// model simplification (see MIGRATION-PLAN-RU "что модель обязана сохранить"):
//
//   - Branch marks a dep taken only under a variation's guard. Branch deps are
//     excluded from a node's core dependency set, so they do not block the parent
//     in blocked-propagation. Collapsing them into plain deps changes statuses.
//   - Role (when set) makes the dep an infra touch and drives the derived L2
//     topology (producedBy / consumedBy / servedBy / calledBy).
type Dep struct {
	// To is the dependency target: a Contract for a plain dep, a Port/Container
	// for an infra dep (Role set).
	To NodeRef
	// Role, when non-empty, marks this as an infrastructure touch.
	Role Role
	// Carries references the Contract whose logical L3 contract this physical
	// infra link realizes — the L3→L2 bridge. Set only when that contract differs
	// from the element's own node (an outgoing link into another service); empty
	// on the side that owns the contract.
	Carries NodeRef
	// Branch is true when this dep is taken only under a variation guard.
	Branch bool
}

// IsInfra reports whether the dep is an infrastructure touch (has a Role).
func (d Dep) IsInfra() bool { return d.Role != "" }
