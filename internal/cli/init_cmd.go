package cli

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/specue/specue/internal/source"
)

// newInitCmd wires `specue init <module-path>`: scaffold a new spec module in the
// target directory (-C, or the cwd). It writes spec.mod.cue + cue.mod/module.cue and
// does not resolve an existing landscape — it creates one module. The --kind flag
// sets the module's role (default service).
func newInitCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	var (
		kind   string
		layout string
		name   string
	)
	c := &cobra.Command{
		Use:   "init <dir> <module-path>",
		Short: "Scaffold a new spec module (spec.mod.cue + cue.mod)",
		Long: "init creates a new spec module in <dir> (created if absent): a spec.mod.cue with\n" +
			"the given kind and a cue.mod pinning the shared schema. <module-path> is the\n" +
			"canonical path/name@vMAJOR the manifest declares.\n\n" +
			"  specue init ../governance x.test/governance@v0 --kind governance\n\n" +
			"With --layout " + source.LayoutDir + ", <dir> is the repo root and the module is\n" +
			"placed at <dir>/" + source.LayoutDir + "/<kind>/[<name>/], keeping all Specue\n" +
			"artifacts under one tree. --kind code in that layout writes a code_root that\n" +
			"climbs back to the repo so the scan reaches your source without the code\n" +
			"module claiming sibling spec modules as its own subpackages.",
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			useLayout, p := parseLayout(layout)
			if p != nil {
				_ = g.renderer(out, err).Fail(*p)
				*code = exitUsage
				return nil
			}
			rep, p := runInit(args[0], args[1], kind, name, useLayout)
			if p != nil {
				_ = g.renderer(out, err).Fail(*p)
				*code = exitUsage
				return nil
			}
			return g.renderer(out, err).Report(rep)
		},
	}
	c.Flags().StringVar(&kind, "kind", string(source.KindService),
		"module role: service | domain | governance | topology | code")
	c.Flags().StringVar(&layout, "layout", "",
		"place the module under <dir>/"+source.LayoutDir+"/<kind>/[<name>/] (use \""+source.LayoutDir+"\")")
	c.Flags().StringVar(&name, "name", "",
		"subfolder name inside "+source.LayoutDir+"/<kind>/ — for repos with several modules of the same kind (only with --layout)")
	return c
}

// parseLayout normalises the --layout flag: empty = flat (the historical
// behaviour); the single accepted value is the LayoutDir constant. Anything
// else is rejected here so a typo (`--layout spcd`) does not silently fall
// back to flat.
func parseLayout(v string) (bool, *Problem) {
	switch v {
	case "":
		return false, nil
	case source.LayoutDir:
		return true, nil
	default:
		p := Errorf("use --layout "+source.LayoutDir+" or drop the flag", "unknown layout %q", v)
		return false, &p
	}
}
