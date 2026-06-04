package migrate

import (
	"fmt"
	"strconv"
	"strings"
)

// normalizeConfidence maps a v1 confidence value to the v2 enum
// (CONFIRMED|LIKELY|SPECULATIVE). v1 also used GAP — an "accepted but unbuilt"
// marker — which v2 expresses as a derived status (asserted), not a confidence;
// at the confidence axis it is the least-certain bucket, so it maps to SPECULATIVE.
func normalizeConfidence(c string) string {
	if c == "GAP" {
		return "SPECULATIVE"
	}
	return c
}

// emitAtoms renders an fr/nfr atom struct. v1 stores a bare ordinal id ("01") and
// only prefixes "fr-"/"nfr-" at reference sites; v2's atom id is the full
// "fr-01" form, so the prefix is applied here, matching how satisfies refs read.
// v2 keys atoms by a CUE-friendly field name (fr_01) so satisfies can reference
// them CUE-natively as need.frs.fr_01.
// field is the v2 container name ("frs"/"nfrs"); kind is the wire-id prefix
// ("fr"/"nfr"). v1 stores both ordinal-only ("01"); v2 keys the struct field
// CUE-friendly ("fr_01") and carries the full id ("fr-01") as the atom's id.
func emitAtoms(b *strings.Builder, field, kind string, atoms []v1Atom) {
	if len(atoms) == 0 {
		return
	}
	fmt.Fprintf(b, "\t%s: {\n", field)
	for _, a := range atoms {
		id := kind + "-" + a.ID
		key := atomFieldKey(kind, a.ID)
		fmt.Fprintf(b, "\t\t%s: {id: %q, text: %s}\n", key, id, cueString(a.Text, 2))
	}
	b.WriteString("\t}\n")
}

// atomFieldKey turns a v1 atom kind+id ("fr", "01") into the v2 struct key
// ("fr_01") satisfies references use.
func atomFieldKey(kind, ordinal string) string {
	return kind + "_" + ordinal
}

// cueString renders s as a CUE string literal. A single-line value uses a quoted
// literal; a multi-line value (a body, mostly) uses a """ block indented at the
// given tab depth, so the emitted CUE stays readable and re-parses identically.
func cueString(s string, depth int) string {
	if !strings.Contains(s, "\n") {
		return strconv.Quote(s)
	}
	indent := strings.Repeat("\t", depth)
	var b strings.Builder
	b.WriteString("\"\"\"\n")
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		b.WriteString(indent + "\t")
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(indent + "\"\"\"")
	return b.String()
}
