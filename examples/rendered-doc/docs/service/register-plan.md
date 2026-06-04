---
title: Open a new Plan as a Plan record plus branches
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Open a new Plan as a Plan record plus branches

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to register a new Plan

## Invariants

### <a id="plan-is-a-branch-set"></a>plan-is-a-branch-set

A new Plan creates identically-named branches in every module it touches and a Plan record in the governance module pointing at them.

Satisfies: [as-planner#fr-01](../domain/as-planner.md#fr-01), [as-decision-keeper#fr-02](../domain/as-decision-keeper.md#fr-02)

Decided by: [ADR-07](../governance/ADR-07.md)

*Proven.*

### <a id="governance-required"></a>governance-required

Without a governance module in the current context, registering a Plan is refused with the next step to take.

Satisfies: [as-planner#fr-06](../domain/as-planner.md#fr-06)

Decided by: [ADR-07](../governance/ADR-07.md)

*Implemented* (no test yet).

### <a id="no-overwrite"></a>no-overwrite

Registering a Plan whose name is already taken is refused; the existing Plan is left untouched.

*Implemented* (no test yet).

### <a id="from-base-only"></a>from-base-only

Registering a Plan from any branch other than the landscape's base branch is refused with the next step to take, so a Plan always forks from a known base and never from another Plan.

*Proven.*


## Postconditions

### —

The Plan record names the branches it points at and the modules they live in.


## Realizes

- [as-planner](../domain/as-planner.md)
- [as-decision-keeper](../domain/as-decision-keeper.md)

