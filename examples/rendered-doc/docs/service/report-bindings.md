---
title: Show a code module's bindable contracts and their state
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Show a code module's bindable contracts and their state

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks what a code module may realize and where it stands

## Invariants

### <a id="scoped-to-code-module"></a>scoped-to-code-module

The report is computed for one code module; asking on a non-code module is refused with a remedy.

Decided by: [ADR-05](../governance/ADR-05.md)

*Proven.*

### <a id="allowed-from-require-closure"></a>allowed-from-require-closure

The contracts the caller may bind are exactly the UseCases reachable through the code module's require closure.

Satisfies: [as-agent-author#fr-02](../domain/as-agent-author.md#fr-02)

Decided by: [ADR-05](../governance/ADR-05.md)

*Implemented* (no test yet).

### <a id="per-element-state"></a>per-element-state

Each row's state (unbound, bound, proven, duplicate, orphan) reflects whether the specific element has a binding and a proving test, not the UseCase as a whole.

Satisfies: [as-agent-author#fr-02](../domain/as-agent-author.md#fr-02)

*Proven.*


## Postconditions

### —

Each row names the contract, the kind of binding, the state and the locations of any code that produced it.


## Realizes

- [as-agent-author](../domain/as-agent-author.md)

