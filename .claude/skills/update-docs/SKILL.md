---
name: update-docs
description: Update README.md, ARCHITECTURE.md, and CLAUDE.md to reflect larger code changes. Use after adding/removing a DNO client, changing the request path or handler behaviour, altering query params or the response shape, adding a package, or any change that makes the existing docs inaccurate.
---

# Update docs

Bring [README.md](../../../README.md), [ARCHITECTURE.md](../../../ARCHITECTURE.md), and [CLAUDE.md](../../../CLAUDE.md) back in sync with the current code after a larger change. For CLAUDE.md, correct anything that has become factually wrong (named functions/constructors, example test names, the conventions-and-gotchas notes) but leave its structure and guidance intact — don't rewrite instructions, only fix stale facts.

## Steps

1. **Find what changed.** Determine the diff scope:
   - If on a feature branch: `git diff main...HEAD --stat` then inspect the relevant diffs.
   - Otherwise review uncommitted work (`git status`, `git diff`) and recent commits (`git log --oneline -10`).
   Focus on changes that affect documented behaviour (see checklist below).

2. **Read the current docs** in full before editing — `README.md`, `ARCHITECTURE.md`, and `CLAUDE.md` — so updates are surgical and the existing structure/voice is preserved.

3. **Verify every claim against the code, don't guess.** For each section that might be affected, open the real source. Never document a command, param, field, or status code you haven't confirmed exists. The docs are explicitly anti-hallucination — keep them that way.

4. **Edit only what's stale.** Prefer targeted edits over rewrites. Keep the established tone: README is for a new developer (runnable examples, accurate); ARCHITECTURE is the design mental model with real `file_path` links and the mermaid flow diagram.

5. **Check the cross-cutting things** that larger changes commonly break (see checklist).

6. **Report** a short summary of what you changed and why, and flag anything you were unsure about rather than silently guessing. Do not commit unless asked.

## What to check against the code

Each item below names a *kind* of change, the code to read to learn the current truth, and the doc surfaces to update — deliberately without restating the current values, so this file can't itself go stale.

- **DNOs** — added/removed? Read the DNO registration points (the client map in the composition root and the canonical DNO list in `model`). Update the supported-DNO list and per-DNO flag list (README), the registration story and the diagram's client nodes (ARCHITECTURE), and the "Adding a DNO" note (CLAUDE.md).
- **Endpoint & query params** — new/changed/removed? Read the query-param parser and the route registration. Update the README parameter table, the opt-out targeting description, and the example `curl`s.
- **Response shape** — fields added/removed/renamed on the canonical outage type? Copy the JSON tags from the source exactly. Update the README response JSON and the request/response summary in ARCHITECTURE.
- **Request path / handler** — changed fan-out, sorting, filtering, pagination, caching, or the partial-failure rule? Read the handler to confirm current behaviour. Update the ARCHITECTURE flow prose + mermaid diagram and the README status-code table.
- **Packages** — package added/removed/repurposed, or its role in the request path changed? Confirm the actual dependency direction and each package's job from the code. Update the package breakdown and dependency-direction note (ARCHITECTURE), the project-layout block (README), and any package-specific gotcha (CLAUDE.md).
- **Build/test/run** — Makefile targets or the listen address changed? Read the Makefile and the server setup. Update the README commands.
- **Prerequisites** — Go version (`go.mod`) or required tooling changed? Update the README prerequisites.
- **CLAUDE.md references** — grep for any identifier the change touched (functions, constructors, test-name examples, package/type names) and fix references that no longer resolve. Fix facts only; leave the guidance and structure intact.

## Conventions to preserve

- Keep file references as markdown links to real paths.
- Keep the mermaid diagram in ARCHITECTURE.md accurate, not just the prose.
- Match the codebase's own doc style (see CLAUDE.md): full-sentence prose in the docs themselves; don't invent sections like "Tips" or "Support" that aren't grounded in the code.
- When editing this skill, describe *what to check and where*, not the current values — point at the code to read rather than restating a fact that will drift.
