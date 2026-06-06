// Package context holds the contracts that set the scene before any read or
// authoring verb runs: the operations over a named context, and the scaffold of
// a new module.
package context

import (
	s "specue.io/schema@v0:spec"
	root "specue.io/service@v0:service"
	agent "specue.io/domain/agent@v0:agent"
	fed "specue.io/domain/federated@v0:federated"
	gov "specue.io/governance@v0:governance"
)

createContext: s.#Contract & {
	slug:        "create-context"
	title:       "Create a new context"
	service:     root.specue
	trigger:     "the caller asks to create a context by name"
	invariants: [{
		id:   "name-is-unique"
		kind: "rejects"
		when: "a context with that name already exists"
		satisfies: [agent.setup.frs."fr-01"]
	}, {
		id:   "starts-empty"
		text: "A new context holds no modules until the caller adds them."
	}, {
		id:   "survives-across-invocations"
		text: "The context survives across invocations."
	}]
}

useContext: s.#Contract & {
	slug:        "use-context"
	title:       "Make a context the active one"
	service:     root.specue
	trigger:     "the caller asks to switch into a context by name"
	invariants: [{
		id:   "context-must-exist"
		kind: "rejects"
		when: "the named context does not exist"
		satisfies: [agent.setup.frs."fr-01"]
	}, {
		id:   "subsequent-verbs-resolve-here"
		text: "Once active, every subsequent read or authoring verb resolves against this context's modules unless overridden for the run."
	}, {
		id:   "active-across-invocations"
		text: "The chosen context is active across invocations until another one is switched in."
	}]
}

readContext: s.#Contract & {
	slug:        "read-context"
	title:       "Read the active context"
	service:     root.specue
	trigger:     "the caller asks which context is active"
	invariants: [{
		id:   "read-returns-current-state"
		kind: "returns"
		text: "Reading the context returns its current state — the same name and module set on every read."
		satisfies: [agent.setup.frs."fr-03"]
		decided_by: [gov.adr14OneInvariantKind]
	}, {
		id:   "names-membership"
		kind: "returns"
		text: "The result names the context and every module it carries."
		satisfies: [agent.setup.frs."fr-03"]
	}, {
		id:   "no-active-context-told"
		kind: "rejects"
		when: "no context is active"
	}]
}

removeContext: s.#Contract & {
	slug:        "remove-context"
	title:       "Remove a context"
	service:     root.specue
	trigger:     "the caller asks to remove a context by name"
	invariants: [{
		id:   "context-must-exist"
		kind: "rejects"
		when: "the named context does not exist"
	}, {
		id:   "removed-until-recreated"
		text: "Once removed the context cannot be switched into until it is created again."
	}]
}

addModuleToContext: s.#Contract & {
	slug:        "add-module-to-context"
	title:       "Add a module to a context by its directory"
	service:     root.specue
	trigger:     "the caller asks to add a module to the current context"
	invariants: [{
		id:   "addressed-by-directory"
		text: "The module is addressed by the directory that holds its manifest, not by its name."
		satisfies: [
			agent.setup.frs."fr-02",
			fed.owner.frs."fr-02",
		]
	}, {
		id:   "must-be-a-module"
		kind: "rejects"
		when: "the directory does not hold a module manifest"
	}, {
		id:   "git-repository-required"
		kind: "rejects"
		when: "the module does not live in a git repository"
		decided_by: [gov.adr03GitNative]
	}, {
		id:   "reachable-until-removed"
		text: "The module is reachable from the context until it is removed."
	}]
}

removeModuleFromContext: s.#Contract & {
	slug:        "remove-module-from-context"
	title:       "Remove a module from a context"
	service:     root.specue
	trigger:     "the caller asks to remove a module from the current context"
	invariants: [{
		id:   "addressed-by-module-path"
		text: "The module is removed by its module path, which is unique within the context."
		satisfies: [agent.setup.frs."fr-02"]
	}, {
		id:   "unreachable-until-readded"
		text: "The module is no longer reachable from the context until it is added again."
	}]
}

initModule: s.#Contract & {
	slug:        "init-module"
	title:       "Start a new module of a known kind"
	service:     root.specue
	trigger:     "the caller asks to scaffold a new module at a directory"
	invariants: [{
		id:   "identity-and-kind-at-creation"
		text: "A new module declares its identity and its kind (service, product, governance or code) when it is created."
		satisfies: [agent.start.frs."fr-01"]
	}, {
		id:   "no-overwrite"
		kind: "rejects"
		when: "a module already exists at the target directory"
		satisfies: [agent.start.frs."fr-02"]
	}, {
		id:   "git-repository-required"
		kind: "rejects"
		when: "the target is outside a git repository"
		decided_by: [gov.adr03GitNative]
	}, {
		id:   "scaffolds-manifest-only"
		kind: "returns"
		text: "The new module is left as a directory with the manifest the kind requires and nothing else."
	}]
}
