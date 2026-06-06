# Target model — where the synthesis points

A consolidated picture of the model Specue is converging toward, drawn from this
session's decisions. Not a new design — a synthesis of conclusions already
reached, each tagged with where it came from. Speculative and pre-release; this
is the destination, not a committed spec.

Legend for provenance: [ADR] manifesto/ADR decision · [fact] verified against
self-spec/schema · [fan-out] found by multi-agent analysis · [roadmap] tracked item.

---

## 1. Node types (the ontology)

| node | role | change from today |
|---|---|---|
| **Contract** (was `UseCase`) | a logical contract: named observable guarantees | **rename** [ADR-13]; carries DbC triad |
| **Need** (+ **Domain**) | audience intent + its atoms (FR/NFR) | unchanged [ADR-10] |
| **Port** | typed transport surface across a boundary | **+kinds**: filesystem, objectstore, secretstore, (clock) [fan-out] |
| **Container** | boundary box (actor/system) | **+invariants** allowed (G3) [fan-out] |
| **Plan** | speculative change as branches | unchanged |
| **ADR** | the why-layer | unchanged |

Rejected as new node types [fan-out]: stocks-as-dynamics (→ monitoring), rate-
limiter / circuit-breaker / retry (→ invariants, behavior not surface),
queue/stream/pubsub as distinct Ports (→ one `channel` + edge role + invariant).

The criterion that holds the ontology closed (manifesto 7.4): **a Port is a
noun-surface; a behavior/policy on the flow is an invariant; WHEN/WHERE/stocks-
dynamics are out (planner / infra / monitoring).** New node types are rare and
must be observable promises, not implementation or runtime.

## 2. Contract elements — the core shape

**The axis: a Contract is a set of invariants. An invariant is either
unconditional or conditional (a `when` guard). Nothing else is a separate
element kind.** [decision this session]

```
Contract {
  invariants: [ {
    id, text, rev?, edges
    when?   : string            // optional guard — folds in the old #variation
    kind?   : "returns" | "rejects"   // optional type; absent = a plain guarantee
  } ]
}
// no preconditions / postconditions / variations as separate slots — all folded in
```

**One element kind: the invariant.** Everything else collapses into it. The
shape is: a text, an optional `when` guard, an optional `kind`. [decision this
session, converged over several rounds]

What folds in, and why:

- **`variation` → optional `when` on the invariant.** A variation was already
  `{id, when, then, rev, edges}` — an invariant with a guard. `when/then` = EARS
  / Given-When-Then, the recognized form for conditional guarantees [fan-out].
  `variations: 0` today = under-used, not unneeded [fact].

- **`postconditions` → gone.** A postcondition is always either a returned value
  (→ an invariant, `kind: returns`), a state change (→ derived from a `write`
  edge), or a branch outcome (→ in a guarded invariant's text). It has no
  irreducible content of its own [fact: post ≈ invariants, 25/25]. Dropped.

- **`preconditions` → gone, absorbed into `rejects`.** A precondition is the
  caller's obligation; what matters observably is *what happens when it is
  violated* — a rejection. "session must be valid" becomes an invariant
  `when: "session is invalid", kind: rejects`. Keeping a precondition slot *and*
  a rejection invariant would duplicate the same text. [user insight] So the
  obligation-direction of DbC survives — as the `rejects` branch, not a slot.

- **`trigger` stays** — activation, not DbC. The only "when invoked" source for
  the 4 addressee-less operation Contracts [fact].

### The two types — and only two

`kind` has exactly two values, because only two natures are both **positively
provable** (manifesto 1.6) **and not derivable from edges** (rule 5.2):

- **`returns`** — a property of what the caller gets back ("result is a single
  verdict"). Provable (bind the producing line); not an edge fact.
- **`rejects`** — a refusal under a condition (23 of 50 invariants in the
  self-spec [fact]). Provable (bind the refusing line); not an edge fact. Closes
  M7 — the error surface is *derived* from the rejects-invariants, and an error
  *code* is NOT authored here (it is implementation detail in code) [user].

Everything else is a **plain invariant (no `kind`)** or **derived**:

| nature | where it lives | why not a `kind` |
|---|---|---|
| `mutates` | derived from edge `write`/`produce` | already in the graph (5.2) |
| `calls` | derived from edge `call`/`serve` | already in the graph (5.2) |
| **negative** ("does not alter", "left untouched") | derived from *absence* of a `write` edge | not positively provable — never an invariant (1.6) [user insight] |
| persistent state ("survives invocations") | plain invariant, no `kind` | provable but needs no type — it is the default |

The two filters that produced this: **derived-vs-declared** (5.2 — mutates/calls
are edge facts) and **positively-provable** (1.6 — a negative guarantee has no
line to bind, so it is derived, never authored). What survives both as an
*authored type* is exactly `returns` and `rejects`.

## 3. Nature at query time — mostly derived

The real need ("filter a long contract by returns vs state-changes vs
rejections") is met by **deriving nature at query time** [fan-out: three agents
converged; 5.2 / 4.2], with only `returns`/`rejects` carried as an authored
`kind`:

| nature | source |
|---|---|
| `mutates` / `calls` | **derived** from infra edge role |
| negative (non-mutation) | **derived** from absence of a write edge (1.6) |
| `returns` / `rejects` | the authored `kind` on the invariant |
| `conditional` | presence of `when` |

So `query "... WHERE nature='mutates'"` works with **zero authoring burden, zero
drift, zero second source of truth** — exactly the posture used for coverage and
status. A closed authored `facet` enum was **rejected** [fan-out]: 4 of its 6
values redeclare edges/slots (4.2/5.2 violation), one is a dumping ground, one is
the rejects gap.

## 4. The one real schema gap — structured rejection (= M7)

`rejects` is the single nature with no home today [fan-out]. A Contract that
refuses under a condition ("name taken → refused", "over quota → 429") can only
say so in prose. This is **the same need as M7 (returnable errors)** [roadmap],
now precisely scoped: a **structured denial/error marker** on a conditional
invariant — `when X → rejects with code E` — from which the Contract's *error
surface* is **derived** (the list of codes it may return), not authored.

## 5. Edges — element-grained and two-sided

- **G2 [roadmap, fact]:** `depends_on`/`carries` should target an *element*
  (a specific invariant), as `satisfies` already targets a Need atom — so a
  Contract depending on one guarantee of another doesn't over-couple. One-line
  type relaxation; the resolver already supports it. Fold into the M3 pass.
- **M-edge [roadmap]:** cross-module edges become two-sided (target sees/assents
  to incoming claims); `carries` becomes a list with status propagation — the
  "responsibility is not atomic" finding.
- **Reuse:** a shared invariant is a CUE `_`-value reused in authoring [fact:
  verified — gives two independent graph nodes, which is correct: an invariant
  belongs to its Contract]. Cross-Contract reliance is `depends_on`+G2, not a
  shared node.

## 6. Boundary & federation (the soft-limit frontier)

- **context-as-node with `decided_by`** [roadmap M-ctx]: the boundary judgment
  (which modules are the system) becomes a versioned, comparable artifact, not
  runtime config — the highest-leverage federation step [manifesto §8].
- **dissent / agreement axis** [roadmap M-dis]: record that a contract was
  contested (residue of negotiation, not its log — manifesto 7.3), so `proven`
  stops masking "but I was against it".
- **attest-bindings** [roadmap M2]: read a spec without the code.

## 7. Schema hygiene (remove the rudiments)

- `legacy_id` — migration alias; migrator deleted, decide drop end-to-end [M4].
- `visibility` — redundant with CUE privacy; derive or drop [M4].
- `migrate` package + `cmd/v1tov2` — **done, removed** (−1159 LOC) [fact].

## 8. What stays OUT (the discipline that keeps it small)

Named so the model does not creep [manifesto 7.4 / 4.2, fan-out]:
- **WHEN** (scheduling, deadlines, ordering) → planner (Jira/Notion).
- **WHERE** (region, node, deployment) → infra (Terraform/k8s).
- **Stocks-as-dynamics** (queue depth now, rates, simulation) → monitoring
  (Grafana) and System-Dynamics tools. Only the *contract about* a stock
  ("bounded", "no loss", "conserves") is in, as an invariant on a Port.
- **Agreement / the negotiation log** → human table; Specue keeps only the
  residue (the agreed contract), holds it steady against code drift.
- **Runtime values, library detail, config** → HOW, via code annotations.

---

## The shape in one paragraph

A **Contract** (renamed from UseCase) is a set of **invariants**, each
unconditional or guarded by a `when`, each optionally typed `returns` or
`rejects` (a refusal-to-serve); a caller obligation survives as the `rejects`
branch that fires when it is violated, not as a separate slot. The two **edge-
or-absence natures** — `mutates`/`calls` (from infra edges) and the negative
non-mutation (from the *absence* of a write edge) — are **derived** at query
time, never authored (MANIFESTO 1.6); only `returns`/`rejects` are an authored
`kind`. Contracts wire
to each other by **element-grained, two-sided edges**, to intent by `satisfies`,
to reasoning by `decided_by`. **Ports** gained a few real surface kinds;
**Containers** can carry invariants. The **boundary** itself becomes a versioned
node. Everything quantitative-over-time, scheduled, placed, or still-being-argued
stays in the tool that owns it — Specue holds the observable contract and proves
the code keeps it.
