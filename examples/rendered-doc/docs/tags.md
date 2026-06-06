---
hide:
    - tags
title: Tags
---

# Tags

Every node carries tags for its type and status. Click a heading or use the table of contents.

## contract { #tag:contract }

- [`service:accept-plan`](service/accept-plan.md) — Apply a Plan to the current spec and close it · *proven*
- [`service:add-module-to-context`](service/add-module-to-context.md) — Add a module to a context by its directory · *implemented*
- [`service:attest-bindings`](service/attest-bindings.md) — Publish a spec module's binding outcomes for readers without code access · *asserted*
- [`service:build-graph`](service/build-graph.md) — Produce a resolved spec graph from the current context · *proven*
- [`service:create-context`](service/create-context.md) — Create a new context · *proven*
- [`service:describe-node`](service/describe-node.md) — Read one node in full by its module-qualified identity · *proven*
- [`service:detect-conflict`](service/detect-conflict.md) — Report conflicts between two open Plans · *proven*
- [`service:diff-refs`](service/diff-refs.md) — Report the typed delta between the spec at two versioned points · *proven*
- [`service:drop-plan`](service/drop-plan.md) — Abandon a Plan without accepting it · *proven*
- [`service:init-module`](service/init-module.md) — Start a new module of a known kind · *proven*
- [`service:list-resources`](service/list-resources.md) — List the kinds of node the spec holds, and the nodes of one kind · *proven*
- [`service:pending-overlay`](service/pending-overlay.md) — Show a Plan against the current spec without switching the working tree · *proven*
- [`service:query-graph`](service/query-graph.md) — Answer a graph query with read-only SQL · *proven*
- [`service:read-context`](service/read-context.md) — Read the active context · *implemented*
- [`service:register-plan`](service/register-plan.md) — Open a new Plan as a Plan record plus branches · *proven*
- [`service:remove-context`](service/remove-context.md) — Remove a context · *implemented*
- [`service:remove-module-from-context`](service/remove-module-from-context.md) — Remove a module from a context · *implemented*
- [`service:render-doc`](service/render-doc.md) — Render the resolved spec as a browsable markdown documentation tree · *proven*
- [`service:report-bindings`](service/report-bindings.md) — Show a code module's bindable contracts and their state · *proven*
- [`service:return-to-base`](service/return-to-base.md) — Leave a Plan and return to the base branch · *proven*
- [`service:scan-code`](service/scan-code.md) — Read code annotations as binding facts · *implemented*
- [`service:use-context`](service/use-context.md) — Make a context the active one · *proven*
- [`service:use-plan`](service/use-plan.md) — Switch the working tree into a Plan · *proven*
- [`service:validate-graph`](service/validate-graph.md) — Report whether the current spec is correct · *proven*
- [`service:warm-schema`](service/warm-schema.md) — Seed the editor's cue cache with the schema and the landscape's modules · *proven*

## need { #tag:need }

- [`domain:as-agent-author`](domain/as-agent-author.md) — Know whether what I authored is correct · *covered*
- [`domain:as-agent-create`](domain/as-agent-create.md) — Add a Contract, Need, ADR or Port to a module · *partial*
- [`domain:as-agent-navigate`](domain/as-agent-navigate.md) — Find my way around an unfamiliar spec · *covered*
- [`domain:as-agent-relate`](domain/as-agent-relate.md) — Wire one thing to another · *partial*
- [`domain:as-agent-review`](domain/as-agent-review.md) — See what I changed between two points · *covered*
- [`domain:as-agent-setup`](domain/as-agent-setup.md) — Choose which spec context I am working against · *partial*
- [`domain:as-agent-start`](domain/as-agent-start.md) — Bring a new module into existence · *partial*
- [`domain:as-author-dx`](domain/as-author-dx.md) — Author with editor support · *partial*
- [`domain:as-decision-keeper`](domain/as-decision-keeper.md) — Keep the why and the what-is-in-flight · *covered*
- [`domain:as-federated-owner`](domain/as-federated-owner.md) — Own my slice of the spec without coordinating every change · *partial*
- [`domain:as-federated-reader`](domain/as-federated-reader.md) — Read the spec without holding the code · *partial*
- [`domain:as-planner`](domain/as-planner.md) — See how a Plan lands before committing to it · *partial*

## domain { #tag:domain }

- [`domain:specue`](domain/specue.md) — Specue — a spec graph derived from CUE modules and the code that realizes them

## adr { #tag:adr }

- [`governance:ADR-01`](governance/ADR-01.md) — Cross-module references resolve through CUE, not a hand-written resolver · *accepted*
- [`governance:ADR-02`](governance/ADR-02.md) — Graph navigation and search is exposed as read-only SQL · *accepted*
- [`governance:ADR-03`](governance/ADR-03.md) — Every module lives in a git repository · *accepted*
- [`governance:ADR-04`](governance/ADR-04.md) — The OCI registry that warms the cue cache runs in-process · *accepted*
- [`governance:ADR-05`](governance/ADR-05.md) — Code is a first-class module of the landscape · *accepted*
- [`governance:ADR-06`](governance/ADR-06.md) — The embedded schema version is fixed; changes ship under the same tag · *accepted*
- [`governance:ADR-07`](governance/ADR-07.md) — A plan is a named branch across every module it touches · *accepted*
- [`governance:ADR-08`](governance/ADR-08.md) — Code-binding outcomes are published alongside the spec for readers without code access · *accepted*
- [`governance:ADR-09`](governance/ADR-09.md) — The rendered document is a derived view, never an authoring source · *accepted*
- [`governance:ADR-10`](governance/ADR-10.md) — The intent node is a Need (with a Domain), not a UserStory (with a Product) · *accepted*
- [`governance:ADR-11`](governance/ADR-11.md) — Code module scans from code_root; repo modules live under spec.d/ · *accepted*
- [`governance:ADR-12`](governance/ADR-12.md) — Render serves two formats — markdown (with knobs) and JSON IR — not a family of presets · *accepted*
- [`governance:ADR-13`](governance/ADR-13.md) — The contract node is named Contract, not UseCase (nor Capability) · *accepted*
- [`governance:ADR-14`](governance/ADR-14.md) — A Contract is a set of invariants; pre/post/variation collapse into one typed kind · *proposed*

## plan { #tag:plan }

- [`governance:plan-accept-from-anywhere`](governance/plan-accept-from-anywhere.md) — accept-from-anywhere · *accepted*
- [`governance:plan-code-root-field`](governance/plan-code-root-field.md) — code-root-field · *accepted*
- [`governance:plan-index-strip-prefix`](governance/plan-index-strip-prefix.md) — index-strip-prefix · *accepted*
- [`governance:plan-link-text-slug`](governance/plan-link-text-slug.md) — link-text-slug · *accepted*
- [`governance:plan-m-elem`](governance/plan-m-elem.md) — Collapse contract elements to one invariant kind · *proposed*
- [`governance:plan-nav-collapse`](governance/plan-nav-collapse.md) — nav-collapse · *accepted*
- [`governance:plan-quick-wins`](governance/plan-quick-wins.md) — quick-wins · *accepted*
- [`governance:plan-readme-progressive`](governance/plan-readme-progressive.md) — readme-progressive · *accepted*
- [`governance:plan-readme-query-step`](governance/plan-readme-query-step.md) — readme-query-step · *accepted*
- [`governance:plan-refs-as-definitions`](governance/plan-refs-as-definitions.md) — refs-as-definitions · *accepted*
- [`governance:plan-render-doc`](governance/plan-render-doc.md) — render-doc · *accepted*
- [`governance:plan-render-index-pages`](governance/plan-render-index-pages.md) — render-index-pages · *accepted*
- [`governance:plan-render-node-as-folder-index`](governance/plan-render-node-as-folder-index.md) — render-node-as-folder-index · *proposed*
- [`governance:plan-render-presets`](governance/plan-render-presets.md) — render-presets · *accepted*
- [`governance:plan-render-status-admonitions`](governance/plan-render-status-admonitions.md) — render-status-admonitions · *accepted*
- [`governance:plan-render-tags-page`](governance/plan-render-tags-page.md) — render-tags-page · *accepted*
- [`governance:plan-userstory-to-need`](governance/plan-userstory-to-need.md) — userstory-to-need · *accepted*

## container { #tag:container }

- [`service:specue`](service/specue.md) — Specue CLI

## proven { #tag:proven }

- [`service:accept-plan`](service/accept-plan.md) — Apply a Plan to the current spec and close it · *proven*
- [`service:build-graph`](service/build-graph.md) — Produce a resolved spec graph from the current context · *proven*
- [`service:create-context`](service/create-context.md) — Create a new context · *proven*
- [`service:describe-node`](service/describe-node.md) — Read one node in full by its module-qualified identity · *proven*
- [`service:detect-conflict`](service/detect-conflict.md) — Report conflicts between two open Plans · *proven*
- [`service:diff-refs`](service/diff-refs.md) — Report the typed delta between the spec at two versioned points · *proven*
- [`service:drop-plan`](service/drop-plan.md) — Abandon a Plan without accepting it · *proven*
- [`service:init-module`](service/init-module.md) — Start a new module of a known kind · *proven*
- [`service:list-resources`](service/list-resources.md) — List the kinds of node the spec holds, and the nodes of one kind · *proven*
- [`service:pending-overlay`](service/pending-overlay.md) — Show a Plan against the current spec without switching the working tree · *proven*
- [`service:query-graph`](service/query-graph.md) — Answer a graph query with read-only SQL · *proven*
- [`service:register-plan`](service/register-plan.md) — Open a new Plan as a Plan record plus branches · *proven*
- [`service:render-doc`](service/render-doc.md) — Render the resolved spec as a browsable markdown documentation tree · *proven*
- [`service:report-bindings`](service/report-bindings.md) — Show a code module's bindable contracts and their state · *proven*
- [`service:return-to-base`](service/return-to-base.md) — Leave a Plan and return to the base branch · *proven*
- [`service:use-context`](service/use-context.md) — Make a context the active one · *proven*
- [`service:use-plan`](service/use-plan.md) — Switch the working tree into a Plan · *proven*
- [`service:validate-graph`](service/validate-graph.md) — Report whether the current spec is correct · *proven*
- [`service:warm-schema`](service/warm-schema.md) — Seed the editor's cue cache with the schema and the landscape's modules · *proven*

## implemented { #tag:implemented }

- [`service:add-module-to-context`](service/add-module-to-context.md) — Add a module to a context by its directory · *implemented*
- [`service:read-context`](service/read-context.md) — Read the active context · *implemented*
- [`service:remove-context`](service/remove-context.md) — Remove a context · *implemented*
- [`service:remove-module-from-context`](service/remove-module-from-context.md) — Remove a module from a context · *implemented*
- [`service:scan-code`](service/scan-code.md) — Read code annotations as binding facts · *implemented*

## asserted { #tag:asserted }

- [`service:attest-bindings`](service/attest-bindings.md) — Publish a spec module's binding outcomes for readers without code access · *asserted*

## covered { #tag:covered }

- [`domain:as-agent-author`](domain/as-agent-author.md) — Know whether what I authored is correct · *covered*
- [`domain:as-agent-navigate`](domain/as-agent-navigate.md) — Find my way around an unfamiliar spec · *covered*
- [`domain:as-agent-review`](domain/as-agent-review.md) — See what I changed between two points · *covered*
- [`domain:as-decision-keeper`](domain/as-decision-keeper.md) — Keep the why and the what-is-in-flight · *covered*

## partial { #tag:partial }

- [`domain:as-agent-create`](domain/as-agent-create.md) — Add a Contract, Need, ADR or Port to a module · *partial*
- [`domain:as-agent-relate`](domain/as-agent-relate.md) — Wire one thing to another · *partial*
- [`domain:as-agent-setup`](domain/as-agent-setup.md) — Choose which spec context I am working against · *partial*
- [`domain:as-agent-start`](domain/as-agent-start.md) — Bring a new module into existence · *partial*
- [`domain:as-author-dx`](domain/as-author-dx.md) — Author with editor support · *partial*
- [`domain:as-federated-owner`](domain/as-federated-owner.md) — Own my slice of the spec without coordinating every change · *partial*
- [`domain:as-federated-reader`](domain/as-federated-reader.md) — Read the spec without holding the code · *partial*
- [`domain:as-planner`](domain/as-planner.md) — See how a Plan lands before committing to it · *partial*

## accepted { #tag:accepted }

- [`governance:ADR-01`](governance/ADR-01.md) — Cross-module references resolve through CUE, not a hand-written resolver · *accepted*
- [`governance:ADR-02`](governance/ADR-02.md) — Graph navigation and search is exposed as read-only SQL · *accepted*
- [`governance:ADR-03`](governance/ADR-03.md) — Every module lives in a git repository · *accepted*
- [`governance:ADR-04`](governance/ADR-04.md) — The OCI registry that warms the cue cache runs in-process · *accepted*
- [`governance:ADR-05`](governance/ADR-05.md) — Code is a first-class module of the landscape · *accepted*
- [`governance:ADR-06`](governance/ADR-06.md) — The embedded schema version is fixed; changes ship under the same tag · *accepted*
- [`governance:ADR-07`](governance/ADR-07.md) — A plan is a named branch across every module it touches · *accepted*
- [`governance:ADR-08`](governance/ADR-08.md) — Code-binding outcomes are published alongside the spec for readers without code access · *accepted*
- [`governance:ADR-09`](governance/ADR-09.md) — The rendered document is a derived view, never an authoring source · *accepted*
- [`governance:ADR-10`](governance/ADR-10.md) — The intent node is a Need (with a Domain), not a UserStory (with a Product) · *accepted*
- [`governance:ADR-11`](governance/ADR-11.md) — Code module scans from code_root; repo modules live under spec.d/ · *accepted*
- [`governance:ADR-12`](governance/ADR-12.md) — Render serves two formats — markdown (with knobs) and JSON IR — not a family of presets · *accepted*
- [`governance:ADR-13`](governance/ADR-13.md) — The contract node is named Contract, not UseCase (nor Capability) · *accepted*
- [`governance:plan-accept-from-anywhere`](governance/plan-accept-from-anywhere.md) — accept-from-anywhere · *accepted*
- [`governance:plan-code-root-field`](governance/plan-code-root-field.md) — code-root-field · *accepted*
- [`governance:plan-index-strip-prefix`](governance/plan-index-strip-prefix.md) — index-strip-prefix · *accepted*
- [`governance:plan-link-text-slug`](governance/plan-link-text-slug.md) — link-text-slug · *accepted*
- [`governance:plan-nav-collapse`](governance/plan-nav-collapse.md) — nav-collapse · *accepted*
- [`governance:plan-quick-wins`](governance/plan-quick-wins.md) — quick-wins · *accepted*
- [`governance:plan-readme-progressive`](governance/plan-readme-progressive.md) — readme-progressive · *accepted*
- [`governance:plan-readme-query-step`](governance/plan-readme-query-step.md) — readme-query-step · *accepted*
- [`governance:plan-refs-as-definitions`](governance/plan-refs-as-definitions.md) — refs-as-definitions · *accepted*
- [`governance:plan-render-doc`](governance/plan-render-doc.md) — render-doc · *accepted*
- [`governance:plan-render-index-pages`](governance/plan-render-index-pages.md) — render-index-pages · *accepted*
- [`governance:plan-render-presets`](governance/plan-render-presets.md) — render-presets · *accepted*
- [`governance:plan-render-status-admonitions`](governance/plan-render-status-admonitions.md) — render-status-admonitions · *accepted*
- [`governance:plan-render-tags-page`](governance/plan-render-tags-page.md) — render-tags-page · *accepted*
- [`governance:plan-userstory-to-need`](governance/plan-userstory-to-need.md) — userstory-to-need · *accepted*

## proposed { #tag:proposed }

- [`governance:ADR-14`](governance/ADR-14.md) — A Contract is a set of invariants; pre/post/variation collapse into one typed kind · *proposed*
- [`governance:plan-m-elem`](governance/plan-m-elem.md) — Collapse contract elements to one invariant kind · *proposed*
- [`governance:plan-render-node-as-folder-index`](governance/plan-render-node-as-folder-index.md) — render-node-as-folder-index · *proposed*

