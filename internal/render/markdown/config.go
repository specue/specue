package markdown

// Config tunes the markdown renderer for the three customer-facing scenarios
// we ship: plain GitHub-style docs (zero config), kovetskiy/mark for
// Confluence, and MkDocs Material. The zero value reproduces the historical
// Default() behaviour byte-for-byte; every field is an opt-in.
//
//   - StripPrefix: removed from module paths when computing directory names
//     AND from visible identifiers in markdown body / link text. The CUE
//     module identifier in frontmatter keeps its full form. A prefix that
//     does not match a particular module is a no-op for that module.
//   - Layout: "flat" keeps the historical single-directory-per-module shape
//     (slashes collapsed to dashes). "tree" splits the module path on `/`
//     into nested directories, dropping the trailing `@vN` from the leaf.
//   - Frontmatter: "full" (default), "minimal", "mark" (kovetskiy/mark for
//     Confluence — PascalCase keys), "mkdocs" (MkDocs Material), "none".
//   - Space: kovetskiy/mark Space key. Empty value omits the field — pair
//     with Frontmatter="mark".
//   - NavSnippetPath: when non-empty, the renderer emits a YAML file at this
//     path holding an MkDocs-style `nav:` snippet over the rendered tree.
//     Path is interpreted relative to the destination directory.
//   - WithIndexPages: when true (requires Layout=tree), the renderer emits an
//     `index.md` at every directory level — root → leaf — instead of the
//     single root README.md. Matches MkDocs Material's `navigation.indexes`
//     feature so a section group lands on its own overview page.
//   - WithTagsPage: when true, the renderer emits a `tags.md` at the root
//     grouping every node by type and by status, with link labels carrying
//     the module-qualified id and a status badge. MkDocs Material's tags
//     plugin recognises `tags.md` and uses it as the tags index page, so
//     the per-node tag pills link straight into our sections by anchor.
//   - WithStatusAdmonitions: when true, every UseCase/Need/ADR/Plan page
//     opens (just after the H1) with a Material admonition block carrying
//     the node's status and a one-line summary, AND every Requirement /
//     Invariant / Variation / named Pre/Post gets an inline status marker.
type Config struct {
	StripPrefix          string
	Layout               string // "" | "flat" | "tree"
	Frontmatter          string // "" | "full" | "minimal" | "mark" | "mkdocs" | "none"
	Space                string // mark only
	NavSnippetPath       string // empty = no nav
	WithIndexPages       bool   // tree layout only
	WithTagsPage         bool
	WithStatusAdmonitions bool
}

// frontmatter shape constants.
const (
	FrontmatterFull    = "full"
	FrontmatterMinimal = "minimal"
	FrontmatterMark    = "mark"
	FrontmatterMkDocs  = "mkdocs"
	FrontmatterNone    = "none"
)

// layout names.
const (
	LayoutFlat = "flat"
	LayoutTree = "tree"
)
