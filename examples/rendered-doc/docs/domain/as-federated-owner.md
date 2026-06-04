---
title: Own my slice of the spec without coordinating every change
icon: material/clipboard-text-outline
tags:
    - need
    - partial
---

# Own my slice of the spec without coordinating every change

!!! warning "Partial — 4/4 covered"
    Some requirements have no proven contract.

Domain: [specue](specue.md)

**Consumer:** a team owning a slice of the spec landscape

to author my UseCases and Ports in my own repository, so that other teams can depend on what I publish without my changes blocking theirs

## Requirements

### <a id="fr-01"></a>fr-01

A UseCase or Port lives in a repository its owner controls.

*Covered by [build-graph#cue-stitches-the-modules](../service/build-graph.md#cue-stitches-the-modules)*

### <a id="fr-02"></a>fr-02

While developing locally my module is reached by its directory; once published it is depended on by name and version.

*Claimed by [add-module-to-context#addressed-by-directory](../service/add-module-to-context.md#addressed-by-directory)* — not proven

### <a id="fr-03"></a>fr-03

What another team may reference is the public part of my contract; everything else is invisible to them.

*Covered by [build-graph#cue-stitches-the-modules](../service/build-graph.md#cue-stitches-the-modules)*

### <a id="fr-04"></a>fr-04

My spec is publishable as a human-readable document for audiences that do not run the tool.

*Covered by [render-doc#cross-links-resolve-as-markdown](../service/render-doc.md#cross-links-resolve-as-markdown)*

