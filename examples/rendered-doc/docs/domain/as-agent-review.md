---
title: See what I changed between two points
icon: material/clipboard-text-outline
tags:
    - need
    - covered
---

# See what I changed between two points

!!! success "Covered — 2/2"
    Every requirement is satisfied by a proven contract.

Domain: [specue](specue.md)

**Consumer:** an agent who has been authoring for a while

to see what the spec became compared to where I started, so that I can review my own work before sharing it

## Requirements

### <a id="fr-01"></a>fr-01

The difference between the spec at two versioned points is reported as a typed delta over UseCases, Needs, Ports and their elements.

*Covered by [diff-refs#typed-over-the-spec-graph](../service/diff-refs.md#typed-over-the-spec-graph)*

### <a id="fr-02"></a>fr-02

Each change names what was added, removed, modified or rewired.

*Covered by [diff-refs#every-change-named](../service/diff-refs.md#every-change-named)*

