package warm

import (
	"os"
	"path/filepath"
	"runtime"
)

// CacheDir returns the cue module cache root, matching cue's own resolution:
// CUE_CACHE_DIR if set, else the OS-conventional cache location. cue lsp reads
// from here, so the warm step must target the same directory.
func CacheDir() (string, error) {
	if d := os.Getenv("CUE_CACHE_DIR"); d != "" {
		return d, nil
	}
	if runtime.GOOS == "darwin" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Caches", "cue"), nil
	}
	// Linux and the rest: $XDG_CACHE_HOME/cue, else $HOME/.cache/cue.
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "cue"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "cue"), nil
}
