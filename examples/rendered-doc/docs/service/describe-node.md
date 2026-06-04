---
title: Read one node in full by its module-qualified identity
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Read one node in full by its module-qualified identity

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks for one node by its module-qualified identity

## Invariants

### <a id="identity-is-module-qualified"></a>identity-is-module-qualified

The node is addressed by its module-qualified identity, which is stable across the landscape.

Satisfies: [as-agent-navigate#fr-02](../domain/as-agent-navigate.md#fr-02)

*Proven.*

### <a id="shown-in-full"></a>shown-in-full

The node's whole contract is returned: its conditions, its invariants, its variations and its declared edges.

Satisfies: [as-agent-navigate#fr-02](../domain/as-agent-navigate.md#fr-02), [as-decision-keeper#fr-01](../domain/as-decision-keeper.md#fr-01)

*Proven.*

### <a id="element-scoped"></a>element-scoped

When the identity carries a named-element suffix, the result is narrowed to that single element — the inquirer reads one invariant or one story FR without scrolling the whole node.

Satisfies: [as-agent-navigate#fr-02](../domain/as-agent-navigate.md#fr-02)

*Proven.*


## Postconditions

### —

The node is returned together with its current status.


## Realizes

- [as-agent-navigate](../domain/as-agent-navigate.md)
- [as-decision-keeper](../domain/as-decision-keeper.md)

