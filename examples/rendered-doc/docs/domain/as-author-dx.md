---
title: Author with editor support
icon: material/clipboard-text-outline
tags:
    - need
    - partial
---

# Author with editor support

!!! warning "Partial — 2/2 covered"
    Some requirements have no proven contract.

Domain: [specue](specue.md)

**Consumer:** a human authoring spec modules in an editor

the editor to understand what I am writing as I write it, so that I produce valid contracts without leaving the file

## Requirements

### <a id="fr-01"></a>fr-01

The fields a Contract, Need, ADR or Port expects are offered while authoring it.

*Covered by [warm-schema#editor-resolves-natively](../service/warm-schema.md#editor-resolves-natively)*

### <a id="fr-02"></a>fr-02

A reference to a Contract, Need, ADR or Port in another module is navigable to its definition.

*Covered by [warm-schema#cross-module-references-resolve](../service/warm-schema.md#cross-module-references-resolve)*

