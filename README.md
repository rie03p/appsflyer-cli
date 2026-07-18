# afcli — AppsFlyer CLI

An unofficial command-line client for the [AppsFlyer](https://www.appsflyer.com/) reporting APIs, written in Go.

Fetch raw data, aggregate data, and Master API reports straight from your terminal — no dashboard clicking, easy to pipe into `jq`, `csvkit`, or your data pipeline.

> **Note:** This project is not affiliated with or endorsed by AppsFlyer.

## Supported APIs

| Command | API | Endpoint |
|---|---|---|
| `afcli raw` | [Pull API raw data](https://dev.appsflyer.com/hc/reference/raw_data_pull_api_tokenv2-overview) | `/api/raw-data/export/app/{app-id}/{report}/v5` |
| `afcli agg` | [Pull API aggregate data](https://support.appsflyer.com/hc/en-us/articles/207034346-Pull-API-aggregate-data) | `/api/agg-data/export/app/{app-id}/{report}/v5` |
| `afcli master` | [Master API](https://dev.appsflyer.com/hc/reference/master_api_get) | `/api/master-agg-data/v4/app/{app-id}` |

## Installation

```sh
go install github.com/rie03p/appsflyer-cli/cmd/afcli@latest
```

Or download a prebuilt binary from the [releases page](https://github.com/rie03p/appsflyer-cli/releases), or build from source:

```sh
git clone https://github.com/rie03p/appsflyer-cli.git
cd appsflyer-cli
go build -o afcli ./cmd/afcli
```

## Authentication

All requests use an **AppsFlyer API V2 token** (Bearer token). An account admin can retrieve it from the AppsFlyer dashboard under **Settings → API tokens**.

Pass the token via the `APPSFLYER_API_TOKEN` environment variable (recommended):

```sh
export APPSFLYER_API_TOKEN="eyJhbGci..."
```

or the `--token` flag on any command.

## Usage

### Raw data reports

```sh
# Installs for the first week of July
afcli raw installs_report --app id123456789 --from 2026-07-01 --to 2026-07-07

# Purchases from Japan, saved to a file
afcli raw in_app_events_report --app id123456789 \
  --from 2026-07-01 --to 2026-07-07 \
  --event-name af_purchase --geo JP -o purchases.csv

# List all available raw report names
afcli raw list
```

Raw reports cover user acquisition, organic, retargeting, ad revenue, Protect360 fraud, and postbacks. Responses are CSV. Useful flags: `--media-source`, `--geo`, `--event-name`, `--timezone`, `--currency`, `--max-rows`, `--additional-fields`.

### Aggregate reports

```sh
# Partner (media source) performance
afcli agg partners_report --app id123456789 --from 2026-07-01 --to 2026-07-07

# Daily geo breakdown as JSON
afcli agg geo_by_date_report --app id123456789 \
  --from 2026-07-01 --to 2026-07-07 --format json

# List all available aggregate report names
afcli agg list
```

Available reports: `partners_report`, `partners_by_date_report`, `daily_report`, `geo_report`, `geo_by_date_report`. Add `--reattr` for retargeting conversions and `--attribution-touch-type impression` for view-through attribution.

### Master API

Cross-app aggregate KPIs with custom groupings, filters, and calculated KPIs:

```sh
afcli master --app id123456789 \
  --from 2026-07-01 --to 2026-07-07 \
  --groupings pid,geo --kpis installs,clicks,impressions

# All apps, filtered to one media source, with a calculated CTR
afcli master --app all \
  --from 2026-07-01 --to 2026-07-07 \
  --groupings app_id,pid --kpis installs,clicks,impressions \
  --filter pid=facebook \
  --calculated-kpi ctr=clicks/impressions \
  --format json
```

The date range is limited to 31 days by the API. Filters accept `pid`, `c`, `af_prt`, `af_channel`, `af_siteid`, and `geo`.

### Global flags

| Flag | Description |
|---|---|
| `--token` | API V2 token (defaults to `$APPSFLYER_API_TOKEN`) |
| `-o, --output` | Write the report to a file instead of stdout |
| `--timeout` | HTTP request timeout (default 5m) |
| `--base-url` | Override the API host (default `https://hq1.appsflyer.com`) |

## Development

```sh
go build -o afcli ./cmd/afcli
go test ./...
go vet ./...
```

The API client lives in `internal/appsflyer` (one file per API family, no CLI dependencies) and is independent of the command tree in `internal/cli` (one file per command); both are covered by tests against `httptest` servers. To support a new AppsFlyer API, add a client file and a matching command file.

`main` is the only long-lived branch ([GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow)): topic branches merge into it via PR once CI passes, and releases are cut by pushing a [semver](https://semver.org/) tag (`vX.Y.Z`), which triggers goreleaser to publish cross-platform binaries to GitHub Releases. Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/) — release notes are generated from them.

## License

[MIT](LICENSE)
