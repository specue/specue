package jsonir

// The JSON IR is split into one schema per node type so each type evolves on
// its own. The common envelope (commonJSON) carries identity and the two
// cross-cutting sections (derived / bindings) every node has; a per-type
// payload (useCaseJSON, needJSON, …) carries the body-specific shape. The
// renderer assembles a flat map[string]any for the file body — flat because
// the wire schema is one JSON object per node, not a tagged union. Splitting
// payloads in Go means adding a new node type touches only its file.

// commonJSON is the envelope every node JSON file opens with — identity, type,
// module, status/confidence/visibility, the long-form body, and rendered_from.
type commonJSON struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Module       string `json:"module"`
	Slug         string `json:"slug"`
	Title        string `json:"title,omitempty"`
	Status       string `json:"status,omitempty"`
	Confidence   string `json:"confidence,omitempty"`
	Visibility   string `json:"visibility,omitempty"`
	Body         string `json:"body,omitempty"`
	RenderedFrom string `json:"rendered_from,omitempty"`
}

// useCaseJSON is the Contract-specific payload: service, binding, interaction,
// trigger, deprecated, and the invariants.
type useCaseJSON struct {
	Service     string     `json:"service,omitempty"`
	Binding     string     `json:"binding,omitempty"`
	Interaction string     `json:"interaction,omitempty"`
	Trigger     string     `json:"trigger,omitempty"`
	Deprecated  string     `json:"deprecated,omitempty"`
	Invariants  []elemJSON `json:"invariants,omitempty"`
}

// needJSON is the Need-specific payload: domain, consumer, description, and
// the FR/NFR atoms split by kind.
type needJSON struct {
	Domain      string     `json:"domain,omitempty"`
	Consumer    string     `json:"consumer,omitempty"`
	Description string     `json:"description,omitempty"`
	FRs         []atomJSON `json:"frs,omitempty"`
	NFRs        []atomJSON `json:"nfrs,omitempty"`
}

// portJSON is the Port-specific payload: kind (channel|rpc|rest|datastore),
// technology, transport, schema ref.
type portJSON struct {
	Kind       string `json:"kind,omitempty"`
	Technology string `json:"technology,omitempty"`
	Transport  string `json:"transport,omitempty"`
	Schema     string `json:"schema,omitempty"`
}

// containerJSON is the Container-specific payload: kind, technology, boundary.
type containerJSON struct {
	Kind       string `json:"kind,omitempty"`
	Technology string `json:"technology,omitempty"`
	Boundary   bool   `json:"boundary"`
}

// govJSON is the Plan/ADR payload: lifecycle, and (Plan only) branch.
type govJSON struct {
	Lifecycle string `json:"lifecycle,omitempty"`
	Branch    string `json:"branch,omitempty"`
}

// elemJSON is one invariant — the only place spec edges attach. Kind carries the
// element's nature (returns / rejects; empty for a plain guarantee).
type elemJSON struct {
	ID        string        `json:"id,omitempty"`
	Kind      string        `json:"kind,omitempty"`
	Text      string        `json:"text,omitempty"`
	When      string        `json:"when,omitempty"`
	Rev       int           `json:"rev,omitempty"`
	DependsOn []depJSON     `json:"depends_on,omitempty"`
	Satisfies []satisfyJSON `json:"satisfies,omitempty"`
	DecidedBy []string      `json:"decided_by,omitempty"`
}

type depJSON struct {
	To      string `json:"to"`
	Role    string `json:"role,omitempty"`
	Carries string `json:"carries,omitempty"`
	Branch  bool   `json:"branch,omitempty"`
}

type satisfyJSON struct {
	Need string `json:"need"`
	Atom string `json:"atom"`
}

type atomJSON struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Text string `json:"text,omitempty"`
}

// derivedJSON carries fields the compiler computes — never authored.
type derivedJSON struct {
	Uses      []string      `json:"uses,omitempty"`
	CoreUses  []string      `json:"core_uses,omitempty"`
	Satisfies []satisfyJSON `json:"satisfies,omitempty"`
	Realizes  []string      `json:"realizes,omitempty"`
	Topology  *topologyJSON `json:"topology,omitempty"`
	Blocked   bool          `json:"blocked,omitempty"`
}

type topologyJSON struct {
	ProducedBy []string `json:"produced_by,omitempty"`
	ConsumedBy []string `json:"consumed_by,omitempty"`
	ServedBy   []string `json:"served_by,omitempty"`
	CalledBy   []string `json:"called_by,omitempty"`
	GrantedBy  []string `json:"granted_by,omitempty"`
}

// bindingsJSON carries scan facts (code annotations attached to this node).
type bindingsJSON struct {
	Req    []bindingJSON      `json:"req,omitempty"`
	Covers []bindingJSON      `json:"covers,omitempty"`
	Infra  []infraBindingJSON `json:"infra,omitempty"`
}

type bindingJSON struct {
	Element      string `json:"element,omitempty"`
	Loc          string `json:"loc"`
	SourceModule string `json:"source_module,omitempty"`
}

type infraBindingJSON struct {
	Role         string `json:"role"`
	To           string `json:"to,omitempty"`
	Element      string `json:"element,omitempty"`
	Loc          string `json:"loc"`
	SourceModule string `json:"source_module,omitempty"`
}

// indexJSON is the root index file's wire shape: every module, every node id,
// every node's file path. Slim — no node bodies.
type indexJSON struct {
	RenderedFrom string            `json:"rendered_from"`
	Modules      []indexModuleJSON `json:"modules"`
	Nodes        []indexNodeJSON   `json:"nodes"`
}

type indexModuleJSON struct {
	Path         string `json:"path"`
	Kind         string `json:"kind,omitempty"`
	NodeCount    int    `json:"node_count"`
	RenderedFrom string `json:"rendered_from,omitempty"`
}

type indexNodeJSON struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
	Title  string `json:"title,omitempty"`
	Path   string `json:"path"`
}
