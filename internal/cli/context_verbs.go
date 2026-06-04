package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/specue/specue/internal/context"
	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/source"
)

// withStore loads the registry, runs mutate, and (if mutate returns no Problem)
// saves it — the shared shape of every context verb that changes the registry.
func withStore(mutate func(*context.Store) (ContextActionReport, *Problem)) (ContextActionReport, *Problem) {
	repo, p := contextRepo()
	if p != nil {
		return ContextActionReport{}, p
	}
	s, err := repo.Load()
	if err != nil {
		p := Errorf("the registry file may be corrupt — fix or remove it", "cannot read contexts: %v", err)
		return ContextActionReport{}, &p
	}
	rep, p := mutate(&s)
	if p != nil {
		return ContextActionReport{}, p
	}
	if err := repo.Save(s); err != nil {
		p := Errorf("check the specue home is writable", "cannot save contexts: %v", err)
		return ContextActionReport{}, &p
	}
	return rep, nil
}

// --- list ---

// ContextListReport is the typed result of `context list`: every context with its
// module count, the active one marked.
type ContextListReport struct {
	Contexts []contextRowJSON `json:"contexts"`
}

type contextRowJSON struct {
	Name    string `json:"name"`
	Modules int    `json:"modules"`
	Active  bool   `json:"active,omitempty"`
}

func runContextList() (ContextListReport, *Problem) {
	repo, p := contextRepo()
	if p != nil {
		return ContextListReport{}, p
	}
	s, err := repo.Load()
	if err != nil {
		p := Errorf("the registry file may be corrupt — fix or remove it", "cannot read contexts: %v", err)
		return ContextListReport{}, &p
	}
	rep := ContextListReport{}
	for _, c := range s.Contexts {
		rep.Contexts = append(rep.Contexts, contextRowJSON{Name: c.Name, Modules: len(c.Modules), Active: c.Name == s.Active})
	}
	return rep, nil
}

func (r ContextListReport) renderHuman(w io.Writer) error {
	if len(r.Contexts) == 0 {
		_, err := fmt.Fprintf(w, "no contexts — create one with `%s`\n", usage(cmdContext, subCreate))
		return err
	}
	for _, c := range r.Contexts {
		mark := "  "
		if c.Active {
			mark = "* "
		}
		if _, err := fmt.Fprintf(w, "%s%s\t%d module(s)\n", mark, c.Name, c.Modules); err != nil {
			return err
		}
	}
	return nil
}

func (r ContextListReport) jsonValue() any {
	if r.Contexts == nil {
		r.Contexts = []contextRowJSON{}
	}
	return r
}

// ContextActionReport is the one-line outcome of a mutating/selecting context verb.
type ContextActionReport struct {
	Message string `json:"message"`
	Active  string `json:"active,omitempty"`
}

func (r ContextActionReport) renderHuman(w io.Writer) error {
	_, err := fmt.Fprintln(w, r.Message)
	return err
}
func (r ContextActionReport) jsonValue() any { return r }

// --- create / use / current / remove ---

// runContextCreate registers a new, empty context. Modules are added with
// `context module add`; nothing is written to the user's tree — the landscape lives
// in the registry.
//
//specue:req:create-context
func runContextCreate(name string) (ContextActionReport, *Problem) {
	return withStore(func(s *context.Store) (ContextActionReport, *Problem) {
		if err := s.Create(name); err != nil {
			var dup *context.DuplicateContextError
			if errors.As(err, &dup) {
				p := Errorf(fmt.Sprintf("pick another name, or inspect it with `%s`", usage(cmdContext, subList)), "%v", err)
				return ContextActionReport{}, &p
			}
			p := Errorf("re-run; if it persists, report it", "%v", err)
			return ContextActionReport{}, &p
		}
		return ContextActionReport{Message: fmt.Sprintf("created context %s — add modules with `%s`", name, usage(cmdContext, subModule, subAdd))}, nil
	})
}

//specue:req:use-context
func runContextUse(name string) (ContextActionReport, *Problem) {
	return withStore(func(s *context.Store) (ContextActionReport, *Problem) {
		if err := s.Use(name); err != nil {
			var unknown *context.UnknownContextError
			if errors.As(err, &unknown) {
				p := Errorf(fmt.Sprintf("see contexts with `%s`, or create it with `%s`",
					usage(cmdContext, subList), usage(cmdContext, subCreate)), "%v", err)
				return ContextActionReport{}, &p
			}
			p := Errorf("re-run; if it persists, report it", "%v", err)
			return ContextActionReport{}, &p
		}
		// Entering a context is the moment authoring begins, so make sure the
		// editor's cue lsp can resolve the schema. Best-effort, no-op when current.
		bestEffortWarm()
		return ContextActionReport{Message: "now using context " + name, Active: name}, nil
	})
}

//specue:req:read-context
func runContextCurrent() (ContextActionReport, *Problem) {
	repo, p := contextRepo()
	if p != nil {
		return ContextActionReport{}, p
	}
	s, err := repo.Load()
	if err != nil {
		p := Errorf("the registry file may be corrupt — fix or remove it", "cannot read contexts: %v", err)
		return ContextActionReport{}, &p
	}
	e, ok := s.ActiveEntry()
	if !ok {
		p := Errorf(fmt.Sprintf("select one with `%s`", usage(cmdContext, subUse)), "no active context")
		return ContextActionReport{}, &p
	}
	return ContextActionReport{Message: fmt.Sprintf("%s (%d module(s))", e.Name, len(e.Modules)), Active: e.Name}, nil
}

//specue:req:remove-context
func runContextRemove(name string) (ContextActionReport, *Problem) {
	return withStore(func(s *context.Store) (ContextActionReport, *Problem) {
		if !s.Remove(name) {
			p := Errorf(fmt.Sprintf("see contexts with `%s`", usage(cmdContext, subList)), "no context named %q", name)
			return ContextActionReport{}, &p
		}
		return ContextActionReport{Message: "removed context " + name}, nil
	})
}

// --- module add / remove / list ---

// targetContext returns the context a module verb acts on: --workspace if given,
// else the active one. A pointer into the store so mutations persist on Save.
func targetContext(g Globals, s *context.Store) (*context.Entry, *Problem) {
	name := g.Workspace
	if name == "" {
		name = s.Active
	}
	if name == "" {
		p := Errorf(fmt.Sprintf("select one with `%s`, or pass --workspace", usage(cmdContext, subUse)), "no active context")
		return nil, &p
	}
	e, ok := s.EntryPtr(name)
	if !ok {
		p := Errorf(fmt.Sprintf("see contexts with `%s`", usage(cmdContext, subList)), "no context named %q", name)
		return nil, &p
	}
	return e, nil
}

// runModuleAdd adds the module at dir to the target context. The module path is read
// from dir's spec.mod.cue (the manifest is the source of truth); the dir is stored
// absolute. If this is the first module and the context has no plan base yet, the
// dir's current git branch is recorded as it (so plans have a baseline).
//
//specue:req:add-module-to-context
func runModuleAdd(g Globals, dir string) (ContextActionReport, *Problem) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		p := Errorf("pass a module directory", "bad path %q: %v", dir, err)
		return ContextActionReport{}, &p
	}
	mf, err := source.ReadManifest(filepath.Join(abs, source.ManifestFile))
	if err != nil {
		p := Errorf(fmt.Sprintf("point at a module directory (one with spec.mod.cue); scaffold one with `%s`", usage(cmdInit)),
			"no module at %s: %v", abs, err)
		return ContextActionReport{}, &p
	}
	// git-native invariant (MANIFESTO P20): a landscape's modules live in git —
	// plan_base and the scanner read git below, so refuse a non-repo module here.
	if p := requireGitRepo(abs); p != nil {
		return ContextActionReport{}, p
	}
	return withStore(func(s *context.Store) (ContextActionReport, *Problem) {
		e, p := targetContext(g, s)
		if p != nil {
			return ContextActionReport{}, p
		}
		if e.PlanBase == "" {
			e.PlanBase = currentGitBranch(abs)
		}
		e.AddModule(string(mf.Path), abs)
		return ContextActionReport{Message: fmt.Sprintf("added %s to context %s", mf.Path, e.Name)}, nil
	})
}

//specue:req:remove-module-from-context
func runModuleRemove(g Globals, modulePath string) (ContextActionReport, *Problem) {
	return withStore(func(s *context.Store) (ContextActionReport, *Problem) {
		e, p := targetContext(g, s)
		if p != nil {
			return ContextActionReport{}, p
		}
		if !e.RemoveModule(modulePath) {
			p := Errorf(fmt.Sprintf("see the context's modules with `%s`", usage(cmdContext, subModule, subList)),
				"context %s has no module %q", e.Name, modulePath)
			return ContextActionReport{}, &p
		}
		return ContextActionReport{Message: fmt.Sprintf("removed %s from context %s", modulePath, e.Name)}, nil
	})
}

// ModuleListReport lists a context's modules.
type ModuleListReport struct {
	Context string           `json:"context"`
	Modules []moduleRowJSON  `json:"modules"`
}

type moduleRowJSON struct {
	Path string `json:"path"`
	Dir  string `json:"dir"`
}

func runModuleList(g Globals) (ModuleListReport, *Problem) {
	repo, p := contextRepo()
	if p != nil {
		return ModuleListReport{}, p
	}
	s, err := repo.Load()
	if err != nil {
		p := Errorf("the registry file may be corrupt — fix or remove it", "cannot read contexts: %v", err)
		return ModuleListReport{}, &p
	}
	e, p := targetContext(g, &s)
	if p != nil {
		return ModuleListReport{}, p
	}
	rep := ModuleListReport{Context: e.Name}
	for _, m := range e.Modules {
		rep.Modules = append(rep.Modules, moduleRowJSON{Path: m.Path, Dir: m.Dir})
	}
	return rep, nil
}

func (r ModuleListReport) renderHuman(w io.Writer) error {
	if len(r.Modules) == 0 {
		_, err := fmt.Fprintf(w, "context %s has no modules — add one with `%s`\n", r.Context, usage(cmdContext, subModule, subAdd))
		return err
	}
	for _, m := range r.Modules {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", m.Path, m.Dir); err != nil {
			return err
		}
	}
	return nil
}

func (r ModuleListReport) jsonValue() any {
	if r.Modules == nil {
		r.Modules = []moduleRowJSON{}
	}
	return r
}

// currentGitBranch returns the checked-out branch of the repo containing dir, or ""
// if dir is not in a git repo. Used to seed a context's plan base from its first
// module. (branch --show-current is valid even on an unborn branch.)
func currentGitBranch(dir string) string {
	git := plan.NewGit(gitBin())
	root, err := git.RepoRoot(dir)
	if err != nil {
		return ""
	}
	branch, err := git.CurrentBranch(root)
	if err != nil {
		return ""
	}
	return branch
}
