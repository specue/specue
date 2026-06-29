# specue

**A specification you navigate like code — but where the only unit is a decision.**

Decisions, linked into a graph, written in free form.

> ⚠️ Experimental and conceptual: a research prototype, redesigned several times,
> distilled down rather than finished. More on that at the end.

---

## Where it started

specue came out of fatigue with AI engineering — staring at piles of code I didn't
understand, repeating "rewrite this, it's not how we do things." If the agent is
going to write the code anyway, the leverage isn't in the code. It's in stating
*what the system must do* and proving it with tests. You can still read the code,
but honestly that's going to hurt.

That flips the question onto the specification. If the spec is what we actually
control, it has to be something you can **navigate** — the way you navigate a
codebase. So borrow what already works in code: decomposition, layering, things
standing on other things — but write the contents in free form, as decisions.

## What a decision is

One uniform record — the problem you're answering, the answer, and what that answer
stands on. No needs, user stories, functional/non-functional layers. Just this:

- **`problem`** — a "How …?" question. This question *is* the decision's identity.
- **body** — the answer, in prose (`README.md`).
- **`contract`** — the decision's public surface and what it stands on (below).
- **`context`** — open, for extra dimensions attached through profiles.

A decision relies on other decisions, so a specification becomes a graph — and
because every node has the same shape, it reads like code written in free language.
On disk: one folder per decision (`spec.cue` + `README.md`), grouped into CUE
modules.

### The contract

A decision's `contract` is a closed struct (`close({...})`) that does two things at
once:

- **public fields are promises** — what other decisions may rely on;
- **hidden `_`-fields are supports** — references into another decision's contract,
  e.g. `_storage: orders.contract.storage`. This is how one decision stands on
  another; the dependency is the reference itself, not a separate link.

Closing the contract is what makes the graph hold: rename or remove a relied-on
field and the dependent breaks loudly (`undefined field`) under `cue vet` — even
when the support is captured in a hidden field. Supports stay private, so nobody
can rely on what a decision merely consumes. (`close()` only closes the top level,
so nested public structs that are relied on are closed too.)

## The ADR instinct

The closest relative is the **Architecture Decision Record**. The way I read ADRs:
code is the source of truth, but not everything important is visible in it — so you
write down the decisions that don't follow directly from the code, the parts that
matter most. specue takes that instinct and makes the decision the *only* record,
with the links between decisions first-class, so the whole thing is a navigable
graph rather than a flat log. It was inspired in large part by the ADR
[ad-guidance-tool](https://github.com/adr/ad-guidance-tool).

## Kept simple, kept loose

Two things mattered while building it: stay simple, and don't bind you to a method.

Simple, because integrity lives in the schema, not the tool — broken links, cycles,
rules like "a link must not point at a withdrawn decision" are all checked by
`cue vet`, so specue reimplements none of it. A decision's identity is just its
package path: no catalog, no ids by hand.

Loose, because you bring your own structure — your own link kinds, your own context
fields through profiles, plain Markdown for the bodies. The model gets out of the
way.

The real stress test was authoring **specue in its own form** without losing track
of what's going on. So far it holds: the `specue/` directory is a decision graph
about specue, authored with specue's own rules — and reading it is the best way to
learn the model.

## Try it

The schema lives in a CUE registry. For local development specue runs its own OCI
registry and publishes the schema + profiles to it (`mise run registry`, then
`cue mod publish`); your module then depends on them normally, no `replace`. You
need Go and the CUE CLI (v0.17+).

Clone and build the tool:

```sh
git clone https://github.com/specue/specue
cd specue && go build -o specue .          # the `specue` binary
```

In **your own project**, start a decision module and declare the schema dependency:

```sh
mkdir spec && cd spec
# your decision module (a CUE module)
cue mod init example.com/myproject@v0

# cue.mod/module.cue — declare the schema dependency:
#
#   deps: {
#		"specue.io/schema@v0": {
#			v: "v0.0.8"
#		}
#	}
#
# resolved from the registry in $CUE_REGISTRY.
```

`specue add <path>` is the one command so far. It finds the surrounding module,
verifies it's a decision module, makes sure the path is free, and writes a
`spec.cue` + `README.md` skeleton — that's authoring a decision:

```sh
/path/to/specue/specue add domain/my-first-decision        
```

Open the generated files and fill them in. A decision's **structure** lives in
`spec.cue`:

```cue
// domain/persist-orders/spec.cue
package persistorders

import s "specue.io/schema@v0:schema"

decision: s.#Decision & {
	problem: "How to persist placed orders"

	// public promise others may rely on
	contract: close({
		storage: "append-only-log"
	})
}
```

Its **body** — the actual answer — lives next to it in `README.md`:

```markdown
<!-- domain/persist-orders/README.md -->
<dec-body>

Orders are stored in an append-only log: every state change is a new event,
nothing is updated in place. This keeps a full audit trail and lets us rebuild
state by replaying events.

</dec-body>
```

A second decision can **rely on** the first — a hidden support in its contract that
references the first's public field. That reference is what turns decisions into a
graph:

```cue
// usecase/refund-order/spec.cue
package refundorder

import (
	s  "specue.io/schema@v0:schema"
	po "example.com/myproject/domain/persist-orders@v0:persistorders"
)

decision: s.#Decision & {
	problem: "How to refund an order"
	contract: close({
		// refunding builds on how orders are stored
		_storage: po.decision.contract.storage
		method:   "reverse the original event"
	})
}
```

The support points at another decision by importing its package and referencing a
public field of its contract. Run `cue vet ./...` to check the whole graph — it
catches broken supports, cycles, and schema violations.

### Bring your own fields

`context` is open, so you attach whatever dimensions you need through a **profile** —
a small CUE definition that mixes extra fields into a decision. Define one once:

```cue
// profiles/maturity.cue
package profiles

#WithMaturity: {
	maturity: "experimental" | "stable" | "deprecated"

	// stay open so profiles compose
	...
}
```

Then apply it to a decision's `context` — and stack several at once:

```cue
decision: s.#Decision & {
	problem: "How to refund an order"
	context: p.#WithMaturity & {maturity: "experimental"}
	contract: close({_storage: po.decision.contract.storage})
}
```

A profile is applied to `context`, not to the decision itself: mixing it into the
whole decision would reopen its closed contract. specue ships ready-made profiles
you reuse the same way — e.g. `stakeholders` (`#WithDrivers`: who wants the decision
and why) and `dimensions`. The core `#Decision` never changes; profiles only add to
`context`.

## Reading specue's own decisions

Start with **`specue/contract/decision-authoring/`** — the public contract for how
a decision is written: the schema (`decision-schema`, `link-schema`,
`status-schema`), profiles, and the extensible context. From there,
**`specue/cli/commands/add/`** (`command` + `output`) shows the command, its
arguments and its output contract. The top-level **`specue/binary`** is the
tool itself (the software: its commands and version). Everything under
**`specue/internal/`** is the tool's own machinery — domain notions, usecases,
tech choices, layout — kept internal because nothing outside relies on it. Each decision is a folder
(`spec.cue` + `README.md`); `mise run vet` validates the whole graph.

The tests follow the same principle: `conformance/` holds black-box scenarios
derived from the decisions, *before and independent of* the implementation. Any
`specue` binary is driven through the full command path and checked against the
output contract (`mise run conformance`); each scenario is tagged with the
decisions it covers.

> The decision bodies are currently in Russian (the author's working language).
> The model and tooling are language-agnostic.

## What works, and what doesn't

Honest trade-offs, since this is a prototype.

**What the approach buys you:**

- **Navigate decisions like code** — in your normal code editor, with the same
  habits: jump to definition, find references, read a folder tree.
- **A mental map like a codebase** — good layout plus decomposition of decisions
  builds the same spatial intuition you get from well-structured code. The same
  problems (where does this belong? how big is one unit?) — familiar to developers.
- **Strong typing via CUE** — fields, enums, and constraints are checked, not just
  conventions.
- **Modularity** — decisions live in modules; other projects can depend on them
  and rely on their decisions.
- **CUE's expressiveness** — the whole tool is built on it: reuse values, schemas,
  constraints across the graph.
- **Markdown's expressiveness** — bodies stay free-form prose.
- **A simple default model** — nothing extra, so far.
- **Tests derive from the decisions** — conformance scenarios are written from the
  spec, before and independent of the code, so the spec itself yields the
  acceptance criteria (see `conformance/`).
- **The spec runs as a prompt** — complete and unambiguous enough that an agent can
  implement from it without round-trips (that's how `specue add` was built).
- **The model held together** — anecdotally (no proof, just the author's
  experience): building it out surfaced no internal contradictions, unlike earlier
  models that ran into too much duplication, too little expressiveness, or too much
  description overhead.

**Where it falls short:**

- **Partial compilation of meaning** — contracts close part of this gap: what a
  decision stands on *is* enforced now, since a support references a public field of
  a closed contract, so renaming or dropping it breaks `cue vet`. But this only
  guards the supports you actually wrote. The *completeness* of a contract is still
  on you: whether every real dependency is captured, whether a support is honest
  (it truly uses that field) or just there for layering, and whether promises and
  supports are well expressed at all. CUE checks that a written reference resolves —
  not that the contract says everything it should, or says it well.
- **"How …?" doesn't always fit** — some decisions don't phrase cleanly as a
  question; usually it's a matter of framing, but not always.
- **Discipline matters a lot** — good decomposition pays off later but, as with
  code, you have to think about it up front. (For some that's neither pro nor con,
  just the deal.)
- **A small barrier for non-programmers** — though an AI can help author decisions.
- **Wording is on you** — there's no settled style guide yet; an AI tends to
  produce overwrought, padded phrasings, so you shape them yourself.
- **Identity is the package path** — moving or renaming a decision rewrites every
  link that points at it. There's no "rename symbol"; restructuring is a
  find-and-replace across the graph.
- **CUE is young** (a mild one) — the tooling is pre-1.0. Using a schema across
  modules wants a registry + publishing; the LSP is static-only, so it doesn't
  evaluate the graph (go-to-definition resolves the published schema, not local
  edits, and contract violations show up in `cue vet`, not inline).

**Not yet built (limitations, not flaws):**

- No semantic search — to avoid duplicate decisions, find things fast, or discover
  what to rely on.
- No MCP server for an AI to read the whole graph — while the graph is small,
  dropping it into a single Markdown file has been enough.

**A data point.** Authoring the core *decisioning* model took ~3 days — mostly
building definitions, choosing the data model, and pinning down needs. Designing
and writing the CLI specification took ~2 more. Then the implementation took under
a minute once the whole thing was handed over "as a prompt." Is it worth it? Open
question.

**Another data point.** The weight is in the spec, not the code. Line counts for
the `specue/` graph (prose = non-blank lines inside `<dec-body>`; CUE =
`spec.cue`):

| Scope                       | Decisions | CUE (`spec.cue`) | Prose (`README.md`) |   Go |
| --------------------------- | --------: | ---------------: | ------------------: | ---: |
| Whole `specue/` graph       |        51 |             2187 |                 533 |    — |
| Of which: backed by code    |        17 |              827 |                 145 |    — |
| `specue add` implementation |         — |                — |                   — |  588 |

Of the 51 decisions, 17 map to code; the other 34 are the method itself (what a
decision is, how one relies on another, file layout, module versioning), written
once and imported thereafter. The 17 code-backed decisions against the 588 lines
of Go, by how much is counted:

| Spec vs code, on the 17 code-backed decisions                 | Ratio |
| ------------------------------------------------------------- | ----: |
| Whole graph / Go                                              |  4.6× |
| Raw: (CUE + prose) / Go                                       | 1.65× |
| Go comments counted, both sides full                          | 1.48× |
| Same, `// decision:` tags excluded                            | 1.71× |
| Spec logic + prose / Go logic *(imports & comments stripped)* | 1.82× |
| Logic only: CUE fields & supports / Go statements             | 1.35× |
| Prose only: decision bodies + CUE comments / Go comments      | 2.75× |

The marginal cost of a feature lands at ~1.3–1.8× lines; the 4.6× is the
un-amortized foundation. Most of the spec is structure — CUE fields and supports,
machine-checked (1.35× on their own). The prose figure is the highest (2.75×):
the "why" lives in the decision bodies rather than in code comments. The
`models/decisioning` foundation adds another 31 decisions / ~950 lines of CUE,
reusable across projects.

Whether ~1.5× pays off is the open question above: it does when the code is
generated (so the spec replaces nothing you'd skip) and the cost is consistency
at scale rather than volume — which a 17-decision tool does not yet exhibit.
Logic = non-blank, non-comment, non-import lines; ratios measured from this repo.
The spec is what you write and own; the code is generated to satisfy it.

**On the line count itself.** Across all 51 `spec.cue` files, only ~21% of lines
carry content (fields, supports, values); ~45% is CUE ceremony — imports (17%),
the fixed `decision: #Decision & {` / `contract: close({` wrappers, and closing
braces (28% together) — with the rest comments and blanks. A thin DSL could erase
most of the ceremony: imports are redundant with the supports that reference them
(`_resolve: rm.decision.contract.resolving.module` already names the target), and
the wrappers are identical in every file. It would roughly halve `spec.cue`. It
is left as plain CUE on purpose: a support is a real CUE import, so go-to
definition, find-references, and `cue vet` work in a normal editor with no
specue-specific tooling — integrity lives in the schema, not in a compiler we'd
have to write. The verbosity is the price of that, not an oversight.

## How it got here, and where it's going

specue is on its *n*-th redesign, and the path matters more than the destination.
It began with use cases and user stories, moved to functional and non-functional
requirements, then to a single "invariant" element. Each stage made the same
discovery: another layer restating the same thought at a different altitude. The
current model is what's left after collapsing all of it into one record and handing
validation to CUE — deliberately minimal, a concept worked out in the open rather
than a finished tool to adopt as-is.

What's next, roughly: the conformance test scenarios; a server so an agent can
explore a specification; revision drift between decisions and code; and richer
bindings from code back to decisions (there's already a primitive form — the
`// decision: <id>` tags in this repo's implementation).

## Layout

```
specue/        the decision graph about specue:
                 binary/        the tool itself (software: commands + version)
                 cli/commands/  CLI commands (each: command + output)
                 contract/      public authoring contract (how a decision is written)
                 internal/      the tool's own machinery (kept internal)
cue/schema     the #Decision / #Link / #DecisionStatus schema (a CUE module)
cue/profiles   reusable context profiles (stakeholders, dimensions)
models/        reusable decision models (decisioning, cli-guideline, cli-contracts)
cmd/, internal/, main.go   the `specue add` implementation (Go)
conformance/   tests derived from the decisions
```
