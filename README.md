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
| `afcli cohort` | [Cohort API](https://dev.appsflyer.com/hc/reference/post_app-id) | `/api/cohorts/v1/data/app/{app-id}` |
| `afcli skan` | [SKAN aggregated performance report](https://dev.appsflyer.com/hc/reference/skan-agg-performance-report-api-get) | `/api/skadnetworks/{v2,v3}/data/app/{app-id}` |
| `afcli freshness` | [Master Freshness API](https://dev.appsflyer.com/hc/reference/master-lastupdate) | `/api/master-agg-data/lastupdate` |
| `afcli onelink` | [OneLink API v2.0](https://dev.appsflyer.com/hc/reference/onelinkapi_v2_overview) | `onelink.appsflyer.com/api/v2.0/shortlinks/...` |

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

Store it once with `auth login` (recommended):

```sh
afcli auth login   # prompts for the token, hidden input
afcli auth status  # shows which token is in use
afcli auth logout  # deletes the stored token
```

The token is saved to your user config directory with `0600` permissions. Alternatively, set the `APPSFLYER_API_TOKEN` environment variable or pass `--token` on any command; precedence is `--token` > `APPSFLYER_API_TOKEN` > stored token.

The OneLink API uses a **separate token** (ask your CSM or dashboard admin). Store it with `afcli auth login --onelink`, or use `ONELINK_API_TOKEN` / `--onelink-token`.

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

### Cohort reports

Retention, LTV, and ROAS by cohort day:

```sh
afcli cohort --app id123456789 --from 2026-06-01 --to 2026-06-30 \
  --cohort-type user_acquisition --kpis users,roas --groupings pid,geo \
  --aggregation-type cumulative

# On-day retention for specific cohort days, as CSV
afcli cohort --app id123456789 --from 2026-06-01 --to 2026-06-30 \
  --cohort-type unified --kpis retention --groupings pid \
  --aggregation-type on_day --filter period=0,1,7,30 --format csv
```

### SKAN performance

iOS SKAdNetwork attribution metrics (the SKAN dashboard as an API):

```sh
afcli skan --app id123456789 --start-date 2026-07-01 --end-date 2026-07-07
```

Defaults to `--version v3` (SKAN 4 postbacks); use `--version v2` for SKAN 3. The date range is limited to 90 days.

### Data freshness

```sh
afcli freshness
```

### OneLink short links

Create, inspect, update, and delete OneLink short links programmatically (requires the OneLink API token, see Authentication):

```sh
afcli onelink create abc123 --param pid=email --param c=summer_sale --ttl 90d
afcli onelink get abc123 qwer9876
afcli onelink update abc123 qwer9876 --param pid=email --param c=autumn_sale
afcli onelink delete abc123 qwer9876
```

`abc123` is the OneLink template ID from the dashboard; `qwer9876` is the short link ID. `--param pid=...` (media source) is mandatory on create.

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
