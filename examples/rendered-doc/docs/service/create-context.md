---
title: Create a new context
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Create a new context

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to create a context by name

## Invariants

### <a id="name-is-unique"></a>name-is-unique

Creating a context with a name that already exists is refused.

Satisfies: [as-agent-setup#fr-01](../domain/as-agent-setup.md#fr-01)

*Proven.*

### <a id="starts-empty"></a>starts-empty

A new context holds no modules until the caller adds them.

*Implemented* (no test yet).


## Postconditions

### —

The context survives across invocations on the same machine.


## Realizes

- [as-agent-setup](../domain/as-agent-setup.md)

