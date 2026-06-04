// The module manifest (spec.mod) — project structure, separate from node content
// (spec.cue). Same CUE package, its own file by domain. A module's kind gates
// which node types it may hold (role-gate, enforced by the compiler).
package spec

#kind: "service" | "domain" | "governance" | "topology" | "code"

// A require declares a dependency for the module-manager to locate and add to the
// closure (replace gives its local path during development). Import naming and the
// scoped-import set are CUE-native now — they live in cue.mod/module.cue and the
// import statement — so a require no longer carries as/use.
#require: {
	module!:  string
	version!: #semver
	replace?: string
}

#code: {
	repo!:    string
	replace?: string
}

#Module: {
	module!:  string
	version!: #semver
	kind!:    #kind
	code?: [...#code]
	require?: [...#require]
	// ignore lists gitignore-style globs the code scanner skips (a kind:code module
	// often tracks testdata/fixtures/generated code that carry foreign annotations).
	ignore?: [...string]
	// code_root is where the code scan begins, relative to spec.mod.cue's own
	// directory. Defaults to "." (the manifest's directory). A code module living
	// in spec.d/code/ inside a repo root sets code_root: "../.." so the scan
	// reaches the repo's Go tree without the code module claiming sibling
	// service modules as its own subpackages — the construct that produces
	// CUE's ambiguous-import error when a service module is registered both
	// nested (as a subpackage) and standalone (in the workspace). Honoured
	// only for kind:code; other kinds ignore it.
	code_root?: string
}

// The workspace (spec.work) — the landscape entry point. It lists every module of
// the landscape and where it lives, so the registry sees the whole graph at once:
// cross-module references resolve while authoring, without a require/replace per
// edge. A module's git repository is derived from its dir (the enclosing git
// toplevel), not declared here. A plan overlays this by reading some modules from a
// branch ref instead of the working dir; the module set itself is always the work.
#Work: {
	// dir each module's path is relative to (the workspace root); defaults to the
	// spec.work file's own directory.
	root?: string
	// plan_base is the branch a plan forks from and diffs against (spec may live on
	// a design branch while HEAD is dev); empty = the repo's current branch.
	plan_base?: string
	modules!: [...#workModule]
}

#workModule: {
	path!: string  // the module's canonical path (its spec.mod `module` line)
	dir!:  string  // its directory, relative to the workspace root
}

#semver: =~"^v[0-9]+\\.[0-9]+\\.[0-9]+$"
