---
title: Cross-module references resolve through CUE, not a hand-written resolver
icon: material/gavel
tags:
    - adr
    - accepted
---

# Cross-module references resolve through CUE, not a hand-written resolver

!!! note "Accepted"
    This decision is in effect.

The whole module set is stitched into one CUE value tree, and CUE resolves
every cross-module reference, version pin and visibility rule. The previous
generation interpreted its own mini-language of string references through a
hand-written resolver, which became the system's bottleneck. Standing on CUE
shifts a class of resolution bugs onto a mature implementation and lets the
compiler do only what CUE cannot — domain constraints (statuses, cycles,
coverage).
