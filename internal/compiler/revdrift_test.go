package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// modWith builds a loaded module with a version and requires (no nodes — rev-drift
// is module-level).
func modWith(path model.ModulePath, version source.Version, reqs ...source.ModuleRequire) source.LoadedModule {
	return source.LoadedModule{Manifest: source.Manifest{
		Path: path, Version: version, Kind: source.KindService, Requires: reqs,
	}}
}

func TestRevDriftPinnedBehindSource(t *testing.T) {
	// consumer pins wallet at v1.0.0; wallet's source is at v1.2.0 → drift gate.
	in := Input{Modules: []source.LoadedModule{
		modWith("consumer", "v0.1.0", source.ModuleRequire{Module: "example", Version: "v1.0.0"}),
		modWith("example", "v1.2.0"),
	}}
	_, diags := New().Compile(in)
	assert.Contains(t, codesOf(diags), RevDrift)
	for _, d := range diags {
		if d.Code == RevDrift {
			assert.Equal(t, Gate, d.Severity(), "rev-drift is a gate")
			assert.Contains(t, d.Message, "v1.0.0")
			assert.Contains(t, d.Message, "v1.2.0")
		}
	}
}

func TestRevDriftPinMatchesSource(t *testing.T) {
	in := Input{Modules: []source.LoadedModule{
		modWith("consumer", "v0.1.0", source.ModuleRequire{Module: "example", Version: "v1.2.0"}),
		modWith("example", "v1.2.0"),
	}}
	_, diags := New().Compile(in)
	assert.NotContains(t, codesOf(diags), RevDrift, "pin == source → no drift")
}

func TestRevDriftPinAheadOfSource(t *testing.T) {
	// A pin newer than the source (source rolled back) is not drift — drift is only
	// the source moving AHEAD of the pin.
	in := Input{Modules: []source.LoadedModule{
		modWith("consumer", "v0.1.0", source.ModuleRequire{Module: "example", Version: "v2.0.0"}),
		modWith("example", "v1.5.0"),
	}}
	_, diags := New().Compile(in)
	assert.NotContains(t, codesOf(diags), RevDrift)
}

func TestRevDriftUnloadedSourceSkipped(t *testing.T) {
	// require to a module not in view → unverifiable, no drift.
	in := Input{Modules: []source.LoadedModule{
		modWith("consumer", "v0.1.0", source.ModuleRequire{Module: "external", Version: "v1.0.0"}),
	}}
	_, diags := New().Compile(in)
	assert.NotContains(t, codesOf(diags), RevDrift)
}

func TestVersionMismatchSpecVsCUEMod(t *testing.T) {
	// spec.mod pins wallet at v1.0.0 but cue.mod deps say v1.1.0 → mismatch gate.
	consumer := modWith("consumer", "v0.1.0", source.ModuleRequire{Module: "example", Version: "v1.0.0"})
	consumer.CUEMod = source.CUEModule{Deps: map[model.ModulePath]source.Version{"example": "v1.1.0"}}
	in := Input{Modules: []source.LoadedModule{consumer, modWith("example", "v1.0.0")}}
	_, diags := New().Compile(in)
	assert.Contains(t, codesOf(diags), VersionMismatch)
}

func TestVersionConsistentNoGate(t *testing.T) {
	consumer := modWith("consumer", "v0.1.0", source.ModuleRequire{Module: "example", Version: "v1.0.0"})
	consumer.CUEMod = source.CUEModule{Deps: map[model.ModulePath]source.Version{"example": "v1.0.0"}}
	in := Input{Modules: []source.LoadedModule{consumer, modWith("example", "v1.0.0")}}
	_, diags := New().Compile(in)
	assert.NotContains(t, codesOf(diags), VersionMismatch, "spec.mod pin == cue.mod dep → consistent")
}

func TestSemverNumericNotLexical(t *testing.T) {
	// v0.10.0 is newer than v0.9.0 (numeric), which a lexical compare would miss.
	assert.Greater(t, cmpSemver("v0.10.0", "v0.9.0"), 0)
	assert.Less(t, cmpSemver("v0.9.0", "v0.10.0"), 0)
	assert.Equal(t, 0, cmpSemver("v1.2.3", "v1.2.3"))
}
