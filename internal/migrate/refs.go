package migrate

import (
	"fmt"
	"strings"
	"unicode"
)

// resolveCtx is what a node needs to turn a v1 reference string into a v2
// cue-native expression: the owning module (for bare refs) and the set so a
// qualified ref finds the target module's import alias and field name.
type resolveCtx struct {
	owner *loadedModule
	set   *loadedSet
	used  map[string]bool // import aliases actually referenced (for the import block)
	errs  []error         // ref-resolution failures, surfaced after emit (no silent drops)
}

// fail records a ref-resolution error; the migrator surfaces it rather than
// emitting garbage or silently dropping the edge.
func (c *resolveCtx) fail(err error) { c.errs = append(c.errs, err) }

// nodeExpr turns a v1 node reference into a CUE expression pointing at that node.
// v1 forms: "slug" (local), "alias:slug" (qualified via require). The result is a
// field reference — "localField" or "importAlias.field" — matching how the
// emitter names node fields and imports.
func (c resolveCtx) nodeExpr(ref string) (string, error) {
	alias, slug, qualified := strings.Cut(ref, ":")
	if !qualified {
		return fieldName(ref), nil
	}
	r, ok := c.aliasRequire(alias)
	if !ok {
		return "", fmt.Errorf("ref %q: unknown require alias %q", ref, alias)
	}
	if c.set.byPath[r.Module] == nil {
		return "", fmt.Errorf("ref %q: module %q not in migration set", ref, r.Module)
	}
	ca := cueAlias(r)
	if c.used != nil {
		c.used[ca] = true
	}
	return ca + "." + fieldName(slug), nil
}

// atomExpr turns a v1 satisfies ref ("alias:story#fr-01" or "story#fr-01") into a
// CUE-native expression pointing at the atom field within the Need's frs/nfrs
// struct (e.g. `need.frs.fr_01`). v2's satisfies edge is a bare reference, so
// the loader recovers both the owning Need and the atom id from the path.
func (c resolveCtx) atomExpr(ref string) (string, error) {
	node, atomID, ok := strings.Cut(ref, "#")
	if !ok {
		return "", fmt.Errorf("satisfies ref %q: missing #atom", ref)
	}
	needExpr, err := c.nodeExpr(node)
	if err != nil {
		return "", err
	}
	kind, ordinal, ok := strings.Cut(atomID, "-")
	if !ok || (kind != "fr" && kind != "nfr") {
		return "", fmt.Errorf("satisfies ref %q: atom id must be fr-NN or nfr-NN", ref)
	}
	bucket := kind + "s" // fr → frs, nfr → nfrs
	return needExpr + "." + bucket + "." + atomFieldKey(kind, ordinal), nil
}

// aliasRequire finds the require a v1 ref alias names. The match is on the RAW v1
// alias (what the ref string carries), not the sanitized CUE identifier.
func (c resolveCtx) aliasRequire(alias string) (v1Require, bool) {
	for _, r := range c.owner.manifest.Require {
		if rawAlias(r) == alias {
			return r, true
		}
	}
	return v1Require{}, false
}

// rawAlias is the v1 import alias a ref string uses: its explicit `as`, else the
// module path's last segment (v1's default). May contain hyphens (service-name).
func rawAlias(r v1Require) string {
	if r.As != "" {
		return r.As
	}
	return lastSegment(r.Module)
}

// cueAlias is the import alias as a valid CUE identifier — rawAlias with hyphens
// stripped (service-name → servicename), since a CUE import alias must be an identifier.
func cueAlias(r v1Require) string {
	return packageName(rawAlias(r))
}

// fieldName converts a kebab/upper slug into the lowerCamelCase CUE field the
// emitter names a node by (e.g. "validate-graph" → "validateGraph",
// "adr-module-role-gate" → "adrModuleRoleGate").
func fieldName(slug string) string {
	parts := strings.FieldsFunc(slug, func(r rune) bool { return r == '-' || r == '/' || r == '_' })
	var b strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(p))
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + strings.ToLower(p[1:]))
	}
	out := b.String()
	if out == "" || unicode.IsDigit(rune(out[0])) {
		out = "n" + out
	}
	return out
}

// packageName sanitizes a module's last segment into a valid CUE package
// identifier (hyphens dropped): "cli-product" → "cliproduct".
func packageName(segment string) string {
	return strings.Map(func(r rune) rune {
		if r == '-' || r == '_' {
			return -1
		}
		return r
	}, segment)
}
