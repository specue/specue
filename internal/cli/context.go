package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/specue/specue/internal/codescan"
	"github.com/specue/specue/internal/context"
	"github.com/specue/specue/internal/engine"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/source"
)

// Globals holds the flags every verb shares. They bind once on the root command
// (persistent flags) and resolve into a Context.
type Globals struct {
	Dir       string // -C: act on the module in this directory (module mode)
	Workspace string // --workspace: act on this named context this run (overrides the active one)
	Attested  bool   // -attested: status from spec.attest, never scan code
	JSON      bool   // -json: machine output
	NoColor   bool   // -no-color: disable color (also honors NO_COLOR / non-TTY)
	Debug     bool   // --debug: trace per-module load (instances, files, errors) to stderr

	// planMode marks a plan verb: when no landscape can be resolved, the fix is
	// plan-specific (bootstrap a landscape WITH a governance module) rather than the
	// generic "scaffold a module" — a plan needs more than any single module.
	planMode bool
}

// runMode is which landscape a run resolved to — surfaced in output so the
// compiler's "world" is never ambiguous (the duality that bit us before). Workspace
// mode is a named context's full landscape; module mode is a single module in
// isolation (resolved via its own require/replace closure).
type runMode int

const (
	modeModule    runMode = iota // a single module, isolated
	modeWorkspace                // a named context's landscape
)

// Context is a resolved run: the landscape (always an in-memory source.Workspace,
// built from a named context's modules or from a single module dir) plus the
// renderer the flags selected. There is no spec.work file — the landscape lives in
// the context registry or is synthesized for one module.
type Context struct {
	Work  *source.Workspace // the resolved landscape, always set on success
	mode  runMode           // how it resolved (printed by the renderer)
	label string            // the context name (workspace mode) or module path (module mode)
	g     Globals
	r     Renderer
}

// Renderer returns the output renderer the flags selected.
func (c Context) Renderer() Renderer { return c.r }

// banner notes the resolved run mode on the side channel, so which landscape the
// command acted on is always visible — workspace `name` or an isolated module.
func (c Context) banner() {
	switch c.mode {
	case modeWorkspace:
		c.r.Note("workspace: " + c.label)
	case modeModule:
		c.r.Note("module: " + c.label + " (isolated)")
	}
}

// resolve turns Globals into a Context, choosing the run mode explicitly so the
// compiler's "world" is never ambiguous. Precedence (highest first):
//
//  1. --workspace <name> → that named context (workspace mode)
//  2. -C <dir>, or cwd holding a spec.mod → that module alone (module mode); this
//     BEATS the active context — "act on this module, ignore the workspace"
//  3. the active context → its landscape (workspace mode)
//  4. nothing            → an actionable error
//
// The landscape is always an in-memory source.Workspace, built from a context's
// registered modules or synthesized for one module dir — there is no spec.work file
// to discover.
func (g Globals) resolve(stdout, stderr io.Writer) (Context, *Problem) {
	r := g.renderer(stdout, stderr)

	// 1: an explicit named context.
	if g.Workspace != "" {
		work, p := g.contextWorkspace(g.Workspace)
		if p != nil {
			return Context{r: r}, p
		}
		return Context{Work: work, mode: modeWorkspace, label: g.Workspace, g: g, r: r}, nil
	}

	// 2: an EXPLICIT -C <dir> module — a deliberate downgrade to one module. Only an
	// explicit flag beats the active context; merely standing in a module directory
	// does not (see 4), so an active context stays in force while you work — crucial
	// with dogfooding, where a code module's spec.mod sits at the repo root you run
	// from.
	if g.Dir != "" {
		mod, ok, p := g.dirModule(g.Dir)
		if p != nil {
			return Context{r: r}, p
		}
		if ok {
			return Context{Work: mod, mode: modeModule, label: string(mod.Modules[0].Path), g: g, r: r}, nil
		}
	}

	// 3: the active context — set globally, so you are in it until you explicitly
	// downgrade with -C.
	repo, p := contextRepo()
	if p != nil {
		return Context{r: r}, p
	}
	s, err := repo.Load()
	if err != nil {
		p := Errorf("the registry file may be corrupt — fix or remove it", "cannot read contexts: %v", err)
		return Context{r: r}, &p
	}
	if e, ok := s.ActiveEntry(); ok {
		return Context{Work: entryWorkspace(e), mode: modeWorkspace, label: e.Name, g: g, r: r}, nil
	}

	// 4: no active context — fall back to an implicit cwd module if one is here.
	if wd, err := os.Getwd(); err == nil {
		mod, ok, p := g.dirModule(wd)
		if p != nil {
			return Context{r: r}, p
		}
		if ok {
			return Context{Work: mod, mode: modeModule, label: string(mod.Modules[0].Path), g: g, r: r}, nil
		}
	}

	// 5: nothing resolved.
	return Context{r: r}, g.noLandscapeProblem()
}

// contextWorkspace builds the landscape for a named context from the registry.
func (g Globals) contextWorkspace(name string) (*source.Workspace, *Problem) {
	repo, p := contextRepo()
	if p != nil {
		return nil, p
	}
	s, err := repo.Load()
	if err != nil {
		p := Errorf("the registry file may be corrupt — fix or remove it", "cannot read contexts: %v", err)
		return nil, &p
	}
	e, ok := s.Find(name)
	if !ok {
		p := Errorf("see contexts with `"+usage(cmdContext, subList)+"`", "no context named %q", name)
		return nil, &p
	}
	return entryWorkspace(e), nil
}

// entryWorkspace builds a source.Workspace from a context entry. Module dirs are
// stored absolute, so Root is empty and each module sits at its own absolute dir.
func entryWorkspace(e context.Entry) *source.Workspace {
	w := &source.Workspace{PlanBase: e.PlanBase}
	for _, m := range e.Modules {
		w.Modules = append(w.Modules, source.WorkModule{Path: model.ModulePath(m.Path), Dir: m.Dir})
	}
	return w
}

// dirModule builds a single-module landscape from dir if it holds a spec.mod;
// ok=false when dir is not a module (the caller falls through). The caller decides
// what dir means — an explicit -C (priority 2) or the implicit cwd fallback
// (priority 4) — so this stays a pure "is this dir a module" probe.
func (g Globals) dirModule(dir string) (*source.Workspace, bool, *Problem) {
	if !fileExists(filepath.Join(dir, source.ManifestFile)) {
		return nil, false, nil
	}
	work, err := singleModuleWorkspace(dir)
	if err != nil {
		p := Errorf("check the module's spec.mod.cue is valid CUE",
			"cannot build a single-module landscape for %s: %v", dir, err)
		return nil, false, &p
	}
	return work, true, nil
}

// noLandscapeProblem is the fix when nothing resolved: pick a context, cd into a
// module, or scaffold. Plan verbs get the plan-specific bootstrap.
func (g Globals) noLandscapeProblem() *Problem {
	if g.planMode {
		p := planBootstrapProblem()
		return &p
	}
	p := Errorf("select a workspace with `"+usage(cmdContext, subUse)+"`, cd into a module, or scaffold one with `"+usage(cmdInit)+"`",
		"no active context and no module here")
	return &p
}

// renderer picks the human or JSON renderer. (Color wiring lands with the verbs
// that emit it; the flag and NO_COLOR/TTY detection are resolved here.)
func (g Globals) renderer(stdout, stderr io.Writer) Renderer {
	if g.JSON {
		return jsonRenderer{out: stdout}
	}
	return humanRenderer{out: stdout, err: stderr}
}

// singleModuleWorkspace builds an in-memory landscape of just the module at modDir
// (module mode). The path comes from its spec.mod.cue; Root is the absolute module
// dir and the module sits at ".".
func singleModuleWorkspace(modDir string) (*source.Workspace, error) {
	abs, err := filepath.Abs(modDir)
	if err != nil {
		return nil, err
	}
	mf, err := source.ReadManifest(filepath.Join(abs, source.ManifestFile))
	if err != nil {
		return nil, err
	}
	return &source.Workspace{
		Root:    abs,
		Modules: []source.WorkModule{{Path: mf.Path, Dir: "."}},
	}, nil
}

// engineConfig builds the engine.Config for this run — the in-memory Workspace (a
// context's modules or one module dir) plus a scan target per code module so the
// graph is derived against real code. A code module's files come from git
// (MANIFESTO P20): tracked files only, .gitignore'd trees skipped.
func (c Context) engineConfig() engine.Config {
	return engine.Config{Workspace: c.Work, ScanTargets: c.scanTargets()}
}

// engineOptions translates the global flags that change the engine's wiring (not
// its inputs) into engine.Option values — currently only --debug, which routes
// per-module load tracing straight to os.Stderr so traces interleave with normal
// progress regardless of which renderer is active (the JSON renderer drops the
// side channel; debug is forensic, not part of the data shape).
func (c Context) engineOptions() []engine.Option {
	var opts []engine.Option
	if c.g.Debug {
		opts = append(opts, engine.WithLoadDebug(os.Stderr))
	}
	return opts
}

// checkModuleDirs verifies every resolved module directory still exists before the
// engine reads it. A vanished dir (a context pointing at a deleted module) would
// otherwise fail opaquely inside CUE; here it names the dir and the remedy — which
// differs by mode: a workspace can drop the stale entry, a single module's dir is
// simply wrong.
func (c Context) checkModuleDirs() *Problem {
	_, dirs, p := c.workspace()
	if p != nil {
		return p
	}
	for path, dir := range dirs {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			var fix string
			if c.mode == modeWorkspace {
				fix = fmt.Sprintf("restore it, or drop it with `%s %s`",
					cmdPath(cmdContext, subModule, subRemove), path)
			} else {
				fix = "restore the directory, or point -C at an existing module"
			}
			pr := Errorf(fix, "module %s: directory %s does not exist", path, dir)
			return &pr
		}
	}
	return nil
}

// scanTargets builds the code scan set for this run by handing the resolved
// landscape to engine.ScanTargetsFor with the cli-edge dependencies: git as the
// file lister (tracked files only — MANIFESTO P20) and os.DirFS as the reader. The
// domain logic (which modules are code, how a target is shaped) lives in engine so
// the server and other callers reuse it. A resolution or listing error yields no
// targets — validate surfaces real problems through the graph.
func (c Context) scanTargets() []codescan.ScanTarget {
	work, dirs, p := c.workspace()
	if p != nil {
		return nil
	}
	git := plan.NewGit(gitBin())
	targets, err := engine.ScanTargetsFor(work, dirs,
		git.ListFiles,
		func(dir string) fs.FS { return os.DirFS(dir) })
	if err != nil {
		return nil
	}
	return targets
}

// workspace returns the resolved landscape and each module's absolute directory —
// the inputs the plan/diff layers take. Module dirs are already absolute (registry)
// or relative to Root (single module); both are resolved here.
func (c Context) workspace() (source.Workspace, map[model.ModulePath]string, *Problem) {
	work := *c.Work
	dirs := make(map[model.ModulePath]string, len(work.Modules))
	for _, m := range work.Modules {
		dir := m.Dir
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(work.Root, dir)
		}
		dirs[m.Path] = dir
	}
	return work, dirs, nil
}

// governanceModule finds the workspace's governance module — where #Plan records
// live — by reading each module's manifest and matching kind: governance. Exactly
// one is expected; none or several is an actionable error (the plan verbs need a
// single, unambiguous record home).
func (c Context) governanceModule() (source.Workspace, map[model.ModulePath]string, model.ModulePath, *Problem) {
	work, dirs, p := c.workspace()
	if p != nil {
		return source.Workspace{}, nil, "", p
	}
	var found []model.ModulePath
	for _, wm := range work.Modules {
		mf, err := source.ReadManifest(filepath.Join(dirs[wm.Path], source.ManifestFile))
		if err != nil {
			continue // a module we can't read its manifest for can't be the gov module
		}
		if mf.Kind == source.KindGovernance {
			found = append(found, wm.Path)
		}
	}
	switch len(found) {
	case 1:
		return work, dirs, found[0], nil
	case 0:
		return source.Workspace{}, nil, "", c.noGovernanceProblem()
	default:
		p := Errorf("keep exactly one governance module in the landscape",
			"%d governance modules found; plan records have one home", len(found))
		return source.Workspace{}, nil, "", &p
	}
}

// noGovernanceProblem builds the "no governance module" error, with a fix that
// differs by situation. With no spec.work at all (a single module resolved via the
// nearest spec.mod), plans need a landscape FIRST, so the fix bootstraps one. With a
// spec.work present (a real landscape lacking a governance module), the fix just
// adds one to it.
func (c Context) noGovernanceProblem() *Problem {
	if c.mode == modeModule {
		// A single module in isolation — plans need a workspace, not a lone module.
		p := planBootstrapProblem()
		return &p
	}
	// A workspace that simply lacks a governance module: add one to it.
	fix := "scaffold one with `" + usage(cmdInit) + " --kind governance`, then add it with `" +
		usage(cmdContext, subModule, subAdd) + "`"
	p := Errorf(fix, "no governance module in this workspace (plan verbs need one)")
	return &p
}

// planBootstrapProblem is the fix for running a plan verb with no workspace (module
// mode or nothing resolved). A plan is not just any module — it needs a workspace
// with a governance module to hold its record — so the hint sets that up directly
// rather than the generic "scaffold a module", which leads into an unrelated module.
func planBootstrapProblem() Problem {
	// Order matters: create, then use (so the next steps act on the active context),
	// then scaffold the governance module and add it.
	return Errorf(
		"set up a workspace: `"+usage(cmdContext, subCreate)+"`, `"+usage(cmdContext, subUse)+
			"`, `"+usage(cmdInit)+" --kind governance`, then `"+usage(cmdContext, subModule, subAdd)+"`",
		"no workspace here (plan verbs need a workspace with a governance module)")
}

// gitBin is the git binary the plan/diff layers drive. It is injectable so the
// host need not have git on PATH under a fixed name: $SPECUE_GIT overrides,
// else "git". The binary is never assumed present — a missing one surfaces as the
// layer's own actionable error when it first runs git.
func gitBin() string {
	if b := os.Getenv("SPECUE_GIT"); b != "" {
		return b
	}
	return "git"
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// requireGitRepo enforces the git-native invariant (MANIFESTO P20): the module at
// dir must live inside a git repository. dir may not exist yet (init creates it),
// so the check runs against its deepest existing ancestor. Returns an actionable
// Problem naming `git init` when no repository encloses it.
//specue:req:init-module#git-repository-required
//specue:req:add-module-to-context#git-repository-required
func requireGitRepo(dir string) *Problem {
	probe := existingAncestor(dir)
	if _, err := plan.NewGit(gitBin()).RepoRoot(probe); err != nil {
		p := Errorf(fmt.Sprintf("run `git init` in %s first — Specue is git-native", probe),
			"%s is not inside a git repository", dir)
		return &p
	}
	return nil
}

// existingAncestor returns dir if it exists, else its nearest existing parent —
// the directory git can actually be asked about before init creates the leaf.
func existingAncestor(dir string) string {
	for {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir // reached the root
		}
		dir = parent
	}
}

