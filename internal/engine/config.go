// Package engine is the resident graph: it derives the spec graph once and
// re-derives only when its inputs change, keyed on a content hash of the two
// inputs (spec + code). v1 had no resident engine — every request rebuilt the
// whole graph from disk (~5s handles); this is the fix. The graph is immutable
// after a build, so concurrent readers are lock-free, and rebuilds are
// single-flighted so N concurrent callers wait on one build.
//
// The engine orchestrates the three layers (source.Loader → codescan.Scanner →
// compiler.Compiler) but takes its config explicitly — which module dirs to load
// and which code trees to scan. Assembling that config from spec.work / spec.code
// is a separate (future) concern, so this layer stays a pure orchestrator and the
// workspace resolver can land later without touching it.
package engine

import (
	"github.com/specue/specue/internal/codescan"
	"github.com/specue/specue/internal/source"
)

// Config is the engine's two inputs: the workspace (the landscape entry point) and
// the code scan targets.
//
// The spec side is a workspace — the explicit list of every landscape module and
// where it lives. The engine resolves the whole set from it, so the CUE registry
// sees the full graph (cross-module references resolve while authoring). It arrives
// either as a spec.work.cue on disk (WorkFile) or already built in memory
// (Workspace) — a single-module landscape the caller synthesized, with no file to
// write. Workspace wins when both are set. Module dirs must be real OS directories:
// CUE resolves the module set itself and a dependency needs a recoverable absolute
// path (see the modules layer). The code side stays fs.FS — scanning is content-
// based and needs no CUE module resolution.
type Config struct {
	// WorkFile is the path to the spec.work.cue that lists the landscape. Used when
	// Workspace is nil.
	WorkFile string
	// Workspace, when non-nil, is the landscape supplied directly — no file is read.
	// Its Root must be an absolute directory (module dirs resolve against it). This
	// is how the CLI runs on a single module without writing a temp spec.work.
	Workspace *source.Workspace
	// ScanTargets are the code trees to scan; each is self-contained (its own FS,
	// kind, and module attribution).
	ScanTargets []codescan.ScanTarget
}
