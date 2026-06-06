---
title: Leave a Plan and return to the base branch
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Leave a Plan and return to the base branch

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to leave the current Plan

## Invariants

### <a id="every-module-returns"></a>every-module-returns

Every module that was switched into the Plan is checked out back to the base branch.

Decided by: [ADR-07](../governance/ADR-07.md)

*Proven.*

### <a id="refuses-on-dirty-tree"></a>refuses-on-dirty-tree

**Rejects** when any affected module's working tree carries uncommitted changes.

*Implemented* (no test yet).

### <a id="authoring-lands-on-base"></a>authoring-lands-on-base

Subsequent authoring lands on the base branch until another Plan is used.

*Implemented* (no test yet).


