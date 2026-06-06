package warm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/specue/specue/internal/source"
)

// CueResolve is the production ResolveFunc: it warms the cache by resolving the
// schema through a throwaway consumer module that imports it, running `cue vet`
// with CUE_REGISTRY pointed at the ephemeral registry. The vet materializes the
// schema's extract in the cache — the whole point of the warm step. The consumer
// is discarded; only the populated cache survives.
//
// cueBin is the cue executable to invoke (so a caller can override the path);
// empty means "cue" on PATH.
func CueResolve(cueBin string) ResolveFunc {
	if cueBin == "" {
		cueBin = "cue"
	}
	return func(registryAddr string) error {
		consumer, err := os.MkdirTemp("", "specue-warm-resolve-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(consumer)
		if err := writeConsumer(consumer); err != nil {
			return err
		}
		cmd := exec.Command(cueBin, "vet", ".")
		cmd.Dir = consumer
		cmd.Env = append(os.Environ(), "CUE_REGISTRY="+registryAddr+"+insecure")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cue vet: %v: %s", err, out)
		}
		return nil
	}
}

// CueResolveClosure is the production ClosureResolveFunc: it warms the cache for
// the whole landscape by running `cue vet ./...` in every root directory with
// CUE_REGISTRY pointed at the ephemeral registry. Each root's vet materializes
// the extracts of every module it transitively imports, so collectively they
// cover the closure. A root with no resolvable package (e.g. the code module
// before any spec lives there) is tolerated — its non-zero exit is reported, but
// later roots still run.
//
// cueBin is the cue executable to invoke; empty means "cue" on PATH.
func CueResolveClosure(cueBin string) ClosureResolveFunc {
	if cueBin == "" {
		cueBin = "cue"
	}
	return func(registryAddr string, rootDirs []string) error {
		for _, dir := range rootDirs {
			cmd := exec.Command(cueBin, "vet", "./...")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "CUE_REGISTRY="+registryAddr+"+insecure")
			if out, err := cmd.CombinedOutput(); err != nil {
				// A root may have no CUE files at all (a code module before any spec).
				// That is not a warm error — it just has nothing to fetch. cue prints
				// this as "no CUE files" / "matched no packages"; either way the cache
				// of the other roots is what we care about.
				if isNoFilesOutput(out) {
					continue
				}
				return fmt.Errorf("cue vet in %s: %v: %s", dir, err, out)
			}
		}
		return nil
	}
}

// isNoFilesOutput reports whether cue's output is the benign "no package files"
// case — happens on a root that holds only a manifest, no nodes yet.
func isNoFilesOutput(out []byte) bool {
	s := string(out)
	return strings.Contains(s, "no CUE files") || strings.Contains(s, "matched no packages")
}

// writeConsumer lays down a minimal module that pins and imports the schema, so
// `cue vet` has a reason to fetch it. Importing one schema definition is enough to
// force resolution of the whole module into the cache.
func writeConsumer(dir string) error {
	if err := os.MkdirAll(filepath.Join(dir, "cue.mod"), 0o755); err != nil {
		return err
	}
	mod := fmt.Sprintf("module: %q\nlanguage: version: %q\ndeps: %q: v: %q\n",
		"specue.io/warm-consumer@v0", "v0.16.1", source.SchemaModulePath, source.SchemaVersion)
	if err := os.WriteFile(filepath.Join(dir, "cue.mod", "module.cue"), []byte(mod), 0o644); err != nil {
		return err
	}
	use := fmt.Sprintf("package warmconsumer\nimport s %q\n_warm: s.#Contract\n", source.SchemaModulePath+":spec")
	return os.WriteFile(filepath.Join(dir, "use.cue"), []byte(use), 0o644)
}
