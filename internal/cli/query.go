package cli

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/specue/specue/internal/query"
)

// QueryReport is the typed result of `query <sql>`: the column order and the rows
// (each a column→value map). Column order is captured so the human table is stable
// (a map alone would not preserve it).
type QueryReport struct {
	Columns []string    `json:"columns"`
	Rows    []query.Row `json:"rows"`
}

// runQuery builds the SQLite projection of the graph and runs the SQL against it.
// The projection is read-only (PRAGMA query_only), so a write is rejected by SQLite
// and surfaced with the fix to read, not mutate.
//
//specue:req:query-graph
func runQuery(ctx Context, sqlText string) (QueryReport, *Problem) {
	res, p := buildGraph(ctx)
	if p != nil {
		return QueryReport{}, p
	}
	db, err := query.Build(res.Graph, res.Diags)
	if err != nil {
		p := Errorf("re-run; if it persists the projection layer has a bug", "cannot build the query projection: %v", err)
		return QueryReport{}, &p
	}
	defer db.Close()

	rows, err := db.Query(sqlText)
	if err != nil {
		p := Errorf("check the SQL against `"+cmdPath(cmdQuery, subTables)+"` (table/column names, read-only)",
			"query failed: %v", err)
		return QueryReport{}, &p
	}
	return QueryReport{Columns: columnsOf(rows), Rows: rows}, nil
}

// columnsOf recovers a stable column order from the rows. SQL result order is not
// preserved by a map, so columns are sorted for a deterministic table; JSON callers
// key by name and do not care.
func columnsOf(rows []query.Row) []string {
	seen := map[string]bool{}
	var cols []string
	for _, r := range rows {
		for c := range r {
			if !seen[c] {
				seen[c] = true
				cols = append(cols, c)
			}
		}
	}
	slices.Sort(cols)
	return cols
}

func (r QueryReport) renderHuman(w io.Writer) error {
	if len(r.Rows) == 0 {
		_, err := fmt.Fprintln(w, "no rows")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, strings.ToUpper(strings.Join(r.Columns, "\t"))); err != nil {
		return err
	}
	for _, row := range r.Rows {
		cells := make([]string, len(r.Columns))
		for i, c := range r.Columns {
			cells[i] = cellString(row[c])
		}
		if _, err := fmt.Fprintln(tw, strings.Join(cells, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (r QueryReport) jsonValue() any {
	if r.Rows == nil {
		r.Rows = []query.Row{}
	}
	return r
}

func cellString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// TablesReport is the result of `query tables`: the projection schema + examples an
// agent reads before writing SQL.
type TablesReport struct {
	Tables string `json:"tables"`
}

//specue:req:query-graph#schema-is-discoverable
func runTables() TablesReport { return TablesReport{Tables: query.Tables()} }

func (r TablesReport) renderHuman(w io.Writer) error {
	_, err := fmt.Fprint(w, r.Tables)
	return err
}

func (r TablesReport) jsonValue() any { return r }
