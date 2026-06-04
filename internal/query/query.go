package query

// Row is one result row, column name → value (string/int64/float64/nil — whatever
// the projection stored). A slice of these renders straight to JSON.
type Row map[string]any

// Query runs read-only SQL against the projection and returns the rows. The
// connection is already PRAGMA query_only, so a write is rejected by SQLite itself;
// a rejected or malformed query comes back as an error for the caller to surface
// with the fix.
func (d *DB) Query(sqlText string) ([]Row, error) {
	rows, err := d.sql.Query(sqlText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var out []Row
	for rows.Next() {
		cells := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range cells {
			ptrs[i] = &cells[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(Row, len(cols))
		for i, c := range cols {
			row[c] = normalize(cells[i])
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Tables returns the discovery text for `query tables`: the schema and examples an
// agent reads before writing SQL.
func Tables() string { return tablesDoc }

// normalize turns driver scan values into JSON-friendly ones — []byte (SQLite text)
// becomes string; the rest pass through.
func normalize(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}
