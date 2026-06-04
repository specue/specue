package query

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite" // pure-Go driver, no CGO

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

// DB is a read-only SQLite projection of a resolved graph. Close it when done.
type DB struct{ sql *sql.DB }

// Build projects the resolved graph (and its diagnostics) into a fresh in-memory
// SQLite database. The connection is opened, the schema created, every node and
// edge inserted, then the connection is flipped read-only (PRAGMA query_only) so a
// caller's SQL can never mutate it. Diagnostics carry the orphan/unbindable
// annotations — code that bound nothing — which hang on no node, so they project
// into the orphans table; this keeps query and the bindings view seeing the same
// truth. The graph is the source of truth; this is a disposable index rebuilt on
// demand.
//specue:req:query-graph#runs-against-projection
//specue:req:query-graph#cannot-mutate
//specue:req:query-graph#pre-joined-views
func Build(g *compiler.ResolvedGraph, diags []compiler.Diagnostic) (*DB, error) {
	sqldb, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := sqldb.Exec(ddl); err != nil {
		sqldb.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	if err := project(sqldb, g, diags); err != nil {
		sqldb.Close()
		return nil, err
	}
	if _, err := sqldb.Exec(`PRAGMA query_only = 1`); err != nil {
		sqldb.Close()
		return nil, fmt.Errorf("set read-only: %w", err)
	}
	return &DB{sql: sqldb}, nil
}

// Close releases the database.
func (d *DB) Close() error { return d.sql.Close() }

// project walks the graph once, inserting each node and its edges/elements/bindings,
// then projects the orphan/unbindable diagnostics into the orphans table.
func project(db *sql.DB, g *compiler.ResolvedGraph, diags []compiler.Diagnostic) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for n := range g.Nodes() {
		if err := insertNode(tx, n); err != nil {
			return err
		}
	}
	if err := insertOrphans(tx, diags); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("project graph: %w", err)
	}
	return nil
}

// insertOrphans projects the orphan/unbindable diagnostics — annotations that bound
// nothing — so query sees them too (they hang on no node, mirroring the bindings
// view's orphan rows).
func insertOrphans(tx *sql.Tx, diags []compiler.Diagnostic) error {
	for _, d := range diags {
		if d.Code != compiler.OrphanBinding && d.Code != compiler.UnbindableTarget {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO orphans (module, slug, reason, loc) VALUES (?,?,?,?)`,
			string(d.Node.Module), string(d.Node.Slug), string(d.Code),
			loc(d.Location.File, d.Location.Line)); err != nil {
			return err
		}
	}
	return nil
}

func insertNode(tx *sql.Tx, n *compiler.ResolvedNode) error {
	id := n.ID()
	node := n.Node()
	idStr := id.String()

	service, domain, portKind, transport := nodeFacets(node)
	if _, err := tx.Exec(`INSERT INTO nodes
		(module, slug, id, type, status, title, visibility, service, domain, port_kind, transport)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		string(id.Module), string(id.Slug), idStr, string(node.Type), string(n.Status),
		node.Title, string(node.Visibility), service, domain, portKind, transport); err != nil {
		return fmt.Errorf("insert node %s: %w", idStr, err)
	}
	if _, err := tx.Exec(`INSERT INTO nodes_fts (id, slug, title, body) VALUES (?,?,?,?)`,
		idStr, string(id.Slug), node.Title, ftsBody(node)); err != nil {
		return err
	}

	for _, dep := range n.CoreUses {
		if _, err := tx.Exec(`INSERT INTO dep_edges (from_id, to_id, core) VALUES (?,?,1)`,
			idStr, dep.String()); err != nil {
			return err
		}
	}
	// Non-core deps: those in Uses but not CoreUses (branch/variation deps).
	for _, dep := range diffNodeIDs(n.Uses, n.CoreUses) {
		if _, err := tx.Exec(`INSERT INTO dep_edges (from_id, to_id, core) VALUES (?,?,0)`,
			idStr, dep.String()); err != nil {
			return err
		}
	}
	for _, a := range n.Satisfies {
		if _, err := tx.Exec(`INSERT INTO satisfies (uc_id, need_id, atom) VALUES (?,?,?)`,
			idStr, a.Need.String(), string(a.Atom)); err != nil {
			return err
		}
	}
	for _, need := range n.Realizes {
		if _, err := tx.Exec(`INSERT INTO realizes (uc_id, need_id) VALUES (?,?)`,
			idStr, need.String()); err != nil {
			return err
		}
	}
	if err := insertElementsAtoms(tx, idStr, node); err != nil {
		return err
	}
	return insertBindings(tx, idStr, n)
}

// nodeFacets pulls the per-type columns (service/domain/port) from the body.
func nodeFacets(node model.Node) (service, domain, portKind, transport string) {
	b := node.Body
	if b == nil {
		return
	}
	switch {
	case b.UseCase != nil:
		service = b.UseCase.Service.String()
	case b.Need != nil:
		domain = b.Need.Domain.String()
	case b.Port != nil:
		portKind = string(b.Port.Kind)
		transport = string(b.Port.Transport)
	}
	return
}

// insertElementsAtoms records a UseCase's named elements (+ infra edges they carry)
// and a Need's atoms.
func insertElementsAtoms(tx *sql.Tx, idStr string, node model.Node) error {
	b := node.Body
	if b == nil {
		return nil
	}
	if b.UseCase != nil {
		for _, e := range b.UseCase.Elements {
			if e.Named() {
				if _, err := tx.Exec(`INSERT INTO elements (node_id, element, kind, text) VALUES (?,?,?,?)`,
					idStr, string(e.ID), string(e.Kind), e.Text); err != nil {
					return err
				}
			}
			for _, dep := range e.Deps {
				if dep.Role == "" {
					continue
				}
				if _, err := tx.Exec(`INSERT INTO infra_edges (from_id, to_id, role, element) VALUES (?,?,?,?)`,
					idStr, dep.To.String(), string(dep.Role), string(e.ID)); err != nil {
					return err
				}
			}
		}
	}
	if b.Need != nil {
		for _, a := range b.Need.Atoms {
			if _, err := tx.Exec(`INSERT INTO atoms (need_id, atom, kind, text) VALUES (?,?,?,?)`,
				idStr, string(a.ID), string(a.Kind), a.Text); err != nil {
				return err
			}
		}
	}
	return nil
}

// insertBindings records the resolved code bindings on a node: req, test, and infra
// proofs, each with its source code module and location.
func insertBindings(tx *sql.Tx, idStr string, n *compiler.ResolvedNode) error {
	ins := func(elem model.ElementID, kind string, bs []compiler.Binding) error {
		for _, b := range bs {
			if _, err := tx.Exec(`INSERT INTO bindings
				(node_id, element, bind_kind, loc, source_module) VALUES (?,?,?,?,?)`,
				idStr, string(elem), kind, loc(b.File, b.Line), string(b.SourceModule)); err != nil {
				return err
			}
		}
		return nil
	}
	for elem, bs := range n.ReqElems {
		if err := ins(elem, "req", bs); err != nil {
			return err
		}
	}
	for elem, bs := range n.CoverElems {
		if err := ins(elem, "test", bs); err != nil {
			return err
		}
	}
	for key, bs := range n.InfraProof {
		if err := ins(key.Element, string(key.Role), bs); err != nil {
			return err
		}
	}
	return nil
}

// ftsBody assembles the searchable text of a node: title plus every element's text
// (and atom text for a Need), so a full-text MATCH reaches the contract prose.
func ftsBody(node model.Node) string {
	var sb strings.Builder
	sb.WriteString(node.Title)
	b := node.Body
	if b == nil {
		return sb.String()
	}
	if b.UseCase != nil {
		for _, e := range b.UseCase.Elements {
			sb.WriteByte(' ')
			sb.WriteString(e.Text)
			if e.When != "" {
				sb.WriteByte(' ')
				sb.WriteString(e.When)
				sb.WriteByte(' ')
				sb.WriteString(e.Then)
			}
		}
	}
	if b.Need != nil {
		sb.WriteByte(' ')
		sb.WriteString(b.Need.Consumer)
		sb.WriteByte(' ')
		sb.WriteString(b.Need.Description)
		for _, a := range b.Need.Atoms {
			sb.WriteByte(' ')
			sb.WriteString(a.Text)
		}
	}
	return sb.String()
}

// loc renders a binding location as a clickable file:line.
func loc(file model.FilePath, line int) string {
	return fmt.Sprintf("%s:%d", file, line)
}

// diffNodeIDs returns the ids in all but not in sub (preserving order).
func diffNodeIDs(all, sub []model.NodeID) []model.NodeID {
	in := make(map[model.NodeID]bool, len(sub))
	for _, id := range sub {
		in[id] = true
	}
	var out []model.NodeID
	for _, id := range all {
		if !in[id] {
			out = append(out, id)
		}
	}
	return out
}
