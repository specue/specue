package compiler

// Compiler turns parser facts (loaded modules) and code facts (the scanner's
// output) into a resolved graph plus diagnostics. An interface, like Parser /
// Loader / Scanner, so callers depend on the capability. The compilation is a
// pure transform — no filesystem — so it is trivially substitutable in tests.
type Compiler interface {
	Compile(in Input) (*ResolvedGraph, []Diagnostic)
}

type compiler struct{}

// New returns a Compiler.
func New() Compiler { return compiler{} }

// Compile runs the passes in dependency order; the graph is immutable once
// returned.
//
// References arrive already resolved by CUE — and CUE has already rejected any
// that dangle, breach a module's visibility, or use an undeclared import — so
// there is no ref-resolution pass here. The compiler checks only the domain
// constraints CUE cannot: role-gate, status, cycles, blocked, coverage.
//
// Pass order:
//  0. index    — place nodes by identity, record module kind
//  1. rolegate — module kind gates node type
//     …  derive, bind, status, cycles, blocked, coverage follow.
func (compiler) Compile(in Input) (*ResolvedGraph, []Diagnostic) {
	g := buildIndex(in)

	var diags []Diagnostic
	diags = append(diags, checkRoleGate(g)...)
	diags = append(diags, checkDangling(g)...)
	diags = append(diags, checkRevDrift(g)...)
	diags = append(diags, checkVersionConsistency(g)...)

	deriveAll(g)
	diags = append(diags, bindFacts(g, in.Facts)...)

	// UC factual status from the fact collision, then the graph-global passes:
	// cycles (needs only edges), blocked-propagation (reads UC readiness), and
	// finally Need coverage (reads blocked, so it runs last).
	assignContractStatus(g)
	diags = append(diags, detectCycles(g)...)
	propagateBlocked(g)
	assignNeedStatus(g)

	return g, diags
}
