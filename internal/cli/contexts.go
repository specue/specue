package cli

import (
	"os"
	"path/filepath"

	"github.com/specue/specue/internal/context"
)

// contextRepo builds the context registry's repository, resolving WHERE it lives
// here at the application edge: $SPECUE_HOME overrides (keeps tests off the real
// home), else ~/.specue. The domain layer takes the resolved path and never
// reads the environment, so every configuration source stays in one place.
func contextRepo() (context.Repository, *Problem) {
	home := os.Getenv("SPECUE_HOME")
	if home == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			p := Errorf("set SPECUE_HOME to a writable directory", "cannot locate the home directory: %v", err)
			return nil, &p
		}
		home = filepath.Join(h, ".specue")
	}
	return context.NewFileRepository(filepath.Join(home, "contexts.json")), nil
}
