// Package governance holds Specue's architecture decision records: the causal
// layer. An ADR is a code-unbound node; Contract elements cite one via decided_by
// to record why a contract is shaped as it is, kept out of the contract's own text.
package governance

import s "specue.io/schema@v0:spec"

adr01CUENativeResolution: s.#ADR & {
	slug:       "ADR-01"
	title:      "Cross-module references resolve through CUE, not a hand-written resolver"
	status:     "accepted"
	body: """
		The whole module set is stitched into one CUE value tree, and CUE resolves
		every cross-module reference, version pin and visibility rule. The previous
		generation interpreted its own mini-language of string references through a
		hand-written resolver, which became the system's bottleneck. Standing on CUE
		shifts a class of resolution bugs onto a mature implementation and lets the
		compiler do only what CUE cannot — domain constraints (statuses, cycles,
		coverage).
		"""
}

adr02SQLQuery: s.#ADR & {
	slug:       "ADR-02"
	title:      "Graph navigation and search is exposed as read-only SQL"
	status:     "accepted"
	body: """
		The graph is projected into an in-memory SQLite database the caller queries
		with SQL — recursive CTEs for walks, full-text search for lookup — instead of
		a fixed set of navigation verbs. A discoverable schema lets one query answer
		what several fixed verbs would, which matters most for the agent caller: fewer
		round-trips, less output to read. The projection is read-only and rebuilt from
		the graph, never a second source of truth.
		"""
}

adr03GitNative: s.#ADR & {
	slug:       "ADR-03"
	title:      "Every module lives in a git repository"
	status:     "accepted"
	body: """
		Plans are branches, scanned code is what git tracks, and a module's history
		comes from its repository — so the tool treats git as infrastructure, not an
		option. A module outside a repository is refused at scaffold time with a
		remedy. This collapses the matrix of "what if there is no git" branches and
		makes plans, diffs and the scanner share a single source of versioned truth.
		"""
}

adr04RegistryInProcess: s.#ADR & {
	slug:       "ADR-04"
	title:      "The OCI registry that warms the cue cache runs in-process"
	status:     "accepted"
	body: """
		The registry that hosts the schema and the landscape's modules for the
		editor's cue lsp is brought up inside the tool's own process from the
		cuelabs library, not as a separate daemon. It lives only long enough to
		publish and warm the cache, then exits. The tool does not depend on a
		development-only CLI it cannot control, and the editor sees a populated
		on-disk cache it can resolve from without anything alive in the background.
		"""
}

adr05CodeAsModule: s.#ADR & {
	slug:       "ADR-05"
	title:      "Code is a first-class module of the landscape"
	status:     "accepted"
	body: """
		A code module is just another module — manifest, requires, kind: code — that
		holds no spec nodes of its own and declares which contracts its source may
		bind through its requires. The previous generation treated code as a sidecar
		attached to a spec module; promoting it makes code participate in the same
		mechanisms (contexts, plans, validation) the rest of the graph uses, with no
		special cases.
		"""
}

adr06FixedSchemaVersion: s.#ADR & {
	slug:       "ADR-06"
	title:      "The embedded schema version is fixed; changes ship under the same tag"
	status:     "accepted"
	body: """
		Every module pins the schema version in its cue.mod deps, so changing the
		version would break every pin in every module on every release. The version
		is fixed; a change to the schema's contents is detected and republished
		under the same tag, so no module pin ever moves. A breaking change to the
		schema is expressed as a NEW major (a sibling module path) embedded
		alongside the old one, so existing modules keep resolving while authors
		migrate at their own pace.
		"""
}

adr07PlansAsBranches: s.#ADR & {
	slug:       "ADR-07"
	title:      "A plan is a named branch across every module it touches"
	status:     "accepted"
	body: """
		A plan's content lives on identically-named branches in every affected
		repository; its identity is a Plan record in a dedicated governance module
		of the landscape — kind: governance — that points at those branches. The
		governance module is where ADRs also live, kept apart from modules that
		hold Contracts, UserStories or Ports. Speculative work is real CUE on a
		real ref the tool can read, diff and overlay, not a separate document
		store. Acceptance merges the branches; conflicts between plans are gates
		derived by overlaying both deltas. The intent axis is git, with
		governance only naming what is in flight.
		"""
}

adr08AttestedBindings: s.#ADR & {
	slug:       "ADR-08"
	title:      "Code-binding outcomes are published alongside the spec for readers without code access"
	status:     "accepted"
	body: """
		The spec and the code that realizes it sit on different sides of an access
		boundary in many real systems: a reader may hold one and not the other.
		Whoever holds the code publishes a small attestation artifact alongside its
		spec module — the binding outcomes per Contract, no source — and a reader
		consumes that instead of scanning. Status is computed by the same rules in
		both paths, so a reader and a code holder see the same picture and the
		federated boundary becomes invisible at the spec level.
		"""
}

adr11CodeRootAndLayout: s.#ADR & {
	slug:       "ADR-11"
	title:      "Code module scans from code_root; repo modules live under spec.d/"
	status:     "accepted"
	body: """
		A code module in a repo root and a service module in a subfolder are, to
		CUE's load resolver, two paths to the same package — registered both
		standalone (workspace) and nested (code module's subtree). CUE refuses
		that with ambiguous-import the moment another module imports the service.

		Two changes together settle it.

		`code_root` (manifest field, relative to spec.mod.cue, default ".") moves
		where the scan begins. A code module in a subfolder points back at the
		repo through code_root, so the manifest is separated from the source it
		scans — and stops claiming sibling spec modules as its own subpackages.

		`spec.d/<kind>/[<name>/]` is the recommended layout: code at spec.d/code/,
		services at spec.d/service/<name>/, etc. Unix drop-in convention (cron.d,
		sudoers.d), visible to agents and shells (no leading dot). `init --layout
		spec.d --kind code` writes it and fills `code_root: "../.."`.

		Recommended, not required: the older flat layout works via code_root
		alone. Self-spec moves to spec.d/ as the living demo.
		"""
}

adr10NeedNotUserStory: s.#ADR & {
	slug:       "ADR-10"
	title:      "The intent node is a Need (with a Domain), not a UserStory (with a Product)"
	status:     "accepted"
	body: """
		The intent node names what the system must provide, not how a team plans an
		iteration. UserStory carries Agile baggage that misleads here: it implies a
		sprint-sized increment authored from a user persona ("as a … I want … so
		that …"), acceptance criteria in Gherkin, and a lifetime bounded by the
		sprint that delivered it. None of that fits the layer Specue models. The
		intent is long-lived, not iterative; its consumer is often non-human (an
		operator, a downstream system, a regulator, an agent); the testable atoms
		(FR/NFR) are the contract, not "acceptance criteria"; and the unit lives as
		long as the system serves it, not until a story is closed.

		Requirements Engineering (ISO/IEC 29148, IREB) calls this unit a Need: an
		objective statement of what a stakeholder requires, independent of any
		delivery cadence. Need carries the right semantics — a consumer and a
		description, with named atoms — and drops the persona-narrative grammar
		that does not generalise. The container above Need is the audience, also
		named in RE-lexicon: Domain. (This is the same Domain DDD codifies; the
		terms align, they do not conflict.) Story statuses follow the same shift:
		delivered/partial/orphan becomes covered/partial/uncovered, since a Need
		is covered by contracts rather than "delivered" by a team.

		Pinning the choice in an ADR settles the question: future authors find
		Need + Domain in the schema and the rationale here, not a recurring debate
		about whether to reintroduce UserStory.
		"""
}

adr12RenderFormatsAndPresets: s.#ADR & {
	slug:       "ADR-12"
	title:      "Render serves two formats — markdown (with knobs) and JSON IR — not a family of presets"
	status:     "accepted"
	body: """
		A single rendered tree cannot satisfy every downstream pipeline: GitHub
		preview wants relative-link markdown, Confluence-via-mark wants a
		PascalCase frontmatter (Title/Space/Parent/Labels), MkDocs Material wants
		lowercase keys plus a nav: snippet, and custom dashboards want structured
		data, not text.

		Render carries two formats and one tunable markdown renderer:

		- `--format markdown` (default) is the publishing target. Knobs:
		  `--layout flat|tree`, `--strip-prefix <s>`, `--frontmatter
		  full|minimal|mark|mkdocs|none`, `--space <key>`, `--nav-snippet <file>`.
		  One renderer, one body, many YAML/path projections.
		- `--format json` emits the structured graph (per-node JSON + index.json)
		  for callers that do their own rendering — a custom Confluence push, an
		  analytics dashboard, a markdown formatter that does not exist yet.

		Rejected: a preset-per-package. Mark and mkdocs differ almost entirely
		in frontmatter shape; >90% would be shared via composition, at the cost
		of a class hierarchy that obscures the tiny real difference. Flags are
		the smaller surface and stay open to future targets without a rename.

		Rejected: ship only JSON IR and let every consumer render markdown
		themselves. GitHub / GitLab preview .md files in-repo with no pipeline —
		that is the cheap-yet-honest publish channel for a self-spec'd tool.
		"""
}

adr09RenderedDocDerived: s.#ADR & {
	slug:       "ADR-09"
	title:      "The rendered document is a derived view, never an authoring source"
	status:     "accepted"
	body: """
		A reader who cannot run the tool still needs the spec to be legible: the
		rendered markdown document is that channel. It is produced from the resolved
		graph and only from the resolved graph — nobody edits the markdown and
		expects it to feed back. Editing happens in CUE; the document is a stale
		artifact the moment the spec moves, so it carries the source revision it was
		rendered from. Any drift between document and spec is closed by re-rendering,
		never by editing the document. This keeps the authoring surface (CUE) the
		only place truth is mutated and treats the markdown as a publishing format.
		"""
}

adr13ContractNotUseCase: s.#ADR & {
	slug:       "ADR-13"
	title:      "The contract node is named Contract, not UseCase (nor Capability)"
	status:     "accepted"
	body: """
		The node names a logical contract a service guarantees. `UseCase` carries
		UML baggage that misleads the way UserStory did for intent (ADR-10): a use
		case is an actor-system interaction scenario. Specue's node is not a
		scenario, it is a guarantee — and "a use case with no actor" is a
		contradiction the operation contracts (the Plan/context verbs, which face no
		external audience) break.

		Rename to `Contract`. It reads right for both a Need-facing contract and an
		internal operation contract (an internal contract is normal in
		Design-by-Contract). Rejected `Capability`: a capability is a bare "can do"
		with no place for the guarantees a contract carries.

		Breaking rename across schema, model, code annotations and the self-spec;
		ships in the pre-release window. Fixes only the node's name, not the shape
		of its guarantees. Symmetric with ADR-10 fixing Need over UserStory.
		"""
}

adr14OneInvariantKind: s.#ADR & {
	slug:       "ADR-14"
	title:      "A Contract is a set of invariants; pre/post/variation collapse into one typed kind"
	status:     "proposed"
	body: """
		The schema carried four element collections — preconditions, postconditions,
		invariants, variations — but none has content the invariant cannot hold: a
		postcondition is an invariant over the result; a variation is an invariant
		with a `when` guard; a precondition's only observable content is the
		rejection when it is violated.

		Collapse to one element kind: an `invariant` with `text`, an optional `when`
		guard, and an optional `kind: "returns" | "rejects"`. `returns` and `rejects`
		are the only two natures worth authoring — both positively provable (rule
		1.6) and not already edge facts (rule 5.2). `mutates`/`calls` come from infra
		edges; a negative guarantee ("does not alter") comes from the absence of a
		write edge — never authored.

		Breaking across schema, model and the self-spec; ships in the pre-release
		window with ADR-13. Pins the element shape so the "how many element kinds"
		question stays settled — symmetric with ADR-10 (Need) and ADR-13 (Contract).
		"""
}

adr15SchemaHygiene: s.#ADR & {
	slug:       "ADR-15"
	title:      "A Contract's service is a Container, a Need's domain is a Domain, a dep is typed by its role, and binding is dropped"
	status:     "proposed"
	body: """
		The schema accreted three ungated or rudimentary spots, all surfaced while
		dogfooding M-elem.

		Ungated typed refs. `#Contract.service` and `#Need.domain` were typed
		`#Node` (the any-node union), so CUE accepted `service: someADR` and the
		compiler only checked the target resolved, not its type — while everywhere
		else the model gates its edges (`satisfies`→atom, binding→Contract). Tighten
		the CUE: `service!: #Container`, `domain!: #Domain`. (`#Port.schema` stays
		`#Node` for now — there is no dedicated IDL node type to point it at.)

		A dep ignored its own role. A plain dep (no `role`) is a logical dependency
		and should target a `#Contract`; an infra dep (`role` set) is a physical
		touch and should target a `#Port | #Container`. Both were `#Node`, and a
		mis-typed infra target was silently skipped by topology. This is expressible
		via the sibling-field idiom (the `#Port if kind ==` pattern): gate `to` on
		whether `role` is set. The `#invariant` option (G2, element-grained deps) is
		preserved on the plain branch.

		`binding` is a rudiment — dropped. The field offered required|optional|
		abstract, but `optional` was dead (treated as `required`) and `abstract` was
		authored by no node — only one synthetic test. Its single effect was one
		fixpoint branch: an abstract contract never blocks and is never a gap. But a
		contract with no promise anyone keeps is not a Contract — everything
		speculative is already its own node type (ADR, Plan, with status proposed).
		Drop `binding`; every Contract without code is now an honest gap. (A closed
		enum exists only when the tool reasons about it; after removing the abstract
		branch there is nothing left to reason about.)

		Breaking across schema, model, source, fixpoint, jsonir, diff and describe;
		ships in the pre-release window. Closes M4.
		"""
}
