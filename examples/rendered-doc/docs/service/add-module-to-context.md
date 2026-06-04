---
title: Add a module to a context by its directory
icon: material/play-circle-outline
tags:
    - usecase
    - implemented
---

# Add a module to a context by its directory

!!! info "Implemented — 0/3 proven"
    Some invariants still lack a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to add a module to the current context

## Invariants

### <a id="addressed-by-directory"></a>addressed-by-directory

The module is addressed by the directory that holds its manifest, not by its name.

Satisfies: [as-agent-setup#fr-02](../domain/as-agent-setup.md#fr-02), [as-federated-owner#fr-02](../domain/as-federated-owner.md#fr-02)

*Implemented* (no test yet).

### <a id="must-be-a-module"></a>must-be-a-module

Adding a directory that does not hold a module manifest is refused with the next step to take.

*Implemented* (no test yet).

### <a id="git-repository-required"></a>git-repository-required

Adding a module that does not live in a git repository is refused with the next step to take.

Decided by: [ADR-03](../governance/ADR-03.md)

*Implemented* (no test yet).


## Postconditions

### —

The module is reachable from the context until it is removed.


## Realizes

- [as-agent-setup](../domain/as-agent-setup.md)
- [as-federated-owner](../domain/as-federated-owner.md)

