---
title: Publish a spec module's binding outcomes for readers without code access
icon: material/play-circle-outline
tags:
    - usecase
    - asserted
---

# Publish a spec module's binding outcomes for readers without code access

!!! warning "Asserted"
    The contract is agreed; no code realises it yet.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the holder of the code asks to publish an attestation for a spec module

## Invariants

### <a id="no-source-in-attestation"></a>no-source-in-attestation

The attestation carries the binding outcome per UseCase and the file:line references; the source content of the code is never included.

Satisfies: [as-federated-reader#fr-02](../domain/as-federated-reader.md#fr-02)

Decided by: [ADR-08](../governance/ADR-08.md)

**Unbound.**

### <a id="lives-with-the-spec"></a>lives-with-the-spec

The attestation is committed alongside the spec module it attests, in the spec module's own repository.

Decided by: [ADR-08](../governance/ADR-08.md)

**Unbound.**

### <a id="status-identical-to-scan"></a>status-identical-to-scan

A UseCase's status computed from the attestation matches the status computed from scanning the code.

Satisfies: [as-federated-reader#fr-01](../domain/as-federated-reader.md#fr-01)

Decided by: [ADR-08](../governance/ADR-08.md)

**Unbound.**


## Postconditions

### —

An audience that holds the spec but not the code can read it and see the same statuses.

Satisfies: [as-federated-owner#fr-04](../domain/as-federated-owner.md#fr-04)


## Realizes

- [as-federated-owner](../domain/as-federated-owner.md)
- [as-federated-reader](../domain/as-federated-reader.md)

