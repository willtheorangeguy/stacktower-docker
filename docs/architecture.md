# Stacktower (Docker) — Architecture

## The pipeline

Upstream's six stages, unchanged by this fork:

1. **Parse** — fetch package metadata from a registry (PyPI, crates.io, npm, Packagist,
   RubyGems), following dependencies breadth-first up to `--max-depth` and `--max-nodes`.
2. **Reduce** — remove transitive edges, so a block rests only on what it directly needs.
3. **Layer** — assign each package to a row by depth.
4. **Order** — minimise edge crossings within each row.
5. **Layout** — compute block widths proportional to how much rests on them.
6. **Render** — emit SVG.

Step 4 is the interesting one. Crossing minimisation is NP-hard, and upstream implements a
branch-and-bound search with PQ-tree pruning that is genuinely optimal for small and medium
graphs, with `--ordering-timeout` to bail out to a heuristic on large ones. Step 5 is what makes
the picture read as a tower rather than a graph: width is downstream dependents, so the
foundation is wide and the application on top is one narrow block.

## Project layout

```
stacktower-docker/
├── main.go               entry point
├── internal/cli/         cobra commands
│   ├── parse.go          registry parsers -> graph JSON
│   ├── render.go         graph JSON -> SVG, all the render flags
│   ├── pqtree.go         the ordering search
│   ├── server.go         ADDED BY THIS FORK: HTTP server
│   └── root.go           command registration
├── pkg/
│   ├── source/           per-language parsers, metadata providers
│   ├── integrations/     registry HTTP clients
│   ├── dag/              graph model, transitive reduction, layering
│   ├── render/tower/     tower layout and SVG styles
│   ├── httputil/         cached HTTP client
│   └── io/               graph JSON read and write
├── blogpost/             the static site, also the server's document root
├── examples/             pre-parsed graphs: real packages and synthetic shapes
├── Dockerfile            ADDED BY THIS FORK
└── docker-compose.yml    ADDED BY THIS FORK
```

## What the fork adds

`internal/cli/server.go`, 200 lines, plus the container files. It is a thin HTTP layer over the
existing CLI rather than a refactor of it, and the shape is worth understanding before changing
it.

### `/api/dependencies` calls the library

The handler holds its own registry map, `parserFactories`, keyed by registry name — `pypi`,
`crates`, `npm`, `rubygems`, `packagist` — where the CLI registers the same five parsers by
language name in `parse.go`. Two independent wirings of one set of parsers: adding a language to
`parse.go` does not add it to the server.

It then calls `runParseForServer`, which the source comments describe as "an adaptation of
runParse from parse.go". It drops the logger and passes no metadata providers, which is why
graphs from this endpoint never carry `repo_*` fields.

### `/api/render` shells out to itself

The render handler does not call the render package. It writes the request body to a temp file,
finds its own binary with `os.Executable()`, and runs it as a subprocess:

```go
cmd := exec.Command(executable, "render", tmpfile.Name(), "-t", "tower", ...)
```

The result is read back off disk and returned as `image/svg+xml`.

Re-invoking the binary gets the whole render path — flag parsing, defaults, styles — without
exposing any of it as a library API, and it isolates a render crash from the server process.
The costs are real, though: a process spawn and two disk round-trips per request, render options
that cannot be varied per request because they are baked into the argument list, and a temp
directory that has to exist. It does not, which is why the endpoint fails as shipped —
[`internal/known-issues.md`](./internal/known-issues.md).

### No server hardening

`http.ListenAndServe(":8080", nil)` uses the default `ServeMux` and a zero-value server: no
read, write, or idle timeouts. Neither handler caps the request body, limits concurrency, or
authenticates. `/api/render` will spawn one 60-second-capable subprocess per request for as long
as requests keep arriving.

## The container

Two stages. `golang:1.25.5-alpine3.23` builds a static binary; `alpine:latest` receives the
binary and `blogpost/`. `WORKDIR /app` matters — the server resolves `blogpost/` relative to the
working directory, so it must be started from the directory containing it.

`docker-compose.yml` builds from the local context and publishes 8080. No volumes, no
environment, no healthcheck.

## Relationship to upstream

Everything except `server.go` and the container files is upstream's, under Apache-2.0. The fork
tracks `github.com/matzehuels/stacktower` as the Go module path, so `server.go` imports the
upstream module path even though it is compiled from this tree.

Layout, ordering, and parsing bugs are upstream's. Server and container bugs are this
repository's.
