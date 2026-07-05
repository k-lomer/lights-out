# Architecture

`lights-out` is a small Go HTTP service that aggregates live electricity
power-cut ("outage") data from the public APIs of the UK's six regional
Distribution Network Operators (DNOs), normalises it into a single canonical
shape, and serves it from one endpoint.

## High-level flow

```mermaid
flowchart TD
    Req["GET /list?..."] --> H[ListHandler]
    H --> QP["ParseQueryParams\n(model)"]
    H -->|fan-out, one goroutine per DNO| CW["clients.ListOutages\n(OutageCache check)"]
    CW -.cache hit.-> Agg
    CW -->|cache miss| C1[EnergyNorthWest]
    CW --> C2[NationalGridDistribution]
    CW --> C3[NorthernPowergrid]
    CW --> C4[SPEnergy]
    CW --> C5[SSE]
    CW --> C6[UKPowerNetwork]
    C1 & C2 & C3 & C4 & C5 & C6 -->|DNO-specific JSON| Conv["per-DNO model\nToOutages()"]
    Conv -->|[]model.Outage, cached| Agg["aggregate → sort → filter → paginate"]
    Agg --> Resp["JSON []Outage"]
```

A request is parsed into `QueryParams`, the handler queries every targeted DNO
concurrently, each client converts its provider-specific payload into the shared
`model.Outage` type, and the handler aggregates, sorts, filters by status and
postcode, and paginates the combined result before encoding it as JSON. A per-DNO
`OutageCache` sits in front of each client, so a fresh cache entry is returned
without touching the provider.

## Packages

The codebase is deliberately flat — five packages with clear, one-directional
dependencies (`cmd → handlers → clients → model`, with both `handlers` and
`clients` also depending on `cache`, and `cache → model`).

### `cmd` — composition root
[cmd/main.go](cmd/main.go) wires everything together and owns process
concerns only:
- Builds the `http.Server` and registers the single route `GET /list`.
- Constructs the `map[model.Dno]clients.DnoClient` and the shared
  `cache.OutageCache` (a 10-minute TTL), and injects both into the handler.
- Owns the two shared `*http.Client` instances. Most DNOs use a standard
  client; SP Energy uses a separate client with `InsecureSkipVerify` because
  its endpoint serves an incomplete certificate chain (missing intermediate
  certificates). This is the one place TLS is relaxed, and it is documented
  inline.

### `handlers` — request orchestration
[handlers/list_handler.go](handlers/list_handler.go) is the heart of the
service. `ListHandler.getOutages` implements the fan-out:
- One goroutine per targeted DNO client, coordinated with a `sync.WaitGroup`.
- Results and errors are written into per-index slices (`dnoOutages`,
  `dnoErrs`), so no shared mutable state needs locking.
- Each goroutine has a `recover()` so a panic in one DNO client cannot bring
  down the request or the process; the panic is recorded as that client's
  error.
- **Partial failure is tolerated**: the request only fails (500) if *every*
  targeted client fails. Otherwise successful DNOs are returned.

After collection the handler: aggregates all results, sorts them by a stable
key for deterministic output, filters by the requested statuses and then
(optionally) the requested postcodes, and applies pagination (`pageSize` /
`pageIndex`, where `pageSize=0` means "return everything").

### `clients` — one adapter per DNO
Every DNO is reached through the [DnoClient](clients/dno_client.go) interface:

```go
type DnoClient interface {
    ListOutages(ctx context.Context) ([]model.Outage, error)
    GetDno() model.Dno
    LastUpdate() *time.Time
    SetUpdated() time.Time
    UpdateLock()
    UpdateUnlock()
}
```

The last four methods back the caching machinery and are satisfied by embedding
`*UpdateTracker` (a mutex plus the last-update timestamp) in each client, so an
implementation only writes `ListOutages` and `GetDno`.

Each implementation (e.g. [clients/ukpn.go](clients/ukpn.go),
[clients/national_grid_distribution.go](clients/national_grid_distribution.go),
[clients/sp_energy.go](clients/sp_energy.go)) is responsible for the quirks of
*its* provider and nothing else. The interface keeps those quirks from leaking
into the handler. Notable provider-specific behaviour includes:
- **UK Power Network** spoofs a browser `User-Agent`/`Accept` to be served the
  light incidents feed.
- **SP Energy** is a two-step call: fetch a count, then request that many
  records in a follow-up POST. It also uses the insecure HTTP client.
- Others are straightforward single `GET`/`POST` calls.

Each client decodes into its DNO-specific model type and returns the converted
`[]model.Outage`; the shared `drainAndClose` helper ensures response bodies are
fully drained and closed so connections can be reused.

### `model` — canonical types, parsing, and per-DNO mapping
This package holds three kinds of thing:
1. **The canonical domain type** — [model/outage.go](model/outage.go) defines
   `Outage` (the JSON shape clients consume) plus aggregation/sorting helpers.
   `model.Dno` and the list of all DNOs live in [model/dno.go](model/dno.go).
2. **Request parsing** — [model/query_params.go](model/query_params.go) turns
   raw `url.Values` into a validated `QueryParams`. DNO targeting is opt-out:
   absent or `true` includes a DNO, `false` excludes it. Status targeting uses
   the same flags but a different default — `Active` is on unless disabled,
   while `Future` and `Resolved` are off unless explicitly enabled.
   [model/postcode.go](model/postcode.go) normalises and validates UK postcodes
   (uppercasing, stripping stray characters, fixing `O`→`0`, inserting the
   space, regex-validating).
3. **Per-DNO payload models** — one file per DNO (e.g.
   [model/national_grid_distribution.go](model/national_grid_distribution.go),
   [model/sp_energy.go](model/sp_energy.go)) describing that provider's JSON and
   a `ToOutage(s)` method that maps it onto the canonical `Outage`. Providers
   differ in subtle ways the models absorb — for example National Grid encodes
   "no time" as a sentinel timestamp (`1900-01-01 00:00:00`) which a custom
   `UnmarshalJSON` translates into a nil `*time.Time`.

### `cache` — per-DNO outage cache
[cache/outage_cache.go](cache/outage_cache.go) is a small thread-safe
(`sync.RWMutex`) in-memory `OutageCache`, keyed by DNO name and holding that
DNO's `[]model.Outage` with a single shared TTL (10 minutes, set in `main()`).
`Get` distinguishes "missing key" from "expired value" via typed errors.

The cache is wired into the request path through the package-level
[`clients.ListOutages`](clients/dno_client.go) helper the handler calls per DNO:

1. Return the entry immediately on a fresh cache hit.
2. On a miss, take the client's update lock, then **re-check** the cache — a
   concurrent request may have refreshed it while we waited — before fetching.
3. Fetch from the client, stamp each outage's `LastUpdated`, store, and return.

The lock and last-update timestamp come from `*UpdateTracker`, embedded in every
client, so only one goroutine refreshes a given DNO at a time while other DNOs
proceed in parallel. A nil cache (as some tests pass) bypasses all of this and
calls the client directly.

## Key design decisions

- **Interface-per-DNO with a shared canonical model.** Adding a DNO means
  implementing `DnoClient`, adding its payload model + `ToOutages`, and
  registering it in the `dnoClients` map in `main()` and `model.AllDnoList` — no
  handler changes. This is the main extension point.
- **Concurrency with isolation.** Fan-out via goroutines makes a request only
  as slow as the slowest DNO, and per-goroutine `recover()` plus
  tolerate-partial-failure means one flaky provider degrades results rather
  than failing the whole request.
- **Determinism.** Results are sorted by a stable `DNO_ID` key before
  pagination so identical requests return identical, page-stable output.
- **Dependency injection at the edges.** HTTP clients, the DNO client map, and
  the cache are built in `cmd` and injected, which is what makes the handler and
  clients straightforward to unit test (see the `_test.go` files and
  `handlers/test_client.go`).
- **Caching with single-flight refresh.** A per-DNO TTL cache spares the
  providers repeated calls, and the per-client update lock with a double-checked
  cache read means a burst of concurrent misses triggers one upstream fetch per
  DNO rather than one per request.

## Request/response summary

- **Endpoint:** `GET /list`
- **Query params:** `pageSize`, `pageIndex`, `postcodes` (comma-separated), one
  boolean flag per DNO name (`EnergyNorthWest`, `NationalGridDistribution`,
  `NorthernPowergrid`, `SPEnergy`, `SSE`, `UKPowerNetwork`), and one boolean flag
  per outage status (`Active`, `Future`, `Resolved`).
- **Response:** `application/json` array of `Outage` objects
  (`dno`, `id`, `start_time`, `estimated_end`, `actual_end`, `postcodes`,
  `last_updated_time`, `status`).
- **Status codes:** `400` for unparseable params, `500` only when all targeted
  DNOs fail, `200` otherwise.
</content>
</invoke>
