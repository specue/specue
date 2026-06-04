# MkDocs template

A ready Material-for-MkDocs config for a `specue render` tree.
Copy these two paths next to your render and `mkdocs serve`.

## Render → serve

```sh
# 1. Render the tree (mkdocs-flavoured frontmatter + nav snippet).
specue render docs \
  --layout tree \
  --strip-prefix <your-module-prefix>/ \
  --frontmatter mkdocs \
  --with-index-pages \
  --with-tags-page \
  --with-status-admonitions \
  --nav-snippet nav.yml

# 2. Drop the template alongside it.
cp examples/mkdocs/mkdocs.yml ./mkdocs.yml
mkdir -p docs/assets && cp examples/mkdocs/assets/*.css docs/assets/

# 3. Install once, then serve.
uv tool install mkdocs --with mkdocs-material
mkdocs serve
```

Open <http://127.0.0.1:8000>.

## What does what

- `--frontmatter mkdocs` writes `tags:` per node — the Material **tags plugin**
  turns each into a pill that links to `/tags/#tag:<name>`.
- `--with-tags-page` emits `tags.md` with `{ #tag:... }` headings (needs the
  `attr_list` extension; the template includes it). Without it, Material's
  auto-tag index would show node URL slugs instead of titles.
- `--with-status-admonitions` writes a Material **admonition** at the top of
  every UC/Need/ADR/Plan and inline marks for each FR/invariant. Needs
  `admonition` + `pymdownx.details` (both in the template).
- `--with-index-pages` writes `index.md` at every directory so
  `navigation.indexes` turns each section header into a landing page.
- `--nav-snippet nav.yml` writes a nav block under `docs/nav.yml`; the template
  pulls it in with MkDocs' `INHERIT:` so re-rendering refreshes navigation
  without touching `mkdocs.yml`.

## Knobs

- **`site_name`** — change in `mkdocs.yml`.
- **Repo link / edit URI** — uncomment `repo_url`, `repo_name` in the template if
  you want the top-right GitHub/GitLab link.
- **Status colours** — `assets/status-badges.css` keys off the tag href
  substring; edit the palette there.
- **Dark backdrop** — drop `assets/dark-bg.css` from `extra_css` to keep
  Material's stock slate grey.

## Build for production

```sh
mkdocs build              # writes ./site/
```

The `site/` directory is self-contained and can be served from any static host.
