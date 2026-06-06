---
title: Produce a resolved spec graph from the current context
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Produce a resolved spec graph from the current context

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** any verb that needs the graph asks for it

## Invariants

### <a id="cue-stitches-the-modules"></a>cue-stitches-the-modules

Every cross-module reference, version pin and visibility rule is resolved by CUE before the graph is handed back.

Satisfies: [as-federated-owner#fr-01](../domain/as-federated-owner.md#fr-01), [as-federated-owner#fr-03](../domain/as-federated-owner.md#fr-03)

Decided by: [ADR-01](../governance/ADR-01.md)

*Implemented* (no test yet).

### <a id="incremental"></a>incremental

the graph is rebuilt

*When* the spec or the code that feeds it has changed since the last build

*Implemented* (no test yet).

### <a id="multi-folder-modules"></a>multi-folder-modules

A module's nodes are loaded from every sub-folder of the module.

Satisfies: [as-agent-create#fr-03](../domain/as-agent-create.md#fr-03)

*Proven.*

### <a id="returns-graph-and-diagnostics"></a>returns-graph-and-diagnostics

*(returns)* The resolved graph is returned together with diagnostics produced while resolving it.

*Implemented* (no test yet).


## Realizes

- [as-federated-owner](../domain/as-federated-owner.md)
- [as-agent-create](../domain/as-agent-create.md)

