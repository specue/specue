---
title: Code module scans from code_root; repo modules live under spec.d/
icon: material/gavel
tags:
    - adr
    - accepted
---

# Code module scans from code_root; repo modules live under spec.d/

!!! note "Accepted"
    This decision is in effect.

A code module in a repo root and a service module in a subfolder are, to
CUE's load resolver, two paths to the same package — registered both
standalone (workspace) and nested (code module's subtree). CUE refuses
that with ambiguous-import the moment another module imports the service.

Two changes together settle it.

`code_root` (manifest field, relative to spec.mod.cue, default ".") moves
where the scan begins. A code module in a subfolder points back at the
repo through code_root, so the manifest is separated from the source it
scans — and stops claiming sibling spec modules as its own subpackages.

`spec.d/<kind>/[<name>/]` is the recommended layout: code at spec.d/code/,
services at spec.d/service/<name>/, etc. Unix drop-in convention (cron.d,
sudoers.d), visible to agents and shells (no leading dot). `init --layout
spec.d --kind code` writes it and fills `code_root: "../.."`.

Recommended, not required: the older flat layout works via code_root
alone. Self-spec moves to spec.d/ as the living demo.
