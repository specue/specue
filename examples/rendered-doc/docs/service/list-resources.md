---
title: List the kinds of node the spec holds, and the nodes of one kind
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# List the kinds of node the spec holds, and the nodes of one kind

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks what kinds the spec holds, or to list one kind

## Invariants

### <a id="kinds-listed-without-prior-knowledge"></a>kinds-listed-without-prior-knowledge

The caller can ask which node kinds exist without naming them.

Satisfies: [as-agent-navigate#fr-01](../domain/as-agent-navigate.md#fr-01)

*Proven.*

### <a id="nodes-of-a-kind"></a>nodes-of-a-kind

Given a node kind, every node of that kind in the current spec is returned.

Satisfies: [as-agent-navigate#fr-01](../domain/as-agent-navigate.md#fr-01)

*Proven.*

### <a id="stable-result-shape"></a>stable-result-shape

*(returns)* The result is one stable shape whether the caller asks for the kinds or for the nodes of one kind.

*Implemented* (no test yet).


## Realizes

- [as-agent-navigate](../domain/as-agent-navigate.md)

