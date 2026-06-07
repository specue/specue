// Package compiler turns authored nodes plus code facts into a resolved, status-
// bearing graph. It is a pure transform: it reads no filesystem (source and scan
// gather the facts) and produces (ResolvedGraph, []Diagnostic). The graph is
// immutable after Compile; identity is (module, slug).
package compiler

import "github.com/specue/specue/internal/model"

// Severity splits the two kinds of finding the redesign locks in (P12): a Gate is
// factual and binary — it turns the graph red (a broken binding, a role
// violation, a sync cycle). An Advisory is a judgement — an FR-coverage gap, a
// migrate hint, a benign async cycle — surfaced for review but never red.
//
// Reference faults (dangling ref, undeclared import, visibility breach, wrong
// target type) are NOT in this set: references are CUE-native, so CUE rejects an
// invalid one as a build error before the compiler ever sees the graph.
type Severity int

const (
	Gate Severity = iota
	Advisory
)

func (s Severity) String() string {
	if s == Gate {
		return "gate"
	}
	return "advisory"
}

// DiagnosticCode is the closed set of diagnostic kinds. Its membership of Gate vs
// Advisory is fixed by gateCodes below, not chosen per-emit, so the split can't
// drift.
type DiagnosticCode string

const (
	RoleGateViolation DiagnosticCode = "role-gate-violation" // node type not allowed by the module's kind
	OrphanBinding     DiagnosticCode = "orphan-binding"      // a code annotation points at no node
	UnbindableTarget  DiagnosticCode = "unbindable-target"   // a req/test annotation on a node type that holds no code (e.g. a Need)
	Unreachable       DiagnosticCode = "unreachable"         // a Contract nothing triggers, satisfies, or invokes
	SyncCycle         DiagnosticCode = "sync-cycle"          // a dependency cycle through a sync contract
	OverlayInvalid    DiagnosticCode = "overlay-invalid"     // an overlaid plan set fails to load/resolve (CUE rejects it)
	// DanglingRef is an edge whose target is empty or names no node in a loaded
	// module. CUE does NOT reject a reference to a missing field on an open struct
	// (it silently yields an incomplete value), so a cross-module ref to a node a
	// plan removed survives load as an empty target — the compiler catches it here.
	DanglingRef DiagnosticCode = "dangling-ref"
	// VersionMismatch is a require pinned to a different version in spec.mod than in
	// cue.mod. Versions live in two files (spec.mod requires drive rev-drift, cue.mod
	// deps drive CUE's import resolution); they must agree, or the graph resolves at
	// one version while the drift check reasons about another.
	VersionMismatch DiagnosticCode = "version-mismatch"

	AsyncCycle    DiagnosticCode = "async-cycle"     // a dependency cycle, all async — tolerable
	RevDrift      DiagnosticCode = "rev-drift"       // a code pin lags the element's rev
	FRCoverageGap DiagnosticCode = "fr-coverage-gap" // a story atom no implemented element discharges
)

// gateCodes is the authoritative gate/advisory membership. Anything absent is an
// advisory.
var gateCodes = map[DiagnosticCode]bool{
	RoleGateViolation: true,
	OrphanBinding:     true,
	UnbindableTarget:  true,
	Unreachable:       true,
	SyncCycle:         true,
	OverlayInvalid:    true,
	DanglingRef:       true,
	RevDrift:          true,
	VersionMismatch:   true,
}

func (c DiagnosticCode) Severity() Severity {
	if gateCodes[c] {
		return Gate
	}
	return Advisory
}

// Location points a diagnostic at a place. Either a code site (File+Line, for a
// binding fault) or a node (for a spec fault); both may be set.
type Location struct {
	File model.FilePath
	Line int
}

// Diagnostic is one finding. Severity is derived from Code so it stays consistent.
type Diagnostic struct {
	Code     DiagnosticCode
	Node     model.NodeID // the owning node; zero value = module/global scope
	Location Location
	Message  string
}

func (d Diagnostic) Severity() Severity { return d.Code.Severity() }

// newDiag builds a diagnostic for a node.
func newDiag(code DiagnosticCode, node model.NodeID, msg string) Diagnostic {
	return Diagnostic{Code: code, Node: node, Message: msg}
}
