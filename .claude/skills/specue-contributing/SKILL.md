---
name: specue-contributing
description: Change the Specue tool's Go code — package layout, conventions, build, test, debug
---

# Contributing to Specue

Use this skill when you are changing the Go source of the tool itself in
this repo — adding a verb, fixing a gate, reworking a layer. This is the
repo-specific operator handbook.

If you are authoring spec nodes, binding code to spec, navigating a
graph or planning a change, reach for the other skills first
(`specue-authoring`, `specue-binding`, `specue-navigation`,
`specue-planning`). This one is about working on the tool's own code.

## The three layers, on the contributor side

You are touching the **HOW** layer. The tool's **WHAT** lives in
`spec.d/service/` — every verb, every gate, every behaviour is a Contract
with named invariants. Read the relevant Contract before you write the code:
`describe specue.io/service@v0:<slug>`. The shape you implement is the
shape the Contract declares.

Your changes carry annotations back (the **HOW→WHAT** link): each
function gets a `//specue:req:<slug>` or
`//specue:req:<slug>#<element-id>`, each test gets a `//specue:test:`.
See `specue-binding` for the binding discipline; this skill covers what
goes where in *this* code base.

## Package map

Map each layer to the Contracts it realizes. Use this to find both "where
do I change X" and "where does Y of the contract live".

| Package | Realizes | Notes |
|---|---|---|
| `internal/cli/` | every verb (`validate`, `get`, `describe`, `bindings`, `query`, `diff`, `plan*`, `context*`, `init`, `registry warm`) | Each verb has a `run<Name>` function with the whole-contract `//specue:req:<slug>`. Verbs render JSON or human via `Renderer`, never `fmt.Println`. Errors are `Errorf(fix, ...)`. |
| `internal/engine/` | `build-graph` | `derive` runs the three layers (load → scan → compile). `Live` adds the content-key cache (`#incremental`). |
| `internal/compiler/` | the gates of `validate-graph` | One file per gate concern: `rolegate.go`, `bind.go` (orphan/unbindable/duplicate), `dangling.go`, `fixpoint.go` (sync-cycle / async-cycle / blocked propagation). New gates go here. |
| `internal/specload/` | `build-graph#cue-stitches-the-modules` and `#multi-folder-modules` | Loads `./...` per module; surfaces "no package files" cleanly. |
| `internal/modules/` | the require closure | Resolver + local replace locator + in-process registry stub. |
| `internal/codescan/` | `scan-code` | Lexical scanner over `(?://|#|--)\s*specue:<verb>:`. Language is decided by extension; test-vs-impl by file name (`_test.go`, `.test.ts`, …). |
| `internal/source/` | parse + schema + workspace | The embedded schema lives here; `MaterializeSchema` writes it to disk for CUE. |
| `internal/query/` | `query-graph` | Builds the SQLite projection from the resolved graph; FTS5 over `nodes_fts`; schema doc in `tables.go`. |
| `internal/diff/` | `diff-refs` core | Pure transform; the CLI snapshots refs and feeds them in. |
| `internal/plan/` | `register-plan`, `use-plan`, `return-to-base`, `drop-plan`, `pending-overlay`, `detect-conflict`, `accept-plan` | Wraps git operations + governance record. |
| `internal/context/` | the six context Contracts + `init-module` storage | Domain layer is pure; persistence sits behind `Repository`. |
| `internal/warm/` | `warm-schema` | `EnsureWarm` (schema only) + `EnsureClosureWarm` (whole landscape). Reads through `ResolveFunc` / `ClosureResolveFunc` so tests stub `cue`. |

## Build, run, test

```
GOWORK=off go build -o ./specue ./cmd/specue/   # build
GOWORK=off go test ./...                              # full suite
GOWORK=off go vet ./...                               # vet
GOWORK=off go test ./compiler -run TestRoleGate -v    # one test
```

`GOWORK=off` is required: there is a `go.work` higher up that excludes this
module; without the env var Go tries to use it and fails.

Run the binary you built with `SPECUE_GIT=$(which git)` for most verbs.
The CLI test suite sets `SPECUE_NO_AUTOWARM=1` in `TestMain` to stay
hermetic — keep this for any new CLI test that exercises a graph build.

## Coding conventions

- **Actionable errors.** Every error renders through `Errorf(fix,
  format, args...)`; the `fix` is the next step the user takes. Bare
  `fmt.Errorf` only inside internal layers, never on the CLI seam.
- **Domain layers stay pure.** No `os`, no `time.Now`, no env reads in
  `internal/compiler/`, `internal/query/`, `internal/diff/`, `internal/plan/` (the wrapper layer is allowed in
  `plan/manager.go` only). Inject filesystems and clocks.
- **Test bindings are first-class.** A new gate without a per-element
  `//specue:test:` will show up as `bound` instead of `proven`. Add the
  matching `<gate>_test.go` and bind it; the dogfood will tell you.
- **JSON shape is stable.** Reports have explicit `jsonValue()` methods
  where needed; do not start returning new fields from a report without
  updating tests that assert the shape.
- **Sub-package CUE in spec changes.** When you add a sub-folder to a
  spec module, the loader already picks it up (`specload` walks `./...`),
  but the warm step needs a re-run on a fresh extract — see "Cache
  troubleshooting" below.

## Where to put a new feature

| You are adding... | Probably touches |
|---|---|
| a new verb | `cli/<verb>.go` (run + renderer) + `cli/<verb>_test.go` + register in `root.go` + constant in `commands.go` + matching Contract in `spec.d/service/<phase>/` |
| a new gate | `compiler/<gate>.go` (the check) + a `Diagnostic` code in `diagnostic.go` + a new invariant on `validate-graph` + a test |
| a new graph projection column | `query/build.go` (insert into SQLite) + `query/tables.go` (doc the column) + update the recipes |
| a new infra verb | `compiler/fact.go` (the verb constant) + `codescan/kind.go` if it has a unique role mapping + binding tests |
| a new module kind | `source/manifest.go` (`ModuleKind`) + `source/schema/module.cue` (`#kind`) + `compiler/rolegate.go` (`kindAllows`) + a Contract in `spec.d/service/` |
| a new spec node type | `source/schema/spec.cue` (`#NewType`) + `internal/model/` (Go type) + `source/load*.go` + `internal/compiler/`-side wiring + `query/build.go` projection columns |

The dogfood is your check at the end: rebuild, `validate`, then `bindings`
and `describe` against the Contract you just touched — the new feature should
show up `proven` once it is built and tested.

## Cache troubleshooting

The `internal/warm/` package publishes the schema and (closure variant) every
landscape module into an in-process OCI registry, then warms the cue
on-disk cache so `cue lsp` resolves natively. Two facts to remember when
something silently drifts:

- **Extract is read-only.** `<CUE_CACHE_DIR>/mod/extract/<module>@<v>/`
  is created `r-x`. `os.RemoveAll` fails on it silently if you do not
  `chmod -R u+rwx` first. `warm/closure.go:removeReadOnlyTree` is the
  helper.
- **Download zip survives republish.** `<CUE_CACHE_DIR>/mod/download/<module>/@v/<v>.zip`
  is consulted before the registry; if the zip is stale, cue serves it
  and never re-fetches. `warm/closure.go:clearModuleCache` clears both
  extract and download — both must go for a re-warm to take.

When the editor's cue lsp stops resolving and you suspect warm, run
`SPECUE_WARM_DEBUG=1 ./specue validate` — the warm error surfaces
on stderr instead of being swallowed. For a hard reset, clear the
specue paths in the cue cache by hand:

```
chmod -R u+rwx ~/Library/Caches/cue/mod/extract/specue.io
rm -rf  ~/Library/Caches/cue/mod/extract/specue.io
rm -rf  ~/Library/Caches/cue/mod/download/specue.io
rm -f   ~/Library/Caches/cue/mod/extract/.specue-warm-*
```

Then re-run `validate` and the warm rebuilds it.

## Git policy

This repo follows the global rule: do not commit unless the user has
explicitly asked. Stage only the files you intend to commit; the user
often keeps untracked dogfood artefacts (a half-written Plan, a temp
governance entry, a scratch `.cue`) you should not pull in with
`git add -A`. Show the diff first.

## Known limitations and pitfalls

- **cue lsp v0.16 does not autocomplete enum values.** Field completion
  works after warm; value completion for `"CONFIRMED" | "LIKELY" | …` or
  `"proposed" | "accepted" | …` does not. Both inline and named-alias
  enums behave the same. Closing this gap is a job for our own LSP
  later, not the schema's shape.
- **Lexical scanner cannot distinguish raw-string-literal annotations
  from real ones.** `codescan/scanner_test.go` and `engine/engine_test.go`
  hold annotation-shaped strings as test data; both files are listed by
  name in `spec.d/code/spec.mod.cue`'s `ignore:`. If you add a similar fixture in a
  new file, extend the ignore list rather than working around it in code.
- **`go.work` is a trap.** Always `GOWORK=off`. CI does this implicitly;
  local commands need it explicitly.
