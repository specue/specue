package model

// Node is an authored spec node. This is the pure authored shape only — derived
// fields (uses, realizes, topology) and scan results (status, bindings) are NOT
// here; they belong to the compiler/engine, which take model.Node as input.
//
// Identity is (Module, Slug): a slug is unique only within its module, like a Go
// symbol in a package. The module is assigned by the source layer from the file's
// location; the model carries the slug and the parsed body.
//
// On open vs closed value types: an enum is closed (a Go const set) only when the
// tool branches on its value — NodeType, Role, ElementKind, PortKind, Visibility,
// Interaction, Binding, Lifecycle drive role-gating, cycle detection, and
// blocked-propagation, so they live in code. A pure label the tool never reasons
// about — Transport, Technology — is an open string; its allowed values live in
// the CUE schema, so adding one is a schema bump, not a tool change.
type NodeType string

const (
	TypeUseCase   NodeType = "UseCase"
	TypeNeed      NodeType = "Need"
	TypeDomain    NodeType = "Domain"
	TypePort      NodeType = "Port"
	TypeContainer NodeType = "Container"
	TypePlan      NodeType = "Plan"
	TypeADR       NodeType = "ADR"
)

// Visibility is a node's reach across module boundaries. v1 derived this from an
// internal/ path segment; v2 carries it as a field (the source layer stamps it
// from the path at load), so the compiler checks it by unification, not by path.
type Visibility string

const (
	Public  Visibility = "public"
	Private Visibility = "private" // module-internal; other modules cannot reference it
)

// Confidence is the author's certainty in a node. Closed: reports surface it.
type Confidence string

const (
	Confirmed   Confidence = "CONFIRMED"
	Likely      Confidence = "LIKELY"
	Speculative Confidence = "SPECULATIVE"
)

type Node struct {
	Slug       Slug
	Type       NodeType
	Title      string
	Visibility Visibility
	Confidence Confidence
	// LegacyID is an optional transitional alias (e.g. GCS-UC-5); refs through it
	// resolve with a migrate warning.
	LegacyID string

	Body *Body // type-specific authored payload; nil only on a malformed node
}

// Body holds the authored payload that varies by node type. One struct with
// per-type sections keeps Node uniform while the compiler reads the section that
// matches Type. A node populates exactly the section for its Type.
type Body struct {
	Prose string // free-text narrative (v1 `body`)

	UseCase   *UseCaseBody
	Need      *NeedBody
	Port      *PortBody
	Container *ContainerBody
	Gov       *GovBody // Plan and ADR
}

// UseCaseBody is a logical contract a service guarantees, bound to code.
type UseCaseBody struct {
	// Service references the service node this UC belongs to.
	Service NodeRef
	// Trigger names an external event that reaches this UC; set only when the UC
	// is event-triggered (otherwise reachability comes from an inbound edge or a
	// satisfies).
	Trigger string
	// Binding is how others may rely on this contract: required (default) means
	// it must be implemented in code; optional / abstract relax that.
	Binding Binding
	// Interaction is how callers invoke it: async (default) or sync. Load-bearing
	// for cycle detection (a sync cycle is broken, an async one is fine) — must
	// survive simplification.
	Interaction Interaction
	// Deprecated, when non-empty, is an advisory retirement note on inbound refs.
	Deprecated string

	Elements []Element // preconditions, postconditions, invariants, variations
}

// NeedBody is the intent seam. It owns the atoms it requires and binds to no
// code; its coverage is computed from the UseCases that satisfy its atoms.
// See ADR-10 for why Need (not UserStory) is the unit here.
type NeedBody struct {
	// Domain references the Domain node this Need belongs to.
	Domain NodeRef
	// Consumer names who or what requires this Need (operator, downstream
	// system, regulator, agent — not necessarily a human).
	Consumer string
	// Description is the stable prose of the need, kept independent of any
	// delivery cadence.
	Description string
	Atoms       []Atom // FR and NFR
}

// PortBody is a typed transport surface a service attaches to (C4 L2).
type PortBody struct {
	Kind       PortKind
	Technology string    // open label
	Transport  Transport // open label; allowed values in the CUE schema
	// Schema (rpc/rest) references the wire IDL (.proto/openapi); bound in code.
	Schema NodeRef
}

// PortKind is closed: the compiler gates fields on it (channel→transport,
// rpc/rest→schema) and routes topology by it.
type PortKind string

const (
	PortChannel   PortKind = "channel"   // async pub/sub
	PortRPC       PortKind = "rpc"       // sync request/response gRPC
	PortREST      PortKind = "rest"      // sync request/response HTTP
	PortDatastore PortKind = "datastore" // persistence
)

// Transport is an open label (kafka, grpc, http, redis-pubsub, websocket, …) the
// tool never branches on; its allowed set lives in the CUE schema so a new
// transport is a schema bump, not a tool change.
type Transport string

// ContainerBody is a boundary box/actor that is not itself an attach-point: an
// external actor (a person, a third-party system). Model the seam, not internals.
type ContainerBody struct {
	Kind       ContainerKind
	Technology string // open label
	Boundary   bool   // third-party: model the seam, not internals
}

type ContainerKind string

const (
	ContainerClient   ContainerKind = "client"
	ContainerExternal ContainerKind = "external"
	ContainerGateway  ContainerKind = "gateway"
	ContainerBroker   ContainerKind = "broker"
	ContainerService  ContainerKind = "service"
	ContainerCronJob  ContainerKind = "cronjob" // a periodic-trigger box (k8s CronJob)
)

// GovBody is the payload of a governance node — a Plan or an ADR. Both bind to no
// code (like a Need). A Plan's speculative content lives on plan/<id>
// branches; this records its identity and lifecycle.
type GovBody struct {
	Lifecycle Lifecycle // proposed | accepted | superseded
	// Branch (Plan only) names the git branch carrying the speculative spec;
	// empty defaults to plan/<slug>.
	Branch string
}

// Binding is closed: required gates implementation-in-code.
type Binding string

const (
	BindingRequired Binding = "required"
	BindingOptional Binding = "optional"
	BindingAbstract Binding = "abstract"
)

// Interaction is closed: sync vs async changes cycle detection.
type Interaction string

const (
	InteractionAsync Interaction = "async"
	InteractionSync  Interaction = "sync"
)

// Lifecycle is closed: governance state transitions.
type Lifecycle string

const (
	LifecycleProposed   Lifecycle = "proposed"
	LifecycleAccepted   Lifecycle = "accepted"
	LifecycleSuperseded Lifecycle = "superseded"
)
