// Command v1tov2 is an ephemeral migrator that translates a set of v1 spec
// modules (YAML nodes + spec.mod) into v2 CUE modules. It is intentionally NOT
// wired into the main specue CLI: v1 is dead and v2 stands on its own, so
// the migrator lives beside the tool only long enough to land the existing
// gameprovider landscape, then can be deleted along with internal/migrate.
//
// Usage:
//
//	v1tov2 -out <dir> <v1-module-dir>...
//
// Each v1-module-dir must hold a v1 `spec.mod` and the node YAMLs that go
// with it. The whole set is migrated as one batch so cross-module refs (an
// alias from one module's `require.use` pointing at a node in another) resolve
// to the right cue-native import. Dropped references (dangling refs, undeclared
// aliases) are reported but do not fail the run — that is v1 debt the v2 graph
// could not accept anyway.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueyaml "cuelang.org/go/encoding/yaml"
	"github.com/specue/specue/internal/migrate"
)

func main() {
	out := flag.String("out", "", "output root directory (each module is written under <out>/<module-path>)")
	initGit := flag.Bool("init-git", false, "after writing each module, `git init` it and commit the migration as one snapshot (v2 is git-native; the loader refuses a module not in a repo)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: v1tov2 -out <dir> [-init-git] <v1-module-dir>...\n\n")
		fmt.Fprintf(os.Stderr, "Each <v1-module-dir> must contain a v1 spec.mod plus *.yaml nodes.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *out == "" || flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	mods, err := resolveModules(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	written, report, err := migrate.Migrate(mods, *out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	// Stable, human-readable summary.
	paths := make([]string, 0, len(written))
	for p := range written {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	fmt.Printf("✓ migrated %d module(s) into %s\n", len(written), *out)
	for _, p := range paths {
		fmt.Printf("  %s → %s\n", p, written[p])
		if *initGit {
			if err := gitInitCommit(written[p]); err != nil {
				fmt.Fprintf(os.Stderr, "  warn: git init %s: %v\n", written[p], err)
			}
		}
	}
	if len(report.Skipped) > 0 {
		fmt.Printf("\n%s", report.String())
		fmt.Fprintln(os.Stderr, "NB: skipped references are v1 debt — the source ref pointed at nothing the migration set defines.")
	}
}

// resolveModules reads each v1 module dir's spec.mod (just enough to grab the
// `module:` line) and pairs it with its directory, so the migrator can resolve
// cross-module references in one pass.
func resolveModules(dirs []string) ([]migrate.Module, error) {
	out := make([]migrate.Module, 0, len(dirs))
	for _, dir := range dirs {
		modPath, err := readModulePath(filepath.Join(dir, "spec.mod"))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", dir, err)
		}
		abs, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		out = append(out, migrate.Module{Path: modPath, Dir: abs})
	}
	return out, nil
}

// gitInitCommit makes the migrated module dir a tracked repo so the v2 loader
// will read it. Snapshots the migration as one commit on `main`.
func gitInitCommit(dir string) error {
	cmds := [][]string{
		{"git", "init", "-q", "-b", "main"},
		{"git", "add", "-A"},
		{"git", "-c", "user.name=v1tov2", "-c", "user.email=v1tov2@local", "commit", "-q", "-m", "migrated from v1"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %w (%s)", c[0], err, out)
		}
	}
	return nil
}

// readModulePath decodes just the `module:` field from a v1 spec.mod YAML.
// The migrator does its own full manifest decode internally; we only need the
// path here to label the input set.
func readModulePath(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	f, err := cueyaml.Extract(path, raw)
	if err != nil {
		return "", err
	}
	val := cuecontext.New().BuildFile(f)
	if err := val.Err(); err != nil {
		return "", err
	}
	mod, err := val.LookupPath(cue.ParsePath("module")).String()
	if err != nil {
		return "", fmt.Errorf("spec.mod has no `module:` field")
	}
	return mod, nil
}
