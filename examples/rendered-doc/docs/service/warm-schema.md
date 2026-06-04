---
title: Seed the editor's cue cache with the schema and the landscape's modules
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Seed the editor's cue cache with the schema and the landscape's modules

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** any verb that knows the active landscape calls into it; the caller can also ask for it explicitly

## Invariants

### <a id="registry-is-ephemeral"></a>registry-is-ephemeral

The registry that hosts the publish is started in this process and torn down once the cache has been populated; no daemon survives the call.

Decided by: [ADR-04](../governance/ADR-04.md)

*Implemented* (no test yet).

### <a id="schema-version-stays-fixed"></a>schema-version-stays-fixed

A change in the schema's contents is republished under the same version tag, so no module pin moves.

Decided by: [ADR-06](../governance/ADR-06.md)

*Implemented* (no test yet).

### <a id="no-op-when-current"></a>no-op-when-current

If the cache already holds the current schema and modules, the call is a no-op.

*Proven.*

### <a id="editor-resolves-natively"></a>editor-resolves-natively

After the call the editor's cue lsp resolves the schema, with fields offered while authoring.

Satisfies: [as-author-dx#fr-01](../domain/as-author-dx.md#fr-01)

*Implemented* (no test yet).

### <a id="cross-module-references-resolve"></a>cross-module-references-resolve

After the call the editor's cue lsp resolves cross-module references and offers go-to-definition between modules.

Satisfies: [as-author-dx#fr-02](../domain/as-author-dx.md#fr-02)

*Implemented* (no test yet).


## Postconditions

### —

The cache state on disk is sufficient for the editor to resolve without anything running in the background.


## Realizes

- [as-author-dx](../domain/as-author-dx.md)

