---
title: Keep the why and the what-is-in-flight
icon: material/clipboard-text-outline
tags:
    - need
    - covered
---

# Keep the why and the what-is-in-flight

!!! success "Covered — 3/3"
    Every requirement is satisfied by a proven contract.

Domain: [specue](specue.md)

**Consumer:** the person or role who keeps the landscape's decisions and open Plans

to record why a contract is shaped as it is, and to name what is being changed, so that the rationale survives the people who authored it and Plans have a place to live

## Requirements

### <a id="fr-01"></a>fr-01

A Contract element that cites an ADR shows the cited ADR among its declared edges.

*Covered by [describe-node#shown-in-full](../service/describe-node.md#shown-in-full)*

### <a id="fr-02"></a>fr-02

A registered Plan carries the branches its content lives on.

*Covered by [register-plan#plan-is-a-branch-set](../service/register-plan.md#plan-is-a-branch-set)*

### <a id="fr-03"></a>fr-03

A node of type ADR or Plan in a module that is not of kind governance is rejected.

*Covered by [validate-graph#role-gate](../service/validate-graph.md#role-gate)*

