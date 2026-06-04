---
title: Show a Plan against the current spec without switching the working tree
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Show a Plan against the current spec without switching the working tree

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to view a Plan against the current spec

## Invariants

### <a id="viewed-without-checkout"></a>viewed-without-checkout

The Plan is projected onto the current spec by reading its branches through git, without touching the working tree.

Satisfies: [as-planner#fr-02](../domain/as-planner.md#fr-02)

Decided by: [ADR-07](../governance/ADR-07.md)

*Proven.*

### <a id="base-side-read-through-git"></a>base-side-read-through-git

The base side of the overlay is read through git from the base branch, not from the working tree; the overlay is the same regardless of which branch is currently checked out.

Decided by: [ADR-07](../governance/ADR-07.md)

*Proven.*

### <a id="overlay-is-a-spec"></a>overlay-is-a-spec

The overlay result is a spec graph with the same shape as the live one, so any read verb works against it.

*Implemented* (no test yet).


## Postconditions

### —

The overlay is returned with the refs and the modules it composed.


## Realizes

- [as-planner](../domain/as-planner.md)

