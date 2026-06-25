# lights-out

A small Go HTTP service that aggregates live electricity power-cut ("outage")
data from the public APIs of the UK's six regional Distribution Network
Operators (DNOs) and serves it through a single, normalised endpoint.

Each DNO publishes outages in its own format; `lights-out` queries them
concurrently, converts every response into one canonical shape, and lets you
filter and paginate the combined results.

Supported DNOs:

- Energy North West
- National Grid Distribution
- Northern Powergrid
- SP Energy
- SSE
- UK Power Network

## Prerequisites

- [Go](https://go.dev/dl/) 1.26.2 or newer (see [go.mod](go.mod))
- [`golangci-lint`](https://golangci-lint.run/) — only needed for `make lint`
- Outbound internet access at runtime (the service calls each DNO's live API)

## Build and run

The [Makefile](Makefile) wraps the common tasks:

```sh
make run     # run the service directly (go run ./cmd)
make build   # compile a binary to bin/lights-out
make test    # run the test suite
make fmt     # go fmt ./...
make vet     # go vet ./...
make lint    # golangci-lint run
make all     # fmt, vet, lint, test, then build
make clean   # remove bin/
```

To run the compiled binary:

```sh
make build
./bin/lights-out
```

The server listens on **`:8080`**.

## Usage

There is a single endpoint:

```
GET /list
```

### Query parameters

| Parameter        | Description                                                                 | Default        |
| ---------------- | --------------------------------------------------------------------------- | -------------- |
| `pageSize`       | Number of outages per page. `0` returns all results.                        | `10`           |
| `pageIndex`      | Zero-based page number.                                                      | `0`            |
| `postcodes`      | Comma-separated UK postcodes to filter by. Omit to return all.              | none           |
| *(DNO name)*     | One boolean flag per DNO to include/exclude it (see below).                 | all included   |

**DNO targeting is opt-out.** Each DNO has its own flag named exactly after it:
`EnergyNorthWest`, `NationalGridDistribution`, `NorthernPowergrid`, `SPEnergy`,
`SSE`, `UKPowerNetwork`. A flag that is absent or `true` includes that DNO;
`false` excludes it. Any other value is a `400`.

### Examples

All outages from every DNO, first page of 10:

```sh
curl 'http://localhost:8080/list'
```

Only SSE and UK Power Network, 50 per page:

```sh
curl 'http://localhost:8080/list?SSE=true&UKPowerNetwork=true&NationalGridDistribution=false&NorthernPowergrid=false&SPEnergy=false&EnergyNorthWest=false&pageSize=50'
```

Filter by postcode, returning everything that matches:

```sh
curl 'http://localhost:8080/list?postcodes=AB12%203CD,EH1%201AB&pageSize=0'
```

### Response

A JSON array of outages:

```json
[
  {
    "dno": "UKPowerNetwork",
    "id": "INC-12345",
    "start_time": "2026-06-25T09:00:00+01:00",
    "end_time": "2026-06-25T13:30:00+01:00",
    "postcodes": ["AB12 3CD", "AB12 3CE"]
  }
]
```

`start_time` and `end_time` may be `null` when a DNO does not report a time.

### Status codes

| Code  | Meaning                                                        |
| ----- | ------------------------------------------------------------- |
| `200` | Success (including when some — but not all — DNOs failed).     |
| `400` | A query parameter could not be parsed.                        |
| `500` | Every targeted DNO failed.                                    |

If one DNO is unavailable the request still succeeds with the results from the
others; the failure is logged server-side.

## Testing

```sh
make test
```

Tests live alongside the code they cover (`*_test.go`) and use
[testify](https://github.com/stretchr/testify). The DNO clients are exercised
against recorded/stubbed HTTP responses, so the suite does not require network
access.

## Project layout

```
cmd/       process entry point and dependency wiring
handlers/  HTTP handler that fans out to DNO clients and assembles the response
clients/   one HTTP client adapter per DNO (implements the DnoClient interface)
model/     canonical Outage type, request parsing, and per-DNO payload models
cache/     in-memory TTL key/value store
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for how these fit together and the
design rationale.
</content>
