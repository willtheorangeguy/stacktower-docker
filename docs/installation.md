# Stacktower (Docker) — Installation

Three routes. Pick by what you want: the web interface needs the container, the CLI does not.

## Docker — the web interface

```bash
git clone https://github.com/willtheorangeguy/stacktower-docker
cd stacktower-docker
docker compose up --build
```

Serves `http://localhost:8080`. The first build compiles the Go binary from source, so expect a
few minutes.

An image is also published to `ghcr.io/willtheorangeguy/stacktower:latest` on every push to
`main` — note the image name has no `-docker` suffix, unlike the repository. It carries only a
`latest` tag, so there is no way to pin a version.

Read [`internal/known-issues.md`](./internal/known-issues.md) before exposing the port: the
render endpoint fails as shipped, and the server has no authentication or request limits.

## `go install` — the CLI

```bash
go install github.com/matzehuels/stacktower@latest
```

This installs **upstream's** CLI, which is the right thing for CLI use: this fork changes
nothing about `parse` or `render`. Requires Go 1.24 or newer.

## From source

```bash
git clone https://github.com/willtheorangeguy/stacktower-docker
cd stacktower-docker
go build -o stacktower .
```

Or `make build`, which writes to `bin/stacktower`. This is the only route that gives you the
`server` subcommand without Docker — and it must be run from the repository root, because the
server resolves its document root relative to the working directory.

## Prerequisites

| Requirement | For |
|---|---|
| Docker with Compose v2 | The container route |
| Go 1.24+ | `go install`, building from source, `make` targets |
| Network access | `parse` and `/api/dependencies` — rendering is entirely offline |

A `GITHUB_TOKEN` is optional and only affects `parse --enrich`; see
[Configuration](./configuration.md).

## Verifying

```bash
stacktower render examples/test/diamond.json -t tower -o diamond.svg
```

A synthetic four-node graph, no network needed. If that produces an SVG, the render path works.

For the server:

```bash
curl -s 'http://localhost:8080/api/dependencies?source=pypi&id=click' | head -c 200
```

JSON back means the parse path works. The render endpoint is tested by the page itself, and
currently fails — [Troubleshooting](./troubleshooting.md).
