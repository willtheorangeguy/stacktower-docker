# Stacktower (Docker) — Troubleshooting

## The web page shows a tower placeholder but no diagram, and the server logs "Error creating temp file"

`/api/render` writes to `blogpost/tmp`, and nothing creates that directory — not the code, not
the Dockerfile, and not the repository, because `.gitignore` excludes `tmp/`. `os.CreateTemp`
does not create parent directories, so every render request fails with a `500`.

Workaround, before starting the server:

```bash
mkdir -p blogpost/tmp
```

In the container it has to happen inside the image or on a mount:

```yaml
services:
  stacktower:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./blogpost/tmp:/app/blogpost/tmp
```

Recorded in [`internal/known-issues.md`](./internal/known-issues.md).

## `500 Error running stacktower render`

The subprocess failed. The reason is on the server's stdout, not in the response:

```bash
docker compose logs stacktower
```

The usual cause is malformed graph JSON — a node id in `edges` that does not appear in `nodes`.
[API](./api.md) has the format.

## `400 Source 'python' not supported`

The server uses registry names where the CLI uses language names. Use `pypi`, `crates`, `npm`,
`rubygems`, or `packagist`. The mapping is in [API](./api.md).

## `stacktower server` serves nothing, or 404s on every page

The document root is `blogpost/`, resolved relative to the working directory. Run it from the
repository root. There is no flag to point it elsewhere.

## Rendering takes a minute, or the server feels stuck

`/api/render` always uses `--ordering optimal`, the exponential crossing-minimisation search,
and passes no timeout, so it falls back only after the built-in 60 seconds. A large graph really
does take that long. From the CLI, `--ordering barycentric` is near-instant and usually just as
readable.

## `--popups` or `--nebraska` shows nothing

Those flags read `repo_*` metadata, which only `parse --enrich` fetches, and `--enrich` needs a
`GITHUB_TOKEN`. Graphs from `/api/dependencies` never carry it — the server requests no metadata
providers at all. [Configuration](./configuration.md).

## `parse` is slow, or returns stale data

Registry responses are cached in `~/.cache/stacktower/` for 24 hours. `--refresh` bypasses it.
In the container the cache does not survive a recreate, so every run re-fetches.

## The graph is enormous and unreadable

Lower `--max-depth` or `--max-nodes` when parsing. Defaults are 10 and 100; `/api/dependencies`
uses 500 nodes, which is a lot of tower.

## Port 8080 already in use

Remap the host side in `docker-compose.yml` — `"9000:8080"`. The container-side port is
hardcoded.

## Docker build fails on `go mod download`

The build runs `go mod tidy` after copying the source, so a `go.sum` that disagrees with
`go.mod` surfaces here. `go mod tidy` locally, commit both files, rebuild.

## Still stuck

For layout, ordering, or registry parsing, upstream is the right place:
[matzehuels/stacktower/issues](https://github.com/matzehuels/stacktower/issues).

For the server or the container,
[open an issue here](https://github.com/willtheorangeguy/stacktower-docker/issues/new/choose).
