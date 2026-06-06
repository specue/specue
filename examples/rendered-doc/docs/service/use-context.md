---
title: Make a context the active one
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Make a context the active one

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to switch into a context by name

## Invariants

### <a id="context-must-exist"></a>context-must-exist

**Rejects** when the named context does not exist.

Satisfies: [as-agent-setup#fr-01](../domain/as-agent-setup.md#fr-01)

*Proven.*

### <a id="subsequent-verbs-resolve-here"></a>subsequent-verbs-resolve-here

Once active, every subsequent read or authoring verb resolves against this context's modules unless overridden for the run.

*Implemented* (no test yet).

### <a id="active-across-invocations"></a>active-across-invocations

The chosen context is active across invocations until another one is switched in.

*Implemented* (no test yet).


## Realizes

- [as-agent-setup](../domain/as-agent-setup.md)

