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

createContext: s.#UseCase & {
	slug:        "create-context"
	title:       "Create a new context"
	service:     root.specue
	trigger:     "the caller asks to create a context by name"
	invariants: [{
		id:   "name-is-unique"
		text: "Creating a context with a name that already exists is refused."
		satisfies: [agent.setup.frs."fr-01"]
	}, {
		id:   "starts-empty"
		text: "A new context holds no modules until the caller adds them."
	}]
	postconditions: [{
		text: "The context survives across invocations on the same machine."
	}]
}

useContext: s.#UseCase & {
	slug:        "use-context"
	title:       "Make a context the active one"
	service:     root.specue
	trigger:     "the caller asks to switch into a context by name"
	invariants: [{
		id:   "context-must-exist"
		text: "Switching into a context that does not exist is refused with the next step to take."
		satisfies: [agent.setup.frs."fr-01"]
	}, {
		id:   "subsequent-verbs-resolve-here"
		text: "Once active, every subsequent read or authoring verb resolves against this context's modules unless overridden for the run."
	}]
	postconditions: [{
		text: "The chosen context is active across invocations until another one is switched in."
	}]
}

readContext: s.#UseCase & {
	slug:        "read-context"
	title:       "Read the active context"
	service:     root.specue
	trigger:     "the caller asks which context is active"
	invariants: [{
		id:   "does-not-mutate"
		text: "Reading the context does not alter it."
		satisfies: [agent.setup.frs."fr-03"]
	}, {
		id:   "names-membership"
		text: "The result names the context and every module it carries."
		satisfies: [agent.setup.frs."fr-03"]
	}]
	postconditions: [{
		text: "If no context is active the caller is told so with the next step to take."
	}]
}

removeContext: s.#UseCase & {
	slug:        "remove-context"
	title:       "Remove a context"
	service:     root.specue
	trigger:     "the caller asks to remove a context by name"
	invariants: [{
		id:   "context-must-exist"
		text: "Removing a context that does not exist is refused with the next step to take."
	}, {
		id:   "modules-survive"
		text: "The directories that held the context's modules are left untouched."
	}]
	postconditions: [{
		text: "Once removed the context cannot be switched into until it is created again."
	}]
}

addModuleToContext: s.#UseCase & {
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
		text: "Adding a directory that does not hold a module manifest is refused with the next step to take."
	}, {
		id:   "git-repository-required"
		text: "Adding a module that does not live in a git repository is refused with the next step to take."
		decided_by: [gov.adr03GitNative]
	}]
	postconditions: [{
		text: "The module is reachable from the context until it is removed."
	}]
}

removeModuleFromContext: s.#UseCase & {
	slug:        "remove-module-from-context"
	title:       "Remove a module from a context"
	service:     root.specue
	trigger:     "the caller asks to remove a module from the current context"
	invariants: [{
		id:   "addressed-by-module-path"
		text: "The module is removed by its module path, which is unique within the context."
		satisfies: [agent.setup.frs."fr-02"]
	}, {
		id:   "module-untouched"
		text: "The directory the module lives in is left as it was."
	}]
	postconditions: [{
		text: "The module is no longer reachable from the context until it is added again."
	}]
}

initModule: s.#UseCase & {
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
		text: "Scaffolding over an existing module is refused; the existing one is left untouched."
		satisfies: [agent.start.frs."fr-02"]
	}, {
		id:   "git-repository-required"
		text: "Scaffolding outside a git repository is refused with the next step to take."
		decided_by: [gov.adr03GitNative]
	}]
	postconditions: [{
		text: "The new module is left as a directory with the manifest the kind requires and nothing else."
	}]
}
