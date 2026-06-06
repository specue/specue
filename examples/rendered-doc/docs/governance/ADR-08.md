---
title: Code-binding outcomes are published alongside the spec for readers without code access
icon: material/gavel
tags:
    - adr
    - accepted
---

# Code-binding outcomes are published alongside the spec for readers without code access

!!! note "Accepted"
    This decision is in effect.

The spec and the code that realizes it sit on different sides of an access
boundary in many real systems: a reader may hold one and not the other.
Whoever holds the code publishes a small attestation artifact alongside its
spec module — the binding outcomes per Contract, no source — and a reader
consumes that instead of scanning. Status is computed by the same rules in
both paths, so a reader and a code holder see the same picture and the
federated boundary becomes invisible at the spec level.
