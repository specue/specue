package modules

import (
	"fmt"
	"os"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// SchemaModule materializes the embedded schema module to a temp OS directory and
// returns it as a ResolvedModule. Every authored spec module imports the schema,
// so it must be in the closure's registry; and because CUE requires a real OS dir
// for a dependency, the embedded tree is written to disk. The caller is
// responsible for removing the dir (Cleanup) when done.
type SchemaModule struct {
	ResolvedModule
	dir string
}

// NewSchemaModule writes the schema to a fresh temp dir and returns it ready to
// add to a closure.
func NewSchemaModule() (SchemaModule, error) {
	dir, err := os.MkdirTemp("", "specue-schema-")
	if err != nil {
		return SchemaModule{}, err
	}
	if err := source.MaterializeSchema(dir); err != nil {
		os.RemoveAll(dir)
		return SchemaModule{}, fmt.Errorf("materialize schema: %w", err)
	}
	return SchemaModule{
		ResolvedModule: ResolvedModule{
			Path:    model.ModulePath(source.SchemaModulePath),
			Version: source.SchemaVersion,
			Dir:     dir,
		},
		dir: dir,
	}, nil
}

// Cleanup removes the materialized schema directory.
func (s SchemaModule) Cleanup() error { return os.RemoveAll(s.dir) }
