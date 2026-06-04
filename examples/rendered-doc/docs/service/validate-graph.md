---
title: Report whether the current spec is correct
icon: material/play-circle-outline
tags:
    - usecase
    - proven
---

# Report whether the current spec is correct

!!! success "Proven"
    All invariants have an implementation and a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** the caller asks to validate the current spec

## Invariants

### <a id="single-verdict"></a>single-verdict

The result is a single verdict over the whole spec: correct, or broken with the list of failures.

Satisfies: [as-agent-author#fr-01](../domain/as-agent-author.md#fr-01)

*Proven.*

### <a id="role-gate"></a>role-gate

A node whose type is not allowed by its module's kind is reported as a failure.

Satisfies: [as-decision-keeper#fr-03](../domain/as-decision-keeper.md#fr-03)

*Proven.*

### <a id="unique-slug-within-module"></a>unique-slug-within-module

Two nodes that share a slug within the same module are reported as a failure.

Satisfies: [as-agent-create#fr-01](../domain/as-agent-create.md#fr-01)

*Proven.*

### <a id="dangling-binding"></a>dangling-binding

A code annotation that does not resolve to a node in the module's require closure is reported as a failure.

*Proven.*

### <a id="unbindable-target"></a>unbindable-target

A code annotation aimed at a node that cannot be bound (anything but a UseCase) is reported as a failure.

*Proven.*

### <a id="duplicate-binding"></a>duplicate-binding

A node bound by more than one code annotation in the same code module is reported as a failure.

*Proven.*

### <a id="unreachable-usecase"></a>unreachable-usecase

A UseCase that no story FR claims, no other contract invokes and no trigger names is reported as a failure.

*Proven.*

### <a id="sync-cycle"></a>sync-cycle

A cycle of synchronous dependencies between contracts is reported as a failure.

*Proven.*


## Postconditions

### —

Each failure carries the next step the caller takes to resolve it.

Satisfies: [as-agent-author#fr-03](../domain/as-agent-author.md#fr-03)


## Realizes

- [as-agent-author](../domain/as-agent-author.md)
- [as-decision-keeper](../domain/as-decision-keeper.md)
- [as-agent-create](../domain/as-agent-create.md)

