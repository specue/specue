---
title: The OCI registry that warms the cue cache runs in-process
icon: material/gavel
tags:
    - adr
    - accepted
---

# The OCI registry that warms the cue cache runs in-process

!!! note "Accepted"
    This decision is in effect.

The registry that hosts the schema and the landscape's modules for the
editor's cue lsp is brought up inside the tool's own process from the
cuelabs library, not as a separate daemon. It lives only long enough to
publish and warm the cache, then exits. The tool does not depend on a
development-only CLI it cannot control, and the editor sees a populated
on-disk cache it can resolve from without anything alive in the background.
