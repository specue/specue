# MANIFESTO

Specue splits a system into three orthogonal layers, and every node, edge
and annotation belongs to exactly one of them.

## WHAT — what the system promises, and to whom

A `Need` carries the audience's intent; its FRs are atomic, observable
guarantees of that intent. A `UseCase` carries a contract the service
guarantees; its invariants are atomic, observable properties of the contract.
This is the graph the tool reads. Authoring lives in CUE.

The text of a WHAT is *observational*: a sentence a third party could check by
watching the system from the outside. "A task added today is visible later in
the same list" — yes. "AddTask writes to Mongo" — no, that is HOW.

## HOW — how the WHAT is realized in source

The function that satisfies an invariant, the test that proves it, the
infrastructure verb that anchors a Port. HOW lives in code, through
`//specue:req:`, `//specue:test:`, `//specue:produces:`,
`//specue:consumes:`, etc. annotations. The text of a WHAT never mentions a
function, a file, a constructor, a library, a build tool. If it does, lift the
HOW out into code where it belongs.

A WHAT with no matching HOW is the honest TODO. A status field, not a wiki
page that promises to be updated.

## WHY — why the contract is shaped this way

An `ADR` records one decision; an invariant cites it through
`decided_by: [<gov>:ADR-NN]`. The reasoning lives on the contract, not on the
intent, and never in an invariant's `text` field. An invariant says what is
guaranteed; the ADR explains why that shape and not another.

ADRs outlive any single plan or implementation. A rewrite that preserves the
invariant preserves the ADR; a rewrite that changes the invariant must either
cite a new ADR or supersede the old one. The why-layer is the layer that
keeps institutional memory honest.

## Why the discipline matters

The discipline keeps each layer doing its own job. A change to the audience
edits WHAT. A change to implementation edits HOW (and the spec only if a
contract genuinely shifts). A change to reasoning edits WHY. Mixing them costs
the graph its value:

- **HOW leaking into WHAT** makes contracts brittle to refactors — every
  rename of an internal function forces a spec edit.
- **WHY leaking into WHAT** makes contracts hard to read — the invariant's
  text drowns in justification when it should state a single observable
  property.
- **WHAT leaking out into code** (annotations on functions that don't
  actually realize a contract) makes the spec drift from the source — the
  status stops being honest.

The split is what lets the tool stay small. The compiler only checks domain
constraints CUE can't (statuses, cycles, blocked propagation, coverage); CUE
checks everything else; code annotations are a lexical scan, not a parser. No
layer reinvents what the layer beside it already does.

## On Need, not UserStory

The intent unit is a `Need` (with a `Domain`), not a `UserStory` (with a
`Product`). UserStory carries Agile baggage that misleads here: it implies a
sprint-sized increment authored from a user persona ("as a … I want … so
that …"), acceptance criteria in Gherkin, and a lifetime bounded by the
sprint that delivered it. None of that fits the layer Specue models.

The intent is long-lived, not iterative; its consumer is often non-human (an
operator, a downstream system, a regulator, an agent); the testable atoms
(FR / NFR) are the contract, not "acceptance criteria"; and the unit lives as
long as the system serves it, not until a story is closed.

Requirements Engineering (ISO/IEC 29148, IREB) calls this unit a Need: an
objective statement of what a stakeholder requires, independent of any
delivery cadence. Need carries the right semantics — a consumer and a
description, with named atoms — and drops the persona-narrative grammar that
does not generalise. The container above Need is the audience, also named in
RE-lexicon: Domain. (This is the same Domain DDD codifies; the terms align,
they do not conflict.)

Story statuses follow the same shift: `delivered / partial / orphan` becomes
`covered / partial / uncovered`, since a Need is covered by contracts rather
than "delivered" by a team. See
[ADR-10](spec.d/governance/adr.cue) for the full rationale.

## On the agent reader

Every verb takes `--json` and returns a stable shape. Errors are JSON too,
with the same `fix` field humans see in `try:`. Identity is always
`module:slug`, the form you copy from `get`, paste into `describe`, and
reference in annotations. The graph the agent reasons over is the same graph
the reviewer reads — there is no agent-friendly view of the spec hiding
behind a human-friendly one.
