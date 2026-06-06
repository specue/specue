---
name: specue-binding
description: Annotate source code with //specue: req / test / infra to tie code to spec
---

# Binding code to spec

Use this skill when you are placing `//specue:req:<slug>`,
`//specue:test:<slug>` or any of the infra verbs (`produces`, `consumes`,
`publishes`, `subscribes`, `serves`, `calls`, `reads`, `writes`, `grants`)
in source code of a code-kind module — any project, any language.

The shared Specue skill (`specue` → `authoring.md`, `maintaining.md`)
covers verb→role semantics, element-scoped vs whole-contract, how status is
computed. This skill is the **discipline of binding**: where to put each
annotation, when not to, how to keep the bindings view honest.

## The three layers, on the code side

Binding is the **HOW** layer talking back to the **WHAT** layer.

- WHAT lives in the spec as a Contract with named invariants.
- HOW lives in the function or method that realizes the invariant; the
  annotation declares the link.
- WHY is not the code's concern; `decided_by` already lives on the spec
  invariant.

An annotation does not duplicate the spec — it points at it. The spec text
is the authoritative WHAT; the code is the authoritative HOW; the
annotation records the correspondence.

## Two scopes per UC: whole and per-element

Every implemented UC carries a **whole-contract** annotation on the main
function/method that realizes it:

```go
//specue:req:validate-graph
func ApplyOperation(...) { ... }
```

That alone is enough for the UC to be `implemented`. Add a **per-element**
annotation when a specific invariant has its own home distinct from the
whole-contract anchor:

```go
//specue:req:validate-graph#idempotent
func reapplyIsNoOp(...) { ... }
```

Stack them when one function realizes the whole contract AND a specific
invariant of it — each on its own line, both at column 0:

```go
//specue:req:validate-graph
//specue:req:validate-graph#idempotent
func ApplyOperation(...) { ... }
```

Per-element binding is what makes the `bindings` view a useful TODO list.
Without it, every invariant rolls up under the whole-contract anchor and
the view cannot tell you which specific invariant needs a test.

## Tests prove a contract

`//specue:test:<slug>` and `//specue:test:<slug>#<element-id>` go on
test functions. A whole-contract test proves the whole UC; a per-element
test proves that specific invariant. A binding reaches `proven` when both
`req` and `test` are present at the same scope (whole or per-element).

Be precise about what the test asserts. Pin the `test:` annotation to the
test that *directly asserts* the invariant, not one that incidentally
touches the code path. A test of "the verb works end-to-end" is a
whole-contract test; a test of "the cycle gate fires" is the per-element
test for `#sync-cycle`. Mixing them up makes the bindings view lie.

## When to leave an element unbound

This is dignity, not negligence. Leave unbound when:

- the invariant is **delegated to the platform** — to CUE itself, to a
  framework, to a library — and you have no host-language code to point
  at;
- the invariant is **not yet realized** — declared in the contract but no
  code emits it today; the `unbound` row in `bindings` is the honest TODO;
- the invariant **diverges** from the code — the spec says "gate" but the
  current implementation reports it as a status only; fix the divergence
  by editing the spec or the code, do not paper over with a wrong
  binding;
- a specific test for that invariant **does not exist**; leave `test:` off
  and accept `bound` instead of `proven`.

The `bindings` view will show these as `unbound`. That is the truthful
state and the right TODO; faking a binding to silence it costs the graph
its value.

## Workflow

1. Pick the UC you are about to implement or prove. Read its spec — every
   invariant text — so you know what you are pointing at.
2. Decide the whole-contract anchor: the function or method that
   realizes the contract end-to-end. Annotate it
   `//specue:req:<slug>`.
3. For each invariant, find the specific function that realizes only that
   invariant. If there is one, annotate it
   `//specue:req:<slug>#<element-id>`. If the whole-contract anchor
   *is* the only home, do not add a per-element duplicate.
4. For each invariant, find the test that asserts it directly. Annotate
   it `//specue:test:<slug>#<element-id>`. Do the same for a
   whole-contract test.
5. Build and run `validate` — an `orphan-binding` means the slug in the
   annotation does not resolve through the code module's require closure.
   Either the spec slug is wrong, or the code module's `require` is
   missing the spec module that owns the slug.
6. `bindings | grep <slug>` (or filtered by `--state`) — read the result
   and confirm it matches your intent: which rows are `proven`, which are
   `bound`, which are honestly `unbound`.

## Infra verbs

`produces`, `consumes`, `publishes`, `subscribes`, `serves`, `calls`,
`reads`, `writes`, `grants` anchor an edge to a Port — they are not
implementations of a UC, they prove an edge the UC declares. Use the verb
that matches the role the UC's spec declared. An anchor whose verb does
not match a declared role is a diagnostic; if you want a new role on the
UC, add it to the invariant first, then bind.

Infra anchors do not have a test counterpart. Their proof is the anchor
itself: `bound` once the anchor exists, never `proven` (there is no
separate test scope for an infra edge).

## Where the code module lives — `spec.d/code/` and `code_root`

A code module's manifest names the **kind** (`code`) and a **`code_root`** —
the path the lexical scan starts from, relative to the manifest's directory.
Default `"."`: the scan walks the manifest's own dir. Set it whenever the
manifest does not live at the root of the source tree it scans.

The recommended layout puts every module of a repo under `spec.d/<kind>/`
(Unix drop-in convention — `cron.d`, `sudoers.d` — visible to agents and
shells; no leading dot). For a code module that means:

```
repo/
├── internal/, cmd/, …             # the Go source the scan covers
└── spec.d/
    ├── code/
    │   ├── spec.mod.cue           # kind: "code", code_root: "../.."
    │   └── cue.mod/module.cue
    └── service/<name>/            # the contracts the code realizes
```

`specue init . <module>@v0 --kind code --layout spec.d` writes exactly
that — including `code_root: "../.."`, which sends the scan two levels up
to the repo root.

**Why a subfolder?** A code module placed at the repo root has, in its own
filesystem subtree, every nested `spec.mod.cue` — including any service
module the same workspace registers standalone. CUE then sees the same
package via two paths (workspace + nested) and refuses with `ambiguous
import` the moment another module imports it. Putting the code manifest in
`spec.d/code/` and pointing `code_root` back at the repo separates the
manifest from the source it scans, so the service module is only reached
via its own registration. See ADR-11 (`gov:adr11CodeRootAndLayout`).

The flat layout (code manifest at the repo root, service module at `./spec/`)
still works for landscapes that do not register a nested module standalone
in the same workspace — `code_root` alone is enough there. New repos should
prefer `spec.d/`.

## Conventions

- Each annotation on its own line at column 0 (no indent), immediately
  above the `func` declaration. If there is a doc comment, the annotation
  goes between the last doc line and the `func` keyword.
- Bare slugs (no alias). The code module's `require` resolves them.
- An annotation quoted inside a comment of its own (e.g. a doc comment
  that mentions `//specue:req:slug` as syntax) is not a binding — the
  scanner ignores it. An annotation appearing inside a raw-string literal
  in the host language *is* taken as a binding by the lexical scanner;
  if the file is a scanner test fixture, exclude it by name in
  `spec.mod.cue`'s `ignore:`.
- Bind on the host-language unit that gives the most precise location:
  the function/method that realizes the invariant, not the file or the
  package as a whole.
