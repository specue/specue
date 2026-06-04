package modules

import (
	"fmt"
	"path/filepath"

	"github.com/specue/specue/internal/source"
)

// replaceLocator resolves a require to the local directory named by its Replace
// path (relative to the requiring module's dir). This is the working-tree /
// local-sibling case; a git-checkout or registry-fetch locator implements the
// same interface for plans and released deps later.
type replaceLocator struct{}

// NewReplaceLocator returns a Locator that uses each require's Replace path.
func NewReplaceLocator() Locator { return replaceLocator{} }

func (replaceLocator) Locate(fromDir string, req source.ModuleRequire) (string, error) {
	if req.Replace == "" {
		return "", fmt.Errorf("require %s has no replace path (only local replace is supported)", req.Module)
	}
	if filepath.IsAbs(req.Replace) {
		return filepath.Clean(req.Replace), nil
	}
	return filepath.Join(fromDir, req.Replace), nil
}
