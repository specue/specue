---
title: Publish a spec module's binding outcomes for readers without code access
icon: material/play-circle-outline
tags:
    - contract
    - asserted
---

# Publish a spec module's binding outcomes for readers without code access

!!! warning "Asserted"
    The contract is agreed; no code realises it yet.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the holder of the code asks to publish an attestation for a spec module

## Invariants

### <a id="attestation-carries-outcomes-only"></a>attestation-carries-outcomes-only

*(returns)* The attestation carries only the binding outcome per Contract and the file:line references.

Satisfies: [as-federated-reader#fr-02](../domain/as-federated-reader.md#fr-02)

Decided by: [ADR-08](../governance/ADR-08.md)

**Unbound.**

### <a id="lives-with-the-spec"></a>lives-with-the-spec

The attestation is committed alongside the spec module it attests, in the spec module's own repository.

Decided by: [ADR-08](../governance/ADR-08.md)

**Unbound.**

### <a id="status-identical-to-scan"></a>status-identical-to-scan

A Contract's status computed from the attestation matches the status computed from scanning the code.

Satisfies: [as-federated-reader#fr-01](../domain/as-federated-reader.md#fr-01)

Decided by: [ADR-08](../governance/ADR-08.md)

**Unbound.**


## Realizes

- [as-federated-reader](../domain/as-federated-reader.md)

