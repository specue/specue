package migrate

import (
	"fmt"
	"strconv"
	"strings"
)

// emitNode renders one v1 node as a v2 CUE field unified against its schema def.
func emitNode(n v1Node, rc *resolveCtx) string {
	// v1 modelled an async channel as its own node type; v2 folds it into a Port
	// with kind: channel (it already carries kind/transport), so normalize here.
	if n.Type == "Channel" {
		n.Type = "Port"
		if n.Kind == "" {
			n.Kind = "channel"
		}
	}
	// v1's intent axis renamed: UserStory → Need, Product → Domain.
	switch n.Type {
	case "UserStory":
		n.Type = "Need"
	case "Product":
		n.Type = "Domain"
	}
	var b strings.Builder
	def := schemaDef(n.Type)
	fmt.Fprintf(&b, "%s: s.%s & {\n", fieldName(n.Slug), def)
	fmt.Fprintf(&b, "\ttype:       %q\n", n.Type)
	fmt.Fprintf(&b, "\tslug:       %q\n", n.Slug)
	fmt.Fprintf(&b, "\ttitle:      %s\n", cueString(n.Title, 1))
	if n.Confidence != "" {
		fmt.Fprintf(&b, "\tconfidence: %q\n", normalizeConfidence(n.Confidence))
	}
	if n.LegacyID != "" {
		fmt.Fprintf(&b, "\tlegacy_id:  %q\n", n.LegacyID)
	}

	switch n.Type {
	case "UseCase":
		emitUseCase(&b, n, rc)
	case "Need":
		emitNeed(&b, n)
	case "Port":
		emitPort(&b, n, rc)
	case "Container":
		emitContainer(&b, n)
	case "Plan", "ADR":
		emitGov(&b, n)
	}

	if n.Body != "" {
		fmt.Fprintf(&b, "\tbody: %s\n", cueString(n.Body, 1))
	}
	b.WriteString("}\n")
	return b.String()
}

func schemaDef(typ string) string {
	switch typ {
	case "UseCase":
		return "#UseCase"
	case "Need":
		return "#Need"
	case "Domain":
		return "#Domain"
	case "Port":
		return "#Port"
	case "Container":
		return "#Container"
	case "Plan":
		return "#Plan"
	case "ADR":
		return "#ADR"
	}
	return "#Node"
}

func emitUseCase(b *strings.Builder, n v1Node, rc *resolveCtx) {
	if n.Service != "" {
		fmt.Fprintf(b, "\tservice:    %s\n", fieldName(n.Service))
	}
	if n.Binding != "" {
		fmt.Fprintf(b, "\tbinding:     %q\n", n.Binding)
	}
	if n.Interaction != "" {
		fmt.Fprintf(b, "\tinteraction: %q\n", n.Interaction)
	}
	if n.Trigger != "" {
		fmt.Fprintf(b, "\ttrigger:     %s\n", cueString(n.Trigger, 1))
	}
	if n.Deprecated != "" {
		fmt.Fprintf(b, "\tdeprecated:  %s\n", cueString(n.Deprecated, 1))
	}
	emitElements(b, "preconditions", n.Preconditions, rc)
	emitElements(b, "postconditions", n.Postconditions, rc)
	emitElements(b, "invariants", n.Invariants, rc)
	emitElements(b, "variations", n.Variations, rc)
}

// emitNed translates a v1 UserStory into a v2 Need. The v1 narrative
// (as/want/so_that) collapses: `as` becomes consumer; `want` and `so_that`
// fold into description as "<want>, so that <so_that>". Empty parts are
// omitted from the joined description.
func emitNeed(b *strings.Builder, n v1Node) {
	fmt.Fprintf(b, "\tdomain:      %s\n", fieldName(n.Product))
	fmt.Fprintf(b, "\tconsumer:    %s\n", cueString(n.Narrative.As, 1))
	fmt.Fprintf(b, "\tdescription: %s\n", cueString(joinDescription(n.Narrative.Want, n.Narrative.SoThat), 1))
	emitAtoms(b, "frs", "fr", n.FR)
	emitAtoms(b, "nfrs", "nfr", n.NFR)
}

// joinDescription folds v1 want + so_that into one Need.description string.
func joinDescription(want, soThat string) string {
	switch {
	case want == "" && soThat == "":
		return ""
	case soThat == "":
		return want
	case want == "":
		return "so that " + soThat
	}
	return want + ", so that " + soThat
}

func emitPort(b *strings.Builder, n v1Node, rc *resolveCtx) {
	fmt.Fprintf(b, "\tkind:       %q\n", n.Kind)
	if n.Technology != "" {
		fmt.Fprintf(b, "\ttechnology: %s\n", cueString(n.Technology, 1))
	}
	if n.Transport != "" {
		fmt.Fprintf(b, "\ttransport:  %q\n", n.Transport)
	}
	// v1 schema: a string file pointer (e.g. "spec/cli/cli-output.schema.json").
	// v2 schema: a #Node ref. There is no schema NODE in the v1 self-spec, so the
	// pointer is dropped (the binding becomes a proto-anchor later, out of scope).
}

func emitContainer(b *strings.Builder, n v1Node) {
	fmt.Fprintf(b, "\tkind:       %q\n", n.Kind)
	if n.Technology != "" {
		fmt.Fprintf(b, "\ttechnology: %s\n", cueString(n.Technology, 1))
	}
	if n.Boundary {
		b.WriteString("\tboundary:   true\n")
	}
}

func emitGov(b *strings.Builder, n v1Node) {
	fmt.Fprintf(b, "\tstatus:     %q\n", n.Status)
	if n.Branch != "" {
		fmt.Fprintf(b, "\tbranch:     %q\n", n.Branch)
	}
}

// emitElements renders an element section (pre/post/invariants/variations).
func emitElements(b *strings.Builder, field string, els []v1Element, rc *resolveCtx) {
	if len(els) == 0 {
		return
	}
	fmt.Fprintf(b, "\t%s: [\n", field)
	for _, e := range els {
		emitElement(b, e, rc)
	}
	b.WriteString("\t]\n")
}

func emitElement(b *strings.Builder, e v1Element, rc *resolveCtx) {
	b.WriteString("\t\t{\n")
	if e.ID != "" {
		fmt.Fprintf(b, "\t\t\tid:   %q\n", e.ID)
	}
	if e.Text != "" {
		fmt.Fprintf(b, "\t\t\ttext: %s\n", cueString(e.Text, 3))
	}
	if e.When != "" {
		fmt.Fprintf(b, "\t\t\twhen: %s\n", cueString(e.When, 3))
	}
	if e.Then != "" {
		fmt.Fprintf(b, "\t\t\tthen: %s\n", cueString(e.Then, 3))
	}
	if e.Rev > 0 {
		fmt.Fprintf(b, "\t\t\trev:  %s\n", strconv.Itoa(e.Rev))
	}
	emitDeps(b, e, rc)
	emitSatisfies(b, e.Satisfies, rc)
	emitDecidedBy(b, e.DecidedBy, rc)
	b.WriteString("\t\t},\n")
}

// emitDeps renders depends_on, invokes, and infra as v2 depends_on entries (v2
// collapses the three into one list; an infra dep just carries a role).
func emitDeps(b *strings.Builder, e v1Element, rc *resolveCtx) {
	var deps []v1Dep
	deps = append(deps, e.DependsOn...)
	deps = append(deps, e.Invokes...)
	deps = append(deps, e.Infra...)
	if len(deps) == 0 {
		return
	}
	b.WriteString("\t\t\tdepends_on: [\n")
	for _, d := range deps {
		to, err := rc.nodeExpr(d.To)
		if err != nil {
			rc.fail(err)
			continue
		}
		b.WriteString("\t\t\t\t{to: " + to)
		if d.Role != "" {
			fmt.Fprintf(b, ", role: %q", d.Role)
		}
		if d.Carries != "" {
			c, err := rc.nodeExpr(d.Carries)
			if err != nil {
				rc.fail(err)
			} else {
				b.WriteString(", carries: " + c)
			}
		}
		b.WriteString("},\n")
	}
	b.WriteString("\t\t\t]\n")
}

func emitSatisfies(b *strings.Builder, refs []string, rc *resolveCtx) {
	if len(refs) == 0 {
		return
	}
	b.WriteString("\t\t\tsatisfies: [\n")
	for _, ref := range refs {
		expr, err := rc.atomExpr(ref)
		if err != nil {
			rc.fail(err)
			continue
		}
		fmt.Fprintf(b, "\t\t\t\t%s,\n", expr)
	}
	b.WriteString("\t\t\t]\n")
}

func emitDecidedBy(b *strings.Builder, refs []string, rc *resolveCtx) {
	if len(refs) == 0 {
		return
	}
	b.WriteString("\t\t\tdecided_by: [")
	first := true
	for _, ref := range refs {
		expr, err := rc.nodeExpr(ref)
		if err != nil {
			rc.fail(err)
			continue
		}
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(expr)
	}
	b.WriteString("]\n")
}
