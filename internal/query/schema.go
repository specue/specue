// Package query projects a resolved spec graph into an in-memory SQLite database so
// callers can navigate and search it with SQL — one query answers what would
// otherwise be a chain of small calls (cheaper in agent turns), and arbitrary
// cross-joins need no new Go verb. The projection is a NAVIGATION index, not a full
// replica: it carries what is needed to find nodes, walk edges, and filter a walk;
// the full node body (contract prose, per-element bindings) is fetched via
// describe/bindings once a slug is found. The database is read-only to callers
// (PRAGMA query_only) and rebuilt from the graph, never mutated in place.
package query

// ddl is the projection schema. Edges are split BY TYPE rather than one table with
// a discriminator, so each carries exactly its own filter columns (no NULL noise):
// a dep edge has none, an infra edge has role/element, satisfies has the atom. The
// schema is small and discoverable (`query tables`), so an agent reads it first,
// then writes SQL — the standard DESCRIBE-then-query pattern.
const ddl = `
CREATE TABLE nodes (
	module     TEXT NOT NULL,
	slug       TEXT NOT NULL,
	id         TEXT NOT NULL,            -- canonical module:slug
	type       TEXT NOT NULL,            -- UseCase | Need | Port | Container | Domain | Plan | ADR
	status     TEXT NOT NULL,            -- proven | implemented | asserted | blocked | broken | covered | partial | uncovered
	title      TEXT,
	visibility TEXT,                     -- public | private
	service    TEXT,                     -- UseCase: its service node id
	domain     TEXT,                     -- Need: its domain node id
	port_kind  TEXT,                     -- Port: channel | rpc | rest | datastore
	transport  TEXT,                     -- Port: open label (kafka, grpc, ...)
	PRIMARY KEY (module, slug)
);

-- A plain contract dependency (UseCase -> UseCase). core = from a core element
-- (not a branch/variation), which is what drives blocked-status propagation.
CREATE TABLE dep_edges (
	from_id TEXT NOT NULL,
	to_id   TEXT NOT NULL,
	core    INTEGER NOT NULL            -- 1 if from a core element
);

-- An infrastructure touch (UseCase -> Port/Container) with its role and the element
-- it sits on. role is the common filter ("who consumes X").
CREATE TABLE infra_edges (
	from_id TEXT NOT NULL,
	to_id   TEXT NOT NULL,
	role    TEXT NOT NULL,              -- produce | consume | serve | call | read | write | ...
	element TEXT                        -- the named element it sits on ("" = whole-contract)
);

-- A UseCase discharges a Need atom (FR/NFR). atom is navigational
-- ("which UCs satisfy fr-01").
CREATE TABLE satisfies (
	uc_id   TEXT NOT NULL,
	need_id TEXT NOT NULL,
	atom    TEXT NOT NULL
);

-- Derived: a UseCase realizes a Need (it satisfies at least one of its atoms).
CREATE TABLE realizes (
	uc_id   TEXT NOT NULL,
	need_id TEXT NOT NULL
);

-- A named element (invariant/variation) of a UseCase — addressable by a scoped
-- binding (slug#element).
CREATE TABLE elements (
	node_id TEXT NOT NULL,
	element TEXT NOT NULL,
	kind    TEXT NOT NULL,              -- pre | post | invariant | variation
	text    TEXT
);

-- An atom owned by a Need.
CREATE TABLE atoms (
	need_id TEXT NOT NULL,
	atom    TEXT NOT NULL,
	kind    TEXT NOT NULL,              -- fr | nfr
	text    TEXT
);

-- A code annotation that bound nothing valid — it resolved to no node (orphan) or
-- to a type that holds no code (unbindable, e.g. a Need). Hangs on no node, so
-- it lives here, not in bindings; this is why nodes can be empty while orphans are
-- not (a code module with annotations but no resolvable contracts).
CREATE TABLE orphans (
	module TEXT NOT NULL,               -- the code module that carried the annotation
	slug   TEXT NOT NULL,               -- the (dangling) target slug
	reason TEXT NOT NULL,               -- orphan-binding | unbindable-target
	loc    TEXT                         -- file:line (clickable)
);

-- A resolved code binding on a node/element: where it is and what kind.
CREATE TABLE bindings (
	node_id       TEXT NOT NULL,
	element       TEXT,                  -- "" = whole-contract
	bind_kind     TEXT NOT NULL,         -- req | test | produce | consume | ...
	loc           TEXT NOT NULL,         -- file:line (clickable)
	source_module TEXT NOT NULL          -- the code module that carried the annotation
);

-- porter stemmer so a search for "single-verdict" matches "idempotently", "apply"
-- matches "applies" — recall over exact token match (the whole point of FTS here).
CREATE VIRTUAL TABLE nodes_fts USING fts5(id, slug, title, body, tokenize = 'porter unicode61');

-- --- views: pre-joined cuts an agent reaches for often ----------------------

-- node_describe: one row per (node, element-or-empty). For a UseCase: one row
-- per element; for any other node a single row with element = ''. Encodes what
-- 'describe' prints in row form, so a query can read the whole node — header
-- plus elements — in one statement.
CREATE VIEW node_describe AS
  SELECT n.id, n.type, n.status, n.title,
         e.element, e.kind AS element_kind, e.text AS element_text
  FROM nodes n
  LEFT JOIN elements e ON e.node_id = n.id
  UNION ALL
  SELECT n.id, n.type, n.status, n.title,
         a.atom AS element, a.kind AS element_kind, a.text AS element_text
  FROM nodes n
  JOIN atoms a ON a.need_id = n.id;

-- fr_coverage: one row per Need atom + every UC that satisfies it + that UC's
-- status. Multiple UCs per atom yield multiple rows; an atom with no satisfier
-- shows up with NULLs in the uc_* columns (LEFT JOIN preserves the gap). The
-- single most-asked question — "is fr-NN proven, by whom?" — is one SELECT.
CREATE VIEW fr_coverage AS
  SELECT a.need_id, a.atom, a.kind AS atom_kind, a.text AS atom_text,
         s.uc_id, n.status AS uc_status
  FROM atoms a
  LEFT JOIN satisfies s ON s.need_id = a.need_id AND s.atom = a.atom
  LEFT JOIN nodes n ON n.id = s.uc_id;
`

// tablesDoc is the human/agent-facing discovery text for `query tables`: the schema
// plus a few worked examples, so the first call teaches the shape and the dialect.
const tablesDoc = `Tables (a navigation projection of the spec graph; query them with SQL):

  nodes(module, slug, id, type, status, title, visibility, service, domain, port_kind, transport)
  dep_edges(from_id, to_id, core)            -- UseCase -> UseCase contract deps
  infra_edges(from_id, to_id, role, element) -- UseCase -> Port/Container infra touches
  satisfies(uc_id, need_id, atom)            -- a UC discharges a Need atom
  realizes(uc_id, need_id)                   -- derived: UC realizes a Need
  elements(node_id, element, kind, text)     -- named invariants/variations
  atoms(need_id, atom, kind, text)           -- Need FR/NFR
  bindings(node_id, element, bind_kind, loc, source_module)  -- loc = file:line
  orphans(module, slug, reason, loc)         -- annotations that bound nothing (no node / unbindable)
  nodes_fts(id, slug, title, body)           -- FTS5 full-text; use: WHERE nodes_fts MATCH 'term'

Views (pre-joined cuts you reach for often):
  node_describe(id, type, status, title, element, element_kind, element_text)
                                             -- one row per node element / atom; what describe prints
  fr_coverage(need_id, atom, atom_kind, atom_text, uc_id, uc_status)
                                             -- per Need atom: every UC that satisfies it + its status

ids are canonical "module:slug". The projection is read-only.

Examples:
  -- everything with no code yet
  SELECT id, type FROM nodes WHERE status = 'asserted';

  -- who consumes a port (one hop)
  SELECT n.id FROM infra_edges e JOIN nodes n ON n.id = e.from_id
  WHERE e.to_id = 'topo:report-channel' AND e.role = 'consume';

  -- transitive dependents of a contract (blast radius)
  WITH RECURSIVE up(id) AS (
    SELECT 'example:validate-graph'
    UNION SELECT from_id FROM dep_edges JOIN up ON to_id = up.id
  ) SELECT id FROM up WHERE id != 'example:validate-graph';

  -- which UCs satisfy a Need's fr-01, and are they proven
  SELECT n.id, n.status FROM satisfies s JOIN nodes n ON n.id = s.uc_id
  WHERE s.need_id = 'example:describe-node' AND s.atom = 'fr-01';

  -- full-text search
  SELECT id, title FROM nodes_fts WHERE nodes_fts MATCH 'idempotent';

  -- one node with its elements in one statement
  SELECT element, element_kind, element_text FROM node_describe
  WHERE id = 'example:validate-graph' AND element != '';

  -- coverage trace for a single FR — every UC that satisfies it + its status
  SELECT uc_id, uc_status FROM fr_coverage
  WHERE need_id = 'example:as-agent-setup' AND atom = 'fr-01';

Recipes (the higher-level views — substitute the scope literal):

  -- MAP: the whole landscape at a glance — Need coverage per domain,
  -- UC implementation per service. One call instead of reading every node.
  SELECT domain, status, count(*) n FROM nodes
  WHERE type = 'Need' GROUP BY domain, status ORDER BY domain;
  SELECT service, status, count(*) n FROM nodes
  WHERE type = 'UseCase' GROUP BY service, status ORDER BY service;

  -- DIGEST: one area in depth. Its nodes...
  SELECT id, type, status FROM nodes WHERE service = 'example:specue' OR id = 'example:specue';
  -- ...plus its one-hop closure (what it depends on / touches):
  SELECT DISTINCT to_id FROM dep_edges WHERE from_id IN
    (SELECT id FROM nodes WHERE service = 'example:specue')
  UNION SELECT DISTINCT to_id FROM infra_edges WHERE from_id IN
    (SELECT id FROM nodes WHERE service = 'example:specue');

  -- COVERAGE GAP: a Need's atoms no proven UC discharges (the honest gap).
  SELECT a.atom, a.text FROM atoms a
  WHERE a.need_id = 'example:describe-node' AND a.atom NOT IN (
    SELECT s.atom FROM satisfies s JOIN nodes n ON n.id = s.uc_id
    WHERE s.need_id = a.need_id AND n.status = 'proven');

  -- DANGLING CODE: annotations that bound nothing (audit a code module).
  SELECT slug, reason, count(*) n, group_concat(loc, ' ') locs
  FROM orphans GROUP BY slug, reason ORDER BY n DESC;
`
