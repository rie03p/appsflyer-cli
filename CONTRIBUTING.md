# Contributing

Thanks for your interest in improving afcli!

## Development setup

Go 1.26 or later is required.

```sh
git clone https://github.com/rie03p/appsflyer-cli.git
cd appsflyer-cli
go build -o afcli ./cmd/afcli
go test ./...
```

## Project layout

- `cmd/afcli` — entry point.
- `internal/appsflyer` — API client, one file per API family (raw data, aggregate, master, ...). No CLI dependencies.
- `internal/cli` — the cobra command tree, one file per command.

To add support for a new AppsFlyer API, add a client file in `internal/appsflyer` with a params struct and a `Client` method, then a matching command file in `internal/cli`. Both layers are tested against `httptest` servers — see the existing `_test.go` files for the pattern.

## Pull requests

- Keep each PR focused on one change.
- Run `gofmt`, `go vet ./...`, and `go test ./...` before pushing (CI enforces all three).
- Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages (`feat:`, `fix:`, `docs:`, ...) — release notes are generated from them.

## Reporting issues

Bug reports and feature requests are welcome via [GitHub issues](https://github.com/rie03p/appsflyer-cli/issues). For bugs, include the command you ran (with the token redacted) and the error output.
