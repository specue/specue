---
title: Start a new module of a known kind
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Start a new module of a known kind

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to scaffold a new module at a directory

## Invariants

### <a id="identity-and-kind-at-creation"></a>identity-and-kind-at-creation

A new module declares its identity and its kind (service, product, governance or code) when it is created.

Satisfies: [as-agent-start#fr-01](../domain/as-agent-start.md#fr-01)

*Implemented* (no test yet).

### <a id="no-overwrite"></a>no-overwrite

**Rejects** when a module already exists at the target directory.

Satisfies: [as-agent-start#fr-02](../domain/as-agent-start.md#fr-02)

*Proven.*

### <a id="git-repository-required"></a>git-repository-required

**Rejects** when the target is outside a git repository.

Decided by: [ADR-03](../governance/ADR-03.md)

*Proven.*

### <a id="scaffolds-manifest-only"></a>scaffolds-manifest-only

*(returns)* The new module is left as a directory with the manifest the kind requires and nothing else.

*Implemented* (no test yet).


## Realizes

- [as-agent-start](../domain/as-agent-start.md)

