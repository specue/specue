---
tags:
    - index
title: service
---

# service

Contract: 25 · Container: 1

**Status:** 19 proven · 5 implemented · 1 asserted

## Contracts

- [accept-plan](accept-plan.md) — Apply a Plan to the current spec and close it · *proven*
- [add-module-to-context](add-module-to-context.md) — Add a module to a context by its directory · *implemented*
- [attest-bindings](attest-bindings.md) — Publish a spec module's binding outcomes for readers without code access · *asserted*
- [build-graph](build-graph.md) — Produce a resolved spec graph from the current context · *proven*
- [create-context](create-context.md) — Create a new context · *proven*
- [describe-node](describe-node.md) — Read one node in full by its module-qualified identity · *proven*
- [detect-conflict](detect-conflict.md) — Report conflicts between two open Plans · *proven*
- [diff-refs](diff-refs.md) — Report the typed delta between the spec at two versioned points · *proven*
- [drop-plan](drop-plan.md) — Abandon a Plan without accepting it · *proven*
- [init-module](init-module.md) — Start a new module of a known kind · *proven*
- [list-resources](list-resources.md) — List the kinds of node the spec holds, and the nodes of one kind · *proven*
- [pending-overlay](pending-overlay.md) — Show a Plan against the current spec without switching the working tree · *proven*
- [query-graph](query-graph.md) — Answer a graph query with read-only SQL · *proven*
- [read-context](read-context.md) — Read the active context · *implemented*
- [register-plan](register-plan.md) — Open a new Plan as a Plan record plus branches · *proven*
- [remove-context](remove-context.md) — Remove a context · *implemented*
- [remove-module-from-context](remove-module-from-context.md) — Remove a module from a context · *implemented*
- [render-doc](render-doc.md) — Render the resolved spec as a browsable markdown documentation tree · *proven*
- [report-bindings](report-bindings.md) — Show a code module's bindable contracts and their state · *proven*
- [return-to-base](return-to-base.md) — Leave a Plan and return to the base branch · *proven*
- [scan-code](scan-code.md) — Read code annotations as binding facts · *implemented*
- [specue](specue.md) — Specue CLI
- [use-context](use-context.md) — Make a context the active one · *proven*
- [use-plan](use-plan.md) — Switch the working tree into a Plan · *proven*
- [validate-graph](validate-graph.md) — Report whether the current spec is correct · *proven*
- [warm-schema](warm-schema.md) — Seed the editor's cue cache with the schema and the landscape's modules · *proven*

