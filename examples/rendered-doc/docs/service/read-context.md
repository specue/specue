---
title: Read the active context
icon: material/play-circle-outline
tags:
    - contract
    - implemented
---

# Read the active context

!!! info "Implemented — 0/2 proven"
    Some invariants still lack a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks which context is active

## Invariants

### <a id="read-returns-current-state"></a>read-returns-current-state

*(returns)* Reading the context returns its current state — the same name and module set on every read.

Satisfies: [as-agent-setup#fr-03](../domain/as-agent-setup.md#fr-03)

Decided by: [ADR-14](../governance/ADR-14.md)

*Implemented* (no test yet).

### <a id="names-membership"></a>names-membership

*(returns)* The result names the context and every module it carries.

Satisfies: [as-agent-setup#fr-03](../domain/as-agent-setup.md#fr-03)

*Implemented* (no test yet).

### <a id="no-active-context-told"></a>no-active-context-told

**Rejects** when no context is active.

*Implemented* (no test yet).


## Realizes

- [as-agent-setup](../domain/as-agent-setup.md)

