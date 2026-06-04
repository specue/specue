// Package cli is the thin command layer over the engine: each verb resolves its
// inputs, calls the compute layers (engine / plan / diff), and hands a typed
// value to a Renderer. Verbs never format domain data themselves — the split
// keeps human, JSON, and (later) HTML output a single swap, and is what the web
// server reuses. The grammar is kubectl-shaped: `get <resource>`,
// `describe <module:slug>`, `graph [resource]`, plus the planning verbs.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// Renderer turns a verb's typed result into output. Human writes a TTY-friendly
// report; JSON writes a stable machine object. A verb calls exactly one method per
// run, chosen by the --json flag. The web server adds an HTML renderer against the
// same typed values without touching any verb.
type Renderer interface {
	// Report renders a successful verb result. The value is verb-specific (a
	// ValidateReport, a node view, a delta); the renderer type-switches on it.
	Report(v any) error
	// Fail renders an error. Every error MUST carry an actionable next step (the Fix
	// field): the diagnosis alone makes the caller guess at the remedy. Fail is the
	// single choke point that enforces this — see Problem.
	Fail(p Problem) error
	// Note writes a one-line aside (e.g. the active run mode) to the side channel —
	// stderr for humans, dropped for JSON so machine output stays clean.
	Note(line string)
}

// Problem is an error paired with what to do about it. The Fix is not optional: a
// CLI error without a concrete next step (a command to run, a flag to add, an edit
// to make) forces the reader — human or agent — to search for the remedy. Every
// construction site supplies one.
type Problem struct {
	// What went wrong, in plain terms.
	Summary string
	// Why, when it adds signal beyond the summary (optional).
	Detail string
	// The obvious next step: a command, a flag, an edit. REQUIRED.
	Fix string
}

func (p Problem) Error() string { return p.Summary }

// Errorf builds a Problem from a summary and a required fix.
func Errorf(fix, format string, a ...any) Problem {
	return Problem{Summary: fmt.Sprintf(format, a...), Fix: fix}
}

// humanRenderer writes a person-facing report to out and errors to err.
type humanRenderer struct {
	out io.Writer
	err io.Writer
}

func (r humanRenderer) Report(v any) error {
	rep, ok := v.(humanReportable)
	if !ok {
		return fmt.Errorf("cli: no human rendering for %T", v)
	}
	return rep.renderHuman(r.out)
}

func (r humanRenderer) Fail(p Problem) error {
	if _, err := fmt.Fprintf(r.err, "error: %s\n", p.Summary); err != nil {
		return err
	}
	if p.Detail != "" {
		if _, err := fmt.Fprintf(r.err, "  %s\n", p.Detail); err != nil {
			return err
		}
	}
	// The actionable line — always present (Problem guarantees a Fix).
	_, err := fmt.Fprintf(r.err, "\ntry: %s\n", p.Fix)
	return err
}

func (r humanRenderer) Note(line string) { fmt.Fprintln(r.err, line) }

// jsonRenderer writes a stable JSON object to out. Errors render as a JSON object
// too (with the fix), so a scripted caller reads the remedy from the same stream.
type jsonRenderer struct {
	out io.Writer
}

func (r jsonRenderer) Report(v any) error {
	rep, ok := v.(jsonReportable)
	if ok {
		v = rep.jsonValue()
	}
	return r.encode(v)
}

func (r jsonRenderer) Fail(p Problem) error {
	return r.encode(map[string]any{
		"error":  p.Summary,
		"detail": p.Detail,
		"fix":    p.Fix,
	})
}

// Note is dropped for JSON — the side channel would corrupt a machine-parsed stream.
func (r jsonRenderer) Note(string) {}

func (r jsonRenderer) encode(v any) error {
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// humanReportable is a verb result that knows how to render itself for a human.
// JSON results implement jsonReportable (or are encoded as-is).
type humanReportable interface {
	renderHuman(w io.Writer) error
}

// jsonReportable lets a result expose a stable JSON shape distinct from its Go
// struct (so the wire format does not track internal field renames).
type jsonReportable interface {
	jsonValue() any
}
