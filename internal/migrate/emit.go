package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	schemaModulePath = "specue.io/schema@v0"
	majorSuffix      = "@v0"
	// normalizedVersion: the migration is a single snapshot, so every module and
	// dep is pinned at one v0 version. v1's varied majors (v1/v2/v5) can't coexist
	// with a uniform @v0 path suffix (CUE ties the suffix to the version major),
	// and cross-module version pinning is a later feature, not golden semantics.
	normalizedVersion = "v0.0.1"
)

// writeModule writes the v2 module tree for one loaded v1 module under dir:
// cue.mod/module.cue, spec.mod.cue, and nodes.cue.
func writeModule(m *loadedModule, set *loadedSet, dir string) ([]SkippedRef, error) {
	nodes, skipped := emitNodes(m, set)
	skipped = append(skipped, droppedRequires(m, set)...)
	files := map[string]string{
		"cue.mod/module.cue": emitCueMod(m, set),
		"spec.mod.cue":       emitSpecMod(m, set),
		"nodes.cue":          nodes,
	}
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return nil, err
		}
	}
	return skipped, nil
}

// emitCueMod writes the CUE module file: the module path (with @v0) plus a dep on
// the schema and on every required module. This is what makes cue-native imports
// resolve.
func emitCueMod(m *loadedModule, set *loadedSet) string {
	var b strings.Builder
	fmt.Fprintf(&b, "module: %q\n", m.manifest.Module+majorSuffix)
	b.WriteString("language: version: \"v0.16.0\"\n")
	fmt.Fprintf(&b, "deps: %q: v: \"v0.0.1\"\n", schemaModulePath)
	for _, r := range keptRequires(m, set) {
		fmt.Fprintf(&b, "deps: %q: v: %q\n", r.Module+majorSuffix, normalizedVersion)
	}
	return b.String()
}

// emitSpecMod writes the v2 manifest: module/version/kind plus require entries
// carrying only module/version/replace (as/use are CUE-native now).
func emitSpecMod(m *loadedModule, set *loadedSet) string {
	var b strings.Builder
	fmt.Fprintf(&b, "module: %q\n", m.manifest.Module+majorSuffix)
	fmt.Fprintf(&b, "version: %q\n", normalizedVersion)
	fmt.Fprintf(&b, "kind:    %q\n", moduleKind(m))
	kept := keptRequires(m, set)
	if len(kept) > 0 {
		b.WriteString("require: [\n")
		for _, r := range kept {
			fmt.Fprintf(&b, "\t{module: %q, version: %q, replace: %q},\n",
				r.Module+majorSuffix, normalizedVersion, replaceFor(m, r))
		}
		b.WriteString("]\n")
	}
	return b.String()
}

// keptRequires are the module's requires whose target is in the migration set; a
// require to an out-of-set module is dropped (the resolver would fail reading a
// module that was never migrated). Such drops are reported by droppedRequires.
func keptRequires(m *loadedModule, set *loadedSet) []v1Require {
	var out []v1Require
	for _, r := range m.manifest.Require {
		if set.byPath[r.Module] != nil {
			out = append(out, r)
		}
	}
	return out
}

// droppedRequires reports each require whose target is not in the migration set.
func droppedRequires(m *loadedModule, set *loadedSet) []SkippedRef {
	var out []SkippedRef
	for _, r := range m.manifest.Require {
		if set.byPath[r.Module] == nil {
			out = append(out, SkippedRef{Module: m.manifest.Module,
				Reason: fmt.Sprintf("require %q dropped: not in migration set", r.Module)})
		}
	}
	return out
}

// moduleKind returns the module's role-gate kind. v2 requires a kind, but the
// original landscape predates the role-gate and declares none, so it is
// inferred from the node types the module holds: a Product/UserStory module
// migrates to a domain, a topology module (Channels/seam Containers) is
// topology, everything else (UseCases/Ports) is a service. An explicit v1
// kind ("product") is normalized to "domain".
func moduleKind(m *loadedModule) string {
	if m.manifest.Kind != "" {
		if m.manifest.Kind == "product" {
			return "domain"
		}
		return m.manifest.Kind
	}
	var hasProduct, hasStory, hasChannel, hasUseCase bool
	for _, n := range m.nodes {
		switch n.Type {
		case "Product":
			hasProduct = true
		case "UserStory":
			hasStory = true
		case "Channel":
			hasChannel = true
		case "UseCase":
			hasUseCase = true
		}
	}
	switch {
	case hasProduct || (hasStory && !hasUseCase):
		return "domain"
	case hasChannel && !hasUseCase:
		return "topology"
	default:
		return "service"
	}
}

// replaceFor is the local path from the requiring module's migrated dir to the
// required module's migrated dir. Both mirror their module path under the out
// root (relPath), so the replace is the relative path between those two — found
// with filepath.Rel, which yields the right number of ../ hops regardless of how
// deep each module sits.
func replaceFor(m *loadedModule, r v1Require) string {
	from := relPath(m.manifest.Module)
	to := relPath(r.Module)
	rel, err := filepath.Rel(from, to)
	if err != nil {
		return to
	}
	return rel
}

// emitNodes writes the package's nodes.cue: package clause, schema + per-require
// imports, synthesized service containers, then every node as a field.
func emitNodes(m *loadedModule, set *loadedSet) (string, []SkippedRef) {
	rc := resolveCtx{owner: m, set: set, used: map[string]bool{}}

	// Render bodies first so rc.used records which import aliases are referenced;
	// CUE rejects an imported-and-unused module, so only used imports are emitted.
	var body strings.Builder
	for _, sc := range serviceContainers(m) {
		body.WriteString(sc)
		body.WriteString("\n")
	}
	for _, n := range m.nodes {
		body.WriteString(emitNode(n, &rc))
		body.WriteString("\n")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", packageName(lastSegment(m.manifest.Module)))
	b.WriteString("import (\n")
	fmt.Fprintf(&b, "\ts %q\n", schemaModulePath+":spec")
	for _, r := range m.manifest.Require {
		alias := cueAlias(r)
		if !rc.used[alias] {
			continue
		}
		fmt.Fprintf(&b, "\t%s %q\n", alias, r.Module+majorSuffix+":"+packageName(lastSegment(r.Module)))
	}
	b.WriteString(")\n\n")
	b.WriteString(body.String())

	var skipped []SkippedRef
	for _, e := range rc.errs {
		skipped = append(skipped, SkippedRef{Module: m.manifest.Module, Reason: e.Error()})
	}
	return b.String(), skipped
}

// serviceContainers synthesizes one #Container per distinct service label the
// module's use cases name (v1 left the service box implicit; v2 makes it a node
// the UseCase points at). The field is the container's fieldName.
func serviceContainers(m *loadedModule) []string {
	seen := map[string]bool{}
	var labels []string
	for _, n := range m.nodes {
		if n.Type == "UseCase" && n.Service != "" && !seen[n.Service] {
			seen[n.Service] = true
			labels = append(labels, n.Service)
		}
	}
	sort.Strings(labels)
	var out []string
	for _, label := range labels {
		field := fieldName(label)
		var b strings.Builder
		fmt.Fprintf(&b, "%s: s.#Container & {\n", field)
		b.WriteString("\ttype:       \"Container\"\n")
		fmt.Fprintf(&b, "\tslug:       %q\n", slugify(label))
		fmt.Fprintf(&b, "\ttitle:      %q\n", label)
		b.WriteString("\tconfidence: \"CONFIRMED\"\n")
		b.WriteString("\tkind:       \"service\"\n")
		b.WriteString("}\n")
		out = append(out, b.String())
	}
	return out
}

// slugify turns a service label into a kebab slug, grouping runs of capitals as
// one acronym word ("SpecueCLI" → "spec-graph-cli", "specue" →
// "specue").
func slugify(label string) string {
	rs := []rune(label)
	isUp := func(r rune) bool { return r >= 'A' && r <= 'Z' }
	var b strings.Builder
	for i, r := range rs {
		if isUp(r) && i > 0 {
			prevLower := !isUp(rs[i-1])
			nextLower := i+1 < len(rs) && !isUp(rs[i+1])
			if prevLower || nextLower {
				b.WriteByte('-')
			}
		}
		if isUp(r) {
			b.WriteRune(r + 32)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func orDefaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
