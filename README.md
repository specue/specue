# Specue

[![status: pre-release](https://img.shields.io/badge/status-pre--release-orange)](#specue)
[![model: speculative](https://img.shields.io/badge/model-speculative%20%2F%20breaking-red)](#what-specue-is-not)
[![Go 1.26+](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD--3--Clause-blue)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/specue/specue.svg)](https://pkg.go.dev/github.com/specue/specue)

**Observable promises, tied to the code that keeps them, with coverage you can
compute instead of assert.** Specue models a system as a graph of contracts —
what each part promises, who it promises to, where its boundary lies — and binds
every promise to the source line that realizes it. Then it tells you, by
derivation rather than by hand, what is kept and what is still owed.

It sits closer to C4 and architecture-as-code than to spec-driven workflow
tools: the unit is a boundary and a promise, not a feature ticket. The spec
lives in CUE files in your repo. The tool prints one node in full (`describe`),
runs SQL over a projection of the whole graph (`query`), tells each code module
what its contracts still owe (`bindings`), publishes the spec as a markdown tree
or a structured JSON graph (`render`), lets you propose changes as branches you
can diff and accept (`plan`), and gates the whole thing on every commit
(`validate`).

Specue answers one question precisely — *does the code do what the contract
says?* — and is honest about the question it does not answer — *is this the
right contract?* See [What Specue is not](#what-specue-is-not) before you reach
for it.

> ⚠️ **Pre-release, and the model is still moving.** Specue is not at a stable
> release. The model itself — node types, the schema, the vocabulary, the
> manifesto — is **speculative and actively changing**. There have already been
> many backwards-incompatible changes, and there will be more: node types get
> renamed (the contract node was just renamed from `UseCase` to `Contract`),
> fields move or are removed, and the embedded schema may break pins between
> versions. Use it to explore the
> idea, dogfood it, and shape it — but do not yet build anything you are not
> prepared to migrate by hand. Nothing here is a compatibility promise until a
> tagged release says otherwise.

This README walks from the smallest possible spec to the full picture. Each
step adds one feature — read until the example you need.

- **Install** — one `go install`.
- **Step 1: One contract** — a service module with a single Contract.
- **Step 2: Bind it to code** — a code module beside it; the contract moves
  `asserted` → `implemented` → `proven`.
- **Step 3: Add an intent (Need)** — who the contract is for, what FRs it
  discharges.
- **Step 4: Record a decision (ADR)** — the why-layer.
- **Step 5: Propose a change (Plan)** — branches across every affected module.
- **Step 6: Query the graph** — one SQL replaces ten `describe` calls.
- **Step 7: Publish the docs** — markdown for GitHub/Confluence/MkDocs, JSON
  for custom pipelines.
- **Concepts**, **Status**, **Layout**, **Working with it as an agent**,
  **What Specue is not**, **License**.

The philosophy — why the spec is split into WHAT, HOW and WHY layers — lives
in [MANIFESTO.md](MANIFESTO.md). Read it once you've tried the examples and
want the model.

## Install

**Prerequisites**: Go 1.26+ ([install](https://go.dev/doc/install)) and git on
`PATH`. That is it — the CUE schema is embedded in the binary, so you do not
install CUE separately. (A `cue` CLI is useful if you want LSP completion in
your editor, but not required to run the tool — see
[cuelang.org](https://cuelang.org/docs/install/).)

**Install the tool**:

```sh
go install github.com/specue/specue/cmd/specue@latest
```

This drops a `specue` binary in `$(go env GOBIN)` (defaults to
`$HOME/go/bin`). Make sure that directory is on your `PATH`:

```sh
export PATH="$(go env GOBIN):$PATH"   # or $HOME/go/bin if GOBIN is unset
specue --help                      # confirms it is installed
```

### Try it on this repo first

The tool is self-spec'd: its own contracts live in `spec.d/`. Clone and poke at a
real graph before writing your own:

```sh
git clone https://github.com/specue/specue && cd specue
specue validate                                   # 68 nodes, green
specue describe specue.io/service@v0:validate-graph
specue query "SELECT id, status FROM nodes WHERE status='asserted'"
specue bindings                                   # what each contract owes
```

## Step 1: One contract

The smallest useful spec is one **service module** with one Contract. No
intent layer yet, no code binding, no context.

```sh
mkdir todo && cd todo && git init -b main
specue init . acme.test/todo@v0 --kind service --layout spec.d
```

This writes `spec.d/service/spec.mod.cue` (module identity) and
`spec.d/service/cue.mod/` (CUE setup). `--layout spec.d` follows the Unix
drop-in convention (`/etc/cron.d/`) — every module of a kind lives under
`spec.d/<kind>/`. The older flat layout (`spec.mod.cue` directly in the dir)
still works; pass `--layout spec.d` to opt in.

Now add a contract:

```cue
// spec.d/service/todo.cue
package todo
import s "specue.io/schema@v0:spec"

api: s.#Container & {slug: "api", title: "Todo API", kind: "service"}

addTask: s.#Contract & {
    slug: "add-task"
    title: "Add a task to a user's list"
    service: api
    trigger: "the user submits a new task"
    postconditions: [{
        text: "The task is durable: a later read of the same list returns it."
    }]
}
```

Read the graph back:

```sh
specue -C ./spec.d/service validate                              # ✓ 2 node(s) valid
specue -C ./spec.d/service describe acme.test/todo@v0:add-task   # contract + status
```

`add-task` shows up `asserted` — the contract is agreed, no code realizes it
yet. That is the honest state. The next step changes that.

## Step 2: Bind it to code

Code lives in a separate module of `kind: "code"` that points at the service
module it realizes. The Go source carries `//specue:req:<slug>` annotations
that bind a function to a Contract, and `//specue:test:<slug>` that bind a
test.

Add the code module beside the service module — both under `spec.d/`:

```sh
specue init . acme.test/todo-code@v0 --kind code --layout spec.d
```

This writes `spec.d/code/spec.mod.cue` with `code_root: "../.."` — the scan
starts two levels up, at the repo root, so it sees `handler.go` and friends.
Without `code_root` (or with a code module placed at the repo root) the same
filesystem subtree would be visible to CUE through two paths — workspace and
nested — and CUE would refuse with `ambiguous import`. The subfolder breaks
the construct. See [ADR-11](spec.d/governance/adr.cue).

The repo now looks like:

```
todo/
├── handler.go                 # //specue:req:add-task
├── handler_test.go            # //specue:test:add-task
└── spec.d/
    ├── code/                  # code module — kind: "code"
    │   ├── spec.mod.cue       # code_root: "../.."
    │   └── cue.mod/module.cue
    └── service/               # service module — kind: "service"
        ├── spec.mod.cue
        ├── cue.mod/module.cue
        └── todo.cue           # (the file from step 1)
```

Point the code module at the service module by editing
`spec.d/code/spec.mod.cue`:

```cue
module:    "acme.test/todo-code@v0"
version:   "v0.1.0"
kind:      "code"
code_root: "../.."
require: [
    {module: "acme.test/todo@v0", version: "v0.1.0", replace: "../service"},
]
```

Bundle both modules into a **context** so the tool sees them as one landscape:

```sh
specue context create todo
specue context use    todo
specue context module add ./spec.d/code      # the code module
specue context module add ./spec.d/service   # the service module
```

Annotate the Go source:

```go
// handler.go
package todo

//specue:req:add-task
func AddTask(text string) error { /* ... */ return nil }
```

```go
// handler_test.go
package todo

import "testing"

//specue:test:add-task
func TestAddTask_Durable(t *testing.T) { /* ... */ }
```

Now:

```sh
specue validate                                   # green
specue describe acme.test/todo@v0:add-task        # status: proven
specue bindings                                   # what each contract owes / proves
```

`add-task` is now **proven**: a function implements it and a test covers it.
Without the test it would be `implemented`. Without either, `asserted` again.

## Step 3: Add an intent (Need)

A `Need` names who or what requires something, and lists the testable atoms
(`fr-NN` / `nfr-NN`) the contracts must satisfy. Needs live in a **domain
module** (`kind: "domain"`) — separate from the service so the audience can be
shared by several services.

```sh
specue init . acme.test/who@v0 --kind domain --layout spec.d
specue context module add ./spec.d/domain
```

```cue
// spec.d/domain/who.cue
package who
import s "specue.io/schema@v0:spec"

users: s.#Domain & {slug: "users", title: "Todo users"}

user: s.#Need & {
    slug: "as-user"
    title: "Capture and complete tasks"
    domain: users
    consumer: "a user with things to remember"
    description: "to write a task down and mark it done later, so that I trust the app to remember what I wrote"
    frs: {
        "fr-01": {id: "fr-01", text: "A task added today is visible later in the same list."}
    }
}
```

Wire the service to satisfy it. In `spec.d/service/spec.mod.cue`, require the
domain:

```cue
require: [
    {module: "acme.test/who@v0", version: "v0.1.0", replace: "../domain"},
]
```

And let the Contract claim the FR:

```cue
// spec.d/service/todo.cue
import (
    s "specue.io/schema@v0:spec"
    d "acme.test/who@v0:who"
)

addTask: s.#Contract & {
    // ...
    postconditions: [{
        text: "The task is durable: a later read of the same list returns it."
        satisfies: [d.user.frs."fr-01"]
    }]
}
```

```sh
specue validate
specue describe acme.test/who@v0:as-user          # status: covered
```

The Need is `covered` because a proven Contract satisfies its only FR. With no
satisfying contract it would be `uncovered`; with some but not all, `partial`.
Coverage is *computed*, not asserted.

## Step 4: Record a decision (ADR)

When a contract is shaped a particular way for a reason, the reason belongs in
an **ADR** — not in the invariant's text. ADRs live in a **governance
module** (`kind: "governance"`).

```sh
specue init . acme.test/gov@v0 --kind governance --layout spec.d
specue context module add ./spec.d/governance
```

```cue
// spec.d/governance/adr.cue
package gov
import s "specue.io/schema@v0:spec"

adr01TaskOrdering: s.#ADR & {
    slug:   "ADR-01"
    title:  "Tasks are returned in insertion order, not by priority"
    status: "accepted"
    body: """
        A user adds tasks in the order they think of them and expects them
        back the same way. Priority sorting was considered and rejected:
        every user invented a different priority scheme, so any default
        misled most users.
        """
}
```

Cite it from the Contract invariant that owes its shape to the decision:

```cue
// spec.d/service/todo.cue
import (
    // ...
    g "acme.test/gov@v0:gov"
)

addTask: s.#Contract & {
    // ...
    invariants: [{
        id: "insertion-order"
        text: "A later read returns tasks in the order they were added."
        decided_by: [g.adr01TaskOrdering]
    }]
}
```

`describe` now prints the ADR ref beside the invariant; `query` joins through
`decided_by` so "what decisions does this Contract rest on?" is one SQL away.

## Step 5: Propose a change (Plan)

A speculative change is a **Plan** — identically-named `plan/<id>` branches
across every module the change touches, plus a Plan record in the governance
module. You can diff, view, validate and merge the plan without ever leaving
base.

```sh
specue plan register add-priority         # creates branches everywhere
specue plan use      add-priority         # checks them out across modules
# ... edit the spec; commit on each affected branch ...
specue plan base                          # return all modules to base
specue diff plan add-priority             # typed delta against base
specue plan accept   add-priority         # validates the overlay, merges, tags
```

`accept` refuses unless the overlay (base + plan) validates. The merge commit
in every affected repo is tagged `plan/add-priority`, so a reader of git
history can enumerate landed Plans without parsing the commit graph.

You can be on any branch when you `accept` — the tool returns each repo to its
base before merging.

## Step 6: Query the graph

Once the graph grows past a handful of nodes, the read verbs (`describe`,
`bindings`) are one node at a time. The whole graph is also projected into an
in-memory SQLite database — one `query` answers what would otherwise be ten
round-trips. The projection is **read-only** and discoverable:

```sh
specue query tables                   # schema + worked examples
```

A few questions you reach for:

```sh
# Every contract with no code behind it yet — the honest TODO list.
specue query "SELECT id, type FROM nodes WHERE status='asserted'"

# Story atoms with no Contract satisfying them — coverage gaps, by Need.
specue query "SELECT need_id, atom FROM fr_coverage WHERE uc_id IS NULL"

# Full-text over titles + bodies (porter-stemmed: 'idempotent' finds 'idempotently').
specue query "SELECT id, title FROM nodes_fts WHERE nodes_fts MATCH 'overlay'"

# Blast radius — every contract that transitively depends on add-task.
specue query "
  WITH RECURSIVE up(id) AS (
    SELECT 'acme.test/todo@v0:add-task'
    UNION SELECT from_id FROM dep_edges JOIN up ON to_id = up.id
  )
  SELECT id FROM up WHERE id != 'acme.test/todo@v0:add-task'"
```

`query tables` prints the schema (`nodes`, `dep_edges`, `infra_edges`,
`satisfies`, `bindings`, …) plus pre-joined views (`node_describe`,
`fr_coverage`) and a recipe section. Add `--json` to consume the rows in a
script; the rows are machine-readable, the projection is read-only, and the
projection is rebuilt from the graph on every run — never a second source of
truth.

## Step 7: Publish the docs

`render <dir>` emits one file per node plus an index. The defaults give a
flat markdown tree with full frontmatter; flags retarget it at the three
publishing pipelines worth naming, or switch to the JSON IR for anything
custom.

```sh
# GitHub-friendly docs — nested dirs mirror module paths, prefix stripped:
specue render --layout tree --strip-prefix gitlab.example.com/myorg/ ./docs

# Publish to Confluence via kovetskiy/mark:
specue render --layout tree --strip-prefix gitlab.example.com/myorg/ \
  --frontmatter mark --space ENGDOCS ./out
mark -u user -p $TOKEN -b https://yourorg.atlassian.net ./out/**/*.md

# Serve with MkDocs Material — see examples/mkdocs/ for a ready template
# (theme, status-pill CSS, INHERIT-d nav.yml):
specue render docs --layout tree \
  --strip-prefix gitlab.example.com/myorg/ \
  --frontmatter mkdocs \
  --with-index-pages --with-tags-page --with-status-admonitions \
  --nav-snippet nav.yml
cp examples/mkdocs/mkdocs.yml ./mkdocs.yml
mkdir -p docs/assets && cp examples/mkdocs/assets/*.css docs/assets/
mkdocs serve   # http://127.0.0.1:8000
```

See [examples/mkdocs/README.md](examples/mkdocs/README.md) for the full setup
and which `--with-*` flag drives which Material feature.

The flags compose: `--format markdown|json`, `--layout flat|tree`,
`--strip-prefix <prefix>`, `--frontmatter full|minimal|mark|mkdocs|none`,
`--space <key>` (with `mark`), `--nav-snippet <file>`,
`--with-index-pages`, `--with-tags-page`, `--with-status-admonitions`. Run
`specue render --help` for the enum listing.

For a custom downstream pipeline, switch to the **structured graph**:

```sh
specue render --format json ./out
# ./out/index.json — modules + flat node list (id, type, status, title, path)
# ./out/<module>/<slug>.json — full per-node payload: authored fields,
#   derived edges (uses, satisfies, realizes, port topology), code bindings
#   with file:line, status.
```

The JSON IR is the same graph the markdown is rendered from; consume it
with `jq` or any script when neither the markdown shapes nor the SQL
projection fits what you need to build.

## Concepts in one paragraph each

**Contract** — a logical contract a service guarantees, with named invariants
the code binds to. The only node that carries code bindings.

**Need** — the intent unit. It names a consumer (who or what requires this)
and a description, and owns the testable atoms (`fr-NN` / `nfr-NN`) Contracts
discharge through `satisfies`. Coverage is *computed* from the contracts that
satisfy its atoms — not asserted by hand. (Need, not "UserStory", because the
unit is long-lived and not always human-consumed — see
[ADR-10](spec.d/governance/adr.cue).)

**Domain** — the audience the system serves. Needs belong to a domain.

**Port** and **Container** — the C4 L2 surface: typed transports (a channel,
an RPC service, a datastore) and boundary boxes (an external actor, a
third-party system). The topology — who produces/consumes/serves/calls a Port
— is derived from edges the Contracts declare, not authored by hand.

**ADR** — the why-layer. Each ADR records one architecture decision; a
Contract invariant cites it through `decided_by`, keeping the contract text
about *what* and the rationale where it belongs.

**Plan** — a speculative change. A Plan record lives in your governance
module and points at identically-named `plan/<id>` branches across every
module it touches.

## Status

Pre-release (see the [pre-release note](#specue) above): the verbs below work,
but the **model** they operate on is still speculative and breaking between
revisions. What is implemented and proven in the tool today:

| Area                                                                                    | State                            |
| --------------------------------------------------------------------------------------- | -------------------------------- |
| CUE-native module resolution + in-process registry (cue lsp completion works)           | ✅                                |
| `validate` / `get` / `describe` / `query` / `bindings`                                  | ✅                                |
| Code binding scanner (`//specue:req:`, `:test:`, infra verbs)                        | ✅                                |
| Plans on branches: `register`, `use`, `base`, `drop`, `diff plan`, `accept`, `conflict` | ✅                                |
| `render <dir>` — markdown tree (flat/tree, mark/mkdocs/…) or JSON IR                    | ✅                                |
| Contexts (named landscapes) + `context add module`                                      | ✅                                |
| Federation: `attest-bindings` for reader-without-code                                   | ◻ contract agreed, not yet built |
| Full LSP for bindings (slug autocomplete inside `//specue:req:`)                     | ◻                                |
| HTML browser (`specue server`)                                                       | ◻                                |

The tool is self-spec'd: Specue's own contracts live in `spec.d/`, and the
Go source binds them. `specue validate` runs against the tool's own spec
every CI build; new features arrive as a Plan and land when the new contracts
prove out.

The self-spec sits under `spec.d/code/`, `spec.d/service/`, `spec.d/domain/`
and `spec.d/governance/` — the recommended drop-in layout introduced by
[ADR-11](spec.d/governance/adr.cue). The older flat layout (`spec/` and
`gov/` directly under the repo root) still works through `code_root` alone;
new repos should prefer `spec.d/`.

## Layout

- `internal/model/` — node/edge types, no I/O
- `internal/source/` — CUE loader; resolves through cue's module system
- `internal/modules/` — module resolver + in-process OCI registry that warms the cue cache
- `internal/compiler/` — domain gates: role-gate, dangling, cycles, coverage; status derivation
- `internal/engine/` — caches the resolved graph by content key
- `internal/codescan/` — lexical scanner over `//specue:` annotations
- `internal/plan/` — git operations + Plan record + pending overlay
- `internal/query/` — SQLite projection of the graph (FTS5 for search)
- `internal/diff/` — typed delta between two snapshots
- `internal/render/` + `render/markdown/` — documentation tree renderer
- `internal/cli/` — every verb, JSON / human renderers
- `spec.d/` — the tool's own spec (`code/`, `service/`, `domain/`, `governance/`)

## Working with it as an agent

Every verb takes `--json` and returns a stable shape. Errors are JSON too,
with the same `fix` field humans see in `try:`. Identity is always
`module:slug`, the form you copy from `get`, paste into `describe`, and
reference in annotations. A run that needs a richer answer than the fixed
verbs offer is one `query` away: the projection schema is itself returned by
`query tables`, so the agent can discover the columns it needs.

## What Specue is not

Specue is a hard-systems tool with a narrow, honest job. Knowing what it
deliberately leaves out is the fastest way to know whether it fits.

- **Not a place to negotiate.** Specue records the *conclusions* of a decision
  (a contract, a boundary, an ADR), never the argument that produced it. The
  table where stakeholders disagree about what the system should do is a human
  process; Specue runs after it, holding the agreed contract steady. `proven`
  means "the code matches the contract," not "everyone agreed this was the right
  contract."

- **Not formal verification.** A bind is a lexical link from a promise to the
  line that keeps it, checked by a code scan — not a proof. Specue tells you a
  contract *has* realizing code and a test that exercises it; it does not prove
  the code is correct. It is lightweight by design: CUE checks structure, the
  compiler checks only what CUE cannot (statuses, cycles, coverage), and the
  scanner never parses. If you need Coq/TLA+-grade guarantees, Specue is the
  wrong layer.

- **Not a model of implementation.** A contract describes what is *observable
  from outside* — "a committed write survives the loss of any one node," not
  "16 shards, 3 replicas, key = tenant_id." The shard count, the replication
  factor, the readiness probe config are HOW: they live in code and attach to a
  promise through an annotation, never as nodes. Specue is intentionally narrow
  in *what it may name*: observable promises in, implementation detail out.

- **Not a language for everything.** That narrowness scales with *size* — one
  service or a thousand, on a fixed small vocabulary — but it bounds *kind*.
  Specue models services that promise something across network and storage
  boundaries, and the code that keeps those promises. A UI layout, an ML model's
  internals, an algorithm's complexity, a business approval flow — you can
  *mention* such a thing in an invariant's text, but Specue has no node to model
  its structure. The seven node types are deliberate: a language for one genus of
  system, not for every domain.

- **Not an objective map of "the system."** A `context` is one observer's
  boundary — which modules are the system and which are the environment. That is
  a judgment, not a fact. Two observers may draw two contexts and get two honest
  verdicts; the tool does not pretend there is a single true one. A dependency
  on something outside your context is reported as an honest gap — *unknown from
  here* — not an error.

- **Not yet federated.** Multiple teams, multiple worldviews, reading a spec
  without holding the code — the model is built toward this, and its Needs name
  it, but the mechanism (`attest-bindings`) is contract-agreed and not yet
  shipped. Today's honest mode is one holder, one landscape, one boundary. The
  [Status](#status) table marks the line between what is proven and what is
  planned; the [manifesto](MANIFESTO.md) states where the computable verdict
  stops and human agreement begins.

## License

[BSD 3-Clause](LICENSE). Use it, fork it, embed it, sell it — just keep the
copyright notice and do not use the author's name to endorse your fork
without asking.

All third-party dependencies are permissive (MIT, BSD-3-Clause, Apache-2.0);
the license set is internally consistent.
