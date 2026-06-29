# Conformance — tests derived from the decisions

Integration tests for the `specue` command, derived **from the decision graph**
and **independent of any one implementation**. They drive any `specue` binary
through the full path: input (a module fixture + arguments) → output (exit code,
stdout/stderr contract, created files, their validity via `cue vet`).

Methodology: tests are written FROM the decisions (before and independent of the
code). An implementation is correct if it passes this suite.

## Running

Reproducible (the go + cue + jq toolchain is pinned in `mise.toml`):

```sh
mise run conformance   # builds the binary, runs the suite with the pinned tools
```

Manually, against any implementation:

```sh
SPECUE_BIN=/path/to/specue/binary \
SCHEMA_DIR=/abs/path/to/cue/schema \
go test ./...
```

- `SPECUE_BIN` — the binary under test (any implementation). Unset → tests skip.
- `SCHEMA_DIR` — absolute path to `cue/schema` for the offline replace.
- `cue` and `jq` must be in PATH (validating generated files / asserting the JSON
  contract). Missing → tests skip.

## What the environment provides

- `go`, `cue`, `jq` — toolchain (cue/jq from PATH, go for build and runner).
- An isolated `$WORK` per scenario (testscript).
- Module fixtures inline in each `.txtar` (`-- cue.mod/module.cue --`, etc.).
- A `link-schema` command that writes `cue.mod/local-module.cue` with a replace
  onto the local schema (for scenarios where the module is valid).
- The `specue` command inside a script == `$SPECUE_BIN`.

## Scenarios and the decisions they cover

Each `.txtar` is tagged `# decision: <id>` — the decisions it checks.

| Scenario | Decisions |
|---|---|
| `add_success` | cli/commands/add/command, cli/commands/add/output, adding-decision, creating-decision-package, decision-file-layout, decision-entrypoint |
| `add_not_a_module` | resolving-module, cli/commands/add/output |
| `add_invalid_module` | loading-module, cli/commands/add/output |
| `add_schema_missing` | loading-module, cli/commands/add/output |
| `add_dir_not_empty` | creating-decision-package, cli/commands/add/output |
| `add_missing_arg` | cli/commands/add/command, applying-cli-guideline, cli/commands/add/output |

When you add a decision that changes command behaviour, add a scenario with its
tag here — so decision coverage by tests stays visible.
