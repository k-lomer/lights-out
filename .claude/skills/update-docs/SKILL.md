---
name: update-docs
description: Update README.md and ARCHITECTURE.md to reflect larger code changes. Use after adding/removing a DNO client, changing the request path or handler behaviour, altering query params or the response shape, adding a package, or any change that makes the existing docs inaccurate.
---

# Update docs

Bring [README.md](../../../README.md) and [ARCHITECTURE.md](../../../ARCHITECTURE.md) back in sync with the current code after a larger change. CLAUDE.md is out of scope here — only touch it if a convention genuinely changed.

## Steps

1. **Find what changed.** Determine the diff scope:
   - If on a feature branch: `git diff main...HEAD --stat` then inspect the relevant diffs.
   - Otherwise review uncommitted work (`git status`, `git diff`) and recent commits (`git log --oneline -10`).
   Focus on changes that affect documented behaviour (see checklist below).

2. **Read the current docs** in full before editing — `README.md` and `ARCHITECTURE.md` — so updates are surgical and the existing structure/voice is preserved.

3. **Verify every claim against the code, don't guess.** For each section that might be affected, open the real source. Never document a command, param, field, or status code you haven't confirmed exists. The docs are explicitly anti-hallucination — keep them that way.

4. **Edit only what's stale.** Prefer targeted edits over rewrites. Keep the established tone: README is for a new developer (runnable examples, accurate); ARCHITECTURE is the design mental model with real `file_path` links and the mermaid flow diagram.

5. **Check the cross-cutting things** that larger changes commonly break (see checklist).

6. **Report** a short summary of what you changed and why, and flag anything you were unsure about rather than silently guessing. Do not commit unless asked.

## What to check against the code

- **DNOs** — added/removed? Update the supported-DNO list (README), the `NewDnoClients` / `AllDnoList` registration story (ARCHITECTURE), the per-DNO flag list, and the architecture diagram's client nodes.
- **Endpoint & query params** — new/changed/removed params? Update the README parameter table, the opt-out targeting description, and the example `curl`s.
- **Response shape** — changes to `model.Outage` struct tags? Update the README response JSON (field names must match the tags exactly) and the request/response summary in ARCHITECTURE.
- **Request path / handler** — changed fan-out, sorting, filtering, pagination, or the partial-failure rule (500 only when all DNOs fail)? Update the ARCHITECTURE flow description, diagram, and the README status-code table.
- **Packages** — package added/removed/repurposed? Update the package breakdown and dependency-direction note in ARCHITECTURE and the project-layout block in README. Note especially whether `cache` is now wired into the request path (currently documented as "not wired in").
- **Build/test/run** — Makefile targets or the listen address (`:8080`) changed? Update README commands.
- **Prerequisites** — Go version (`go.mod`) or tooling changed? Update README prerequisites.

## Conventions to preserve

- Keep file references as markdown links to real paths.
- Keep the mermaid diagram in ARCHITECTURE.md accurate, not just the prose.
- Match the codebase's own doc style (see CLAUDE.md): full-sentence prose in the docs themselves; don't invent sections like "Tips" or "Support" that aren't grounded in the code.
