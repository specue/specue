---
title: See how a Plan lands before committing to it
icon: material/clipboard-text-outline
tags:
    - need
    - partial
---

# See how a Plan lands before committing to it

!!! warning "Partial — 6/6 covered"
    Some requirements have no proven contract.

Domain: [specue](specue.md)

**Consumer:** an agent proposing a Plan that is not yet accepted

to see how the Plan lands on the current spec, so that I can decide whether to accept it without breaking the system or other open Plans

## Requirements

### <a id="fr-01"></a>fr-01

A Plan is a named, retrievable object distinct from the current spec.

*Covered by [register-plan#plan-is-a-branch-set](../service/register-plan.md#plan-is-a-branch-set)*

### <a id="fr-02"></a>fr-02

A Plan is viewable against the current spec without altering the working tree.

*Covered by [pending-overlay#viewed-without-checkout](../service/pending-overlay.md#viewed-without-checkout)*

### <a id="fr-03"></a>fr-03

Two Plans whose changes cannot both apply (one removes what the other modifies, both rewire the same edge, etc.) are blocked before either is accepted.

*Covered by [detect-conflict#structural-conflict-blocks](../service/detect-conflict.md#structural-conflict-blocks)*

### <a id="fr-04"></a>fr-04

Two Plans that touch the same Contract or Port but could both apply are surfaced for human or agent review rather than blocked.

*Covered by [detect-conflict#co-touch-surfaces-for-review](../service/detect-conflict.md#co-touch-surfaces-for-review)*

### <a id="fr-05"></a>fr-05

Accepting a Plan applies its changes to the current spec and closes the Plan.

*Covered by [accept-plan#merge-only-if-valid](../service/accept-plan.md#merge-only-if-valid)* (+3 more)

### <a id="fr-06"></a>fr-06

Planning requires a governance module in the current context; without one, the verb refuses with the next step to take.

*Covered by [register-plan#governance-required](../service/register-plan.md#governance-required)*

