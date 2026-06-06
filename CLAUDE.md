# Specue repo guide

This is the **Specue v2 tool** — a CLI that derives a spec graph from CUE
modules plus the code that realizes them. It is **described in its own graph**
(self-spec): the spec lives here under `spec.d/code/`, `spec.d/service/`,
`spec.d/domain/` and `spec.d/governance/` — the drop-in layout introduced by
ADR-11. The repo root no longer holds a `spec.mod.cue` or `cue.mod/`; the code
module sits at `spec.d/code/` with `code_root: "../.."` so the Go scan reaches
`internal/`, `cmd/`, etc. The graph the tool is supposed to produce is the
graph it produces *of itself*.

If you are not yet familiar with Specue as a model (Contract, Need, Domain,
Port, Container, ADR, satisfies, decided_by, the binding lifecycle), read the
shared Specue skill first — it is the model layer. (It is normally
registered as a Claude skill named `specue`; the local skills assume you
have read it.)

The skills in `.claude/skills/` here are the **repo-specific layer** — what is
different about this repo on top of the model. Reach for them by task:

- `specue-authoring/` — writing nodes in CUE (Contract/Need/Domain/ADR/Port).
- `specue-binding/` — `//specue:req:` / `//specue:test:` /
  `//specue:produces:` etc. in the tool's Go source.
- `specue-navigation/` — `get` / `describe` / `query` / `bindings`, reading
  any graph including this one.
- `specue-planning/` — `plan register` / `use` / `diff` / `accept` for
  speculative changes.
- `specue-contributing/` — changing the Go code of the tool itself
  (package layout, conventions, build, test, debug).

## Build and run

The repo lives next to a `go.work` further up that does not include it. Always
build with `GOWORK=off` from the repo root:

```
GOWORK=off go build -o ./specue ./cmd/specue/
GOWORK=off go test ./...
```

(The examples below call it `./specue` — substitute whatever you built or
the binary on your PATH.) Most verbs need `SPECUE_GIT` pointing at git, so
the git-native invariant (MANIFESTO P20) is honoured:

```
SPECUE_GIT=$(which git) ./specue validate
SPECUE_GIT=$(which git) ./specue bindings
SPECUE_GIT=$(which git) ./specue query "SELECT id, status FROM nodes WHERE type='Contract'"
SPECUE_GIT=$(which git) ./specue describe specue.io/service@v0:validate-graph
```

## Environment switches

- `SPECUE_GIT=<path>` — the git binary to call; required by most verbs.
- `SPECUE_NO_AUTOWARM=1` — skip the implicit `warm` on validate/context-use.
  Use it in CI and hermetic tests so the suite does not touch the user's cue
  cache. The CLI test suite sets it in `TestMain`.
- `SPECUE_WARM_DEBUG=1` — surface a warm error to stderr instead of
  swallowing it. Useful when the editor's cue lsp stops resolving and you
  suspect the warm path; see the contributing skill for the read-only extract
  / stale download story.

## Read the graph the tool produces

The self-spec context is registered as `specue-self`. Switch into it
and the read verbs resolve against the four modules of this repo:

```
SPECUE_GIT=$(which git) ./specue context use specue-self
SPECUE_GIT=$(which git) ./specue query "SELECT status, count(*) FROM nodes WHERE type='Contract' GROUP BY status"
SPECUE_GIT=$(which git) ./specue bindings | grep -v '^proven'   # what is left to bind/prove
SPECUE_GIT=$(which git) ./specue describe specue.io/service@v0:build-graph
```

This is the fastest way to understand the tool: read its own contracts.

## Repo layout (short)

- `cmd/specue/` — the CLI entry point.
- `cli/` — verbs (`validate`, `get`, `describe`, `query`, `bindings`, `diff`,
  `plan`, `context`, `init`, `registry warm`). Each verb has a `run<Name>`
  function annotated with `//specue:req:<slug>`.
- `engine/`, `compiler/`, `specload/`, `modules/`, `codescan/`, `source/`,
  `query/`, `diff/`, `plan/`, `context/`, `warm/` — the layers underneath.
  The contributing skill maps each one to its Contract.
- `spec.d/` — the self-spec (`code/`, `service/`, `domain/`, `governance/`).
- `spec.d/governance/adr.cue` — the ADRs that justify the tool's shape.

## Do not commit without asking

Edits to this repo wait for explicit confirmation before being committed. Show
the diff first; never use `git add -A` blindly — the user often keeps untracked
dogfood artefacts you should not stage.
