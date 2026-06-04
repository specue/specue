---
title: Read the spec without holding the code
icon: material/clipboard-text-outline
tags:
    - need
    - partial
---

# Read the spec without holding the code

!!! warning "Partial — 1/3 covered"
    Some requirements have no proven contract.

Domain: [specue](specue.md)

**Consumer:** a reader who has the spec but not the code that realizes it

to see the same UseCases, their statuses and their code-binding outcomes a holder of the code would see, so that I can review and reason about the system across access boundaries

## Requirements

### <a id="fr-01"></a>fr-01

A UseCase's status is determined the same way whether or not the code is reachable.

*Claimed by [attest-bindings#status-identical-to-scan](../service/attest-bindings.md#status-identical-to-scan)* — not proven

### <a id="fr-02"></a>fr-02

Source content of the code is never required to render the spec.

*Claimed by [attest-bindings#no-source-in-attestation](../service/attest-bindings.md#no-source-in-attestation)* — not proven

### <a id="fr-03"></a>fr-03

The spec is consumable as a human-readable document without running the tool.

*Covered by [render-doc#one-file-per-node](../service/render-doc.md#one-file-per-node)* (+1 more)

