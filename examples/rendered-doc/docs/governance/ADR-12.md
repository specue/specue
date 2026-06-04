---
title: Render serves two formats — markdown (with knobs) and JSON IR — not a family of presets
icon: material/gavel
tags:
    - adr
    - accepted
---

# Render serves two formats — markdown (with knobs) and JSON IR — not a family of presets

!!! note "Accepted"
    This decision is in effect.

A single rendered tree cannot satisfy every downstream pipeline: GitHub
preview wants relative-link markdown, Confluence-via-mark wants a
PascalCase frontmatter (Title/Space/Parent/Labels), MkDocs Material wants
lowercase keys plus a nav: snippet, and custom dashboards want structured
data, not text.

Render carries two formats and one tunable markdown renderer:

- `--format markdown` (default) is the publishing target. Knobs:
  `--layout flat|tree`, `--strip-prefix <s>`, `--frontmatter
  full|minimal|mark|mkdocs|none`, `--space <key>`, `--nav-snippet <file>`.
  One renderer, one body, many YAML/path projections.
- `--format json` emits the structured graph (per-node JSON + index.json)
  for callers that do their own rendering — a custom Confluence push, an
  analytics dashboard, a markdown formatter that does not exist yet.

Rejected: a preset-per-package. Mark and mkdocs differ almost entirely
in frontmatter shape; >90% would be shared via composition, at the cost
of a class hierarchy that obscures the tiny real difference. Flags are
the smaller surface and stay open to future targets without a rename.

Rejected: ship only JSON IR and let every consumer render markdown
themselves. GitHub / GitLab preview .md files in-repo with no pipeline —
that is the cheap-yet-honest publish channel for a self-spec'd tool.
