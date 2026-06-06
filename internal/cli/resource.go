package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/specue/specue/internal/model"
)

// resource is one selectable node kind: its canonical name, aliases, and the node
// type it filters. The CLI grammar is kubectl-shaped — `get <resource>` filters by
// type, so the resource set is exactly the node-type set, named for humans.
type resource struct {
	name    string
	aliases []string
	typ     model.NodeType
}

// allResource is the pseudo-resource that lists every node of every type — the
// kubectl `get all` pattern. It is not in the registry (it has no single type); the
// get verb special-cases it.
const allResource = "all"

// resources is the closed registry. Order is the listing order in help.
var resources = []resource{
	{"contract", []string{"ct"}, model.TypeContract},
	{"need", []string{"story", "us", "userstory"}, model.TypeNeed},
	{"domain", []string{"prod", "product", "dom"}, model.TypeDomain},
	{"port", nil, model.TypePort},
	{"container", []string{"box"}, model.TypeContainer},
	{"plan", nil, model.TypePlan},
	{"adr", nil, model.TypeADR},
}

// resolveResource maps a user-supplied resource word (canonical or alias) to its
// registry entry. Returning the entry (not just the type) lets callers record the
// CANONICAL name, so an alias never leaks into output — `get ct` and `get contract`
// produce identical results. An unknown word is a Problem naming the valid set.
func resolveResource(word string) (resource, *Problem) {
	w := strings.ToLower(word)
	for _, r := range resources {
		if r.name == w {
			return r, nil
		}
		for _, a := range r.aliases {
			if a == w {
				return r, nil
			}
		}
	}
	p := Errorf("use one of: "+resourceList(), "unknown resource %q", word)
	return resource{}, &p
}

// resourceList renders the valid resources with their aliases, for error fixes and
// help text: "contract (con), userstory (story, us), …".
func resourceList() string {
	parts := make([]string, len(resources))
	for i, r := range resources {
		if len(r.aliases) == 0 {
			parts[i] = r.name
			continue
		}
		parts[i] = fmt.Sprintf("%s (%s)", r.name, strings.Join(r.aliases, ", "))
	}
	return strings.Join(parts, ", ")
}

// parseNodeID parses a `module:slug` reference into a NodeID. A bare slug (no
// colon) is rejected with the fix to qualify it — identity is always
// module-scoped, and the CLI refuses the ambiguity rather than guessing a module.
// The module path itself may contain no colon, so the split is on the LAST colon.
func parseNodeID(ref string) (model.NodeID, *Problem) {
	i := strings.LastIndex(ref, ":")
	if i < 0 {
		p := Errorf("qualify it as module:slug (copy the full reference from `get` output)",
			"%q is not a module:slug reference", ref)
		return model.NodeID{}, &p
	}
	mod, slug := ref[:i], ref[i+1:]
	if mod == "" || slug == "" {
		p := Errorf("qualify it as module:slug, e.g. specue/example:validate-graph",
			"%q is missing a module or a slug", ref)
		return model.NodeID{}, &p
	}
	return model.NodeID{Module: model.ModulePath(mod), Slug: model.Slug(slug)}, nil
}

// parseNodeAtElement parses `module:slug` or `module:slug#element`. The element
// suffix is optional; when absent the returned ElementID is empty and the caller
// reads the whole node. Identity is always module-scoped.
func parseNodeAtElement(ref string) (model.NodeID, model.ElementID, *Problem) {
	nodeRef, elem := ref, ""
	if i := strings.IndexByte(ref, '#'); i >= 0 {
		nodeRef, elem = ref[:i], ref[i+1:]
	}
	id, p := parseNodeID(nodeRef)
	if p != nil {
		return model.NodeID{}, "", p
	}
	return id, model.ElementID(elem), nil
}

// sortByID orders node IDs for stable output.
func sortByID(ids []model.NodeID) {
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
}
