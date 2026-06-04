---
title: Remove a context
icon: material/play-circle-outline
tags:
    - usecase
    - implemented
---

# Remove a context

!!! info "Implemented — 0/2 proven"
    Some invariants still lack a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to remove a context by name

## Invariants

### <a id="context-must-exist"></a>context-must-exist

Removing a context that does not exist is refused with the next step to take.

*Implemented* (no test yet).

### <a id="modules-survive"></a>modules-survive

The directories that held the context's modules are left untouched.

*Implemented* (no test yet).


## Postconditions

### —

Once removed the context cannot be switched into until it is created again.


