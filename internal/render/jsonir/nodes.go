package jsonir

// One file struct per node type: the common envelope, the type-specific
// payload, and the cross-cutting derived/bindings sections, embedded so the
// JSON output is one flat object. Adding a new node type means adding one new
// file struct here and one builder in render.go — nothing else moves.

type fileContract struct {
	commonJSON
	useCaseJSON
	Derived  *derivedJSON  `json:"derived,omitempty"`
	Bindings *bindingsJSON `json:"bindings,omitempty"`
}

type fileNeed struct {
	commonJSON
	needJSON
	Derived  *derivedJSON  `json:"derived,omitempty"`
	Bindings *bindingsJSON `json:"bindings,omitempty"`
}

type fileDomain struct {
	commonJSON
	Derived  *derivedJSON  `json:"derived,omitempty"`
	Bindings *bindingsJSON `json:"bindings,omitempty"`
}

type filePort struct {
	commonJSON
	portJSON
	Derived  *derivedJSON  `json:"derived,omitempty"`
	Bindings *bindingsJSON `json:"bindings,omitempty"`
}

type fileContainer struct {
	commonJSON
	containerJSON
	Derived  *derivedJSON  `json:"derived,omitempty"`
	Bindings *bindingsJSON `json:"bindings,omitempty"`
}

type fileGov struct {
	commonJSON
	govJSON
	Derived  *derivedJSON  `json:"derived,omitempty"`
	Bindings *bindingsJSON `json:"bindings,omitempty"`
}
