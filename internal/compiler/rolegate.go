package compiler

import (
	"fmt"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/source"
)

// checkRoleGate enforces that a module's kind permits each node's type (P17, the
// intent axis: who owns a node decides its home). A violation is a gate and marks
// the node broken.
//
//specue:req:validate-graph#role-gate
func checkRoleGate(g *ResolvedGraph) []Diagnostic {
	var diags []Diagnostic
	for n := range g.Nodes() {
		info, ok := g.mods[n.ID().Module]
		if !ok {
			continue
		}
		if !kindAllows(info.Kind, n.Node().Type) {
			n.Status = StatusBroken
			diags = append(diags, newDiag(RoleGateViolation, n.ID(), roleGateMessage(info.Kind, n.Node().Type)))
		}
	}
	return diags
}

// kindAllows reports whether a module of the given kind may hold a node of the
// given type. service → Contract/Port/Container; domain → Domain/Need;
// governance → Plan/ADR; topology → Port/Container (the cross-service seam: shared
// channels/datastores and the broker/gateway boxes that own them); code → nothing
// (a code module is manifest + require only — it imports the contracts its source
// binds and holds no spec nodes of its own). An empty kind allows any type — real
// modules always declare one (the schema requires it), but an unkinded module
// (hand-built input) is not gated.
func kindAllows(kind source.ModuleKind, t model.NodeType) bool {
	switch kind {
	case "":
		return true
	case source.KindService:
		return t == model.TypeContract || t == model.TypePort || t == model.TypeContainer
	case source.KindDomain:
		return t == model.TypeDomain || t == model.TypeNeed
	case source.KindGovernance:
		return t == model.TypePlan || t == model.TypeADR
	case source.KindTopology:
		return t == model.TypePort || t == model.TypeContainer
	case source.KindCode:
		return false
	}
	return false
}

// roleGateMessage tailors the violation to the kind. A code module is the special
// case: the rule is "no spec nodes at all" (it only declares which contracts its
// source may bind), so the message says that rather than naming the type.
func roleGateMessage(kind source.ModuleKind, t model.NodeType) string {
	if kind == source.KindCode {
		return fmt.Sprintf("a code module holds no spec nodes (found %s); it carries only a manifest and requires", t)
	}
	return fmt.Sprintf("node type %s not allowed in a %q module", t, kind)
}
