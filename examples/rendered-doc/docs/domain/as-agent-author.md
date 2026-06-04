---
title: Know whether what I authored is correct
icon: material/clipboard-text-outline
tags:
    - need
    - covered
---

# Know whether what I authored is correct

!!! success "Covered — 3/3"
    Every requirement is satisfied by a proven contract.

Domain: [specue](specue.md)

**Consumer:** an agent authoring a spec and binding it to code

to know whether what I just wrote is correct, so that I can iterate without re-reading the landscape between changes

## Requirements

### <a id="fr-01"></a>fr-01

The spec as a whole is reported as correct or broken in a single check.

*Covered by [validate-graph#single-verdict](../service/validate-graph.md#single-verdict)*

### <a id="fr-02"></a>fr-02

For a code module, every UseCase it may realize and its current binding state are listed.

*Covered by [report-bindings#allowed-from-require-closure](../service/report-bindings.md#allowed-from-require-closure)* (+1 more)

### <a id="fr-03"></a>fr-03

Every failure carries the next step the caller takes to resolve it.

*Covered by [validate-graph](../service/validate-graph.md)*

