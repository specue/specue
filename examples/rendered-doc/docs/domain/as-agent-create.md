---
title: Add a UseCase, Need, ADR or Port to a module
icon: material/clipboard-text-outline
tags:
    - need
    - partial
---

# Add a UseCase, Need, ADR or Port to a module

!!! warning "Partial — 2/3 covered"
    Some requirements have no proven contract.

Domain: [specue](specue.md)

**Consumer:** an agent authoring inside a module

to add a UseCase, Need, ADR or Port and arrange it where it belongs, so that the module carries the contracts and needs it owns, in a structure I can navigate

## Requirements

### <a id="fr-01"></a>fr-01

A new node carries an identity that is unique within its module.

*Covered by [validate-graph#unique-slug-within-module](../service/validate-graph.md#unique-slug-within-module)*

### <a id="fr-02"></a>fr-02

The node kinds a module of a given kind may hold are visible in the schema the modules import.

**Uncovered.**

### <a id="fr-03"></a>fr-03

Nodes within a module can be organized into sub-folders.

*Covered by [build-graph#multi-folder-modules](../service/build-graph.md#multi-folder-modules)*

