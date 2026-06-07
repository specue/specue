# Roadmap

Pre-release. The model is speculative and breaking changes are expected (see the
[pre-release note](README.md#specue)). This roadmap filters a backlog of ideas
through the [manifesto](MANIFESTO.md): an item earns a place only if it fits the
foundation, and items that violate it are listed as **rejected**, with the rule
they break — so the "no" is as documented as the "yes".

Two axes, kept separate on purpose:

- **Model** — changes to semantics / schema / vocabulary. Breaking; each needs
  an ADR before it ships.
- **Tooling** — LSP, MCP, CI, performance, editor. Does not touch the model.

Horizon: **Now** is in focus and reasoned out. **Next / Later** are parked with
a one-line rationale, not yet designed.

---

# MODEL

## Now (foundation-critical — the manifesto already points here)

### M1. Federation that does not block `validate`
A contract depending on a module outside the active context must report an
**honest gap**, never a hard failure. *Fits manifesto 8.2 ("unknown from here" is
a legitimate state) — this is making rule 8.2 true in code.* Today's behaviour is
closer to error than gap; this is the first federation step and unblocks M2.

### M2. Code attestation (`attest-bindings`)
A code holder publishes per-Contract binding outcomes (no source); a reader
without the code consumes that instead of scanning. *Fits 8.3 — it is already
"contract-agreed, not built" in Status. Closes the `as-federated-reader` Need.*
Depends on M1.

### M3. Rename `UseCase` → `Contract` (ADR-13)
The node is a set of observable invariants a service guarantees; the term
`UseCase` imports UML interaction-scenario baggage it does not honour — the same
objection ADR-10 raised against UserStory. `Contract` reads correctly for a
guarantee (and for an internal operation contract, normal in Design-by-Contract,
where "use case without an actor" is an oxymoron). *Fits 1.3 / 3.1.* Breaking;
write ADR-13 first. Alternatives considered: Capability — rejected, it answers
"what the system *can* do", not "what it *guarantees*". *Note: the DbC-triad
argument is weaker than it first looks — pre/post are being removed by M-elem
(invariants are the substance); ADR-13's body should rest on "Contract = a set of
provable invariants", not on a full pre/post/invariant triad.*

### M4. Schema consistency — remove rudimentary / confusing fields — DONE except `visibility` (ADR-14, ADR-15)
*`legacy_id` removed end-to-end on `plan/m-elem` (schema + model + source +
jsonir + the `migrate-legacy` diagnostic; ADR-14 window).*
*`binding`, the ungated typed refs, and the role-untyped `#dep.to` all closed on
`plan/m4-schema-hygiene` (ADR-15): `binding` dropped entirely (it was 1 bit —
`abstract` vs not — and `abstract` is what ADR/Plan nodes already express, so a
Contract with no code is now an honest gap); `service!: #Container` and
`domain!: #Domain` typed so CUE rejects a mis-aimed edge; `#dep.to` gated by
role (plain → `#Contract | #invariant`, infra → `#Port | #Container`). The
role-gate surfaced two test fixtures that pointed an infra `role:"call"` at a
Contract — exactly the upside-down edge M-edge-invert removes — reshaped to plain
deps. Self-spec validates 72 nodes / 0 advisories with all three in force.*
*`visibility` is KEPT: it is load-bearing (the bindings gate filters Contracts to
Public, stamped from the `internal/` path); dropping or deriving it from CUE
visibility is a separate, larger change — still open below.*
****
Stray fields make the model **ambiguous to read** — an author hits a field and
wonders if it is theirs to set. While the model is still breaking, removing
confusion is cheaper now than after adoption. *Fits the manifesto's drive for an
honest, minimal vocabulary.* Concrete targets:

- **`legacy_id`** — a transitional alias from the (now-deleted) v1→v2 migration.
  The `internal/migrate` package and `cmd/v1tov2` are gone, but `legacy_id`
  still threads through `model.Node`, `source/node.go`, `jsonir/schema.go`,
  `query`, and the `migrate-legacy` diagnostic in `compiler`. With the migrator
  removed, the question is sharper: does any live spec still resolve refs
  through a `legacy_id` alias? If no, drop it end to end (schema + model +
  diagnostic). If a few legacy aliases remain in real specs, keep only the
  resolver path and drop it from the authored surface.

- **`visibility`** — a `public|private` field on nodes (carried in `model.Node`,
  the query projection, `jsonir`, ~25 refs across 13 files) that is **not in the
  authored schema** and is **not set anywhere in the self-spec**. Privacy is
  already a CUE concern — unexported (`_`-prefixed) definitions and CUE's own
  visibility rules (ADR-01) are the real source of truth, and M6
  (`_someReusableContract`) leans on exactly that. So `visibility` is a second,
  redundant privacy mechanism shadowing CUE's. Decision: derive it from CUE
  visibility if it must be exposed at all (computed, per 5.2), or drop the field
  outright. Either way it should stop being a standalone node attribute.

- **`binding: required | optional | abstract`** on `#Contract` — DONE (ADR-15):
  dropped entirely. *Was:* every self-spec
  contract uses the default `required`; **none** authors `optional` or `abstract`.
  Worse, the tool only branches on `abstract` (one site: `fixpoint.go` — an
  abstract contract never blocks and is never a gap) and **never branches on
  `optional`** at all: an `optional` contract is treated identically to a
  `required` one. So `optional` is a dead value (violates the node.go rule that a
  closed enum exists only when the tool reasons about it), and `abstract` is wired
  but exercised by no self-spec instance and no test. Decision: either give
  `optional` real meaning (relax the "no code = gap" pressure — nothing implements
  this today) or drop it, leaving `required | abstract`; and add an `abstract`
  self-spec instance + test if the value is kept. Found dogfooding M-elem.

- **Typed refs accept any node (`service`/`domain`/`schema`)** — DONE for
  `service`/`domain` (ADR-15: `service!: #Container`, `domain!: #Domain`);
  `#Port.schema` stays `#Node` (no dedicated IDL node type yet). *Was:* `#Contract.service`,
  `#Need.domain` and `#Port.schema` are all typed `#Node` (the any-node union), so
  CUE accepts `service: someADR` or `domain: aContract`, and the compiler only
  checks the target *resolves*, not its type (`dangling.go` adds `uc.Service`
  without a kind check). This is an ungated edge where the model elsewhere gates
  (`satisfies`→atom, binding→Contract). **Fix: tighten the CUE types** —
  `service!: #Container`, `domain!: #Domain`, `schema?` to the IDL node — so CUE
  itself rejects the mistake. *Verified de-risked: tightening `service!: #Container`
  keeps the self-spec validating (70 nodes, 0 advisories) and cross-module
  Container refs resolve — it is a one-line type change per field, no compiler
  work needed.* Found dogfooding M-elem.

- **`#dep.to` accepts any node, regardless of role** — DONE (ADR-15): gated by
  role exactly as drafted below. *Was:* a plain dep (no `role`)
  should target a `#Contract`; an infra dep (`role` set) should target a
  `#Port | #Container`. Today `to!: #Node | #invariant` accepts either for both,
  and the compiler silently ignores a mis-typed infra target (`topology.go` skips
  a non-Port). **This IS expressible in CUE** (the type depends on a sibling
  field, exactly the `#Port if kind ==` idiom) — `role` is optional with no
  default, but `if role == _|_ {...}` / `if role != _|_ {...}` gates it:
  ```
  #dep: {
      role?: #role
      if role == _|_ { to!: #Contract | #invariant }
      if role != _|_ { to!: #Port | #Container }
      carries?: #Node | #invariant
  }
  ```
  *Verified de-risked: this shape keeps the self-spec validating (70 nodes, 0
  advisories).* Found dogfooding M-elem.

- **`carries` is wired but unused — and points at the edge being inverted.** It
  is read only by `dangling.go` (resolve-check) and echoed to JSON; `topology.go`
  (its supposed L3→L2 consumer) never reads it — L2 is built from `to`+`role`
  alone. Zero instances in the self-spec. But the deeper issue is *why* it feels
  bolted-on → see **M-edge-invert** below. (The G2 widening of `carries` is
  harmless but moot if the field dissolves.)

### M-edge-invert. An infra dep targets a Contract; the transport is `over` (DESIGN)
The current infra edge is upside-down. Today `depends_on` puts the **mechanism**
in `to` (a Port) with `role`, and bolts the **real target** (the Contract being
relied on) onto a side field `carries`. That is why `carries` reads awkwardly: it
is the actual dependency, demoted to an afterthought, while `to` holds an
implementation detail.

Invert it. `to` always names **what** is depended on — a Contract (the logical
dependency, primary). *How* that dependency is realized — the physical transport —
moves to an optional `over`:
```
depends_on: [{
    to: b.someContract                    // WHAT — a logical Contract dependency
    over?: { port: somePort, role: produce }  // HOW — the physical realization (optional)
}]
```
- a plain `depends_on` (no `over`) = a pure Contract→Contract dependency.
- with `over` = the same dependency, realized through a Port with a role.
- **`carries` dissolves** — it *was* the real `to`; inverting makes it the `to`.
- L2 topology derives from `over.port` (as it derives from `to` today), but the
  semantic edge now points at a model node, never an infra node — consistent with
  the §service/domain/schema gate (every typed ref targets a model type).

Bigger than the M4 hygiene items: touches schema (`#dep`), loader, `topology.go`
(read `over.port` not `to`), diff sig, jsonir. Supersedes M-edge's "carries as a
list" framing (there is no carries to listify). *Found dogfooding M-elem — the
question "why carries?" has no good answer because the edge is inverted.*

Audit the remaining optional fields in the same pass for the same "could an
author be confused?" test. Best done with M3's schema pass (same breaking
window).

### M-elem. Collapse contract elements to one shape: invariant{when?, kind?} — DONE (ADR-14)
*Landed on `plan/m-elem` together with M4 and G2; ADR-14 records it. The
self-spec re-typing is documented per-contract in
[M-ELEM-MIGRATION.md](M-ELEM-MIGRATION.md).*

The central model decision of this design pass. Today a Contract carries four
element slots — `invariants`, `variations`, `preconditions`, `postconditions`.
They collapse to **one**: the invariant, with an optional `when` guard and an
optional `kind`. See [MODEL-TARGET.md](MODEL-TARGET.md) §2 for the full shape.
- **`variation` → optional `when`** on the invariant (it was already an
  invariant with a guard; `when/then` = EARS / Given-When-Then).
- **`postconditions` → gone** — a postcondition is always a returned value
  (invariant), a state change (derived from a `write` edge), or a branch outcome
  (in a guarded invariant's text). No irreducible content [fact: post≈inv 25/25].
- **`preconditions` → gone**, absorbed into a `rejects` branch: "session must be
  valid" becomes `when:"session invalid", kind:rejects`. A precondition slot
  *and* a rejection would duplicate the text [user insight].
- **`kind` has exactly two values — `returns` | `rejects`** — the only natures
  that are both positively provable (manifesto 1.6) and not derivable from edges
  (5.2). `mutates`/`calls` are derived from infra roles; negative guarantees
  ("does not alter") are derived from the *absence* of a write edge (1.6) and are
  never authored; persistent-state properties are plain invariants (no kind).
*Breaking; the heart of the M3 schema-pass. Fold M3/M4/G2/M-elem into one window.
This supersedes the earlier "use the pre/post slots" idea — they are removed, not
populated.*

### M5. Emergent system promises — node + derived check (DESIGN OPEN)
A promise of the *whole* that no single Contract carries ("the system stays
available under single-node loss"), which under federation has no obvious home.
*Tension with 1.2 / 5.2: a hand-declared "emergent" node is a contradiction —
emergence must arise from parts, not be postulated.* So this is **not** "just add
a node." Open design question to resolve in an ADR before building:
- an **anchor node** (names the whole-system promise) whose **status is
  derived** from the Contracts/edges that compose it — emergence as a *function*,
  like coverage, with a node only as the addressable handle;
- vs. a declared node (rejected unless it carries a derivation, else it is an
  unverifiable Contract with a loud name).
This is the deepest item; it gates how far Specue reaches beyond one service.

### M-Q. The query model is a public contract and under-designed (DESIGN OPEN)
The SQL projection is exposed to users/agents (the `query` verb teaches CTEs,
FTS5, `query tables`), so its schema is **part of the interface**, not an
internal detail — yet it is rough: table names and columns are easy to get
wrong (`dep_edges`/`satisfies`/`infra_edges` instead of an `edges` table; no
`src`/`dst`). Fixing it is **breaking** for anyone who wrote queries. Decide the
target before churning it. Three options, by ascending power and descending
fidelity to "small tool" (manifesto 4.2):
1. **Tidy the SQL projection in place** — one coherent edge model (a unified
   `edges(from,to,kind,...)` view over the three tables), stable documented
   columns, keep in-memory SQLite (pure-Go, no CGO, FTS5 free). Cheapest, keeps
   the single binary.
2. **Embedded graph query** — a Cypher-like / native graph traversal over the
   in-memory model, no DB server. Natural for a node/edge model (blast-radius,
   transitive deps stop needing recursive CTEs) but is a new query surface to
   build and document.
3. **Embedded graph DB — Dgraph/modusGraph (CGO gate PASSED; two new trades).**
   `github.com/hypermodeinc/modusgraph` runs Dgraph **in-process, file-based**
   (`NewEngine(NewDefaultConfig(dir))`), **no server, no daemon**. Apache-2.0,
   no license key as of Dgraph v25. Real graph semantics + DQL/Cypher traversals,
   edges first-class. (Bonus: Dgraph v25 ships an MCP server, adjacent to T3.)
   **Spike result (verified, not assumed):** builds clean with `CGO_ENABLED=0`
   both natively and cross-compiled to linux/amd64; `runtime/cgo` is absent from
   the dep tree; the no-CGO binary opens the embedded engine at runtime. So the
   CGO gate that would have killed this is **passed — it stays pure-Go and
   single-binary (no conflict with 4.2).** Two trades surfaced by the spike
   instead:
   - **Footprint.** ~42–46 MB binary and **~798 transitive deps** (gRPC, OTel,
     Badger, protobuf) vs today's light SQLite projector. Real weight for graph
     semantics.
   - **Persistence breaks the ephemeral model.** Today's projection is
     `:memory:`, rebuilt from the graph in ms — the graph is the only source of
     truth, the query DB is throwaway. modusGraph is **file-based on disk**, so
     there is now DB state to own (when to rebuild, how to invalidate). Not a
     blocker, but an architectural shift, not a drop-in.
   *Net: viable and the strongest long-term graph option — but earns its weight
   only if traversal power / Cypher / MCP matter enough to justify +40 MB and a
   persistence story.*
4. **External graph server (Memgraph / KùzuDB-CGO / etc.)** — Cypher, edges
   first-class, but a **server** (bolt/Docker) or CGO. Breaks "one `go install`,
   no daemon," loses free FTS5, turns an ephemeral ms-rebuild into
   spin-up-and-load — a direct conflict with 4.2. Reconsider *only if* T1 daemon
   mode lands and amortizes it. Strictly worse than option 3 for our footprint.

The engine swap itself is cheap (the projection is one isolated package rebuilt
from the graph each run) — so the decision is about **footprint and query
surface**, not migration cost.
*Recommendation (post-spike): option 1 now — tidy the SQL projection, keep the
light pure-Go `:memory:` deployment, freeze a documented stable schema. The CGO
spike on option 3 passed, so **option 4 (external server) is dropped** —
embedded Dgraph dominates it on footprint. Option 3 is the serious long-term
target the day graph traversals, Cypher, or the Dgraph MCP server are worth
+40 MB and a persistence story; until then its weight is not justified. Decide 1
vs 3 on need, not on feasibility — feasibility is settled.* Whatever wins, the
query schema becomes a documented, stable contract so it stops being a moving
target.

## Next (model, fits — not yet urgent)

- **M6. Private/reusable contract references.** Resolve CUE-native references to
  unexported contracts (`_someReusableContract`) so one contract can build on
  another without exposing it. *Fits 1.1 (visibility is already in the model);
  mostly a resolver capability.*
- **M7. Returnable errors — SUBSUMED by M-elem `kind:rejects`.** *Superseded:
  the original "errors layer with reusable codes" is now just the `rejects`
  invariants from M-elem. A contract's error surface (which conditions it refuses
  under) is **derived** from its `rejects`-kind invariants — not a separate slot.
  An error **code** is NOT authored in the spec: it is implementation detail that
  lives in code [user]. So M7 needs no new structure beyond M-elem's `kind`;
  what remains is a derived "error surface" view in query/describe.*
- **M8. File-backed node text.** Let `body`/ADR text point at a file (large doc,
  own format) instead of inline string. *Tooling-adjacent but touches schema;
  keep the rendered graph the source of truth (ADR-09).*
- **M9. Built-in tags on nodes.** First-class tag field for grouping/filtering.
  *Small schema add; check it does not become a second, informal type system.*

### M-addr. Compiler gate: every Contract must have an addressee
A Contract must satisfy a Need **or** be invoked by another Contract
(`depends_on`); a Contract addressing no one is the leaked-HOW case (manifesto
1.1 / 1.3). *Why now: this is the gate that makes the manifesto's WHAT-purity
claim true by construction instead of by prose — the self-spec has 4 such
orphans today (the Plan/context operation verbs), which must then either gain an
addressee or be marked operation-Contracts explicitly.* Pairs with M3 (the
rename) and M4 (the consistency pass).

## Next (model — agreement & responsibility, fits the federation direction)

These came out of the foundation discussion; they make the soft-limits named in
manifesto §8 into real structure. None is urgent, all fit.

- **M-dis. A `dissent` / agreement axis on contracts.** Record that a contract
  was *contested* and by whom — the residue of a negotiation, not its log
  (manifesto 7.3). Makes `proven` stop masking "but I was against it." *Cheap
  schema add; the hard part is resisting scope creep into hosting the argument.*
- **M-ctx. Context as a first-class node with `decided_by`.** Today a `context`
  is runtime config in `$SPECUE_HOME`, invisible and unversioned. Promote it to a
  graph node so a boundary judgment (manifesto §8) becomes comparable across
  observers and carries its rationale. *The single highest-leverage federation
  item — it turns "where the system ends" from a private lens into a shared,
  diffable artifact.*
- **M-edge. Two-sided cross-module edges.** A `satisfies` edge is a one-sided
  claim (manifesto 8.3); give the target visibility and a way to assent/dissent,
  and propagate status along dependency edges so a broken upstream contract
  reddens its dependents and responsibility stops being hidden in one invariant.
  *Fits the "responsibility is not atomic" finding; needs M-ctx's boundary work
  first.* (The earlier "carries as a list" framing is dropped — see
  **M-edge-invert**, which removes `carries` rather than listifying it.)

## Later (model — parked)

- **M10. Custom properties + custom renderers.** User-defined fields and render
  hooks. *Powerful but risks reopening the "what may be a node" discipline (the
  size-vs-kind boundary); design last, after the core vocabulary settles.*

- **M11. Schema migration is incremental and idempotent (WHY now, WHAT later).**
  ADR-06 already says a breaking schema change ships as a new major and authors
  migrate at their own pace. The properties "a migration is **idempotent**
  (running it twice equals once) and **incremental** (v1→v3 goes through v2, not
  a leap)" are observable contracts (manifesto 1.2) — a third party checks them
  by behaviour, not by reading code. But there is **no migrator today** (we
  deleted v1→v2), so a `migrate-schema` Contract now would be an addressee-less
  asserted contract with no HOW. Correct placement, by layer:
  - **WHY (do this when the question comes up):** extend/添 an ADR off ADR-06
    stating migrations must be incremental + idempotent, with the reasoning
    (authors migrate apart; a re-run must not corrupt). This is the durable part.
  - **WHAT (defer until a migrator exists):** when a schema-migration mechanism
    is actually built, add a `migrate-schema` Contract whose invariants are
    `idempotent` / `incremental`, `decided_by` that ADR, bound to the code.
  *Do not author the Contract before the code — that is a promise with no
  addressee (rule 1.3). Record the WHY; materialize the WHAT with the migrator.*

## Expressiveness gaps (found by fan-out analysis + schema check)

Five gaps surfaced by enumerating Port kinds, stock/resource types, and
cross-cutting behaviors against the model, then attacking each adversarially.
Four of five survive; most "missing" features were rejected (see below) as
already-present, already-roadmapped, or out-of-scope by rule 7.4. The surviving
gaps share **one theme**: the model is *qualitatively* rich (what is guaranteed)
but *quantitatively* mute (how much / how long / how many at once). G1, G4, G5
are facets of that; G2 is a cheap structural fix; G3 is small.

### G1. Quantified capacity has no slot (strongest)
Every "bounded to N", "≤ N concurrent", "1000 rps", "retained 90 days" lives as
**prose inside an invariant's `text`** — yet it *is* a number-with-a-unit (the
defining trait of a System-Dynamics stock). Consequence: two contracts can't be
compared, a limit can't be checked against code, tightening it reads as a text
edit. The WHAT here is **too vague to check** — the mirror of rule 2.2. Touches
rate-limit, quota, pool size, retention, timeout. *Open design: a typed
quantity/bound/unit shape on an invariant, without dragging in runtime values
(those stay in monitoring, 7.4).*

### G2. Element-grained `depends_on` / `carries` (cheapest — one-line fix) — DONE
*Landed on `plan/m-elem`. `#dep.to`/`#dep.carries` now accept `#Node | #invariant`;
`mapRef` recovers the owning node when a dep targets an element, so existing
node-target deps resolve unchanged. Full element-grained resolution (carrying the
target element id through the model) is a follow-up; the type now permits it.*

### G2. Element-grained `depends_on` / `carries` (cheapest — one-line fix)
`satisfies` can point at an atom (`frs."fr-01"`); `depends_on.to` and `carries`
are typed `#Node` — whole-node only. So Contract A depending on **one invariant**
of B over-couples: any change to B looks like a risk to A. **Fix:** widen the
`to`/`carries` reference type to accept an element (`#invariant`/`#condition`)
exactly as `satisfies` already accepts `#atom` — the CUE-native resolver already
dereferences these. No new node/edge, just relaxing a target type. *Fold into
the M3 schema-pass — it is nearly free.*

### G3. `Container` carries no invariants (small)
Invariants live only on Contracts (verified: `#Container` has no `invariants[]`).
So genuinely container-scoped observable promises — "this service runs in a
bounded working set / does not leak memory" (OOM is third-party-observable →
WHAT per 7.4) — have nowhere to attach. *Fix: allow invariants on Container, or
decide such promises must be a Contract about the service.*

### G4. No `lease`/`acquire`/`release` role; no leasable-resource Port kind
`#role` has read/write/produce/consume/grant but no **acquire → hold → release**.
The borrow-and-return discipline at the heart of concurrency stocks (connection
pool, lock, semaphore) can't be expressed as topology, and there is no Port
`kind` for a held, bounded-count resource — so locks/pools are homeless as
nodes. *Open design: a `lease` role family + possibly a `resource` Port kind.
Related to G1 (the bound) and to the responsibility work (M-edge).*

### G5. Liveness-with-deadline — re-examined under rule 1.6 (likely NOT a gap)
"X is erased after 90 days", "PII gone within 30 days of a delete request" fire
**without an interaction** and are verified by the *passage of time*. Originally
flagged as a missing element kind. **Reconsidered:** rule 1.6 (a WHAT must be
positively provable by binding the line that realizes it) makes these suspect —
"erased after 90 days" is verified by watching time pass, with no single line to
bind and no test that proves the guarantee (only one run). That is the same
non-provability that rules out negative guarantees. The provable, observable
re-statement is positive and interaction-shaped: "a *retention sweep* deletes
records older than 90 days" (bind the sweep) or "a *delete request* removes PII
within 30 days" (bind the request handler) — a guarded invariant
(`when: delete requested → then: PII removed`), already expressible. *Verdict:
probably NOT a new element kind — restate the time-promise as a provable
interaction. Keep only if a genuinely interaction-less, bindable temporal
guarantee is found; none so far.*

### Port-kind additions (from the same analysis)
The fan-out also vetted Port kinds. Accept as genuine new transport-surfaces the
current four cannot express: **`filesystem`** (path-addressed byte store),
**`objectstore`** (bucket/key blobs — distinct consistency contract),
**`secretstore`** (the `grant` role already presupposes it), and narrowly
**`clock`** (only when time crosses a boundary). *Reject* queue/stream/pubsub/
eventlog/websocket/SSE as new kinds — they are one surface (`channel`) with a
different **traffic shape**, already expressed by edge role + invariants; and
cache/searchindex/vectorstore/CDN/lock/ledger as `datastore`-or-`channel` +
invariant. Fold the accepted kinds into the M3/M4 schema-pass.

## Rejected / fold-in (model)

- **Explicit mutating vs read-only contracts.** *Rejected as a new field:
  already derivable — whether a contract mutates is visible from its infra edges
  (`write`/`produce`). Computed, not declared (5.2). Surface it in `query`, do
  not author it.*
- **Per-edge ADR / authoring provenance (who published, when).** *Rejected:
  largely git history already (manifesto's git-native stance). Add only the thin
  structural owner if M2 federation needs it, never a hand-kept audit log.*
- **Rejected expressiveness candidates (seemed like gaps, are not):**
  error/failure modes (= M7); SLO numbers (runtime → monitoring, 7.1); payload
  schema (already `Port.schema` for rpc/rest); trust boundary (already
  `Container.boundary`); ownership (= federation + module identity); deprecation
  (already `deprecated` field + confidence/status); temporal ordering of
  contracts (= WHEN, 7.4 → planner); cost (no observable grain); test-vectors on
  an invariant (= HOW/code → T9). *The model is tighter than it looks.*

---

# TOOLING (does not touch the model)

## Now

### T1. Performance / daemon mode
250 nodes across folders already validates and queries slowly. Investigate a
**server/daemon mode** (warm resolved graph, incremental revalidation) so
`validate`/`query` are fast on real-size landscapes. *Blocks adoption more than
any model gap; the graph is cached by content key already — extend that.*

## Next

- **T2. LSP server for bindings + VSCode extension.** Slug autocomplete inside
  `//specue:req:`, go-to-node, diagnostics. *Closes `as-author-dx`.*
- **T3. MCP server for the agent.** Expose the read verbs as MCP tools so an
  agent reaches the graph natively. *Aligns with manifesto §6 (agent is a
  first-class reader); the `--json`/stable-shape work is the substrate.*
- **T4. File-level bindings (not line-level).** Bind a whole file (prompts,
  generated/immutable files where a line number is meaningless). *Scanner change,
  not a model change.*
- **T5. CI hardening — reusable actions/jobs.** Package the validate/bindings
  gate as reusable CI. *Note: CI is currently GitLab; the repo is GitHub — pick
  the canonical CI home before investing.*

## Later

- **T6. Plan rollback.** Undo an accepted plan if feasible. *Depends on git
  mechanics; scope after federation settles, since plans span modules.*
- **T7. Plan-merge protection hooks.** Git hooks that refuse a manual merge of a
  plan branch without `plan accept`. *Guards the Plan lifecycle integrity.*
- **T8. Known-issues layer.** A surface for known defects/limitations per node.
  *Decide model vs tooling: if it is "this contract has an open caveat," it may
  belong to the model (a status/annotation), not tooling — revisit when reached.*
- **T9. Test-run references.** Link to test-run reports rather than running
  tests. *Running tests is out of scope (that is Alloy/CI territory); a
  reference/link is the honest, cheap version.*

## Rejected (tooling)

- **Auto-running tests inside Specue.** *Out of scope: Specue scans for the
  binding, it does not execute. Test execution belongs to CI; carry a reference
  to the run (T9), not the run itself.*

---

## How items graduate

An item moves from a list into a Plan (branches across affected modules) when:
its ADR is written (Model items), it has a Need it closes, and — for breaking
Model changes — a migration path exists. The Plan lifecycle is itself the
mechanism; this file only decides *what* is worth a Plan and *why it fits*.
