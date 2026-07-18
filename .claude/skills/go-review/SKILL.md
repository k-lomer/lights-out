---
name: go-review
description: Perform a full senior-Go-reviewer pass over this repository (architecture, idiomatic Go, concurrency, correctness, security, tests) and write a prioritized findings report to a file. Use when asked to "review the repo", "do a full code review", or "run the Go review".
---

# Go review

Act as a senior Go reviewer performing a full review of this repository.

Module: `github.com/k-lomer/lights-out` — an HTTP service that polls six UK
electricity DNOs (Distribution Network Operators) for power-cut data, normalizes
each provider's bespoke JSON into a canonical model, and serves it from
`GET /list`. Dependency direction is one-way: `cmd → handlers → clients → model`,
with a standalone cache package. Consult [README.md](../../../README.md),
[ARCHITECTURE.md](../../../ARCHITECTURE.md), and [CLAUDE.md](../../../CLAUDE.md)
for the intended design and documented behaviour before judging it — and flag
where the code and these docs disagree.

## Ground rules

- **Read the actual code before claiming anything.** Every finding must cite a
  real `file:line` you have opened. Do not infer behaviour from names or from
  this skill — if you haven't read it, don't report it.
- **Quote the offending lines.** If a claim depends on runtime behaviour you
  can't verify statically, run it: `make test`, `go test -race ./...`,
  `go vet ./...`, `golangci-lint run`.
- **Don't flag deliberate, documented quirks** unless you can show they're
  implemented incorrectly: SPEnergy's `InsecureSkipVerify` (incomplete cert
  chain) + two-step count-then-fetch; UK Power Network's spoofed
  `User-Agent`/`Accept`; National Grid's `1900-01-01` sentinel → nil `*time.Time`;
  opt-out DNO targeting; partial-success returns 200, all-fail returns 500.
- **No speculative refactors.** Only report what the code as written actually
  needs. If something is fine, say so briefly rather than inventing concerns.

## Review scope

1. **Architecture & package design**
   - Are the model / clients / cache / handlers / cmd boundaries clean, or is
     there leaky abstraction or coupling against the one-way direction?
   - Is `DnoClient` ([clients/dno_client.go](../../../clients/dno_client.go)) the
     right seam, or do per-DNO clients duplicate logic that should be shared
     (`drainAndClose`, request setup, error wrapping)?
   - Is the cache package's concurrency model sound on its own terms?
   - Should hardcoded config in [cmd/main.go](../../../cmd/main.go) (timeouts,
     TLS skip-verify, server addr/port) be externalized? Justify cost/benefit
     rather than asserting.

2. **Idiomatic Go / correctness under concurrency**
   - Error handling: wrapping with `%w`, sentinel vs. typed errors, swallowed
     errors. Assess every `//nolint:errcheck` (e.g. in `drainAndClose`) — is it
     justified at each call site?
   - Context propagation: is `ctx` threaded into each DNO client's HTTP calls and
     respected for cancellation/timeout? Flag any `http.Get` / request built
     without the request context.
   - Handler fan-out: goroutine leaks, races, per-index slice writes vs. shared
     state. Back this with `go test -race` output.
   - Interface minimalism and consumer-defined placement.

3. **Correctness & bugs**
   - Per-DNO JSON decoding (`clients/*.go`, `model/*.go`): nil derefs, off-by-one,
     silent failures when an upstream changes shape, custom `UnmarshalJSON` edge
     cases.
   - Postcode normalization/validation ([model/postcode.go](../../../model/postcode.go))
     and [model/outage.go](../../../model/outage.go).
   - Handler status codes ([handlers/list_handler.go](../../../handlers/list_handler.go)):
     confirm one DNO failure can't fail the whole request, and that the
     `recover()`/all-fail logic holds.
   - Resource leaks: unclosed/undrained response bodies, missing defers.

4. **Security**
   - Confirm `InsecureSkipVerify` is scoped to the single SPEnergy client and not
     reused by other clients or the default transport.
   - Input validation on user-supplied query params
     ([model/query_params.go](../../../model/query_params.go)): pagination
     bounds, postcode injection, unbounded values.
   - DoS hardening: server read/write/idle timeouts, outbound client timeouts.

5. **Tests**
   - Coverage gaps: which exported functions/branches lack tests? Prefer
     `go test -cover` / `-coverprofile` evidence over guessing.
   - Test quality: appropriate `assert` vs. `require`, and whether
     `test_helpers.go` / `test_client.go` mock real behaviour faithfully.

## Output

Write the review to a file named `claude_review_<datetime>.out` in the repo root,
where `<datetime>` is the current timestamp (e.g. `date +%Y%m%d_%H%M%S`).

- **Raw text, not Markdown.** No `#` headings, backticks, or bullet markup —
  plain prose and simple indentation only.
- **Hard-wrap every line at 150 characters.**
- A single prioritized list, grouped Critical → Should Fix → Nit. For each
  finding give: `file:line`; the issue with the quoted lines; why it matters
  (impact, not restatement); and a concrete fix (diff or precise change).
- End with a one-paragraph overall assessment. If a scope area has no real
  issues, say so in one line instead of padding it.

After writing the file, tell the user its path and give a two-or-three-line
summary of the headline findings.
