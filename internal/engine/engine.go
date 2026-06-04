package engine

import (
	"fmt"
	"io"
	"sync"

	"golang.org/x/sync/singleflight"

	"github.com/specue/specue/internal/codescan"
	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/modules"
	"github.com/specue/specue/internal/source"
	"github.com/specue/specue/internal/specload"
)

// Result is one build's output: the immutable graph and the diagnostics produced
// alongside it. Cached together — a build yields both, and a reader needs both.
type Result struct {
	Graph *compiler.ResolvedGraph
	Diags []compiler.Diagnostic
}

// Engine derives the spec graph and serves it from memory, re-deriving only when
// its inputs change. Live returns the current graph; the first call (or a call
// after an input change) builds, the rest hit the cache.
type Engine interface {
	Live() (Result, error)
	// Close releases lifetime resources (the materialized schema directory).
	Close() error
}

type engine struct {
	cfg       Config
	parser    source.Parser
	resolver  modules.Resolver
	loader    specload.Loader
	schema    modules.SchemaModule // materialized once; in the closure of every build
	scanner   codescan.Scanner
	compiler  compiler.Compiler
	specKeyer SourceKeyer
	codeKeyer SourceKeyer

	mu     sync.RWMutex
	cur    Result
	curKey inputKey
	has    bool

	flight singleflight.Group
}

// New builds an Engine for cfg. By default it wires the standard layers and
// content-hash keyers (correct over any fs.FS); options swap in stat keyers (for
// a live filesystem) or test doubles.
func New(cfg Config, opts ...Option) (Engine, error) {
	parser, err := source.NewCUEParser()
	if err != nil {
		return nil, fmt.Errorf("engine: %w", err)
	}
	schema, err := modules.NewSchemaModule()
	if err != nil {
		return nil, fmt.Errorf("engine: %w", err)
	}
	e := &engine{
		cfg:      cfg,
		parser:   parser,
		resolver: modules.NewResolver(parser, modules.NewReplaceLocator()),
		loader:   specload.New(),
		schema:   schema,
		scanner:  codescan.NewScanner(),
		compiler: compiler.New(),
	}
	e.specKeyer, e.codeKeyer = newContentKeyers(cfg)
	for _, opt := range opts {
		opt(e)
	}
	return e, nil
}

// Close releases resources held for the engine's lifetime (the materialized
// schema directory).
func (e *engine) Close() error { return e.schema.Cleanup() }

// Option customizes an Engine.
type Option func(*engine)

// WithStatKeyers swaps in the cheap stat-based keyers — use on a real filesystem
// with monotonic mtime (the live tree), not an in-memory FS.
func WithStatKeyers() Option {
	return func(e *engine) { e.specKeyer, e.codeKeyer = newStatKeyers(e.cfg) }
}

// WithLoadDebug routes specload's per-module trace (what CUE returned: instances,
// files, errors) to w. Wired from the CLI's --debug flag for one-off forensics
// when a module resolves to something unexpected.
func WithLoadDebug(w io.Writer) Option {
	return func(e *engine) { e.loader = specload.New(specload.WithDebug(w)) }
}

// Live returns the current graph, building only if the inputs changed since the
// last build. A failed rebuild returns the error and leaves the last good build
// in place (it is not clobbered by a failure).
//specue:req:build-graph#incremental
func (e *engine) Live() (Result, error) {
	k, err := e.liveKey()
	if err != nil {
		return Result{}, err
	}
	e.mu.RLock()
	if e.has && e.curKey == k {
		r := e.cur
		e.mu.RUnlock()
		return r, nil
	}
	e.mu.RUnlock()
	return e.rebuild(k)
}

// liveKey computes the current content key from both inputs (outside the lock).
func (e *engine) liveKey() (inputKey, error) {
	spec, err := e.specKeyer.Key()
	if err != nil {
		return inputKey{}, err
	}
	code, err := e.codeKeyer.Key()
	if err != nil {
		return inputKey{}, err
	}
	return inputKey{spec: spec, code: code}, nil
}
