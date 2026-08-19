# Stacktower (Docker) — API

Two HTTP endpoints, added by this fork in `internal/cli/server.go`. Everything else on port 8080
is static files served out of `blogpost/`, with `/` redirecting to `/dependencies.html`.

**Neither endpoint requires authentication, and neither imposes any limit** — see
[`internal/known-issues.md`](./internal/known-issues.md) before exposing the port.

## `GET /api/dependencies`

Parses a package's dependency graph from a registry and returns it as graph JSON.

| Parameter | Required | Description |
|---|---|---|
| `source` | yes | Registry key — see the table below |
| `id` | yes | Package name, e.g. `fastapi` or `monolog/monolog` |

```bash
curl 'http://localhost:8080/api/dependencies?source=pypi&id=fastapi'
```

### Registry keys

The server uses registry names, while the CLI uses language names. They are different
vocabularies for the same five parsers:

| Server `source` | CLI subcommand |
|---|---|
| `pypi` | `python` |
| `crates` | `rust` |
| `npm` | `javascript` |
| `rubygems` | `ruby` |
| `packagist` | `php` |

Anything else returns `400 Source '<x>' not supported`.

The server hardcodes `maxDepth: 10` and `maxNodes: 500` — five times the CLI's default node
budget — and requests no repository metadata, so graphs fetched this way never carry the
`repo_*` fields that `--popups` and `--nebraska` use.

| Status | Meaning |
|---|---|
| `400` | Missing `source` or `id`, or an unknown registry |
| `500` | Parser construction failed, the registry fetch failed, or JSON encoding failed |

## `POST /api/render`

Takes graph JSON as the request body and returns an SVG.

```bash
curl -X POST --data-binary @fastapi.json http://localhost:8080/api/render > fastapi.svg
```

Returns `image/svg+xml`. Any method other than POST gets `405`.

Render options are **not** configurable over HTTP. The handler always invokes:

```
render <tmpfile> -t tower --style handdrawn --width 982 --height 500
       --ordering optimal --merge --randomize
```

`--ordering optimal` is the expensive algorithm, and the handler passes no
`--ordering-timeout`, so it uses the 60-second default per request.

| Status | Meaning |
|---|---|
| `405` | Method other than POST |
| `500` | Temp file creation, the render subprocess, or reading the result failed |

The `500` covers several distinct failures behind one message; the detail is on the server's
stdout, not in the response. On a fresh checkout it is always temp file creation —
[Troubleshooting](./troubleshooting.md).

## The graph JSON format

Both endpoints speak the same format the CLI reads and writes. It describes **any** directed
graph, so you can hand-write one for a component diagram or a call graph and render it as a
tower.

```json
{
  "nodes": [
    { "id": "app" },
    { "id": "lib-a" },
    { "id": "lib-b" }
  ],
  "edges": [
    { "from": "app", "to": "lib-a" },
    { "from": "lib-a", "to": "lib-b" }
  ]
}
```

### Required fields

| Field | Type | Description |
|---|---|---|
| `nodes[].id` | string | Unique identifier, also used as the label |
| `edges[].from` | string | Source node id |
| `edges[].to` | string | Target node id |

### Optional fields

| Field | Type | Description |
|---|---|---|
| `nodes[].row` | int | Pre-assigned layer; computed if omitted |
| `nodes[].kind` | string | Internal: `"subdivider"` or `"auxiliary"` |
| `nodes[].meta` | object | Freeform metadata for display features |

### Recognised `meta` keys

Read by specific render flags. All optional — a missing key disables the corresponding feature
rather than erroring.

| Key | Type | Used by |
|---|---|---|
| `repo_url` | string | Clickable blocks, `--popups`, `--nebraska` |
| `repo_stars` | int | `--popups` |
| `repo_owner` | string | `--nebraska` |
| `repo_maintainers` | []string | `--nebraska`, `--popups` |
| `repo_last_commit` | date string | `--popups`, brittle detection |
| `repo_last_release` | date string | `--popups` |
| `repo_archived` | bool | `--popups`, brittle detection |
| `summary` | string | `--popups` (falls back to `description`) |

`--detailed`, on node-link diagrams only, prints every meta key in the label.

## External services

`parse` and `/api/dependencies` fetch from PyPI, crates.io, npm, Packagist, and RubyGems.
`--enrich` additionally calls the GitHub or GitLab API. Responses are cached — see
[Configuration](./configuration.md).
