---
title: Read the active context
icon: material/play-circle-outline
tags:
    - usecase
    - implemented
---

# Read the active context

!!! info "Implemented — 0/2 proven"
    Some invariants still lack a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks which context is active

## Invariants

### <a id="does-not-mutate"></a>does-not-mutate

Reading the context does not alter it.

Satisfies: [as-agent-setup#fr-03](../domain/as-agent-setup.md#fr-03)

*Implemented* (no test yet).

### <a id="names-membership"></a>names-membership

The result names the context and every module it carries.

Satisfies: [as-agent-setup#fr-03](../domain/as-agent-setup.md#fr-03)

*Implemented* (no test yet).


## Postconditions

### —

If no context is active the caller is told so with the next step to take.


## Realizes

- [as-agent-setup](../domain/as-agent-setup.md)

