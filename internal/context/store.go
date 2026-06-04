// Package context is the registry of named workspaces — the spec graph's
// equivalent of kubectl contexts. A context binds a human name to a spec.work.cue
// path; one is active at a time, and the active one is the landscape that
// workspace-mode commands resolve against, so work is not tied to the current
// directory or to discovering a spec.work up the tree.
//
// This is a domain layer: it knows nothing of the CLI, cobra, or rendering. The
// Store here is pure data plus its invariants (no I/O); persistence lives behind a
// Repository so the model is testable without a disk and the storage format can
// change independently.
//
// Note the deliberate name overlap: an Entry points at a workspace, where a
// source.Workspace is the parsed CONTENT of one spec.work. This package registers
// and selects among such workspaces; it does not parse them.
package context

import "sort"

// ModuleRef is one module in a context's landscape: its canonical module path and
// the absolute directory it lives in. Paths are absolute — the registry is a local
// machine config (like kubectl's), not a portable artifact, so there is no file
// "home" to resolve relative dirs against.
type ModuleRef struct {
	Path string `json:"path"`
	Dir  string `json:"dir"`
}

// Entry is one named context: a self-contained landscape. The context manager owns
// this — there is no user-managed spec.work.cue file. PlanBase is the git branch
// plans fork from and merge back to; Modules is the landscape's module set.
type Entry struct {
	Name     string      `json:"name"`
	PlanBase string      `json:"plan_base,omitempty"`
	Modules  []ModuleRef `json:"modules,omitempty"`
}

// AddModule inserts or updates a module by path, keeping the list sorted. Reports
// whether it was newly added (false = updated an existing dir).
func (e *Entry) AddModule(path, dir string) bool {
	for i := range e.Modules {
		if e.Modules[i].Path == path {
			e.Modules[i].Dir = dir
			return false
		}
	}
	e.Modules = append(e.Modules, ModuleRef{Path: path, Dir: dir})
	sort.Slice(e.Modules, func(i, j int) bool { return e.Modules[i].Path < e.Modules[j].Path })
	return true
}

// RemoveModule drops a module by path; reports whether one was removed.
func (e *Entry) RemoveModule(path string) bool {
	for i, m := range e.Modules {
		if m.Path == path {
			e.Modules = append(e.Modules[:i], e.Modules[i+1:]...)
			return true
		}
	}
	return false
}

// Store is the registry: every known context plus which one is active. It is a
// value type with no I/O — the invariants (active never dangles, names unique and
// sorted) are enforced by its methods.
type Store struct {
	Active   string  `json:"active"`
	Contexts []Entry `json:"contexts"`
}

// Find returns the context with the given name, or false.
func (s Store) Find(name string) (Entry, bool) {
	for _, c := range s.Contexts {
		if c.Name == name {
			return c, true
		}
	}
	return Entry{}, false
}

// ActiveEntry returns the active context, or false if none is active.
func (s Store) ActiveEntry() (Entry, bool) {
	if s.Active == "" {
		return Entry{}, false
	}
	return s.Find(s.Active)
}

// Create adds a new empty context, keeping the list sorted. It errors if the name
// is taken so create never silently clobbers an existing landscape; it does not
// change the active selection (that is Use's job).
//specue:req:create-context#name-is-unique
func (s *Store) Create(name string) error {
	if _, ok := s.Find(name); ok {
		return &DuplicateContextError{Name: name}
	}
	s.Contexts = append(s.Contexts, Entry{Name: name})
	sort.Slice(s.Contexts, func(i, j int) bool { return s.Contexts[i].Name < s.Contexts[j].Name })
	return nil
}

// EntryPtr returns a mutable pointer to the named context, or false. Mutating it
// (e.g. AddModule) and then Save'ing the Store persists the change.
func (s *Store) EntryPtr(name string) (*Entry, bool) {
	for i := range s.Contexts {
		if s.Contexts[i].Name == name {
			return &s.Contexts[i], true
		}
	}
	return nil, false
}

// Use makes name the active context, erroring if no such context exists so the
// active selection can never dangle.
//specue:req:use-context#context-must-exist
func (s *Store) Use(name string) error {
	if _, ok := s.Find(name); !ok {
		return &UnknownContextError{Name: name}
	}
	s.Active = name
	return nil
}

// Remove deletes a context by name, clearing the active selection if it pointed
// there. Reports whether a context was removed.
func (s *Store) Remove(name string) bool {
	for i, c := range s.Contexts {
		if c.Name == name {
			s.Contexts = append(s.Contexts[:i], s.Contexts[i+1:]...)
			if s.Active == name {
				s.Active = ""
			}
			return true
		}
	}
	return false
}

// UnknownContextError reports a reference to a context that is not in the registry.
// A typed error so the cli layer can attach the right fix (run `context list`).
type UnknownContextError struct{ Name string }

func (e *UnknownContextError) Error() string {
	return "no context named " + e.Name
}

// DuplicateContextError reports a create of a name that already exists.
type DuplicateContextError struct{ Name string }

func (e *DuplicateContextError) Error() string {
	return "context " + e.Name + " already exists"
}
