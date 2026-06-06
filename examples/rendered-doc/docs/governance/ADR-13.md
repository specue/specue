---
title: The contract node is named Contract, not UseCase (nor Capability)
icon: material/gavel
tags:
    - adr
    - accepted
---

# The contract node is named Contract, not UseCase (nor Capability)

!!! note "Accepted"
    This decision is in effect.

The node names a logical contract a service guarantees. `UseCase` carries
UML baggage that misleads the way UserStory did for intent (ADR-10): a use
case is an actor-system interaction scenario. Specue's node is not a
scenario, it is a guarantee — and "a use case with no actor" is a
contradiction the operation contracts (the Plan/context verbs, which face no
external audience) break.

Rename to `Contract`. It reads right for both a Need-facing contract and an
internal operation contract (an internal contract is normal in
Design-by-Contract). Rejected `Capability`: a capability is a bare "can do"
with no place for the guarantees a contract carries.

Breaking rename across schema, model, code annotations and the self-spec;
ships in the pre-release window. Fixes only the node's name, not the shape
of its guarantees. Symmetric with ADR-10 fixing Need over UserStory.
