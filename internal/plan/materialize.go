package plan

import (
	"fmt"
	"os"
)

// Materialized is a module's tree extracted from a git ref into a temp directory.
// CUE requires a real OS directory for a dependency (it needs a recoverable
// absolute path), so reading a module from a branch means archiving its subtree to
// disk — the same approach Go takes for a VCS dependency. Cleanup removes it.
type Materialized struct {
	Dir string
}

// Cleanup removes the materialized directory.
func (m Materialized) Cleanup() error { return os.RemoveAll(m.Dir) }

// Materializer extracts a module's tree from a git ref into a fresh temp dir. It
// is the read-only counterpart to a checkout: the working tree is never touched,
// so a plan's branch can be projected from base without disturbing it.
type Materializer struct {
	git Git
}

// NewMaterializer returns a Materializer driving git.
func NewMaterializer(git Git) Materializer { return Materializer{git: git} }

// Subtree extracts subdir of ref from the repo at root into a new temp dir and
// returns it. The caller owns the directory (Cleanup when done).
func (m Materializer) Subtree(root, ref, subdir string) (Materialized, error) {
	dest, err := os.MkdirTemp("", "specue-ref-")
	if err != nil {
		return Materialized{}, err
	}
	if err := m.git.ArchiveSubtree(root, ref, subdir, dest); err != nil {
		os.RemoveAll(dest)
		return Materialized{}, fmt.Errorf("materialize %s@%s: %w", subdir, ref, err)
	}
	return Materialized{Dir: dest}, nil
}
