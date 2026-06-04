package cli

import (
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// Command names live here as constants, referenced by the cobra `Use:` line that
// defines a command and by every hint that points at one. The cobra command's
// `Use:` field is the single source of truth for a command's ARGUMENT signature;
// hints read it back through usage() so neither name nor args can drift. A deleted
// constant makes the compiler flag every stale reference.
const (
	cmdValidate = "validate"
	cmdGet      = "get"
	cmdDescribe = "describe"
	cmdDiff     = "diff"
	cmdPlan     = "plan"
	cmdInit     = "init"
	cmdContext  = "context"
	cmdBindings = "bindings"
	cmdQuery    = "query"
	cmdRegistry = "registry"
	cmdRender   = "render"
)

// Subcommand names (the second word of a nested command).
const (
	subList   = "list"
	subAdd    = "add"
	subRemove = "remove"
	subUse    = "use"
	subCreate = "create"
	subModule = "module"
	subTables = "tables"
	subWarm   = "warm"
)

// usage returns the full invocation for a command path — program name, command
// path, and its argument signature — read from the live command tree's `Use:`
// field, so a hint always matches the command's actual arguments.
// usage(cmdContext, subModule, subAdd) → "specue context module add <dir>". A
// path that does not resolve falls back to the bare path (so a hint never blanks
// out), but TestHintedCommandsExist guards against that case.
func usage(path ...string) string {
	c := findCommand(path...)
	if c == nil {
		return "specue " + strings.Join(path, " ")
	}
	// CommandPath includes the program name ("specue work add"); append the
	// argument tail from Use ("<name> <args…>") so the hint shows the real signature.
	out := c.CommandPath()
	if i := strings.IndexByte(c.Use, ' '); i >= 0 {
		out += c.Use[i:]
	}
	return out
}

// cmdPath returns just the command path (no argument signature) — for hints that
// reference a command without spelling out its args ("run `specue plan list`").
func cmdPath(path ...string) string {
	if c := findCommand(path...); c != nil {
		return c.CommandPath()
	}
	return "specue " + strings.Join(path, " ")
}

// findCommand walks the freshly built tree to the command at path. The tree is
// cheap to build and hints are rare, so this needs no caching.
func findCommand(path ...string) *cobra.Command {
	var g Globals
	code := exitOK
	cur := newRootCmd(&g, io.Discard, io.Discard, &code)
	for _, name := range path {
		next := childNamed(cur, name)
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

func childNamed(parent *cobra.Command, name string) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
