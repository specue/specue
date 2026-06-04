---
title: Bring a new module into existence
icon: material/clipboard-text-outline
tags:
    - need
    - partial
---

# Bring a new module into existence

!!! warning "Partial — 2/2 covered"
    Some requirements have no proven contract.

Domain: [specue](specue.md)

**Consumer:** an agent extending the landscape

to start a new module of a known kind, so that subsequent authoring has a place to live

## Requirements

### <a id="fr-01"></a>fr-01

A new module declares its identity and its kind (service, domain, governance or code) at creation.

*Covered by [init-module#identity-and-kind-at-creation](../service/init-module.md#identity-and-kind-at-creation)*

### <a id="fr-02"></a>fr-02

Creating a module over an existing one is refused.

*Covered by [init-module#no-overwrite](../service/init-module.md#no-overwrite)*

