---
name: specue-planning
description: Propose, view, conflict-check and accept Plans — speculative changes to the spec
---

# Planning a speculative change

Use this skill when a change to the spec is **not yet accepted** — an
experiment, a proposal, a multi-step migration you want to land in pieces.
Plans are how Specue makes intent first-class without forking the
landscape.

The shared Specue skill (`specue` → `authoring.md`,
`maintaining.md`) covers the model: a Plan is a node in the governance
module pointing at git branches, conflicts are derived from overlay,
acceptance merges. This skill is the **operator handbook**: which verb to
run when, what each one does to the working tree, how to read the result.

## The three layers, on the planning side

A Plan adds an **intent axis** to WHAT.

- WHAT — the *current* spec (every node on the base branches) plus *each
  open Plan's projected spec* (the same nodes with the Plan's branches
  swapped in for the modules it touches). Two overlapping WHATs that
  coexist as long as the Plan is open.
- HOW — code lives on Plan branches too, so a code change can ride a
  spec change. The same `//specue:req:` rules apply on the branch.
- WHY — a Plan record carries no rationale of its own; if a Plan changes a
  UC's invariant, the new invariant carries its own `decided_by` to an
  ADR (existing or new). A migration Plan often opens a new ADR first.

## When a Plan is the right tool

Reach for a Plan when:

- the change is **breaking** at the contract level (an invariant tightens,
  an edge rewires) and someone else may depend on the current shape;
- the change spans **multiple modules** and you want to land them
  together — a single commit would mean coordinating writes across repos;
- the change is **speculative** — you want to *see* how it lands before
  committing to it (diff, conflict-check, then accept or drop);
- you want to **review** before merging, with a typed diff over Contracts
  and edges instead of a line-by-line patch.

Reach for a direct commit when the change is additive on one module and
nothing breaks (a new UC, a new ADR, a new annotation). Plans are
overhead — use them where the overhead pays for itself.

## The verbs

Plans require a governance module in the current context. Without one
every plan verb refuses with the next step to take.

- `plan register <id>` — creates branches `plan/<id>` in every module the
  Plan will touch and writes a Plan record in the governance module
  pointing at them. **Mutates the working tree** (it checks out the new
  branches across affected repos so you can start authoring).
- `plan use <id>` — switches an existing Plan in: checks out
  `plan/<id>` in every module that has the branch. Refuses on a dirty
  tree. **Mutates the working tree.**
- `plan base` — leaves the current Plan: checks every affected module
  back to the base branch. Refuses on a dirty tree. **Mutates the
  working tree.**
- `plan drop <id>` — closes the Plan: deletes the `plan/<id>` branches
  and closes the record. **Mutates the working tree** (returns to base
  and removes branches).
- `diff plan <id>` — typed delta of the Plan against the current spec
  (added/removed/modified/rewired over Contracts, Needs, Ports and
  their elements). **Does not touch the working tree** — reads through
  git.
- `plan conflict <a> <b>` — overlays both Plans together and reports
  structural failures (gates: a removed node referenced, an edge
  rewired two ways) and co-touch advisories. Read-only.
- `plan accept <id>` — overlays the Plan onto the current spec,
  validates the result, and on success merges the Plan's branches into
  the base everywhere. **Mutates the working tree.** On failure the
  merge is rolled back and the working tree is left as it was.

## Structural vs co-touch conflicts

Two Plans can both edit the same Contract and still both apply cleanly —
one adds an invariant, the other tightens a different one. That is a
**co-touch advisory**: the tool surfaces the pair for review but does not
block. Two Plans that cannot both apply — one removes what the other
modifies, both rewire the same edge to different targets, both rename the
same slug — are a **structural gate** and acceptance of either blocks
until one is dropped.

When `plan conflict` reports an advisory, decide as a human or as an
agent which to accept first; the second often needs a small rebase.
When it reports a gate, one of the two Plans needs to change its shape
or be dropped.

## Workflow

1. Decide: is this a Plan-worthy change? Breaking, multi-module,
   speculative or wants review — yes; additive on one module — no.
2. `plan register <id>` — start. The working tree is on the new
   branches.
3. Author the change (spec + code as needed). Same authoring and
   binding discipline as on base; the dogfood self-spec will tell you if
   something broke. `validate` works on the Plan's tree just like on
   base.
4. `diff plan <id>` from another shell or after `plan base`, to read the
   typed delta. This is what reviewers see.
5. `plan conflict <id> <other>` for every other open Plan you might
   collide with.
6. When ready: `plan accept <id>`. On failure the report tells you why;
   the tree is left untouched. Fix and re-run.
7. If the Plan turns out wrong: `plan drop <id>` — the base branches
   are exactly as they were.

## Caveats

- Plans mutate the working tree on register/use/base/drop/accept.
  Commit anything you do not want to lose before running them; the tool
  refuses on a dirty tree where it can, but `register` is a creation
  step and is less forgiving.
- The Plan record lives in the governance module; if the landscape has
  no governance module, the verbs refuse with the next step to take.
- `accept` runs `validate` on the overlay before merging. A broken
  overlay never lands.
- A Plan that touches a module someone else's Plan also touches is
  fine — `conflict` is the way to find out which pairs are advisory
  and which are blocking.

## Worked example — render-doc

A real Plan run in this repo, end to end. The change: add a `render-doc`
Contract that closes two federated FRs (`as-federated-owner#fr-04` and
`as-federated-reader#fr-03`) plus a new ADR justifying the shape. Two
modules touched — `spec.d/service/` and `spec.d/governance/` — so a Plan is
the right tool.

1. **register** — `plan register render-doc` created `plan/render-doc` in
   both modules and wrote `spec.d/governance/plan-render-doc.cue` (the Plan
   record). The working tree did not switch; `register` only creates.
2. **use** — `plan use render-doc` checked the plan branch out across
   both modules. Now editing on the Plan.
3. **author** — added the UC to
   `spec.d/service/federation/federation.cue` and ADR-09 to
   `spec.d/governance/adr.cue`. Six invariants, two of them with
   `satisfies` to the federated FRs. `validate` stayed green on the
   branch (49 nodes, the new UC `asserted` — honest: no code yet).
4. **describe** — `describe specue.io/service@v0:render-doc` showed
   the resolved node with `realizes` pointing at both federated stories,
   confirming the wiring landed as intended.
5. **commit** — committed the UC + ADR on the plan branch.
6. **diff** — `diff plan render-doc` returned the typed delta showing the
   added UC and ADR, read through git from base and plan/<id> — the
   working tree did not need to move.
7. **accept** — `plan accept render-doc` overlaid the plan, ran
   `validate` on the overlay (green), merged the branch into base, and
   flipped the Plan record to `accepted`.
8. **result** — `as-federated-owner` moved uncovered → partial (fr-04 is
   now satisfied by an asserted UC; the rest by proven attestation).
   `as-federated-reader` stayed uncovered because both UCs that satisfy its
   FRs are still asserted — true to the GAP: contract agreed, code not
   yet written. That is the right state and the right next TODO.

A Plan that lands without code is not a failed Plan; it agrees a contract
that subsequent work will discharge. The federated reader Need staying
uncovered is the truthful state.
