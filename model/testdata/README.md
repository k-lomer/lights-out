# DNO data: capture commands, overview, and gotchas

These JSON files are real responses captured from each DNO's public power-cut
endpoint, used as fixtures for the model conversion tests (`model/*_test.go`,
embedded via `//go:embed`). This document records how to regenerate them and the
data-format quirks we found while doing so, for future reference.

The fixtures keep the **full field set** of each real record — every field the
provider returns, not just the ones the model currently decodes — so new fields can
be built from real examples later. Each is trimmed to **no more than 10** outages
chosen to cover the variety of states the provider distinguishes (planned, current,
and resolved outages, plus each status-code value present). Fields we actually use
(notably the postcode lists) are kept **in full**; only large *unused* arrays are
slimmed (see "Regenerating fixtures"). The commands below return the full live
payloads.

## Capture commands

Run from the repo root. Each provider has its own quirks (see the gotcha notes).

### National Grid Distribution
```sh
curl -s -m 30 \
  "https://powercuts.nationalgrid.co.uk/__powercuts/getTabularView"
```

### Northern Powergrid
```sh
curl -s -m 30 \
  "https://power.northernpowergrid.com/Powercut_API/rest/powercuts/getall"
```

### SSE — requires `--compressed`
```sh
curl -s --compressed -m 30 \
  "https://ssen-powertrack-api.opcld.com/gridiview/reporter/info/livefaults"
```

### UK Power Networks — sits behind Akamai, needs a full browser header set
```sh
curl -s --compressed -m 30 \
  -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:150.0) Gecko/20100101 Firefox/150.0" \
  -H "Accept: application/json, text/plain, */*" \
  -H "Accept-Language: en-GB,en;q=0.9" \
  -H "Referer: https://www.ukpowernetworks.co.uk/power-cut/" \
  -H "Sec-Fetch-Dest: empty" -H "Sec-Fetch-Mode: cors" -H "Sec-Fetch-Site: same-origin" \
  "https://www.ukpowernetworks.co.uk/api/power-cut/all-incidents-light"
```
The Go client's `User-Agent` + `Accept: text/plain` spoof is enough from Go's HTTP
transport, but a raw `curl` with only those headers gets a **403 Access Denied** from
Akamai — hence the fuller header set above.

### Energy North West — POST, with paging query params
```sh
curl -s --compressed -m 30 -X POST \
  "https://www.enwl.co.uk/api/power-outages/search?pageSize=200&pageNumber=1&includeCurrent=true&includeResolved=true&includeTodaysPlanned=true&includeFuturePlanned=true&includeCancelledPlanned=false"
```
The client refetches with a larger `pageSize` if `TotalResults` exceeds the page.

### SP Energy — two-step, self-signed cert (`-k`), needs `Content-Type`
Step 1, get the count:
```sh
curl -s -k -m 30 -X POST \
  -H "Content-Type: application/json" \
  "https://powercuts.spenergynetworks.co.uk/webruntime/api/apex/execute?language=en-US&asGuest=true&htmlEncode=false" \
  --data-raw '{"namespace":"","classname":"@udd/01pSr000002yGTp","method":"getImpactDataCount","isContinuation":false,"params":{"postcode":"","statuses":[]},"cacheable":false}'
```
Step 2, fetch `count + buffer` outages (substitute the returned count, e.g. 64):
```sh
curl -s -k -m 30 -X POST \
  -H "Content-Type: application/json" \
  "https://powercuts.spenergynetworks.co.uk/webruntime/api/apex/execute?language=en-US&asGuest=true&htmlEncode=false" \
  --data-raw '{"namespace":"","classname":"@udd/01pSr000002yGTp","method":"getImpactData","isContinuation":false,"params":{"paramsJson":"{\"postcode\":\"\",\"pageNumber\":1,\"pageSize\":64,\"statuses\":[]}"},"cacheable":false}'
```

## Data overview

| DNO | Method | Top-level shape | Outage array | Postcode source |
|-----|--------|-----------------|--------------|-----------------|
| Energy North West | POST | `{Items, TotalResults}` | `Items` | `AffectedPostcodes` (comma string) |
| National Grid | GET | `{lastUpdated, incidents}` | `incidents` | `postcodes` (array) |
| Northern Powergrid | GET | bare array | (root) | `Postcode` (single) |
| SP Energy | POST | `{returnValue, cacheable}` | `returnValue` | `postcodeList` (comma string) |
| SSE | GET | `{Faults, …}` | `Faults` | `affectedAreas` (string array) |
| UK Power Networks | GET | bare array | (root) | `FullPostcodeData` (string array) |

### Time formats and zones
All `Outage` times are normalised to **UTC** in `ToOutage` (see `toUTC` and the
`Outage` doc comment in `model/outage.go`). The source formats differ widely:

| DNO | Source layout(s) | Source zone |
|-----|------------------|-------------|
| Energy North West | `2006-01-02T15:04:05` | naked → **assumed UK local** |
| National Grid | `2006-01-02 15:04:05` | naked → **assumed UK local** |
| Northern Powergrid | RFC3339Nano (`…Z`) | explicit offset (UTC) |
| SP Energy (start) | `2006-01-02 15:04:05` **or** `2/1/2006, 15:04` (D/M/Y) | naked → **assumed UK local** |
| SP Energy (end) | `1/2/2006, 3:04 PM` (M/D/Y, 12h) **or** `2/1/2006, 15:04` (D/M/Y, 24h) | naked → **assumed UK local** |
| SSE | `2006-01-02T15:04:05.999-0700` | explicit offset |
| UK Power Networks | `2006-01-02T15:04:05` and `.999` ms variant | naked → **assumed UK local** |

The "assumed UK local" parse is exactly that — an assumption. If such a provider
actually emits UTC, its summer (BST) times will be an hour out. See the `ukLocation`
comment in `model/outage.go`.

### End-time semantics (what `Outage.End` means per DNO)
| DNO | End derivation |
|-----|----------------|
| Energy North West | actual restoration → else estimated → else nil |
| SP Energy | actual restoration → else estimated (estimated always present) |
| UK Power Networks | restored → else estimated → else nil |
| National Grid | `etr` only (nil via sentinel) |
| Northern Powergrid | `EstimatedTimeTillResolution` only (nil via sentinel) |
| SSE | `estimatedRestoration` only (required field) |

## Gotchas and edge cases

**Transport / request quirks**
- **SSE** serves gzip and returns garbled bytes unless decompressed — use
  `curl --compressed` (Go's transport does this transparently).
- **UK Power Networks** returns 403/empty without a browser-like `User-Agent` and
  `Accept` header. Go's transport gets through with just those two, but Akamai blocks a
  raw `curl` unless it also sends `Referer`/`Accept-Language`/`Sec-Fetch-*` (see the
  capture command). Its payload also includes `Multiple`/type-0 summary rows that carry
  the nested `UnplannedIncidents`/`RestoredIncidents`/`PlannedIncidents` arrays — the
  only place those fields are populated.
- **SP Energy** returns an **empty body** unless the request sets
  `Content-Type: application/json`, uses a **self-signed/incomplete cert chain**
  (`curl -k`; the client uses `InsecureSkipVerify`), and requires a **two-step**
  count-then-fetch call.
- **Energy North West** is a POST whose page size must be grown if `TotalResults`
  exceeds the requested `pageSize`.

**Decoding quirks**
- **`Postcodes.UnmarshalJSON`** splits the raw JSON bytes on commas and strips
  brackets/quotes, so it happens to handle **both** a JSON array (`["A","B"]`) and a
  comma-string (`"A, B"`). Fragile but effective; a postcode containing a comma would
  break it (none do).
- **One bad record fails the whole DNO.** Clients decode the entire array in one
  `Decode` call, and the custom time `UnmarshalJSON` methods return an error on an
  unparseable date — so a single malformed record discards *all* outages for that
  provider. The handler's per-DNO `recover()` only isolates failures *between*
  providers, not within one.

**Data-quality / value edge cases**
- **Sentinel "no time" values** map to a nil end time:
  National Grid uses `1900-01-01 00:00:00` (space), Northern Powergrid uses
  `1900-01-01T00:00:00` (RFC3339-style). Both appear in real data.
- **Northern Powergrid sends one row per affected postcode** — heavy duplication
  (observed one incident repeated 117 times). `NorthernPowergridToOutages` dedupes on
  `Reference + LoggedTime + ETR` and merges postcodes. Verified across a full payload:
  the merge key is constant within each incident, so incidents do **not** fragment.
- **Dirty postcodes are common and silently skipped.** Real SP Energy / UKPN data
  contained bare outward codes (` SY4`, ` ML1`) and junk (`UK`, `UMS`); UKPN omits
  spaces entirely (`N166RJ` → normalised to `N16 6RJ`). Invalid entries are dropped
  via non-strict `ParsePostcodes`, so an outage can end up with **zero** postcodes and
  still be emitted.
- **`End` collapses fact and prediction.** It holds the *actual* restoration time for
  resolved outages but the *estimated* time for ongoing ones, with no flag to tell
  them apart, and the canonical model carries **no status field** even though the raw
  payloads are rich with one (SSE `resolved`, ENW `Type`, NPg `IncidentStatus`).

## Regenerating fixtures

After capturing a full payload, trim it to **no more than 10** records while keeping
each record's **full field set**. Choose the records to cover the variety the provider
distinguishes — some with end times, some without, sentinels, planned vs current vs
resolved, and every status-code value present (e.g. ENW `Type`/`WebTMSFaultType`/
`faultStatus`, UKPN `IncidentType`/`IncidentTypeName`/`PowerCutType`/`StatusId`) — and
for Northern Powergrid include a duplicate-`Reference` group so the merge path is
exercised. Keep the **complete top-level envelope**, not just the decoded fields.

Keep anything the model uses in full — in particular the postcode lists (ENW
`AffectedPostcodes`, UKPN `FullPostcodeData`, SSE `affectedAreas`, etc.). Only slim
large *unused* arrays, preserving their shape: UKPN `Steps`/`Logs`/`PostcodeGeometryData`
trim to ~3 elements each; ENW `addressMpanList` to ~3. Fields that are null/empty across
the whole snapshot can't be given real values — leave them null. If a `*_RealData*` test
hardcodes an ID/postcode from the old fixture, repoint it at data in the new one. Then
re-run `go test ./model/`.
