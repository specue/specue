// Package conformance holds integration tests for the specue command,
// derived FROM THE DECISIONS, ahead of and independent from any one
// implementation.
//
// The tests drive ANY specue binary: its path comes from env SPECUE_BIN.
// Each scenario (testdata/*.txtar) checks the full command path:
// input (module fixture + arguments) -> output (exit code, stdout/stderr
// contract, created files, their validity via cue vet).
//
// Each .txtar is tagged with the decisions it covers (# decision: <id>).
package conformance

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestConformance(t *testing.T) {
	bin := os.Getenv("SPECUE_BIN")
	if bin == "" {
		t.Skip("SPECUE_BIN is not set — path to the specue binary under test")
	}
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue not found in PATH — needed to validate generated files")
	}
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not found in PATH — needed to assert the JSON output contract")
	}
	schema := os.Getenv("SCHEMA_DIR")
	if schema == "" {
		t.Fatal("SCHEMA_DIR is not set — absolute path to cue/schema for replace")
	}

	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(e *testscript.Env) error {
			// 'specue' inside scripts == the binary under test
			e.Vars = append(e.Vars,
				"SPECUE_BIN="+bin,
				"SCHEMA_DIR="+schema,
				// offline resolution via replace onto the local schema
				"CUE_REGISTRY=localhost:9999+insecure",
			)
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			// specue <args...> — run the binary under test
			"specue": func(ts *testscript.TestScript, neg bool, args []string) {
				err := ts.Exec(ts.Getenv("SPECUE_BIN"), args...)
				if neg && err == nil {
					ts.Fatalf("expected non-zero exit, got 0")
				}
				if !neg && err != nil {
					ts.Fatalf("command failed: %v", err)
				}
			},
			// link-schema — writes cue.mod/local-module.cue with a replace
			// onto the local schema (absolute path from SCHEMA_DIR). Called
			// from a scenario after module.cue so offline resolution finds it.
			"link-schema": func(ts *testscript.TestScript, neg bool, args []string) {
				dir := filepath.Join(ts.MkAbs("cue.mod"))
				content := `deps: {
	"specue.io/schema@v0": {
		v:       "v0.0.4"
		replace: "` + ts.Getenv("SCHEMA_DIR") + `"
	}
}
`
				if err := os.WriteFile(filepath.Join(dir, "local-module.cue"), []byte(content), 0o644); err != nil {
					ts.Fatalf("link-schema: %v", err)
				}
			},
		},
	})
}
