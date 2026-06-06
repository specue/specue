---
title: Report whether the current spec is correct
icon: material/play-circle-outline
tags:
    - contract
    - proven
---

# Report whether the current spec is correct

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to validate the current spec

## Invariants

### <a id="single-verdict"></a>single-verdict

*(returns)* The result is a single verdict over the whole spec: correct, or broken with the list of failures.

Satisfies: [as-agent-author#fr-01](../domain/as-agent-author.md#fr-01)

*Proven.*

### <a id="role-gate"></a>role-gate

**Rejects** when a node's type is not allowed by its module's kind.

Satisfies: [as-decision-keeper#fr-03](../domain/as-decision-keeper.md#fr-03)

*Proven.*

### <a id="unique-slug-within-module"></a>unique-slug-within-module

**Rejects** when two nodes share a slug within the same module.

Satisfies: [as-agent-create#fr-01](../domain/as-agent-create.md#fr-01)

*Proven.*

### <a id="dangling-binding"></a>dangling-binding

**Rejects** when a code annotation does not resolve to a node in the module's require closure.

*Proven.*

### <a id="unbindable-target"></a>unbindable-target

**Rejects** when a code annotation is aimed at a node that cannot be bound (anything but a Contract).

*Proven.*

### <a id="duplicate-binding"></a>duplicate-binding

**Rejects** when a node is bound by more than one code annotation in the same code module.

*Proven.*

### <a id="unreachable-contract"></a>unreachable-contract

**Rejects** when a Contract is claimed by no story FR, invoked by no other contract and named by no trigger.

*Proven.*

### <a id="sync-cycle"></a>sync-cycle

**Rejects** when a cycle of synchronous dependencies between contracts exists.

*Proven.*

### <a id="failure-carries-next-step"></a>failure-carries-next-step

*(returns)* Each failure carries the next step the caller takes to resolve it.

Satisfies: [as-agent-author#fr-03](../domain/as-agent-author.md#fr-03)

*Proven.*


## Realizes

- [as-agent-author](../domain/as-agent-author.md)
- [as-decision-keeper](../domain/as-decision-keeper.md)
- [as-agent-create](../domain/as-agent-create.md)

