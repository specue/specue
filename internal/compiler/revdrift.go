package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// checkRevDrift flags a require whose pinned version has fallen behind the version
// the resolved source module actually declares. The pin is the consumer's claim
// "I am built against vX"; if the source has since moved to a later vY, the
// consumer may be relying on a contract that changed underneath it — a gate, with
// the freshness (pinned vs current) in the message. A require to a module not in
// view (no resolved source here) is skipped — its version is unverifiable.
func checkRevDrift(g *ResolvedGraph) []Diagnostic {
	var diags []Diagnostic
	for path, info := range g.mods {
		for _, req := range info.Requires {
			src, ok := g.mods[req.Module]
			if !ok {
				continue // source not loaded — can't compare
			}
			if cmpSemver(src.Version, req.Version) > 0 {
				diags = append(diags, newDiag(RevDrift, model.NodeID{Module: path},
					fmt.Sprintf("%s pins %s at %s, but the source is at %s — re-pin to pick up the changed contract",
						path, req.Module, req.Version, src.Version)))
			}
		}
	}
	return diags
}

// checkVersionConsistency flags a require whose spec.mod pin disagrees with the
// same dependency's version in cue.mod. Versions live in two files: spec.mod
// requires (which drive rev-drift) and cue.mod deps (which drive CUE's import
// resolution). If they disagree, CUE resolves the graph at one version while the
// drift check reasons about another — a silent split. This gate makes it loud. A
// dep absent from cue.mod is skipped (CUE would have failed to resolve it anyway).
func checkVersionConsistency(g *ResolvedGraph) []Diagnostic {
	var diags []Diagnostic
	for path, info := range g.mods {
		for _, req := range info.Requires {
			cueVer, ok := info.CUEMod.Deps[req.Module]
			if !ok {
				continue
			}
			if cueVer != req.Version {
				diags = append(diags, newDiag(VersionMismatch, model.NodeID{Module: path},
					fmt.Sprintf("%s pins %s at %s in spec.mod but %s in cue.mod — the two must agree",
						path, req.Module, req.Version, cueVer)))
			}
		}
	}
	return diags
}

// cmpSemver compares two vMAJOR.MINOR.PATCH versions numerically (so v0.10.0 >
// v0.9.0, unlike a lexical compare). Returns >0 if a is newer, 0 if equal, <0 if a
// is older. A malformed version sorts as zero, so a parse fault never falsely
// trips drift.
func cmpSemver(a, b source.Version) int {
	am, an, ap := parseSemver(a)
	bm, bn, bp := parseSemver(b)
	switch {
	case am != bm:
		return am - bm
	case an != bn:
		return an - bn
	default:
		return ap - bp
	}
}

func parseSemver(v source.Version) (major, minor, patch int) {
	s := strings.TrimPrefix(string(v), "v")
	parts := strings.SplitN(s, ".", 3)
	get := func(i int) int {
		if i >= len(parts) {
			return 0
		}
		n, _ := strconv.Atoi(parts[i])
		return n
	}
	return get(0), get(1), get(2)
}
