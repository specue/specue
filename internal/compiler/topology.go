package compiler

import "github.com/specue/specue/internal/model"

// deriveTopology aggregates each Port's L2 topology from the infra deps use cases
// declare: an infra dep (Role set) on a UC points at a Port, and the role places
// that UC in one of the Port's role lists. The topology is never authored — it is
// the reverse index of the infra edges, the L3→L2 bridge.
func deriveTopology(g *ResolvedGraph) {
	for n := range g.Nodes() {
		uc := n.Node().Body.UseCase
		if uc == nil {
			continue
		}
		from := n.ID().Module
		for _, el := range uc.Elements {
			for _, dep := range el.Deps {
				if !dep.IsInfra() {
					continue
				}
				attachToPort(g, from, n.ID(), dep)
			}
		}
	}
}

// attachToPort records a use case under the target Port's role list.
func attachToPort(g *ResolvedGraph, from model.ModulePath, uc model.NodeID, dep model.Dep) {
	port, ok := resolveTarget(g, from, dep.To)
	if !ok {
		return
	}
	pn, ok := g.Node(port)
	if !ok || pn.Node().Type != model.TypePort {
		return
	}
	pn.Topology.attach(dep.Role, uc)
}

// attach files a use case under the role list its role maps to.
func (t *TopologyRoles) attach(role model.Role, uc model.NodeID) {
	switch role {
	case model.RoleProduce, model.RolePublish:
		t.ProducedBy = append(t.ProducedBy, uc)
	case model.RoleConsume, model.RoleSubscribe:
		t.ConsumedBy = append(t.ConsumedBy, uc)
	case model.RoleServe:
		t.ServedBy = append(t.ServedBy, uc)
	case model.RoleCall, model.RoleRead, model.RoleWrite:
		t.CalledBy = append(t.CalledBy, uc)
	case model.RoleGrant:
		t.GrantedBy = append(t.GrantedBy, uc)
	}
}
