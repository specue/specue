---
title: Find my way around an unfamiliar spec
icon: material/clipboard-text-outline
tags:
    - need
    - covered
---

# Find my way around an unfamiliar spec

!!! success "Covered — 4/4"
    Every requirement is satisfied by a proven contract.

Domain: [specue](specue.md)

**Consumer:** an agent exploring a spec I did not author

to find my way around it on demand, so that I can answer questions about the system without reading every file

## Requirements

### <a id="fr-01"></a>fr-01

A Contract, Need, Port, ADR or code binding can each be listed.

*Covered by [list-resources#kinds-listed-without-prior-knowledge](../service/list-resources.md#kinds-listed-without-prior-knowledge)* (+1 more)

### <a id="fr-02"></a>fr-02

Any one of them can be read in full by its module-qualified identity, which is stable across the landscape.

*Covered by [describe-node#identity-is-module-qualified](../service/describe-node.md#identity-is-module-qualified)* (+2 more)

### <a id="fr-03"></a>fr-03

How they are related to each other is retrievable as machine-readable data.

*Covered by [query-graph#runs-against-projection](../service/query-graph.md#runs-against-projection)*

### <a id="fr-04"></a>fr-04

Nodes matching a stated criterion can be found without naming each one.

*Covered by [query-graph#matches-stated-criterion](../service/query-graph.md#matches-stated-criterion)* (+1 more)

