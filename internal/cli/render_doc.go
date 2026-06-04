package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/render"
	"github.com/specue/specue/internal/render/jsonir"
	"github.com/specue/specue/internal/render/markdown"
)

// renderFormat picks the per-format Renderer. The output kind differs (text
// vs JSON), so this is render's own flag rather than the global --json switch
// (which means "machine-readable summary of the verb's result"). Adding a new
// format means adding one case and one package.
const (
	formatMarkdown = "markdown"
	formatJSON     = "json"
)

// newRenderCmd wires `specue render <dir>`: rebuilds the graph, reads each
// module's HEAD sha for frontmatter, and writes a self-contained markdown
// documentation tree under <dir>. Refuses to write into a non-empty directory.
func newRenderCmd(g *Globals, out, err io.Writer, code *int) *cobra.Command {
	var (
		format         string
		stripPrefix    string
		layout                string
		fmShape               string
		space                 string
		navSnippet            string
		withIndexPages        bool
		withTagsPage          bool
		withStatusAdmonitions bool
	)
	cmd := &cobra.Command{
		Use:   "render <dir>",
		Short: "Render the spec as a documentation tree (markdown or JSON IR)",
		Long: "render writes one file per node plus a root index. The default markdown\n" +
			"format emits .md files (frontmatter + describe-style body) with a README;\n" +
			"--format json emits .json files with an index.json — the JSON IR a\n" +
			"downstream tool consumes without re-parsing markdown. The destination must\n" +
			"not exist or must be empty — the verb never overwrites siblings.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(g, out, err, code, func(ctx Context) (any, *Problem) {
				return runRender(ctx, args[0], format, markdown.Config{
					StripPrefix:          stripPrefix,
					Layout:               layout,
					Frontmatter:          fmShape,
					Space:                space,
					NavSnippetPath:       navSnippet,
					WithIndexPages:       withIndexPages,
					WithTagsPage:         withTagsPage,
					WithStatusAdmonitions: withStatusAdmonitions,
				})
			})
		},
	}
	cmd.Flags().StringVar(&format, "format", formatMarkdown,
		"output format: markdown | json")
	cmd.Flags().StringVar(&stripPrefix, "strip-prefix", "",
		"strip this prefix from module paths in directory layout and link text")
	cmd.Flags().StringVar(&layout, "layout", "",
		"layout: flat | tree (default: flat)")
	cmd.Flags().StringVar(&fmShape, "frontmatter", "",
		"frontmatter shape: full | minimal | mark | mkdocs | none (default: full)")
	cmd.Flags().StringVar(&space, "space", "",
		"Confluence Space key (used with --frontmatter mark)")
	cmd.Flags().StringVar(&navSnippet, "nav-snippet", "",
		"path inside the destination to write an MkDocs-compatible nav.yml")
	cmd.Flags().BoolVar(&withIndexPages, "with-index-pages", false,
		"emit an index.md at every tree directory (mkdocs navigation.indexes); requires --layout tree")
	cmd.Flags().BoolVar(&withTagsPage, "with-tags-page", false,
		"emit a tags.md grouping nodes by type and status (for MkDocs Material's tags plugin)")
	cmd.Flags().BoolVar(&withStatusAdmonitions, "with-status-admonitions", false,
		"emit a Material admonition with the node's status at the top of every UseCase/Need/ADR/Plan page; mark each Requirement and Invariant inline with its own status")
	return cmd
}

// RenderReport is the typed result of `specue render <dir>`: where the tree
// landed and how many files were written. The renderer prints/jsons it.
type RenderReport struct {
	Destination string
	Files       int
}

// runRender builds the graph, reads each module's HEAD sha for frontmatter,
// renders the tree through the default markdown renderer, and writes it under
// destDir. Destination must be empty or not yet exist.
//
//specue:req:render-doc
//specue:req:render-doc#destination-is-explicit
func runRender(ctx Context, destArg, format string, mdCfg markdown.Config) (RenderReport, *Problem) {
	if strings.TrimSpace(destArg) == "" {
		p := Errorf("pass the destination as `specue render <dir>`",
			"render needs an explicit destination directory")
		return RenderReport{}, &p
	}
	if p := validateMarkdownConfig(mdCfg); p != nil {
		return RenderReport{}, p
	}
	renderer, p := pickRenderer(format, mdCfg)
	if p != nil {
		return RenderReport{}, p
	}
	dest, err := filepath.Abs(destArg)
	if err != nil {
		p := Errorf("check the destination path is valid",
			"resolve %q: %v", destArg, err)
		return RenderReport{}, &p
	}
	if p := ensureEmptyDestination(dest); p != nil {
		return RenderReport{}, p
	}

	res, p := buildGraph(ctx)
	if p != nil {
		return RenderReport{}, p
	}
	revs, p := moduleRevisions(ctx)
	if p != nil {
		return RenderReport{}, p
	}

	tree, err := renderer.Render(render.Input{
		Graph:     res.Graph,
		Revisions: revs,
	})
	if err != nil {
		p := Errorf("this is an internal error — re-run; if it persists, report it",
			"render: %v", err)
		return RenderReport{}, &p
	}
	// Two-phase collision handling. DetectCollisions describes every
	// node-vs-folder name conflict; passed to EmitIndexPages so it skips the
	// dirs a node will fill. nav-snippet and tags-page are emitted next while
	// the tree still has the original `<dir>/<slug>.md` files — their
	// references match the layout. Apply runs LAST: it moves each colliding
	// file and walks the whole tree (including the just-emitted nav-snippet
	// and tags.md) rewriting cross-file references in one pass.
	collisions := markdown.DetectCollisions(tree)
	// Index pages first: nav-snippet must see them in the tree so it can attach
	// each directory's index.md as the section's landing page (Material's
	// navigation.indexes feature). Reversing the order leaves the snippet
	// without index entries even when the files exist on disk.
	ctxRender := render.Context{Graph: res.Graph, Revisions: revs, Layout: renderer.Layout()}
	if entries, ok := markdown.EmitIndexPages(renderer, tree, ctxRender, collisions); ok {
		for p, c := range entries {
			tree[p] = c
		}
	}
	if path, content, ok := markdown.EmitNavSnippet(renderer, tree); ok {
		tree[path] = content
	}
	if path, content, ok := markdown.EmitTagsPage(renderer, tree, ctxRender, collisions); ok {
		tree[path] = content
	}
	collisions.Apply(tree)

	if err := writeTree(dest, tree); err != nil {
		p := Errorf("check the destination is writable",
			"write tree under %s: %v", dest, err)
		return RenderReport{}, &p
	}
	return RenderReport{Destination: dest, Files: len(tree)}, nil
}

// ensureEmptyDestination refuses a destination that already holds content. A
// path that does not exist is fine — it is created. A non-directory at the path
// is an error. The presence of any sibling refuses the run; the renderer never
// overwrites unrelated files.
//
//specue:req:render-doc#refuses-non-empty-destination
func ensureEmptyDestination(dest string) *Problem {
	info, err := os.Stat(dest)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		p := Errorf("check the destination path is accessible",
			"stat %s: %v", dest, err)
		return &p
	}
	if !info.IsDir() {
		p := Errorf("pass a directory path, not a file",
			"%s is a file, not a directory", dest)
		return &p
	}
	entries, err := os.ReadDir(dest)
	if err != nil {
		p := Errorf("check the destination path is accessible",
			"read %s: %v", dest, err)
		return &p
	}
	if len(entries) > 0 {
		p := Errorf("pass an empty (or not-yet-existing) directory — the renderer refuses to overwrite siblings",
			"%s is not empty (%d entries)", dest, len(entries))
		return &p
	}
	return nil
}

// moduleRevisions reads each workspace module's git HEAD sha (frontmatter
// `rendered_from`). A module with no commits yet gets the empty string — the
// frontmatter omits the field rather than failing the whole render.
func moduleRevisions(ctx Context) (map[model.ModulePath]string, *Problem) {
	work, dirs, p := ctx.workspace()
	if p != nil {
		return nil, p
	}
	git := plan.NewGit(gitBin())
	out := make(map[model.ModulePath]string, len(work.Modules))
	for _, wm := range work.Modules {
		root, err := git.RepoRoot(dirs[wm.Path])
		if err != nil {
			p := Errorf("ensure each module sits inside a git repository",
				"locate repo for %s: %v", wm.Path, err)
			return nil, &p
		}
		sha, err := git.Head(root)
		if err != nil {
			out[wm.Path] = "" // a repo with no commits has no HEAD
			continue
		}
		out[wm.Path] = sha
	}
	return out, nil
}

// writeTree writes every (path, content) pair under dest, creating directories
// as needed.
func writeTree(dest string, tree render.Tree) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	for rel, content := range tree {
		full := filepath.Join(dest, filepath.FromSlash(string(rel)))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// pickRenderer maps the --format flag to its Renderer. Unknown values produce
// an actionable error rather than silently falling back, so a typo never
// produces the wrong shape on disk.
//
//specue:req:render-doc#format-is-explicit
func pickRenderer(format string, cfg markdown.Config) (*render.Renderer, *Problem) {
	switch format {
	case "", formatMarkdown:
		return markdown.New(cfg), nil
	case formatJSON:
		return jsonir.Default(), nil
	default:
		p := Errorf("pass --format markdown or --format json",
			"unknown render format %q", format)
		return nil, &p
	}
}

// validateMarkdownConfig rejects unknown enum values up front, so a typo never
// silently degrades into the default shape.
func validateMarkdownConfig(cfg markdown.Config) *Problem {
	switch cfg.Layout {
	case "", markdown.LayoutFlat, markdown.LayoutTree:
	default:
		p := Errorf("pass --layout flat or --layout tree",
			"unknown layout %q", cfg.Layout)
		return &p
	}
	if cfg.WithIndexPages && cfg.Layout != markdown.LayoutTree {
		p := Errorf("use --layout tree with --with-index-pages, or drop the flag",
			"--with-index-pages requires --layout tree")
		return &p
	}
	switch cfg.Frontmatter {
	case "",
		markdown.FrontmatterFull,
		markdown.FrontmatterMinimal,
		markdown.FrontmatterMark,
		markdown.FrontmatterMkDocs,
		markdown.FrontmatterNone:
	default:
		p := Errorf("pass one of: full | minimal | mark | mkdocs | none",
			"unknown frontmatter shape %q", cfg.Frontmatter)
		return &p
	}
	return nil
}

func (r RenderReport) renderHuman(w io.Writer) error {
	_, err := fmt.Fprintf(w, "✓ rendered %d file(s) under %s\n", r.Files, r.Destination)
	return err
}

func (r RenderReport) jsonValue() any {
	return map[string]any{"destination": r.Destination, "files": r.Files}
}
