---
title: Answer a graph query with read-only SQL
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Answer a graph query with read-only SQL

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller runs a query against the graph

## Invariants

### <a id="runs-against-projection"></a>runs-against-projection

The query runs against a projection of the graph, not the graph itself.

Satisfies: [as-agent-navigate#fr-03](../domain/as-agent-navigate.md#fr-03)

Decided by: [ADR-02](../governance/ADR-02.md)

*Proven.*

### <a id="cannot-mutate"></a>cannot-mutate

The query cannot mutate the graph.

Decided by: [ADR-02](../governance/ADR-02.md)

*Proven.*

### <a id="schema-is-discoverable"></a>schema-is-discoverable

The shape of the projection (its tables and columns) is retrievable by the caller without prior knowledge.

Decided by: [ADR-02](../governance/ADR-02.md)

*Implemented* (no test yet).

### <a id="matches-stated-criterion"></a>matches-stated-criterion

Nodes matching a criterion stated in the query are returned without the caller naming each one.

Satisfies: [as-agent-navigate#fr-04](../domain/as-agent-navigate.md#fr-04)

*Implemented* (no test yet).

### <a id="pre-joined-views"></a>pre-joined-views

The projection exposes pre-joined views for the questions a caller asks most often (a node with its elements, a story FR with the contracts that cover it), so common reads are one statement instead of a chain of joins.

Satisfies: [as-agent-navigate#fr-04](../domain/as-agent-navigate.md#fr-04)

*Proven.*


## Postconditions

### —

Matching rows are returned as machine-readable data.


## Realizes

- [as-agent-navigate](../domain/as-agent-navigate.md)

