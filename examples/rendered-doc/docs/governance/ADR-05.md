---
title: Code is a first-class module of the landscape
icon: material/gavel
tags:
    - adr
    - accepted
---

# Code is a first-class module of the landscape

!!! note "Accepted"
    This decision is in effect.

A code module is just another module — manifest, requires, kind: code — that
holds no spec nodes of its own and declares which contracts its source may
bind through its requires. The previous generation treated code as a sidecar
attached to a spec module; promoting it makes code participate in the same
mechanisms (contexts, plans, validation) the rest of the graph uses, with no
special cases.
