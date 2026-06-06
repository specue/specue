---
title: Create a new context
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Create a new context

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to create a context by name

## Invariants

### <a id="name-is-unique"></a>name-is-unique

**Rejects** when a context with that name already exists.

Satisfies: [as-agent-setup#fr-01](../domain/as-agent-setup.md#fr-01)

*Proven.*

### <a id="starts-empty"></a>starts-empty

A new context holds no modules until the caller adds them.

*Implemented* (no test yet).

### <a id="survives-across-invocations"></a>survives-across-invocations

The context survives across invocations.

*Proven.*


## Realizes

- [as-agent-setup](../domain/as-agent-setup.md)

