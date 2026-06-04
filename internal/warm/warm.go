// Package warm keeps the editor's stock `cue lsp` able to resolve the specue
// schema. The schema lives as an embedded CUE module that authored spec files
// import; cue lsp resolves it through the on-disk module cache, which is
// consulted before CUE_REGISTRY. So the tool only has to seed that cache once:
// publish the schema into an ephemeral in-memory OCI registry and run one resolve
// to materialize the extract. After that the editor needs neither our process nor
// CUE_REGISTRY — it reads the cache natively. We never hold a daemon.
//
// The schema version is fixed (source.SchemaVersion) because every module pins it
// in cue.mod deps; a content change is republished under the same version. A
// content key (hash of the schema bytes) tells us whether the cached extract is
// stale and a re-warm is due.
package warm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cuelabs.dev/go/oci/ociregistry/ocimem"
	"cuelabs.dev/go/oci/ociregistry/ociserver"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/mod/modregistry"
	"cuelang.org/go/mod/modzip"
	"cuelang.org/go/mod/module"

	"github.com/specue/specue/internal/source"
)

// Warmer seeds and refreshes the schema in the cue module cache. cueRun resolves
// CUE against a live registry (the warm step) — injected so the edge wires the
// real `cue` invocation and tests can stub it. cacheDir is the cue cache root.
type Warmer struct {
	cacheDir string
	cueRun   ResolveFunc
}

// ResolveFunc performs one CUE resolve that fetches the schema from registryAddr,
// materializing its extract in the cache. It returns the resolve error, if any.
// The edge implements this by invoking `cue` with CUE_REGISTRY pointed at addr.
type ResolveFunc func(registryAddr string) error

// New builds a Warmer. cacheDir defaults to the cue cache location when empty.
func New(cacheDir string, resolve ResolveFunc) (*Warmer, error) {
	if cacheDir == "" {
		dir, err := CacheDir()
		if err != nil {
			return nil, err
		}
		cacheDir = dir
	}
	return &Warmer{cacheDir: cacheDir, cueRun: resolve}, nil
}

// EnsureWarm makes the schema resolvable from the cache, doing nothing when the
// cache already holds the current schema content. It returns whether a (re)warm
// actually ran, so callers can report it.
//specue:req:warm-schema#no-op-when-current
func (w *Warmer) EnsureWarm() (rewarmed bool, err error) {
	key, err := source.SchemaContentKey()
	if err != nil {
		return false, fmt.Errorf("schema content key: %w", err)
	}
	if w.fresh(key) {
		return false, nil
	}
	if err := w.warm(key); err != nil {
		return false, err
	}
	return true, nil
}

// fresh reports whether the extract for the current schema content is present:
// the extract dir exists AND our recorded content key matches. A version-only
// check is not enough — the version is fixed, so content can change underneath it.
func (w *Warmer) fresh(key string) bool {
	if _, err := os.Stat(w.extractDir()); err != nil {
		return false
	}
	got, err := os.ReadFile(w.keyStamp())
	return err == nil && string(got) == key
}

// warm clears any stale extract, publishes the schema into an ephemeral in-memory
// registry, runs one resolve to materialize the extract, then records the content
// key. The registry lives only for the duration of this call.
//specue:req:warm-schema#registry-is-ephemeral
func (w *Warmer) warm(key string) error {
	if err := w.clearExtract(); err != nil {
		return fmt.Errorf("clear stale extract: %w", err)
	}

	schema, err := source.NewSchemaDir()
	if err != nil {
		return fmt.Errorf("materialize schema: %w", err)
	}
	defer os.RemoveAll(schema)

	srv, addr, stop, err := startRegistry()
	if err != nil {
		return err
	}
	defer stop()

	if err := publishSchema(srv, schema); err != nil {
		return fmt.Errorf("publish schema: %w", err)
	}
	if err := w.cueRun(addr); err != nil {
		return fmt.Errorf("warm resolve: %w", err)
	}
	if err := os.WriteFile(w.keyStamp(), []byte(key), 0o644); err != nil {
		return fmt.Errorf("record content key: %w", err)
	}
	return nil
}

// clearExtract removes the schema's extract dir. The cue cache makes extract dirs
// read-only, so a plain RemoveAll fails silently-ish — chmod the tree writable
// first. Absence is not an error.
func (w *Warmer) clearExtract() error {
	dir := w.extractDir()
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err := chmodTreeWritable(dir); err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

// extractDir is where cue extracts the schema module:
// <cache>/mod/extract/<base-path>@<version>, where base-path is the module path
// without its @vN major suffix (cue keys the extract by base path + full version).
func (w *Warmer) extractDir() string {
	base := source.SchemaModulePath
	if prefix, _, ok := ast.SplitPackageVersion(base); ok {
		base = prefix
	}
	return filepath.Join(w.cacheDir, "mod", "extract", base+"@"+source.SchemaVersion)
}

// keyStamp records the content key of the last successful warm, alongside the
// extract, so a stale extract is detected even though the version is fixed.
func (w *Warmer) keyStamp() string {
	return filepath.Join(w.cacheDir, "mod", "extract", schemaStampName())
}

func schemaStampName() string {
	return ".specue-warm-" + source.SchemaVersion
}

// chmodTreeWritable makes dir and everything under it user-writable so RemoveAll
// can delete cue's read-only extract.
func chmodTreeWritable(dir string) error {
	return filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chmod(p, info.Mode()|0o200)
	})
}

// startRegistry brings up an in-memory OCI registry on a free localhost port and
// returns its address and a stop func. This is the same machinery `cue mod
// registry` uses, driven from the library directly — no dev-only CLI dependency.
func startRegistry() (reg *modregistry.Client, addr string, stop func(), err error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, "", nil, fmt.Errorf("listen: %w", err)
	}
	mem := ocimem.New()
	httpSrv := &http.Server{Handler: ociserver.New(mem, nil)}
	go func() { _ = httpSrv.Serve(l) }()
	stop = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(ctx)
	}
	return modregistry.NewClient(mem), l.Addr().String(), stop, nil
}

// publishSchema zips the materialized schema module and puts it into the registry
// under the fixed version, mirroring `cue mod publish` for a self-sourced module.
func publishSchema(reg *modregistry.Client, schemaDir string) error {
	mv, err := module.NewVersion(source.SchemaModulePath, source.SchemaVersion)
	if err != nil {
		return fmt.Errorf("module version: %w", err)
	}
	zf, err := os.CreateTemp("", "specue-schema-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(zf.Name())
	defer zf.Close()

	if err := modzip.CreateFromDir(zf, mv, schemaDir); err != nil {
		return fmt.Errorf("zip schema: %w", err)
	}
	info, err := zf.Stat()
	if err != nil {
		return err
	}
	if _, err := zf.Seek(0, 0); err != nil {
		return err
	}
	return reg.PutModule(context.Background(), mv, zf, info.Size())
}
