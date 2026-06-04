---
title: Apply a Plan to the current spec and close it
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Apply a Plan to the current spec and close it

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to accept a Plan

## Invariants

### <a id="merge-only-if-valid"></a>merge-only-if-valid

The Plan is accepted only when overlaying it on the current spec produces a graph that validates; otherwise the merge is refused and nothing is changed.

Satisfies: [as-planner#fr-05](../domain/as-planner.md#fr-05)

Decided by: [ADR-07](../governance/ADR-07.md)

*Proven.*

### <a id="branches-merged-everywhere"></a>branches-merged-everywhere

Acceptance merges the Plan's branches into the base branch in every module it touches.

Satisfies: [as-planner#fr-05](../domain/as-planner.md#fr-05)

Decided by: [ADR-07](../governance/ADR-07.md)

*Proven.*

### <a id="plan-record-closes"></a>plan-record-closes

Once merged, the Plan's record in the governance module is closed.

Satisfies: [as-planner#fr-05](../domain/as-planner.md#fr-05)

*Proven.*

### <a id="works-from-anywhere"></a>works-from-anywhere

Acceptance succeeds regardless of which branch the caller is currently on: a repo found on the Plan's branch is switched to base before merging, so the caller does not have to leave the Plan to land it.

*Proven.*

### <a id="tags-the-landing"></a>tags-the-landing

Acceptance marks the merge commit of every affected repo with a tag named after the Plan, so a reader of git history can enumerate landed Plans without parsing the commit graph.

*Proven.*


## Postconditions

### —

On refusal the working tree is left exactly as it was before the attempt.


## Realizes

- [as-planner](../domain/as-planner.md)

