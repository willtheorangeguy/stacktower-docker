# Stacktower (Docker) — Development

## Build

```bash
make build          # -> bin/stacktower
make install        # go install
go build -o stacktower .
```

Go 1.24 or newer. The module path is upstream's, `github.com/matzehuels/stacktower`, so imports
inside this tree reference it rather than the fork.

## The Makefile

| Target | What it does |
|---|---|
| `make check` | `fmt` then `lint` then `test` — run this before pushing |
| `make fmt` | `gofmt -s -w .` and `goimports -w -local stacktower .` |
| `make lint` | `go vet ./...` and `staticcheck ./...` |
| `make test` | `go test -race -timeout=2m ./...` |
| `make cover` | Coverage profile plus a per-function summary |
| `make e2e` | `scripts/test_e2e.sh all` — needs a build first |
| `make e2e-test` | Synthetic fixtures only, no network |
| `make e2e-real` | Real packages — hits the registries |
| `make blog` | Regenerate the diagrams and showcase plots under `blogpost/` |
| `make snapshot` | GoReleaser snapshot build, no publish |
| `make install-tools` | Fetch `goimports` and `staticcheck` |

`make test` runs with `-race`, so the tests are meaningfully concurrent — keep them that way.

## Linting

`.golangci.yml` configures golangci-lint for CI; the Makefile uses `go vet` and `staticcheck`
directly. Both must pass. Formatting is `goimports` with a local prefix, so import blocks group
project imports separately — running `make fmt` before committing avoids a diff argument.

## Tests

Unit tests sit beside the code (`pkg/source/php/php_test.go` and so on). End-to-end tests are
shell, in `scripts/test_e2e.sh`, and split so the offline half can run without touching the
network:

```bash
make build && make e2e-test     # synthetic fixtures, offline
make build && make e2e-real     # real packages, network
```

The web server added by this fork has no tests.

## CI and release workflows

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | push, PR | Build, vet, test |
| `docker-image.yml` | push to `main` | Builds and pushes `ghcr.io/{owner}/stacktower:latest` |
| `docs.yml` | push to `main` touching `blogpost/**` | Deploys `blogpost/` to GitHub Pages |
| `release.yml` | tag push | GoReleaser — binaries and archives |
| `homebrew.yml` | published release | Bumps a Homebrew formula |

Three of these were inherited from upstream and still point at upstream's targets: `docs.yml`
deploys a `blogpost/CNAME` naming a domain this repository does not control, and `homebrew.yml`
bumps `matzehuels/homebrew-tap`. Both are recorded in
[`internal/known-issues.md`](./internal/known-issues.md).

## Working on the server

`internal/cli/server.go` is the fork's own code and the place most changes belong. Two things
to know before editing it:

- It holds a **second** registry map, `parserFactories`, keyed by registry name rather than by
  language. Adding a parser to `internal/cli/parse.go` does not add it to the server.
- `/api/render` shells out to the binary rather than calling the render package, so changing
  render behaviour over HTTP means changing an argument list, not a function call. See
  [Architecture](./architecture.md).

## Contributing

Changes to parsing, layout, ordering, or rendering belong
[upstream](https://github.com/matzehuels/stacktower) — this fork deliberately does not diverge
there. Changes to the server, the Dockerfile, or the Compose file belong here.

For this repository, see the org-wide
[Contributing Guide](https://github.com/willtheorangeguy/.github/blob/main/CONTRIBUTING.md).
