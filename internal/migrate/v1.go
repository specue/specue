package migrate

import "encoding/json"

// The v1 authored shape, as decoded from YAML. Only the fields the v2 model
// carries are mapped; v1-only fields the v2 schema dropped (slice, purpose,
// audience) are read but not emitted.

type v1Manifest struct {
	Module  string      `json:"module"`
	Version string      `json:"version"`
	Kind    string      `json:"kind"`
	Require []v1Require `json:"require"`
}

type v1Require struct {
	Module  string   `json:"module"`
	Version string   `json:"version"`
	As      string   `json:"as"`
	Use     []string `json:"use"`
	Replace string   `json:"replace"`
}

type v1Node struct {
	Slug       string `json:"slug"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Confidence string `json:"confidence"`
	Body       string `json:"body"`
	LegacyID   string `json:"legacy_id"`

	// UseCase
	Service        string        `json:"service"`
	Binding        string        `json:"binding"`
	Interaction    string        `json:"interaction"`
	Trigger        string        `json:"trigger"`
	Deprecated     string        `json:"deprecated"`
	Preconditions  []v1Element   `json:"preconditions"`
	Postconditions []v1Element   `json:"postconditions"`
	Invariants     []v1Element   `json:"invariants"`
	Variations     []v1Element   `json:"variations"`

	// UserStory
	Product   string      `json:"product"`
	Narrative v1Narrative `json:"narrative"`
	FR        []v1Atom    `json:"fr"`
	NFR       []v1Atom    `json:"nfr"`

	// Port / Container
	Kind       string `json:"kind"`
	Technology string `json:"technology"`
	Transport  string `json:"transport"`
	Schema     string `json:"schema"`
	Boundary   bool   `json:"boundary"`

	// Plan / ADR
	Status string `json:"status"`
	Branch string `json:"branch"`
}

// v1Element is a WHAT-element. Like a dep, v1 accepts a bare string (a plain
// condition with only text) or a struct; UnmarshalJSON folds both.
type v1Element struct {
	ID        string   `json:"id"`
	Text      string   `json:"text"`
	When      string   `json:"when"`
	Then      string   `json:"then"`
	Rev       int      `json:"rev"`
	DependsOn []v1Dep  `json:"depends_on"`
	Invokes   []v1Dep  `json:"invokes"`
	Infra     []v1Dep  `json:"infra"`
	Satisfies []string `json:"satisfies"`
	DecidedBy []string `json:"decided_by"`
}

func (e *v1Element) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		return json.Unmarshal(b, &e.Text)
	}
	type raw v1Element
	return json.Unmarshal(b, (*raw)(e))
}

// v1Dep is a dependency edge. v1 accepts two authored shapes: a bare string
// ("alias:slug", a plain L3 dep with no role) and a struct ({to, role, carries}).
// UnmarshalJSON folds both into this struct so the rest of the migrator sees one
// form. (CUE's decoder uses encoding/json semantics, so this hook fires.)
type v1Dep struct {
	To      string `json:"to"`
	Role    string `json:"role"`
	Carries string `json:"carries"`
}

func (d *v1Dep) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		return json.Unmarshal(b, &d.To)
	}
	type raw v1Dep // avoid recursing into this method
	return json.Unmarshal(b, (*raw)(d))
}

type v1Narrative struct {
	As     string `json:"as"`
	Want   string `json:"want"`
	SoThat string `json:"so_that"`
}

type v1Atom struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
