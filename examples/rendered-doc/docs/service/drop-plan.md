---
title: Abandon a Plan without accepting it
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Abandon a Plan without accepting it

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to drop a Plan

## Invariants

### <a id="branches-and-record-removed"></a>branches-and-record-removed

The Plan record is closed and every branch it pointed at is removed.

Decided by: [ADR-07](../governance/ADR-07.md)

*Proven.*

### <a id="dropped-until-reregistered"></a>dropped-until-reregistered

Once dropped the Plan cannot be used again under the same name until it is registered again.

*Implemented* (no test yet).


