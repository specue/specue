package plan

import (
	"fmt"
	"os"
)

// writeRecord writes the plan's anchor — a #Plan node in the governance module,
// authored CUE-natively like any other node. It names the plan and points at the
// branch its content lives on; status starts proposed. The field name is the
// slug's camel form; the file is plan-<id>.cue.
func (m *Manager) writeRecord(id, title string) error {
	if title == "" {
		title = id
	}
	src := fmt.Sprintf(`package %s

import s "specue.io/schema@v0:spec"

%s: s.#Plan & {
	type:       "Plan"
	slug:       %q
	title:      %q
	confidence: "CONFIRMED"
	status:     "proposed"
	branch:     %q
}
`, m.planPackage(), planField(id), "plan-"+id, title, branch(id))
	return os.WriteFile(m.recordFile(id), []byte(src), 0o644)
}

// removeRecord deletes the plan's record file. A missing file is not an error
// (drop is idempotent).
func (m *Manager) removeRecord(id string) error {
	err := os.Remove(m.recordFile(id))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// planField is the CUE field name for a plan record: plan<Id> with the id's
// separators stripped (plan id "gp-1076" → field "planGp1076").
func planField(id string) string {
	field := "plan"
	upper := true
	for _, r := range id {
		if r == '-' || r == '_' || r == '/' {
			upper = true
			continue
		}
		if upper && r >= 'a' && r <= 'z' {
			r = r - 32
		}
		upper = false
		field += string(r)
	}
	return field
}
