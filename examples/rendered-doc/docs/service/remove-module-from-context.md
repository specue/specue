---
title: Remove a module from a context
icon: material/play-circle-outline
tags:
    - usecase
    - implemented
---

# Remove a module from a context

!!! info "Implemented — 0/2 proven"
    Some invariants still lack a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to remove a module from the current context

## Invariants

### <a id="addressed-by-module-path"></a>addressed-by-module-path

The module is removed by its module path, which is unique within the context.

Satisfies: [as-agent-setup#fr-02](../domain/as-agent-setup.md#fr-02)

*Implemented* (no test yet).

### <a id="module-untouched"></a>module-untouched

The directory the module lives in is left as it was.

*Implemented* (no test yet).


## Postconditions

### —

The module is no longer reachable from the context until it is added again.


## Realizes

- [as-agent-setup](../domain/as-agent-setup.md)

