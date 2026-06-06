---
title: Read code annotations as binding facts
icon: material/play-circle-outline
tags:
    - contract
    - implemented
---

# Read code annotations as binding facts

!!! info "Implemented — 0/5 proven"
    Some invariants still lack a test.

Service: [specue](specue.md)  •  binding: required  •  interaction: sync

**Trigger.** any verb that needs to know which code realizes which Contracts asks for the scan

## Invariants

### <a id="language-agnostic-match"></a>language-agnostic-match

An annotation is recognised by its lexical shape, independent of the host language's syntax.

Decided by: [ADR-05](../governance/ADR-05.md)

*Implemented* (no test yet).

### <a id="annotation-is-the-only-binding-channel"></a>annotation-is-the-only-binding-channel

Code is bound to a Contract only by an annotation in its source; nothing else (a file name, a path convention) counts as a binding.

Satisfies: [as-agent-relate#fr-04](../domain/as-agent-relate.md#fr-04)

Decided by: [ADR-05](../governance/ADR-05.md)

*Implemented* (no test yet).

### <a id="ignored-by-comment-context"></a>ignored-by-comment-context

An annotation that sits as quoted prose inside another comment is not taken as a binding.

*Implemented* (no test yet).

### <a id="scan-rooted-at-code-root"></a>scan-rooted-at-code-root

The scan begins at the code module's declared code_root (relative to its manifest), defaulting to the manifest's own directory, so a code module may live in a subfolder of the repository it scans without claiming sibling spec modules as its own subpackages.

Decided by: [ADR-11](../governance/ADR-11.md)

*Implemented* (no test yet).

### <a id="binding-fact-carries-location"></a>binding-fact-carries-location

*(returns)* Each binding fact carries the file and line that produced it.

*Implemented* (no test yet).


## Realizes

- [as-agent-relate](../domain/as-agent-relate.md)

