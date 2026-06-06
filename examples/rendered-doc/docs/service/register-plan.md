---
title: Open a new Plan as a Plan record plus branches
icon: material/play-circle-outline
tags:
    - contract
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

**Rejects** when there is no governance module in the current context.

Satisfies: [as-planner#fr-06](../domain/as-planner.md#fr-06)

Decided by: [ADR-07](../governance/ADR-07.md)

*Implemented* (no test yet).

### <a id="no-overwrite"></a>no-overwrite

**Rejects** when a Plan with that name is already taken.

*Implemented* (no test yet).

### <a id="from-base-only"></a>from-base-only

**Rejects** when registering from any branch other than the landscape's base branch (so a Plan always forks from a known base, never another Plan).

*Proven.*

### <a id="record-names-branches"></a>record-names-branches

*(returns)* The Plan record names the branches it points at and the modules they live in.

*Implemented* (no test yet).


## Realizes

- [as-planner](../domain/as-planner.md)
- [as-decision-keeper](../domain/as-decision-keeper.md)

