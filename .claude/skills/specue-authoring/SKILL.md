---
name: specue-authoring
description: Author spec nodes (Contract, Need, ADR, Port) in CUE for this repo
---

# Authoring spec nodes in this repo

Use this skill when you are creating or editing a `s.#Contract`,
`s.#Need`, `s.#ADR`, `s.#Port`, `s.#Container` or `s.#Domain` in this
repo's `spec.d/` modules.

The shared Specue skill (`specue` → `authoring.md`) is the model layer:
it covers slugs, qualified imports, satisfies vs realizes, infra edges, the
rules `validate` enforces. This skill is the **repo-specific dialect** of that
model.

## The three layers — WHAT, HOW, WHY

Everything below follows from this split. Keep it in mind when you write a
single line of text.

**WHAT** — what the system promises and to whom. Need carries an
audience's intent; its FRs are atomic, observable guarantees of that intent.
Contract carries a contract the service guarantees; its invariants are
atomic, observable properties of the contract. WHAT is the graph; it is what
`validate`, `describe`, `query` read.

**HOW** — how the WHAT is realized in source: the Go function, the test,
the binding annotation. It lives in code (`//specue:req:`,
`//specue:test:`, the infra verbs). HOW does not appear in WHAT — the
text of an FR or an invariant does not name a function, a file, a
constructor, a library, a build tool. If it does, lift it out.

**WHY** — why the contract is shaped this way. ADRs live in the governance
module and record a single decision each. WHY does not appear in WHAT
either — an invariant's `text` states what is guaranteed, not the reasoning
that led to that shape. The reasoning goes onto the same invariant through
`decided_by: [gov.adrNN...]`, which links to the ADR that justifies it.

Two failure modes to watch for while you author:

- HOW leaking into WHAT — "the validator iterates the graph and calls
  checkRoleGate, which …". Strip it down to "a node whose type is not
  allowed by its module's kind is reported as a failure."
- WHY leaking into WHAT — "the result is correct or broken so that the
  agent can iterate quickly". The reasoning goes to ADR (or, briefly, into
  the Need's `description` "…so that…" tail — never an invariant's `text`).

You can verify the discipline by reading the invariant aloud and asking: is
this an observable property of the contract that any implementation would
have to satisfy? If yes it is WHAT. If it sounds like an implementation
choice or a justification, it is HOW or WHY mislabelled.

## Language

The self-spec is authored in **English** — it describes the tool itself, not
the user's business. Prose, Need descriptions, FR text and ADR bodies are
all English even though the user's prompts and other landscapes are often
Russian.

## Module layout

Four modules live in this repo, all under `spec.d/` (the recommended
drop-in layout — see ADR-11):

- `spec.d/domain/` — `kind: domain`. Holds the audience-facing Needs.
  Sub-packages by audience: `agent/`, `human/`, `federated/`, `governance/`.
- `spec.d/service/` — `kind: service`. Holds the Contract contracts and the
  service Container. Sub-packages by phase: `graph-build/`, `validation/`,
  `navigation/`, `binding/`, `planning/`, `context/`, `federation/`.
- `spec.d/governance/` — `kind: governance`. Holds the ADRs (and any
  future Plan records).
- `spec.d/code/` — `kind: code`. Holds no nodes; declares which contracts
  the tool's Go source binds. Its `spec.mod.cue` sets `code_root: "../.."`
  so the scan reaches the repo root (`internal/`, `cmd/`, …).

A module is one CUE module with **sub-packages by sub-folder**. Files in
`spec.d/service/navigation/` are `package navigation`; files in
`spec.d/service/planning/` are `package planning`. Both load into one module
because `specload` walks `./...`.

## Cross-folder and cross-module imports

CUE-native. The version goes **at the end** of the path, once:

```
import (
    s     "specue.io/schema@v0:spec"                       // schema
    root  "specue.io/service@v0:service"                   // root package of own module
    agent "specue.io/domain/agent@v0:agent"                // a sub-package of another module
    gov   "specue.io/governance@v0:governance"             // another module
    fed   "specue.io/domain/federated@v0:federated"
)
```

The mistake to avoid: `specue.io/domain@v0/agent@v0` (two versions). CUE
rejects it.

## Don't repeat the defaults

A node's `type` is fixed by its CUE definition (`s.#Contract` ⇒ `type:
"Contract"`), so don't write the `type:` line — CUE concretizes it through
unification, and `s.#Port & {type: "Contract"}` is a unification error
(exactly the diagnostic the convention buys).

The same goes for any field whose default matches what most nodes carry:
the schema concretizes it for you. Today that is:

- `type` (every node) — set by the definition
- `confidence: "CONFIRMED"` — the default
- `interaction: "sync"` (Contract) — the default

So a typical Contract shrinks to:

```
diffRefs: s.#Contract & {
    slug:    "diff-refs"
    title:   "Report the typed delta between two refs"
    service: root.specue
    trigger: "the caller asks for a diff"
    invariants: [...]
}
```

Write a field only when it overrides a default. Write `confidence` when
it is not `CONFIRMED`, `interaction` when it is not `sync`. Less authored
text means less drift and less to read.

## The invariant shape

A Contract's guarantees are `invariants`. An invariant is:

```
{
    id:   "kebab-id"                       // required, unique within the Contract
    text: "an observable guarantee"        // required for plain/returns; optional for rejects
    when: "a guard condition"              // optional — a conditional guarantee
    kind: "plain" | "returns" | "rejects"  // defaults to "plain"
    // satisfies / depends_on / decided_by attach here
}
```

- **plain** (the default): an always-holds guarantee.
  `{ id: "survives", text: "the context survives across invocations" }`.
- **`returns`**: a property of what the caller gets back.
  `{ id: "returns-status", kind: "returns", text: "the node is returned with its current status" }`.
- **`rejects`**: a refusal-to-serve under a condition. `when` is **required**;
  `text` is **optional** (the meaning is "when `<when>`, refused" — write `text`
  only for extra detail).
  `{ id: "name-is-unique", kind: "rejects", when: "a context with that name already exists" }`.
- **`when` without a `kind`**: a conditional behaviour.
  `{ id: "incremental", when: "the spec changed since the last build", text: "the graph is rebuilt" }`.

**Never author a negative guarantee** ("does not alter", "is left untouched"):
it has no line to bind and cannot reach `proven` (MANIFESTO 1.6). Read-only-ness
is *derived* from the absence of a write edge. Reformulate positively (what the
call *does* return or refuse), or drop it and let the graph derive it.

## Authoring discipline (lessons from the dogfood)

- **A Need's description is one coherent want, not a list.** If you write
  `description: "to X, see Y, and check Z, so that …"` you have three Needs
  or three FRs smuggled into one. The dogfood blind-comprehension test
  caught this; pull it apart into separate FRs.
- **FR text names concrete domain entities** (Contract, Need, ADR, Port,
  Plan, code binding), not abstractions like "node" or "relationship". The
  audience does not come to manipulate nodes — it comes for ADRs to cite,
  Contracts to implement, Plans to land.
- **Atomic invariants only.** One guarantee per invariant. A composite ("the
  query runs against a projection and cannot mutate the graph") splits into
  two named invariants.
- **Schema is not re-stated.** "A new node has an identity unique within its
  module" is a property the schema enforces; an invariant says it as
  something *the tool does*: "Two nodes that share a slug within the same
  module are reported as a failure."

## Needs and their atoms

A Need declares its FRs and NFRs as **named CUE fields**, not list
entries — a satisfies edge points at the atom's *definition*, so renaming
the atom updates every reference, and the editor's go-to-definition jumps
straight to the source. The struct key is opaque (use `"fr-01"` or
`fr_idempotent` — read what reads well); the wire id lives in the atom's
own `id` field.

```
frs: {
    "fr-01": {id: "fr-01", text: "A named context can be created and switched between."},
    "fr-02": {id: "fr-02", text: "A module is added to or removed from the current context."},
}
```

## satisfies and decided_by

`satisfies` ties a specific Contract invariant to
a specific Need FR — as a **bare cue-native reference** into the Need's
`frs`/`nfrs` struct. The loader recovers both the owning Need and the wire
atom id from the reference; the author never repeats them.

```
satisfies: [
    agent.navigate.frs."fr-02",
    govaud.decisionKeeper.frs."fr-01",
]
```

`decided_by` cites the ADR that justifies the invariant. It is element-scoped,
not Need-scoped — the rationale lives on the contract, not on the intent.

Pure-rationale invariants (no satisfies) are fine and common: they are
internal guarantees with `decided_by` only, justifying the contract's shape
without being explicitly required by a Need.

## Asserted is honest

A Contract with no code annotation is `asserted` — a declared contract
waiting on code. A Need whose FRs are all satisfied by `asserted` Contracts
is `uncovered` until at least one of them turns `proven`. Do not chase the
colours: an honest `asserted` Contract is the right state when the contract
has been agreed but not yet built (e.g. `attest-bindings` here).

## Workflow

1. Edit or create a `.cue` file under the right module/folder.
2. `validate` for fast structural check.
3. `describe <module:slug>` to see how the tool resolved the new node —
   satisfies, realizes, edges. This is the smoke test that the wiring is
   what you intended.
4. `query "SELECT id, status FROM nodes WHERE ..."` for coverage spot-checks.
5. If the change touches what the editor's cue lsp resolves, the implicit
   warm runs on validate; `SPECUE_WARM_DEBUG=1` surfaces a warm error if
   resolution silently drifts.

## A reference triple

The smallest working unit is a domain Need + a service Contract + a governance
ADR, with `satisfies` from a Contract invariant to the Need FR and `decided_by`
pointing at the ADR. `spec.d/domain/agent/agent.cue:as-agent-navigate` plus
`spec.d/service/navigation/navigation.cue:queryGraph` plus
`spec.d/governance/adr.cue:adr02SQLQuery` is the live example in this repo —
read it as a template.
