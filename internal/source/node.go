package source

import (
	"fmt"

	"cuelang.org/go/cue"

	"github.com/specue/specue/internal/model"
)

// Attributor maps a source filename (where a node value was authored) to the
// module that owns it. The specload layer builds it from the resolved closure: a
// reference's target file, matched against each module's directory, names the
// target's module. The node mapper uses it to turn a recovered CUE reference into
// a resolved NodeRef.
type Attributor func(filename string) (model.ModulePath, bool)

// MapNode maps one resolved CUE node value into the authored model. Unlike v1 it
// reads a value already stitched into the whole module set: every cross-module
// reference is resolved by CUE, so MapNode recovers each reference's target
// (module + slug) from the value's expression rather than parsing a string. The
// module/file/visibility of the node itself are stamped by the caller (specload),
// which knows the node's own location.
func MapNode(v cue.Value, attrib Attributor) (model.Node, error) {
	typ, err := lookupString(v, "type")
	if err != nil {
		return model.Node{}, err
	}
	n := model.Node{
		Slug:       model.Slug(mustString(v, "slug")),
		Type:       model.NodeType(typ),
		Title:      mustString(v, "title"),
		Confidence: model.Confidence(mustString(v, "confidence")),
		Visibility: model.Public, // specload stamps Private from the path
		Body:       &model.Body{Prose: optString(v, "body")},
	}

	switch n.Type {
	case model.TypeContract:
		n.Body.Contract = mapContract(v, attrib)
	case model.TypeNeed:
		n.Body.Need = mapNeed(v, attrib)
	case model.TypeDomain:
		// A domain carries only common fields + prose; no type-specific body.
	case model.TypePort:
		n.Body.Port = mapPort(v, attrib)
	case model.TypeContainer:
		n.Body.Container = mapContainer(v)
	case model.TypePlan, model.TypeADR:
		n.Body.Gov = mapGov(v)
	default:
		return model.Node{}, fmt.Errorf("unknown node type %q", n.Type)
	}
	return n, nil
}

func mapContract(v cue.Value, attrib Attributor) *model.ContractBody {
	return &model.ContractBody{
		Service:     mapRef(v.LookupPath(cue.ParsePath("service")), attrib),
		Trigger:     optString(v, "trigger"),
		Binding:     model.Binding(orDefault(optString(v, "binding"), string(model.BindingRequired))),
		Interaction: model.Interaction(orDefault(optString(v, "interaction"), string(model.InteractionAsync))),
		Deprecated:  optString(v, "deprecated"),
		Elements:    mapElements(v, attrib),
	}
}

// mapElements reads the single `invariants` section into an ordered slice. A
// guarded invariant (When set) has branch deps — a conditional branch must not
// block the parent's main contract — so branch is derived per element from the
// presence of `when`.
func mapElements(v cue.Value, attrib Attributor) []model.Element {
	list := v.LookupPath(cue.ParsePath("invariants"))
	if !list.Exists() {
		return nil
	}
	var out []model.Element
	for it, _ := list.List(); it.Next(); {
		out = append(out, mapElement(it.Value(), attrib))
	}
	return out
}

func mapElement(e cue.Value, attrib Attributor) model.Element {
	when := optString(e, "when")
	return model.Element{
		ID:        model.ElementID(optString(e, "id")),
		Text:      optString(e, "text"),
		Kind:      model.ElementKind(optString(e, "kind")),
		When:      when,
		Rev:       optInt(e, "rev"),
		Deps:      mapDeps(e, when != "", attrib),
		Satisfies: mapSatisfies(e, attrib),
		DecidedBy: mapDecidedBy(e, attrib),
	}
}

func mapDeps(e cue.Value, branch bool, attrib Attributor) []model.Dep {
	list := e.LookupPath(cue.ParsePath("depends_on"))
	if !list.Exists() {
		return nil
	}
	var out []model.Dep
	for it, _ := list.List(); it.Next(); {
		dep := it.Value()
		out = append(out, model.Dep{
			To:      mapRef(dep.LookupPath(cue.ParsePath("to")), attrib),
			Role:    model.Role(optString(dep, "role")),
			Carries: mapRef(dep.LookupPath(cue.ParsePath("carries")), attrib),
			Branch:  branch,
		})
	}
	return out
}

func mapSatisfies(e cue.Value, attrib Attributor) []model.AtomRef {
	list := e.LookupPath(cue.ParsePath("satisfies"))
	if !list.Exists() {
		return nil
	}
	var out []model.AtomRef
	for it, _ := list.List(); it.Next(); {
		out = append(out, mapAtomRef(it.Value(), attrib))
	}
	return out
}

// mapAtomRef recovers a satisfies' atom reference into a typed AtomRef. The
// `atom` value is a cue-native reference into a Need's frs/nfrs struct
// (`agent.navigate.frs."fr-01"`). Both the owning Need and the wire atom id
// are reachable from it directly:
//   - need: the prefix of the reference's path minus the last two segments
//     (frs/nfrs and the atom key), looked up against the reference's root
//   - atom id: the dereferenced target's `id` field
//
// The author never repeats the Need on the satisfies edge — CUE knows it.
func mapAtomRef(atom cue.Value, attrib Attributor) model.AtomRef {
	ref := referenceOperand(atom)
	root, path := ref.ReferencePath()
	selectors := path.Selectors()
	var needVal cue.Value
	if len(selectors) >= 2 {
		needVal = root.LookupPath(cue.MakePath(selectors[:len(selectors)-2]...))
	}
	needSlug, _ := needVal.LookupPath(cue.ParsePath("slug")).String()
	needMod, _ := attrib(needVal.Pos().Filename())
	target := cue.Dereference(ref)
	atomID, _ := target.LookupPath(cue.ParsePath("id")).String()
	return model.AtomRef{
		Need: model.NodeRef{Module: needMod, Slug: model.Slug(needSlug)},
		Atom: model.AtomID(atomID),
	}
}

func mapDecidedBy(e cue.Value, attrib Attributor) []model.NodeRef {
	list := e.LookupPath(cue.ParsePath("decided_by"))
	if !list.Exists() {
		return nil
	}
	var out []model.NodeRef
	for it, _ := list.List(); it.Next(); {
		out = append(out, mapRef(it.Value(), attrib))
	}
	return out
}

func mapNeed(v cue.Value, attrib Attributor) *model.NeedBody {
	return &model.NeedBody{
		Domain:      mapRef(v.LookupPath(cue.ParsePath("domain")), attrib),
		Consumer:    mustString(v, "consumer"),
		Description: mustString(v, "description"),
		Atoms:       mapAtoms(v),
	}
}

func mapAtoms(v cue.Value) []model.Atom {
	var out []model.Atom
	out = append(out, mapAtomsOf(v, "frs", model.KindFR)...)
	out = append(out, mapAtomsOf(v, "nfrs", model.KindNFR)...)
	return out
}

// mapAtomsOf reads a struct of named atoms (frs: { "fr-01": {...}, ... }) and
// returns them in declaration order. The CUE field key is opaque — the wire id
// lives in the atom's own `id` field; the schema enforces it.
func mapAtomsOf(v cue.Value, field string, kind model.AtomKind) []model.Atom {
	st := v.LookupPath(cue.ParsePath(field))
	if !st.Exists() {
		return nil
	}
	var out []model.Atom
	it, err := st.Fields()
	if err != nil {
		return nil
	}
	for it.Next() {
		a := it.Value()
		out = append(out, model.Atom{Kind: kind, ID: model.AtomID(mustString(a, "id")), Text: mustString(a, "text")})
	}
	return out
}

func mapPort(v cue.Value, attrib Attributor) *model.PortBody {
	pb := &model.PortBody{
		Kind:       model.PortKind(mustString(v, "kind")),
		Technology: optString(v, "technology"),
		Transport:  model.Transport(optString(v, "transport")),
	}
	if schema := v.LookupPath(cue.ParsePath("schema")); schema.Exists() {
		pb.Schema = mapRef(schema, attrib)
	}
	return pb
}

func mapContainer(v cue.Value) *model.ContainerBody {
	b, _ := v.LookupPath(cue.ParsePath("boundary")).Bool()
	return &model.ContainerBody{
		Kind:       model.ContainerKind(mustString(v, "kind")),
		Technology: optString(v, "technology"),
		Boundary:   b,
	}
}

func mapGov(v cue.Value) *model.GovBody {
	return &model.GovBody{
		Lifecycle: model.Lifecycle(mustString(v, "status")),
		Branch:    optString(v, "branch"),
	}
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
