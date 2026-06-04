---
title: Switch the working tree into a Plan
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Switch the working tree into a Plan

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to work on a Plan

## Invariants

### <a id="checks-out-every-branch"></a>checks-out-every-branch

Every module the Plan touches is checked out onto the Plan's branch in a single step.

Decided by: [ADR-07](../governance/ADR-07.md)

*Proven.*

### <a id="refuses-on-dirty-tree"></a>refuses-on-dirty-tree

Switching is refused when any affected module's working tree carries uncommitted changes, so nothing is silently overwritten.

*Proven.*


## Postconditions

### —

Subsequent authoring lands on the Plan's branches until the caller returns to base.


