---
title: Remove a context
icon: material/play-circle-outline
tags:
    - contract
    - implemented
---

# Remove a context

!!! info "Implemented — 0/1 proven"
    Some invariants still lack a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to remove a context by name

## Invariants

### <a id="context-must-exist"></a>context-must-exist

**Rejects** when the named context does not exist.

*Implemented* (no test yet).

### <a id="removed-until-recreated"></a>removed-until-recreated

Once removed the context cannot be switched into until it is created again.

*Implemented* (no test yet).


