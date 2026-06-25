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
    H -->|fan-out, one goroutine per DNO| C1[EnergyNorthWest]
    H --> C2[NationalGridDistribution]
    H --> C3[NorthernPowergrid]
    H --> C4[SPEnergy]
    H --> C5[SSE]
    H --> C6[UKPowerNetwork]
    C1 & C2 & C3 & C4 & C5 & C6 -->|DNO-specific JSON| Conv["per-DNO model\nToOutages()"]
    Conv -->|[]model.Outage| Agg["aggregate → sort → filter → paginate"]
    Agg --> Resp["JSON []Outage"]
```

A request is parsed into `QueryParams`, the handler queries every targeted DNO
concurrently, each client converts its provider-specific payload into the shared
`model.Outage` type, and the handler aggregates, sorts, postcode-filters, and
paginates the combined result before encoding it as JSON.

## Packages

The codebase is deliberately flat — four packages with clear, one-directional
dependencies (`cmd → handlers → clients → model`, and `clients → model`).

### `cmd` — composition root
[cmd/main.go](cmd/main.go) wires everything together and owns process
concerns only:
- Builds the `http.Server` and registers the single route `GET /list`.
- Constructs the `map[model.Dno]clients.DnoClient` in `NewDnoClients()` and
  injects it into the handler.
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
key for deterministic output, optionally filters by the requested postcodes,
and applies pagination (`pageSize` / `pageIndex`, where `pageSize=0` means
"return everything").

### `clients` — one adapter per DNO
Every DNO is reached through the [DnoClient](clients/dno_client.go) interface:

```go
type DnoClient interface {
    ListOutages(ctx context.Context) ([]model.Outage, error)
    GetDno() model.Dno
}
```

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
   absent or `true` includes a DNO, `false` excludes it.
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

### `cache` — TTL key/value store (present, not yet wired in)
[cache/kv_store.go](cache/kv_store.go) is a small thread-safe (`sync.RWMutex`)
in-memory `KvStore` with per-entry TTL, distinguishing "missing key" from
"expired value" via typed errors. It is **not currently used by the request
path** — it exists as the building block for response/upstream caching and
should be treated as forthcoming rather than active behaviour.

## Key design decisions

- **Interface-per-DNO with a shared canonical model.** Adding a DNO means
  implementing `DnoClient`, adding its payload model + `ToOutages`, and
  registering it in `NewDnoClients()` and `model.AllDnoList` — no handler
  changes. This is the main extension point.
- **Concurrency with isolation.** Fan-out via goroutines makes a request only
  as slow as the slowest DNO, and per-goroutine `recover()` plus
  tolerate-partial-failure means one flaky provider degrades results rather
  than failing the whole request.
- **Determinism.** Results are sorted by a stable `DNO_ID` key before
  pagination so identical requests return identical, page-stable output.
- **Dependency injection at the edges.** HTTP clients and the DNO client map are
  built in `cmd` and injected, which is what makes the handler and clients
  straightforward to unit test (see the `_test.go` files and
  `handlers/test_client.go`).

## Request/response summary

- **Endpoint:** `GET /list`
- **Query params:** `pageSize`, `pageIndex`, `postcodes` (comma-separated), and
  one boolean flag per DNO name (`EnergyNorthWest`, `NationalGridDistribution`,
  `NorthernPowergrid`, `SPEnergy`, `SSE`, `UKPowerNetwork`).
- **Response:** `application/json` array of `Outage` objects
  (`dno`, `id`, `start_time`, `end_time`, `postcodes`).
- **Status codes:** `400` for unparseable params, `500` only when all targeted
  DNOs fail, `200` otherwise.
</content>
</invoke>
