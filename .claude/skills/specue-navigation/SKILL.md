---
name: specue-navigation
description: Read a spec graph — list, describe, query, inspect bindings
---

# Navigating a spec graph

Use this skill when you need to understand a graph you did not author: find
what exists, read one node in full, ask how nodes relate, or check where
the code stands against the contracts.

The shared Specue skill (`specue` → `reverse.md`) covers reading a
landscape end-to-end (which modules, which contracts, what is delivered).
This skill is the **operator handbook**: the read verbs the tool gives
you (`get`, `describe`, `query`, `bindings`, `render`), when to use
which, and how to phrase a question as a SQL query.

## The three layers, on the reader side

When you read a graph you are walking the **WHAT** layer:

- WHAT (UseCase, Need, ADR, Port, Container, Plan) is what `get`,
  `describe` and `query` return.
- HOW (the code that realizes WHAT) appears as `bindings` rows and as
  status (`proven` / `implemented` / `asserted` / `broken`).
- WHY (the rationale) appears as `decided_by` edges on invariants pointing
  at ADRs; follow them when an invariant looks arbitrary.

Picking the right verb saves output. `get` for "what kinds of thing
exist"; `describe` for "show me this one"; `query` for "all the things
matching some criterion or the relations between them"; `bindings` for
"what does this code module still owe"; `render` for "hand the whole
graph to a reader who will not run the tool".

## get — discover the surface

`get` with no argument lists the selectable resource types. `get <resource>`
lists nodes of that type. `get <resource> <module:slug>` narrows to one.

```
get                                    # what kinds of node exist
get usecase                            # every UseCase in the active landscape
get usecase example:validate-graph     # one UseCase (same row, narrowed)
get all                                # every node, every type (rarely useful)
```

Always pair with `--json` for machine reading: the column shape is stable
and small, and downstream tooling parses it cleanly.

## describe — read one node in full

`describe <module:slug>` is the single-node read. It prints the contract:
trigger, preconditions, every invariant (with its `satisfies` and
`decided_by`), every variation, postconditions, the derived edges, and
the current status.

Use it after editing a node to confirm the tool resolved it the way you
intended (qualified imports, satisfies wiring, decided_by). It is also the
fastest way to understand an unfamiliar node — far faster than reading the
CUE file, because it shows the *resolved* form, with references already
followed.

## query — SQL over the projection

The graph is projected into an in-memory read-only SQLite database;
`query <sql>` runs against it. `query tables` prints the schema, examples
and recipes — read it first, every time you write a query you have not
run before, so the column shapes are current.

Typical questions and their SQL:

```sql
-- What is the spec's coverage by status?
SELECT type, status, count(*) FROM nodes GROUP BY type, status ORDER BY type, status;

-- Which UseCases are still asserted (declared but not built)?
SELECT id, title FROM nodes WHERE type = 'UseCase' AND status = 'asserted';

-- Which UseCases satisfy a given Need FR?
SELECT uc_id, atom FROM satisfies WHERE need_id = '<module:slug>' AND atom = 'fr-NN';

-- Which Needs are still uncovered and why?
SELECT n.id, s.atom FROM nodes n
JOIN atoms s ON s.node_id = n.id
LEFT JOIN satisfies sat ON sat.need_id = n.id AND sat.atom = s.atom
WHERE n.type = 'Need' AND n.status = 'uncovered' AND sat.need_id IS NULL;

-- Full-text search on every node body and title.
SELECT id FROM nodes_fts WHERE nodes_fts MATCH 'idempotent';
```

The projection is read-only — a write fails fast with an actionable hint.
Recursive CTEs are how you walk graph relations; `nodes_fts` is how you
search by phrase. Combine the two when you want "every UseCase mentioning
X and the stories they satisfy".

## bindings — what the code owes

`bindings [<code-module>]` is the code-module view: every UseCase the
module may realize, and for each one a row per binding kind (`req` plus
any infra roles) with a state — `unbound` / `bound` / `proven` /
`duplicate` / `orphan`.

It is the working TODO of a code module. Filter it to see what is left or
what is broken:

```
bindings --state unbound        # what still needs to be written
bindings --state proven         # what is done end-to-end (req + test)
bindings --state orphan         # an annotation that resolves to nothing — fix it
bindings --state duplicate      # the same contract bound twice — pick one
bindings --kind req             # only the implementation rows, no infra
```

Pair with `--json` to drive an automated authoring loop: read `unbound`
rows, decide which to tackle, write code, re-run.

## render — publish the graph

`render <dir>` is the bulk-read verb: one file per node plus an index,
written into an empty directory. Use it to hand a graph to a reader who
will not run the tool, or to feed a downstream pipeline that wants the
graph as data rather than as SQL.

Two output shapes:

- **markdown** (default) — one `.md` per node, frontmatter + describe-style
  body, cross-links by relative path. The defaults (`--layout flat`,
  `--frontmatter full`) give a single directory of files good for `grep`.
- **json** — `index.json` (modules + flat node list) plus
  `<module>/<slug>.json` (authored fields, derived edges, bindings with
  `file:line`, status). The structured graph for custom pipelines, no
  markdown parsing required.

Markdown has knobs for the three publishing targets worth naming:

```
render ./docs                                                   # default: flat, full
render --layout tree --strip-prefix gitlab.example.com/org/ ./docs   # GitHub
render --frontmatter mark   --space ENG ./out                        # kovetskiy/mark → Confluence
render --frontmatter mkdocs --nav-snippet nav.yml ./docs             # MkDocs Material
render --frontmatter minimal ./docs                                  # title/type/status only
render --frontmatter none    ./docs                                  # body only
```

`--layout tree` mirrors module paths as nested directories (the leaf drops
`@vN`); `--strip-prefix` shortens visible identifiers in directory names
and link text. The full enum list lives in `render --help`.

## Identity is module-qualified

Every node lives in a module and has a slug; its **identity is
`module:slug`**. Use that form everywhere — in `describe`, in SQL `id`
columns, in `bindings` rows. Bare slugs are accepted only inside a single
module's authoring context (annotations); the moment you cross modules
the prefix is required.

## Workflow

1. Enter the right landscape. `context use <name>` if working against a
   named context; otherwise `-C <dir>` for single-module reads.
2. `get` to discover; `describe` to read; `query tables` then `query` for
   anything richer than a listing; `bindings` for code-module questions.
3. Add `--json` whenever something downstream consumes the output.
4. If a verb feels missing, think SQL first — most navigation questions
   are one query. Only reach for a new verb when the same query repeats.
