---
title: Render the resolved spec as a browsable markdown documentation tree
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Render the resolved spec as a browsable markdown documentation tree

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** an owner asks to render the spec as a documentation tree into a destination directory

## Invariants

### <a id="derived-from-resolved-graph"></a>derived-from-resolved-graph

Every file written is produced from the same resolved graph the other verbs read.

Decided by: [ADR-09](../governance/ADR-09.md)

*Proven.*

### <a id="destination-is-explicit"></a>destination-is-explicit

The caller names the destination directory the render writes into.

*Proven.*

### <a id="refuses-non-empty-destination"></a>refuses-non-empty-destination

**Rejects** when the destination directory exists and is not empty.

*Proven.*

### <a id="one-file-per-node"></a>one-file-per-node

Each spec node is rendered as one markdown file under <module>/<slug>.md; an index at the destination root lists the modules.

Satisfies: [as-federated-reader#fr-03](../domain/as-federated-reader.md#fr-03)

*Proven.*

### <a id="machine-readable-frontmatter"></a>machine-readable-frontmatter

*(returns)* Each file opens with YAML frontmatter carrying the node's type, module, status, confidence, decided_by and satisfies, plus the source revision the document was rendered from.

Decided by: [ADR-09](../governance/ADR-09.md)

*Proven.*

### <a id="cross-links-resolve-as-markdown"></a>cross-links-resolve-as-markdown

References between nodes are written as relative markdown links to the target file, with an anchor on a named element where one is addressed; the destination is self-contained — every link a reader follows stays inside the tree.

Satisfies: [as-federated-reader#fr-03](../domain/as-federated-reader.md#fr-03), [as-federated-owner#fr-04](../domain/as-federated-owner.md#fr-04)

*Proven.*

### <a id="format-is-explicit"></a>format-is-explicit

The caller selects the output format explicitly before the render runs.

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="json-emits-one-file-per-node-plus-index"></a>json-emits-one-file-per-node-plus-index

each node is written as one JSON file and a single index.json at the root carries the modules and a flat node list.

*When* the format is json

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="layout-is-chosen-per-run"></a>layout-is-chosen-per-run

the caller picks the layout (flat or tree) for that run.

*When* the format is markdown

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="tree-layout-mirrors-module-path"></a>tree-layout-mirrors-module-path

the module path is split into nested directories so the on-disk tree mirrors the logical module path.

*When* the tree layout is selected

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="strip-prefix-shortens-paths"></a>strip-prefix-shortens-paths

it is dropped from directory names and from the visible identifiers rendered in the body.

*When* the format is markdown and the caller strips a common module-path prefix

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="frontmatter-shape-is-chosen"></a>frontmatter-shape-is-chosen

the caller picks one frontmatter shape for the whole run from a fixed set the renderer knows.

*When* the format is markdown

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="nav-snippet-on-request"></a>nav-snippet-on-request

the renderer also writes an MkDocs-compatible nav: file mirroring the rendered tree.

*When* the caller names a nav-snippet path

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="index-pages-on-request"></a>index-pages-on-request

the renderer emits an index.md at every tree directory carrying a child listing and a status summary, suitable as a MkDocs Material section landing page.

*When* the caller asks for index pages and the layout is tree

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="tags-page-on-request"></a>tags-page-on-request

the renderer emits a tags.md at the root grouping every node by type and by status, with link labels carrying the module-qualified id and a status badge.

*When* the caller asks for a tags page

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="status-admonitions-on-request"></a>status-admonitions-on-request

every node page opens with a Material admonition block carrying the node's status and a one-line summary, and every requirement and invariant carries an inline status of its own.

*When* the caller asks for status admonitions

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="node-name-collisions-resolved"></a>node-name-collisions-resolved

the node is written as that directory's index.md so its content is not hidden by an auto-generated index in URL routing.

*When* a node's slug equals the name of a sibling directory holding other nodes

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="link-text-is-slug-only"></a>link-text-is-slug-only

The visible text of a cross-link in the body carries only the target's slug — with #element for an atom — while the URL keeps the full module-qualified path.

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="defaults-need-no-flags"></a>defaults-need-no-flags

Running render with no customisation flags produces a sensible publishable tree — the default needs no flag to be useful.

Decided by: [ADR-12](../governance/ADR-12.md)

*Proven.*

### <a id="returns-self-contained-tree"></a>returns-self-contained-tree

*(returns)* The destination holds a self-contained, internally linked documentation tree the audience reads through a markdown forge — or a per-node JSON tree a downstream pipeline transforms further.

*Proven.*


## Realizes

- [as-federated-reader](../domain/as-federated-reader.md)
- [as-federated-owner](../domain/as-federated-owner.md)

