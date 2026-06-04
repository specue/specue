package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// BindingsReport is the typed result of `bindings`: a code module's view of the
// contracts it may implement and their binding state. One row per bindable element,
// uniform for a human table and a JSON array.
type BindingsReport struct {
	Module string        `json:"module"`
	Rows   []bindingRow  `json:"rows"`
}

type bindingRow struct {
	Target    string   `json:"target"`
	Element   string   `json:"element,omitempty"`
	Kind      string   `json:"kind"`
	State     string   `json:"state"`
	Locations []string `json:"locations,omitempty"` // file:line
}

// runBindings computes the bindings view for a code module. The module is the arg
// if given, else the single code module the run resolves to (module mode), else the
// single code module in the landscape; ambiguity or none is an actionable error.
// --state and --kind filter the rows (repeatable, comma-joined).
//
//specue:req:report-bindings
func runBindings(ctx Context, moduleArg string, states, kinds []string) (BindingsReport, *Problem) {
	res, p := buildGraph(ctx)
	if p != nil {
		return BindingsReport{}, p
	}

	mod, p := ctx.resolveCodeModule(moduleArg)
	if p != nil {
		return BindingsReport{}, p
	}

	view, ok := res.Graph.BindingsFor(mod, res.Diags)
	if !ok {
		p := Errorf(fmt.Sprintf("point at a code module (kind: code); see modules with `%s`", usage(cmdContext, subModule, subList)),
			"%s is not a code module — bindings is a code module's view of what it implements", mod)
		return BindingsReport{}, &p
	}

	want, p := wantedStates(states)
	if p != nil {
		return BindingsReport{}, p
	}
	wantKind := wantedKinds(kinds)

	rep := BindingsReport{Module: string(view.Module)}
	for _, r := range view.Rows {
		if want != nil && !want[r.State] {
			continue
		}
		if wantKind != nil && !wantKind[r.Kind] {
			continue
		}
		rep.Rows = append(rep.Rows, bindingRow{
			Target:    r.Target.String(),
			Element:   string(r.Element),
			Kind:      string(r.Kind),
			State:     string(r.State),
			Locations: locationsOf(r.Locations),
		})
	}
	return rep, nil
}

// resolveCodeModule picks the code module the view is for: the explicit arg, the
// resolved single module if it is code, or the landscape's sole code module.
func (c Context) resolveCodeModule(arg string) (model.ModulePath, *Problem) {
	if arg != "" {
		return model.ModulePath(arg), nil
	}
	// A code module implements OTHER modules' contracts, so its bindings are only
	// meaningful when those required modules are loaded — i.e. in a workspace. An
	// isolated module-mode run can't see them (every annotation would read as
	// orphan), so refuse it with the workspace remedy rather than emit noise.
	if c.mode == modeModule {
		p := Errorf(fmt.Sprintf("a code module's bindings need its required contracts loaded — add it to a workspace and `%s`",
			cmdPath(cmdContext, subUse)),
			"bindings needs a workspace, not an isolated module (the required contracts aren't loaded)")
		return "", &p
	}
	// Otherwise the landscape's single code module, if unambiguous.
	var code []model.ModulePath
	_, dirs, p := c.workspace()
	if p != nil {
		return "", p
	}
	for path, dir := range dirs {
		mf, err := source.ReadManifest(dir + "/" + source.ManifestFile)
		if err == nil && mf.Kind == source.KindCode {
			code = append(code, path)
		}
	}
	switch len(code) {
	case 1:
		return code[0], nil
	case 0:
		p := Errorf(fmt.Sprintf("scaffold one with `%s --kind code`, or pass a module path", usage(cmdInit)),
			"no code module in this landscape")
		return "", &p
	default:
		sort.Slice(code, func(i, j int) bool { return code[i] < code[j] })
		p := Errorf(fmt.Sprintf("name one: `%s <module>` (e.g. %s)", usage(cmdBindings), code[0]),
			"%d code modules — bindings is per code module", len(code))
		return "", &p
	}
}

// wantedStates turns repeatable --state flags (comma-joined accepted) into a set,
// validating each against the known states so a typo is an actionable error rather
// than an always-empty result.
func wantedStates(states []string) (map[compiler.BindState]bool, *Problem) {
	if len(states) == 0 {
		return nil, nil
	}
	valid := map[compiler.BindState]bool{
		compiler.BindUnbound: true, compiler.BindBound: true, compiler.BindProven: true,
		compiler.BindDuplicate: true, compiler.BindOrphan: true,
	}
	want := map[compiler.BindState]bool{}
	for _, s := range states {
		for _, part := range strings.Split(s, ",") {
			st := compiler.BindState(strings.TrimSpace(part))
			if !valid[st] {
				p := Errorf("use one of: unbound, bound, proven, duplicate, orphan",
					"unknown --state %q", part)
				return nil, &p
			}
			want[st] = true
		}
	}
	return want, nil
}

// wantedKinds turns repeatable --kind flags into a set, validating each against the
// known kinds (req + the infra roles) so a typo is actionable, not an empty result.
func wantedKinds(kinds []string) map[compiler.BindKind]bool {
	if len(kinds) == 0 {
		return nil
	}
	want := map[compiler.BindKind]bool{}
	for _, k := range kinds {
		for _, part := range strings.Split(k, ",") {
			want[compiler.BindKind(strings.TrimSpace(part))] = true
		}
	}
	return want
}

func locationsOf(bs []compiler.Binding) []string {
	var out []string
	for _, b := range bs {
		out = append(out, fmt.Sprintf("%s:%d", b.File, b.Line))
	}
	return out
}

func (r BindingsReport) renderHuman(w io.Writer) error {
	if len(r.Rows) == 0 {
		_, err := fmt.Fprintf(w, "no bindings for %s\n", r.Module)
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "STATE\tKIND\tCONTRACT\tWHERE"); err != nil {
		return err
	}
	for _, row := range r.Rows {
		target := row.Target
		if row.Element != "" {
			target += "#" + row.Element
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", row.State, row.Kind, target, strings.Join(row.Locations, " ")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (r BindingsReport) jsonValue() any {
	if r.Rows == nil {
		r.Rows = []bindingRow{}
	}
	return r
}
