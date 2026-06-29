// decision: presentation/add-output
// decision: internal/presentation/rendering-output
// decision: internal/presentation/output-language
package presentation

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/specue/specue/internal/usecase"
)

// addSuccess is the machine success object.
//
// decision: presentation/add-output
type addSuccess struct {
	OK       bool        `json:"ok"`
	Module   addModule   `json:"module"`
	Decision addDecision `json:"decision"`
}

type addModule struct {
	ID   string `json:"id"`
	Root string `json:"root"`
}

type addDecision struct {
	Dir           string `json:"dir"`
	StructureFile string `json:"structureFile"`
	BodyFile      string `json:"bodyFile"`
}

// addFailure is the machine error object.
//
// decision: presentation/add-output
type addFailure struct {
	OK    bool     `json:"ok"`
	Error errorObj `json:"error"`
}

type errorObj struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RenderAddSuccess writes a success result.
// Default human-readable form goes to stdout; --format json gives the
// single-line machine object. All text is English, no decoration/emoji.
//
// decision: presentation/add-output
// decision: internal/presentation/rendering-output
// decision: internal/presentation/output-language
func RenderAddSuccess(stdout io.Writer, format string, r *usecase.AddResult) error {
	if format == "json" {
		// decision: internal/presentation/rendering-output
		b, err := json.Marshal(addSuccess{
			OK:     true,
			Module: addModule{ID: r.Module, Root: r.Root},
			Decision: addDecision{
				Dir:           r.ID,
				StructureFile: r.StructureFile,
				BodyFile:      r.BodyFile,
			},
		})
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, string(b))
		return nil
	}

	// Human-readable: report what was created, absolute module root.
	// decision: presentation/add-output
	fmt.Fprintf(stdout, "created decision %s in %s\n", r.ID, r.Root)
	fmt.Fprintf(stdout, "  %s\n", r.StructureFile)
	fmt.Fprintf(stdout, "  %s\n", r.BodyFile)
	return nil
}

// RenderAddError writes a failure to stderr.
// The error explains the cause; machine form carries a stable code.
//
// decision: presentation/add-output
func RenderAddError(stderr io.Writer, format string, err error) {
	code := "unknown"
	msg := err.Error()
	if re, ok := err.(*usecase.ResolveError); ok {
		code = re.Code
		msg = re.Message
	}

	if format == "json" {
		b, _ := json.Marshal(addFailure{OK: false, Error: errorObj{Code: code, Message: msg}})
		fmt.Fprintln(stderr, string(b))
		return
	}
	// decision: internal/presentation/output-language
	fmt.Fprintf(stderr, "error: %s\n", msg)
}
