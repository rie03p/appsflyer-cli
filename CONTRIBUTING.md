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

## Branching and releases

This project uses [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow):

- `main` is the only long-lived branch and is always releasable. It is protected against force pushes and deletion.
- All changes land on `main` through short-lived topic branches (`feat/cohort-api`, `fix/timeout-flag`, ...) merged via pull request once CI is green. There is no `develop` branch.
- Releases are cut by tagging `main` with a [semver](https://semver.org/) tag (`v0.2.0`). Pushing the tag triggers the release workflow, which builds binaries for Linux/macOS/Windows (amd64/arm64) with goreleaser and publishes a GitHub release with generated notes. Nothing is released from branches.
- Breaking CLI changes (renamed commands/flags) bump the minor version while pre-1.0, and the major version after.

CI (`gofmt`, `go vet`, `go test`, `go build`) runs on every pull request and on every push to `main`, across Linux, macOS, and Windows.

## Pull requests

- Keep each PR focused on one change.
- Run `gofmt`, `go vet ./...`, and `go test ./...` before pushing (CI enforces all three).
- Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages (`feat:`, `fix:`, `docs:`, ...) — release notes are generated from them.

## Reporting issues

Bug reports and feature requests are welcome via [GitHub issues](https://github.com/rie03p/appsflyer-cli/issues). For bugs, include the command you ran (with the token redacted) and the error output.
