package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/modules"
	"github.com/specue/specue/internal/source"
	"github.com/specue/specue/internal/warm"
)

// newRegistryCmd wires `specue registry warm`: it seeds the schema into the
// cue module cache so the editor's stock `cue lsp` can resolve specue.io/schema
// (and autocomplete schema fields) natively, with no daemon and no CUE_REGISTRY in
// the editor's environment. Idempotent — a no-op when the cache already holds the
// current schema content. Needs no spec tree (it only touches the embedded schema).
func newRegistryCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	c := &cobra.Command{
		Use:   "registry",
		Short: "Manage the local schema cache the editor's cue lsp reads",
		Long: "registry seeds the specue schema into the cue module cache so the editor's\n" +
			"stock `cue lsp` resolves it natively — schema-field autocomplete with no daemon.",
	}
	c.AddCommand(newRegistryWarmCmd(g, out, err, code))
	return c
}

func newRegistryWarmCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	return &cobra.Command{
		Use:   "warm",
		Short: "Seed the schema into the cue cache (idempotent)",
		Long: "warm publishes the schema into an ephemeral in-memory registry and runs one\n" +
			"resolve to materialize its extract in the cue cache. After this the editor's\n" +
			"cue lsp resolves the schema from the cache with no registry alive. Re-run after\n" +
			"clearing the cache or changing the schema; it is a no-op when already current.",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runRegistryWarm(g, out, err, code)
		},
	}
}

// WarmReport is the typed result of `registry warm`: whether a (re)warm ran, or
// the cache was already current.
type WarmReport struct {
	Rewarmed bool `json:"rewarmed"`
}

func (r WarmReport) renderHuman(w io.Writer) error {
	if r.Rewarmed {
		_, err := fmt.Fprintln(w, "schema warmed into the cue cache; cue lsp resolves it natively now")
		return err
	}
	_, err := fmt.Fprintln(w, "cache already current; nothing to do")
	return err
}

// resolveClosureForWarm builds the landscape closure (every module of the active
// workspace plus the schema) and the list of root dirs needed to drive a warm.
// Self-contained: it does not touch the engine — warm runs alongside it, not
// through it. Returns ok=false when there is no usable landscape (no active
// context, resolve failure); the caller falls back to a schema-only warm.
func resolveClosureForWarm(ctx Context) (modules.Closure, []string, bool) {
	work, dirs, p := ctx.workspace()
	if p != nil || len(work.Modules) == 0 {
		return modules.Closure{}, nil, false
	}
	parser, err := source.NewCUEParser()
	if err != nil {
		return modules.Closure{}, nil, false
	}
	closure, err := modules.NewResolver(parser, modules.NewReplaceLocator()).ResolveWork(work, dirs)
	if err != nil {
		return modules.Closure{}, nil, false
	}
	schema, err := modules.NewSchemaModule()
	if err != nil {
		return modules.Closure{}, nil, false
	}
	closure.Modules = append(closure.Modules, schema.ResolvedModule)
	// roots: every resolved module's dir — running `cue vet ./...` in each of them
	// covers the closure's extracts. (The schema is excluded — its dir is the
	// materialized temp tree, not a root the user owns.)
	roots := make([]string, 0, len(closure.Modules)-1)
	for _, m := range closure.Modules {
		if m.Path == model.ModulePath(source.SchemaModulePath) {
			continue
		}
		roots = append(roots, m.Dir)
	}
	return closure, roots, true
}

// bestEffortWarm seeds the schema cache as a side-effect of the normal cycle
// (validate, context use), so the editor's cue lsp keeps resolving even after the
// cache is cleared — without the user knowing a warm was ever needed. It is
// strictly best-effort: any failure (no `cue` on PATH, no network, a transient
// registry error) is swallowed. It must NEVER break the command it rides along
// with — the tool's own resolution does not depend on this cache. EnsureWarm is a
// cheap no-op when the cache is already current, so this is safe to call often.
func bestEffortWarm() {
	// A kill switch for environments that don't want the implicit cue invocation:
	// CI, or tests that must stay hermetic (not touch the user's real cue cache).
	if os.Getenv("SPECUE_NO_AUTOWARM") != "" {
		return
	}
	w, err := warm.New("", warm.CueResolve(""))
	if err != nil {
		return
	}
	_, _ = w.EnsureWarm()
}

// bestEffortWarmClosure is the warm path called from a verb that already has a
// resolved Context: it seeds not only the schema but every local module of the
// landscape, so the editor's cue lsp resolves cross-module references and offers
// go-to-definition between modules. If the context cannot be resolved here (no
// active landscape, build error), it falls back to the schema-only warm — never
// gates the verb. Each module is a no-op when its source has not changed.
func bestEffortWarmClosure(ctx Context) {
	if os.Getenv("SPECUE_NO_AUTOWARM") != "" {
		return
	}
	w, err := warm.New("", warm.CueResolve(""))
	if err != nil {
		return
	}
	closure, roots, ok := resolveClosureForWarm(ctx)
	if !ok {
		// No usable landscape (e.g. resolve failed); the schema warm is still useful.
		_, _ = w.EnsureWarm()
		return
	}
	if _, err := w.EnsureClosureWarm(closure, roots, warm.CueResolveClosure("")); err != nil && os.Getenv("SPECUE_WARM_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "warm-closure: %v\n", err)
	}
}

// runRegistryWarm builds a Warmer with the production cue-backed resolve and runs
// EnsureWarm. It is not a graph verb, so it renders directly rather than going
// through dispatch.
//
//specue:req:warm-schema
func runRegistryWarm(g *Globals, out, errw io.Writer, code *int) error {
	r := g.renderer(out, errw)
	w, err := warm.New("", warm.CueResolve(""))
	if err != nil {
		*code = exitUsage
		return r.Fail(Errorf("ensure `cue` is installed and on PATH", "set up schema cache: %v", err))
	}
	rewarmed, err := w.EnsureWarm()
	if err != nil {
		*code = exitUsage
		return r.Fail(Errorf("ensure `cue` is installed and on PATH, then retry", "warm schema cache: %v", err))
	}
	return r.Report(WarmReport{Rewarmed: rewarmed})
}
