// decision: internal/usecase/checking-path-free
package usecase

import (
	"fmt"
	"os"
)

// CheckPathFree verifies the target folder is empty or does not exist.
// A directory that exists and contains files is refused: it might hold
// something we would overwrite.
//
// decision: internal/usecase/checking-path-free
func CheckPathFree(dir, displayPath string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // does not exist yet — free
		}
		return err
	}
	if len(entries) > 0 {
		// decision: presentation/add-output
		return &ResolveError{
			Code:    "dir_not_empty",
			Message: fmt.Sprintf("%s is not empty", relForMessage(displayPath)),
		}
	}
	return nil
}
