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
The node carries the Design-by-Contract triad (pre/post/invariants); the term
`UseCase` imports UML interaction-scenario baggage it does not honour — the same
objection ADR-10 raised against UserStory. `Contract` matches the schema and
keeps internal (operation) contracts legal. *Fits 1.3 / 3.1.* Breaking; write
ADR-13 first (alternatives considered: Capability — rejected, no place for
pre/post; it answers "what the system *can*", not "what it *guarantees*").

### M4. Schema consistency — remove rudimentary / confusing fields
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

Audit the remaining optional fields in the same pass for the same "could an
author be confused?" test. Best done with M3's schema pass (same breaking
window).

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
- **M7. Returnable-errors layer.** Which error codes a contract may return, as
  reusable CUE values. *Fits as an observable property of the contract (WHAT) —
  "callers see code E on condition X" is third-party-checkable.* Needs a schema
  slot; ADR-light.
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
- **M-edge. Two-sided cross-module edges + `carries` as a list.** A `satisfies`
  edge is a one-sided claim (manifesto 8.3); give the target visibility and a
  way to assent/dissent. Make `carries` a list (an invariant can rest on several
  hands — app-code *and* infra) and propagate status along it, so a broken
  upstream contract reddens its dependents and responsibility stops being hidden
  in one invariant. *Fits the "responsibility is not atomic" finding; needs
  M-ctx's boundary work first.*

## Later (model — parked)

- **M10. Custom properties + custom renderers.** User-defined fields and render
  hooks. *Powerful but risks reopening the "what may be a node" discipline (the
  size-vs-kind boundary); design last, after the core vocabulary settles.*

## Rejected / fold-in (model)

- **Explicit mutating vs read-only contracts.** *Rejected as a new field:
  already derivable — whether a contract mutates is visible from its infra edges
  (`write`/`produce`). Computed, not declared (5.2). Surface it in `query`, do
  not author it.*
- **Per-edge ADR / authoring provenance (who published, when).** *Rejected:
  largely git history already (manifesto's git-native stance). Add only the thin
  structural owner if M2 federation needs it, never a hand-kept audit log.*

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
