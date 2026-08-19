# Stacktower (Docker) — Quickstart

## The container

```bash
git clone https://github.com/willtheorangeguy/stacktower-docker
cd stacktower-docker
docker compose up --build
```

The server starts on `http://localhost:8080` and redirects to `/dependencies.html`, the
interactive page served out of `blogpost/`.

> **Before you use it:** the render endpoint fails on a fresh checkout — the temp directory it
> writes to is never created. There is a one-line workaround in
> [Troubleshooting](./troubleshooting.md), and the defect is recorded in
> [`internal/known-issues.md`](./internal/known-issues.md). The server also has no
> authentication and no request limits, so keep it on localhost.

## The CLI, which needs no container

If you only want a tower, the CLI is the faster path and has none of the above problems.

```bash
go install github.com/matzehuels/stacktower@latest
```

Render one of the graphs shipped with the repository:

```bash
stacktower render examples/real/flask.json -t tower -o flask.svg
```

Or start from a real package:

```bash
stacktower parse python fastapi -o fastapi.json
stacktower render fastapi.json -t tower --style handdrawn -o fastapi.svg
```

`parse` fetches from the registry — PyPI here — and writes a graph as JSON. `render` turns that
JSON into an SVG. The two stages are deliberately separate: parsing hits the network, rendering
does not, so you can re-render as often as you like without re-fetching.

## What success looks like

An SVG whose widest blocks are at the bottom, each block resting on what it depends on, with
your package a single narrow block at the top. Open it in any browser.

## Next

- [Usage](./usage.md) — every flag, including `--nebraska`, `--popups`, and the ordering
  algorithms
- [API](./api.md) — the JSON format, if you want to visualise a graph that has nothing to do
  with packages
