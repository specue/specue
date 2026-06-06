---
title: Report conflicts between two open Plans
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Report conflicts between two open Plans

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks whether two Plans conflict

## Invariants

### <a id="structural-conflict-blocks"></a>structural-conflict-blocks

the pair is reported as blocking.

*When* overlaying both Plans together produces a graph that cannot resolve (a removed node is referenced, the same edge is rewired two ways)

Satisfies: [as-planner#fr-03](../domain/as-planner.md#fr-03)

Decided by: [ADR-07](../governance/ADR-07.md)

*Proven.*

### <a id="co-touch-surfaces-for-review"></a>co-touch-surfaces-for-review

they are reported as advisory for human or agent review, not blocked.

*When* two Plans touch the same Contract or Port but both apply cleanly

Satisfies: [as-planner#fr-04](../domain/as-planner.md#fr-04)

*Implemented* (no test yet).

### <a id="conflict-names-plans-and-node"></a>conflict-names-plans-and-node

*(returns)* Each conflict names the two Plans, the shared Contract or Port, and whether it is blocking or advisory.

*Implemented* (no test yet).


## Realizes

- [as-planner](../domain/as-planner.md)

