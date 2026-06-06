# MANIFESTO

Specue sorts a system into three layers — **WHAT** it promises, **HOW** the
code keeps the promise, **WHY** the promise is shaped that way. The layers are
*mostly* separable, and that separation is what makes coverage computable rather
than asserted. They are not perfectly orthogonal: the line between WHAT and HOW
is observer-relative and leaks at the edges (rule 1.4). The discipline below is
the set of rules that keep the leak small enough for the graph to stay honest.

Each rule is stated as a rule, with one line of *why* — because a rule you can
scan is worth more than a paragraph you skim, and a rule without its reason
rots into ritual.

---

## 1. WHAT — the promise, and to whom

**1.1 — A WHAT is a promise to an addressee.**
A `Need` carries an audience's intent (its FRs are atomic, observable
guarantees); a `UseCase` carries a contract a service guarantees (its invariants
are atomic, observable properties). *Why: a promise with no one to keep it for
is not a contract, it is a note to self.*

**1.2 — A WHAT is observational.**
Its text is a sentence a third party could check by watching the system from
outside. "A task added today is visible later in the same list" — yes. "AddTask
writes to Mongo" — no, that is HOW. *Why: an observable promise survives a
refactor; an implementation sentence breaks every time the code moves.*

**1.3 — A UseCase is two-natured; name which nature it is.**
A *contract UseCase* faces a `Need` — it is true WHAT. An *operation UseCase*
has no external audience; its addressee is the tool itself. Both are legal, but
an operation UseCase must still name its addressee, not pose as an external
contract. *Why: the self-spec has operation UseCases (the Plan/context verbs);
pretending they face an audience they do not is the dishonesty rule 1.1 exists
to catch.*

**1.4 — The WHAT/HOW line is observer-relative, and that is admitted, not
hidden.**
"working tree", "branch" are HOW to a user of a service — but domain vocabulary
to a tool whose subject *is* git. When the tool is its own addressee (1.3) its
domain words may appear in a WHAT. *Why: the boundary of "observable" depends on
who observes; a manifesto that denies this lies more than the leak it forbids.*

**1.5 — A WHAT with no matching HOW is the honest TODO.**
It is a status field, not a wiki page that promises to be updated. *Why: an
unkept promise that says so is information; one that pretends to be kept is rot.*

**1.6 — A WHAT must be positively provable; a negative guarantee is not a WHAT.**
An invariant is proven by binding the line that *realizes* it — so it must name
something that *happens*. "Reading does not alter the context", "the directory
is left untouched", "the call does not mutate" name an *absence* — there is no
line to bind, and a passing test proves one run, never the guarantee. Such a
property is never authored as an invariant. It is **derived**: a contract that
holds no `write`/`produce` infra edge is read-only by construction; the graph
already knows it. *Why: a non-event has no realizing HOW, so it can never reach
`proven` honestly — it would sit asserted forever, pretending. The model states
what the system does and derives what it does not; it never asks an author to
promise an absence it cannot prove.*

## 2. HOW — the realization in source

**2.1 — HOW lives in code, through annotations.**
`//specue:req:`, `//specue:test:`, `//specue:produces:`, `//specue:consumes:`
and kin bind the function that satisfies an invariant, the test that proves it,
the infra verb that anchors a Port. *Why: the realization belongs where it can
drift — in the source — so the binding catches the drift.*

**2.2 — A WHAT's text never names a function, file, constructor, library or
build tool.**
If it does, lift the HOW down into code where it belongs. (Subject to 1.4 for a
self-addressed tool.) *Why: every such name welds the contract to today's
implementation and makes tomorrow's refactor a spec edit.*

**2.3 — A binding aimed at anything but a UseCase is an error.**
Only a UseCase carries code bindings; an annotation on a Need or ADR resolves to
nothing. *Why: status is computed from contract↔code; binding the non-bindable
would fake a status that means nothing.*

## 3. WHY — the reason for the shape

**3.1 — An ADR records one decision; an invariant cites it via `decided_by`.**
The reasoning lives on the contract, never in the invariant's `text`. An
invariant says *what* is guaranteed; the ADR says *why this shape and not
another*. *Why: justification in the promise text drowns the single observable
property the promise exists to state.*

**3.2 — ADRs outlive any plan or implementation.**
A rewrite that preserves the invariant preserves its ADR; a rewrite that changes
the invariant must cite a new ADR or supersede the old one. *Why: the why-layer
is the only one that keeps institutional memory honest across rewrites.*

## 4. Why the separation is load-bearing, not decorative

**4.1 — Each change touches its own layer.**
Audience changes edit WHAT; implementation changes edit HOW (and the spec only
if a contract genuinely shifts); reasoning changes edit WHY. *Why: mixing them
costs the graph its value —*
- HOW into WHAT → contracts brittle to refactors;
- WHY into WHAT → invariants unreadable under justification;
- WHAT out into code (binding functions that realize no contract) → the status
  stops being honest.

**4.2 — The separation is what lets the tool stay small.**
CUE checks structure; the compiler checks only what CUE cannot (statuses,
cycles, blocked propagation, coverage); annotations are a lexical scan, not a
parser. *Why: no layer reinvents what the layer beside it already does.*

## 5. The intent unit is a Need

**5.1 — Intent is a `Need` (with a `Domain`).**
A Need is an objective statement of what a stakeholder requires, independent of
delivery cadence — a consumer, a description, named atoms (FR/NFR). Domain is
the audience above it. *Why: the term comes from Requirements Engineering
(ISO/IEC 29148, IREB) and the same Domain DDD codifies — a long-lived, often
non-human intent, not a sprint-sized increment. The alternatives considered
(notably UserStory + Product) and the full rationale live in
[ADR-10](spec.d/governance/adr.cue), where a why-decision belongs (rule 3.1).*

**5.2 — Coverage is computed, not declared.**
A Need's status is `covered / partial / uncovered`, derived from the contracts
that satisfy its atoms — never set by hand. *Why: a Need is covered by contracts
that prove out, not "delivered" by a team that says so.*

## 6. The agent is a first-class reader

**6.1 — Every verb is machine-addressable.**
Each takes `--json` with a stable shape; errors are JSON too, carrying the same
`fix` field humans see in `try:`. *Why: the agent reasons over the same graph
the reviewer reads — there is no agent view hiding behind a human one.*

**6.2 — Identity is always `module:slug`.**
The form you copy from `get`, paste into `describe`, reference in annotations.
*Why: one identity across read, write and bind removes a class of mismatch
errors.*

## 7. What Specue measures — and the axis it does not

**7.1 — Specue is a hard-systems tool: it computes code ⟷ contract.**
`proven` means the code does what the contract says — the whole of what the
verdict asserts. *Why: a single computable verdict is the thing Specue is
precise about, and it pays to say exactly how far that precision reaches.*

**7.2 — It does not measure contract ⟷ agreement.**
Whether the contract is the *right* one — whether the people it serves agree —
is not computed. `proven` means "built exactly," never "built the right thing."
*Why: that argument happens at a human table; Specue runs after the table,
holding the agreed contract steady so code cannot drift from it.*

**7.3 — Specue keeps the residue of a negotiation, not its log.**
A boundary, a decision, an intent land as nodes; the deliberation that produced
them does not. *Why: a graph makes the invisible visible and queryable, but a
tool that tries to host the argument becomes a worse whiteboard.*

**7.4 — The layers are WHAT/HOW/WHY — not WHEN, not WHERE.**
*When* a thing gets done (deadlines, sprints, ordering) and *where* it runs
(region, node, cluster, deployment) are not layers of the model. They split, and
neither half asks for a new layer:
- their *observable* grain is **already WHAT** — "a token is valid for one
  hour" (a temporal invariant), "a committed write survives the loss of any one
  node" (a placement-resilience invariant), the Port/Container topology. These
  are contracts, expressible today.
- their *process* grain belongs to other tools — scheduling ("when we will ship
  it") to a planner (Jira/Notion); physical placement ("where it actually runs")
  to infrastructure (Terraform/k8s/a runbook); and the *how* of either to code,
  through annotations.

*Why: WHAT/HOW/WHY is the complete, minimal cut of one question — what is
promised, what realizes it, why it is shaped so. WHEN and WHERE are not a
different question about the promise; they are the same promise from another
angle (→ WHAT) or not about the promise at all (→ a planner, an infra tool). A
contract may carry a `reference` outward to a ticket or a deployment, but it
does not absorb their model — the same discipline that chose Need over UserStory
(rule 5.1).*

## 8. The boundary is a judgment — and named as one

**8.1 — A `context` is one observer's answer to where the system ends.**
Which modules are the system, which are the environment — a judgment, not a fact
about the world. *Why: the same thing sits inside or outside the boundary
depending on who is looking. To Specue, git is part of the environment — a
`Port` it reaches across; the system ends before git. But git is itself a
system, and from its standpoint Specue is just one more consumer in its
audience. Neither view is wrong; the boundary is drawn by the observer, so there
is no single objective one to pretend about.*

**8.2 — *Unknown from here* is a legitimate state, not an error.**
A contract that depends on a module outside your context is an honest gap.
*Why: not knowing of a neighbour is normal; reporting it as failure would punish
honesty.*

**8.3 — Where the hard verdict stops, say so in the discipline's own words.**
Stated plainly so no reader is misled:
- **One verdict assumes one boundary.** Today's model carries a single `Domain`
  and renders cross-module claims from the asserting side; with several teams
  and worldviews, "the verdict" is *a verdict from one boundary* — true, but
  partial. Making boundaries visible, comparable and owned is the direction, not
  a present property.
- **A `satisfies` edge is a claim, not a handshake.** Owned by the module that
  declares it; the target need not yet agree. The assent/dissent of the target
  is not yet data.
- **Federation is contract-agreed, not built.** `attest-bindings` (reading a
  spec without the code) is designed, not shipped; until then the honest mode is
  one holder, one landscape.

*Why: `federation`, `context`, `Plan` are real structure for many boundaries —
not soft-systems theatre — and the tool must promise only what it proves.*
