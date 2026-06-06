// Package federation holds the contracts that span access boundaries: publishing
// binding outcomes for readers without code, and seeding the editor's cue cache.
package federation

import (
	s "specue.io/schema@v0:spec"
	root "specue.io/service@v0:service"
	humanaud "specue.io/domain/human@v0:human"
	fed "specue.io/domain/federated@v0:federated"
	gov "specue.io/governance@v0:governance"
)

attestBindings: s.#Contract & {
	slug:        "attest-bindings"
	title:       "Publish a spec module's binding outcomes for readers without code access"
	service:     root.specue
	trigger:     "the holder of the code asks to publish an attestation for a spec module"
	invariants: [{
		id:   "attestation-carries-outcomes-only"
		kind: "returns"
		text: "The attestation carries only the binding outcome per Contract and the file:line references."
		satisfies: [fed.reader.frs."fr-02"]
		decided_by: [gov.adr08AttestedBindings]
	}, {
		id:   "lives-with-the-spec"
		text: "The attestation is committed alongside the spec module it attests, in the spec module's own repository."
		decided_by: [gov.adr08AttestedBindings]
	}, {
		id:   "status-identical-to-scan"
		text: "A Contract's status computed from the attestation matches the status computed from scanning the code."
		satisfies: [fed.reader.frs."fr-01"]
		decided_by: [gov.adr08AttestedBindings]
	}]
}

renderDoc: s.#Contract & {
	slug:        "render-doc"
	title:       "Render the resolved spec as a browsable markdown documentation tree"
	service:     root.specue
	trigger:     "an owner asks to render the spec as a documentation tree into a destination directory"
	invariants: [{
		id:   "derived-from-resolved-graph"
		text: "Every file written is produced from the same resolved graph the other verbs read."
		decided_by: [gov.adr09RenderedDocDerived]
	}, {
		id:   "destination-is-explicit"
		text: "The caller names the destination directory the render writes into."
	}, {
		id:   "refuses-non-empty-destination"
		kind: "rejects"
		when: "the destination directory exists and is not empty"
	}, {
		id:   "one-file-per-node"
		text: "Each spec node is rendered as one markdown file under <module>/<slug>.md; an index at the destination root lists the modules."
		satisfies: [fed.reader.frs."fr-03"]
	}, {
		id:   "machine-readable-frontmatter"
		kind: "returns"
		text: "Each file opens with YAML frontmatter carrying the node's type, module, status, confidence, decided_by and satisfies, plus the source revision the document was rendered from."
		decided_by: [gov.adr09RenderedDocDerived]
	}, {
		id:   "cross-links-resolve-as-markdown"
		text: "References between nodes are written as relative markdown links to the target file, with an anchor on a named element where one is addressed; the destination is self-contained — every link a reader follows stays inside the tree."
		satisfies: [fed.reader.frs."fr-03", fed.owner.frs."fr-04"]
	}, {
		id:   "format-is-explicit"
		text: "The caller selects the output format explicitly before the render runs."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "json-emits-one-file-per-node-plus-index"
		when: "the format is json"
		text: "each node is written as one JSON file and a single index.json at the root carries the modules and a flat node list."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "layout-is-chosen-per-run"
		when: "the format is markdown"
		text: "the caller picks the layout (flat or tree) for that run."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "tree-layout-mirrors-module-path"
		when: "the tree layout is selected"
		text: "the module path is split into nested directories so the on-disk tree mirrors the logical module path."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "strip-prefix-shortens-paths"
		when: "the format is markdown and the caller strips a common module-path prefix"
		text: "it is dropped from directory names and from the visible identifiers rendered in the body."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "frontmatter-shape-is-chosen"
		when: "the format is markdown"
		text: "the caller picks one frontmatter shape for the whole run from a fixed set the renderer knows."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "nav-snippet-on-request"
		when: "the caller names a nav-snippet path"
		text: "the renderer also writes an MkDocs-compatible nav: file mirroring the rendered tree."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "index-pages-on-request"
		when: "the caller asks for index pages and the layout is tree"
		text: "the renderer emits an index.md at every tree directory carrying a child listing and a status summary, suitable as a MkDocs Material section landing page."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "tags-page-on-request"
		when: "the caller asks for a tags page"
		text: "the renderer emits a tags.md at the root grouping every node by type and by status, with link labels carrying the module-qualified id and a status badge."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "status-admonitions-on-request"
		when: "the caller asks for status admonitions"
		text: "every node page opens with a Material admonition block carrying the node's status and a one-line summary, and every requirement and invariant carries an inline status of its own."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "node-name-collisions-resolved"
		when: "a node's slug equals the name of a sibling directory holding other nodes"
		text: "the node is written as that directory's index.md so its content is not hidden by an auto-generated index in URL routing."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "link-text-is-slug-only"
		text: "The visible text of a cross-link in the body carries only the target's slug — with #element for an atom — while the URL keeps the full module-qualified path."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "defaults-need-no-flags"
		text: "Running render with no customisation flags produces a sensible publishable tree — the default needs no flag to be useful."
		decided_by: [gov.adr12RenderFormatsAndPresets]
	}, {
		id:   "returns-self-contained-tree"
		kind: "returns"
		text: "The destination holds a self-contained, internally linked documentation tree the audience reads through a markdown forge — or a per-node JSON tree a downstream pipeline transforms further."
	}]
}

warmSchema: s.#Contract & {
	slug:        "warm-schema"
	title:       "Seed the editor's cue cache with the schema and the landscape's modules"
	service:     root.specue
	trigger:     "any verb that knows the active landscape calls into it; the caller can also ask for it explicitly"
	invariants: [{
		id:   "registry-is-ephemeral"
		text: "The registry that hosts the publish is started in this process and torn down once the cache has been populated."
		decided_by: [gov.adr04RegistryInProcess]
	}, {
		id:   "schema-version-stays-fixed"
		text: "A change in the schema's contents is republished under the same version tag, so no module pin moves."
		decided_by: [gov.adr06FixedSchemaVersion]
	}, {
		id:   "no-op-when-current"
		when: "the cache already holds the current schema and modules"
		text: "the call is a no-op."
	}, {
		id:   "editor-resolves-natively"
		text: "After the call the editor's cue lsp resolves the schema, with fields offered while authoring."
		satisfies: [humanaud.editorDX.frs."fr-01"]
	}, {
		id:   "cross-module-references-resolve"
		text: "After the call the editor's cue lsp resolves cross-module references and offers go-to-definition between modules."
		satisfies: [humanaud.editorDX.frs."fr-02"]
	}, {
		id:   "cache-self-sufficient-on-disk"
		text: "The cache state on disk is sufficient for the editor to resolve with nothing running in the background."
	}]
}
