---
title: Report the typed delta between the spec at two versioned points
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Report the typed delta between the spec at two versioned points

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks for the difference between two refs of a module

## Invariants

### <a id="typed-over-the-spec-graph"></a>typed-over-the-spec-graph

The delta is over Contracts, Needs, Ports and their elements.

Satisfies: [as-agent-review#fr-01](../domain/as-agent-review.md#fr-01)

*Proven.*

### <a id="every-change-named"></a>every-change-named

*(returns)* Each change is labelled added, removed, modified or rewired.

Satisfies: [as-agent-review#fr-02](../domain/as-agent-review.md#fr-02)

*Proven.*

### <a id="two-snapshots"></a>two-snapshots

The diff is computed between two snapshots produced from the refs the caller named.

*Implemented* (no test yet).

### <a id="returns-delta-with-refs"></a>returns-delta-with-refs

*(returns)* The delta is returned together with the two refs it was computed against.

*Implemented* (no test yet).


## Realizes

- [as-agent-review](../domain/as-agent-review.md)

