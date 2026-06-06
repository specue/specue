---
title: A plan is a named branch across every module it touches
icon: material/gavel
tags:
    - adr
    - accepted
---

# A plan is a named branch across every module it touches

!!! note "Accepted"
    This decision is in effect.

A plan's content lives on identically-named branches in every affected
repository; its identity is a Plan record in a dedicated governance module
of the landscape — kind: governance — that points at those branches. The
governance module is where ADRs also live, kept apart from modules that
hold Contracts, UserStories or Ports. Speculative work is real CUE on a
real ref the tool can read, diff and overlay, not a separate document
store. Acceptance merges the branches; conflicts between plans are gates
derived by overlaying both deltas. The intent axis is git, with
governance only naming what is in flight.
