# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

All common tasks go through the [Makefile](Makefile):

```sh
make         # default target (all): fmt, vet, lint, test, build
make run     # go run ./cmd  — serves on :8080
make build   # build binary to bin/lights-out
make test    # go clean -testcache && go test ./...
make fmt     # go fmt ./...
make vet     # go vet ./...
make lint    # golangci-lint run
make all     # fmt, vet, lint, test, build
```

**Validate changes with `make` (no args)** — the default target runs fmt, vet,
lint, test, and build. `make test` always clears the test cache first. To run a single package or test
directly, bypass the Makefile:

```sh
go test ./handlers/                       # one package
go test ./clients/ -run Test_UKPN         # one test (matches by name)
go test ./... -v                          # verbose
```

`golangci-lint` is required for `make lint` but not for building or testing.

## Architecture

A Go HTTP service that aggregates UK power-cut data from six DNOs (Distribution
Network Operators) and serves it from one endpoint, `GET /list`. See
[ARCHITECTURE.md](ARCHITECTURE.md) for the full picture; the essentials:

- **Dependency direction is one-way:** `cmd → handlers → clients → model`.
  [cmd/main.go](cmd/main.go) is the composition root — it builds the HTTP
  clients, constructs the `map[model.Dno]clients.DnoClient`, and injects it into
  the handler. Nothing else constructs dependencies.

- **Every DNO is an adapter behind one interface**,
  [`DnoClient`](clients/dno_client.go) (`ListOutages` + `GetDno`). A client
  decodes its provider's bespoke JSON into a per-DNO model type, then converts
  it to the canonical `model.Outage` via a `ToOutage(s)` method. The handler
  only ever sees `model.Outage`, so provider quirks stay inside the client +
  its model file.

- **The request path** ([handlers/list_handler.go](handlers/list_handler.go)):
  parse `QueryParams` → fan out one goroutine per targeted DNO (`sync.WaitGroup`,
  results written to per-index slices, no shared mutable state) → aggregate →
  **stable sort by `DNO_ID` key** → status filter → optional postcode filter →
  paginate.

- **Failure handling is deliberate:** each goroutine has a `recover()`, and a
  request returns `500` **only if every** targeted DNO fails. Any partial
  success returns `200`. Preserve this behaviour when editing the handler.

## Code style

- **Comments are full, capitalised sentences ending in a full stop** (e.g.
  `// Sort to ensure determinism.`). This applies to test-function doc comments
  too (`// Test the postcode filter.`).
- **Error strings are lowercase and have no trailing full stop**, per Go
  convention (e.g. `fmt.Errorf("unexpected return code from %s, %d", ...)`,
  `errors.New("no DNOs targeted")`). Wrap underlying errors with `%w`.
- **Tests use `testify`.** Default to `assert.*`; use `require.*` only when the
  test cannot meaningfully continue after the failure (e.g. checking `NoError`
  before using the returned value, or asserting status before decoding a body).
- **Test functions are named `Test_<Subject>_<Case>`** with underscores
  (e.g. `Test_ListHandler_PageSize`, `Test_OutageCache_GetExpired`). Each test has a
  one-line doc comment describing what it asserts.
- **Commit messages are capitalised, imperative, and have no trailing full
  stop** (e.g. `Handle panics in DNO clients`, `Standardise DNO model names`).
- Shared test helpers live in `handlers/test_helpers.go` /
  `handlers/test_client.go` — reuse them (`requireStatus`, `decodeOutages`,
  `addQueryParams`, `NewTestDnoClients`) rather than re-implementing setup.

## Conventions and gotchas

- **Adding a DNO** touches exactly: a new `clients/<dno>.go` implementing
  `DnoClient`; a new `model/<dno>.go` with the payload struct and `ToOutages`;
  and registration in both the `dnoClients` map in `main()`
  ([cmd/main.go](cmd/main.go)) and `model.AllDnoList`
  ([model/dno.go](model/dno.go)). The query-param flag for
  the DNO is derived automatically from `AllDnoList` — its name is the
  `model.Dno` string value (e.g. `UKPowerNetwork`).

- **Constructors use `Make*`** (e.g. `MakeUKPowerNetworkClient`,
  `MakeOutageCache`), except the handler which uses `NewListHandler`. Match the
  surrounding file.

- **DNO targeting is opt-out**: in `ParseQueryParams`, a DNO flag that is absent
  or `true` includes it, `false` excludes it, anything else is a `400`. **Status
  targeting** uses the same flag semantics but a different default — `Active` is
  on unless set `false`, whereas `Future` and `Resolved` are off unless set
  `true`. It is a `400` if no DNOs *or* no statuses end up targeted.

- **Provider-specific quirks already encoded** (don't "fix" them blindly):
  SP Energy uses a dedicated `InsecureSkipVerify` HTTP client (incomplete cert
  chain) and a two-step count-then-fetch call; UK Power Network spoofs a browser
  `User-Agent`/`Accept`; National Grid encodes "no time" as the sentinel
  `1900-01-01 00:00:00`, mapped to a nil `*time.Time` in a custom
  `UnmarshalJSON`.

- **Postcodes are normalised** in `model.NewPostcode` (uppercase, strip stray
  chars, `O`→`0` in the inward code, insert the space, regex-validate) — reuse
  it rather than handling postcode strings ad hoc.

- **Clients must drain and close response bodies** via the shared
  `drainAndClose` helper so connections are reused.

- **`cache.OutageCache` sits in front of every DNO client** in the request
  path, keyed by DNO name with a shared TTL (10 minutes, set in `main()`). The
  package-level `clients.ListOutages` helper does the cache check, per-client
  update lock, and double-checked refresh; a nil cache bypasses caching and
  calls the client directly. See [ARCHITECTURE.md](ARCHITECTURE.md) for the
  flow.

- **Tests** live beside the code (`*_test.go`), use `testify`, handler tests rely on
  stubbed HTTP responses / test client maps (see
  [handlers/test_client.go](handlers/test_client.go)) — the suite needs no
  network access. Client tests are live and do require network access.
</content>
