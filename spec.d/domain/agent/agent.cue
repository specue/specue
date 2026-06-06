// Package agent holds the needs of the agent audience: an automated caller
// that authors, navigates and plans against the spec and wants cheap,
// machine-readable feedback.
package agent

import (
	s "specue.io/schema@v0:spec"
	d "specue.io/domain@v0:domain"
)

setup: s.#Need & {
	slug:        "as-agent-setup"
	title:       "Choose which spec context I am working against"
	domain:      d.specue
	consumer:    "an agent starting work on this machine"
	description: "to declare which spec context I am working against right now, so that every subsequent command resolves against the same picture without me passing the set each time"
	frs: {
		"fr-01": {id: "fr-01", text: "A named context can be created and switched between."},
		"fr-02": {id: "fr-02", text: "A module is added to or removed from the current context by its directory."},
		"fr-03": {id: "fr-03", text: "The current context is readable on demand."},
	}
}

start: s.#Need & {
	slug:        "as-agent-start"
	title:       "Bring a new module into existence"
	domain:      d.specue
	consumer:    "an agent extending the landscape"
	description: "to start a new module of a known kind, so that subsequent authoring has a place to live"
	frs: {
		"fr-01": {id: "fr-01", text: "A new module declares its identity and its kind (service, domain, governance or code) at creation."},
		"fr-02": {id: "fr-02", text: "Creating a module over an existing one is refused."},
	}
}

create: s.#Need & {
	slug:        "as-agent-create"
	title:       "Add a Contract, Need, ADR or Port to a module"
	domain:      d.specue
	consumer:    "an agent authoring inside a module"
	description: "to add a Contract, Need, ADR or Port and arrange it where it belongs, so that the module carries the contracts and needs it owns, in a structure I can navigate"
	frs: {
		"fr-01": {id: "fr-01", text: "A new node carries an identity that is unique within its module."},
		"fr-02": {id: "fr-02", text: "The node kinds a module of a given kind may hold are visible in the schema the modules import."},
		"fr-03": {id: "fr-03", text: "Nodes within a module can be organized into sub-folders."},
	}
}

navigate: s.#Need & {
	slug:        "as-agent-navigate"
	title:       "Find my way around an unfamiliar spec"
	domain:      d.specue
	consumer:    "an agent exploring a spec I did not author"
	description: "to find my way around it on demand, so that I can answer questions about the system without reading every file"
	frs: {
		"fr-01": {id: "fr-01", text: "A Contract, Need, Port, ADR or code binding can each be listed."},
		"fr-02": {id: "fr-02", text: "Any one of them can be read in full by its module-qualified identity, which is stable across the landscape."},
		"fr-03": {id: "fr-03", text: "How they are related to each other is retrievable as machine-readable data."},
		"fr-04": {id: "fr-04", text: "Nodes matching a stated criterion can be found without naming each one."},
	}
}

relate: s.#Need & {
	slug:        "as-agent-relate"
	title:       "Wire one thing to another"
	domain:      d.specue
	consumer:    "an agent authoring the spec"
	description: "to wire one thing to another as part of the contract, so that the relationship is enforced, not just described in prose"
	frs: {
		"fr-01": {id: "fr-01", text: "A Contract invariant can declare which Need FR it satisfies."},
		"fr-02": {id: "fr-02", text: "A Contract element can declare which ADR justifies it."},
		"fr-03": {id: "fr-03", text: "A Contract element can declare which other Contract or Port it depends on."},
		"fr-04": {id: "fr-04", text: "A code file can declare which Contract a line of code realizes."},
	}
}

author: s.#Need & {
	slug:        "as-agent-author"
	title:       "Know whether what I authored is correct"
	domain:      d.specue
	consumer:    "an agent authoring a spec and binding it to code"
	description: "to know whether what I just wrote is correct, so that I can iterate without re-reading the landscape between changes"
	frs: {
		"fr-01": {id: "fr-01", text: "The spec as a whole is reported as correct or broken in a single check."},
		"fr-02": {id: "fr-02", text: "For a code module, every Contract it may realize and its current binding state are listed."},
		"fr-03": {id: "fr-03", text: "Every failure carries the next step the caller takes to resolve it."},
	}
}

review: s.#Need & {
	slug:        "as-agent-review"
	title:       "See what I changed between two points"
	domain:      d.specue
	consumer:    "an agent who has been authoring for a while"
	description: "to see what the spec became compared to where I started, so that I can review my own work before sharing it"
	frs: {
		"fr-01": {id: "fr-01", text: "The difference between the spec at two versioned points is reported as a typed delta over Contracts, Needs, Ports and their elements."},
		"fr-02": {id: "fr-02", text: "Each change names what was added, removed, modified or rewired."},
	}
}

planner: s.#Need & {
	slug:        "as-planner"
	title:       "See how a Plan lands before committing to it"
	domain:      d.specue
	consumer:    "an agent proposing a Plan that is not yet accepted"
	description: "to see how the Plan lands on the current spec, so that I can decide whether to accept it without breaking the system or other open Plans"
	frs: {
		"fr-01": {id: "fr-01", text: "A Plan is a named, retrievable object distinct from the current spec."},
		"fr-02": {id: "fr-02", text: "A Plan is viewable against the current spec without altering the working tree."},
		"fr-03": {id: "fr-03", text: "Two Plans whose changes cannot both apply (one removes what the other modifies, both rewire the same edge, etc.) are blocked before either is accepted."},
		"fr-04": {id: "fr-04", text: "Two Plans that touch the same Contract or Port but could both apply are surfaced for human or agent review rather than blocked."},
		"fr-05": {id: "fr-05", text: "Accepting a Plan applies its changes to the current spec and closes the Plan."},
		"fr-06": {id: "fr-06", text: "Planning requires a governance module in the current context; without one, the verb refuses with the next step to take."},
	}
}
