---
title: Choose which spec context I am working against
icon: material/clipboard-text-outline
tags:
    - need
    - partial
---

# Choose which spec context I am working against

!!! warning "Partial — 3/3 covered"
    Some requirements have no proven contract.

Domain: [specue](specue.md)

**Consumer:** an agent starting work on this machine

to declare which spec context I am working against right now, so that every subsequent command resolves against the same picture without me passing the set each time

## Requirements

### <a id="fr-01"></a>fr-01

A named context can be created and switched between.

*Covered by [create-context#name-is-unique](../service/create-context.md#name-is-unique)* (+1 more)

### <a id="fr-02"></a>fr-02

A module is added to or removed from the current context by its directory.

*Claimed by [remove-module-from-context#addressed-by-module-path](../service/remove-module-from-context.md#addressed-by-module-path)* (+1 more) — not proven

### <a id="fr-03"></a>fr-03

The current context is readable on demand.

*Claimed by [read-context#read-returns-current-state](../service/read-context.md#read-returns-current-state)* (+1 more) — not proven

