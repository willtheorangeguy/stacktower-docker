# Stacktower (Docker) — Usage

Stacktower works in two stages: **parse** dependency data from a package registry, then
**render** a visualisation. They are separate commands because parsing hits the network and
rendering does not.

## Parsing

```bash
stacktower parse python fastapi -o fastapi.json      # PyPI
stacktower parse rust serde -o serde.json            # crates.io
stacktower parse javascript yup -o yup.json          # npm
stacktower parse php monolog/monolog -o monolog.json # Packagist
stacktower parse ruby rspec -o rspec.json            # RubyGems
```

Add `--enrich` with a `GITHUB_TOKEN` set to pull repository metadata — stars, maintainers, last
commit — which several render features depend on. See [Configuration](./configuration.md).

### Parse options

| Flag | Description |
|---|---|
| `--max-depth N` | Maximum dependency depth (default: 10) |
| `--max-nodes N` | Maximum packages to fetch (default: 100) |
| `--enrich` | Add repository metadata (requires a token) |
| `--refresh` | Bypass the HTTP cache |

## Rendering

```bash
# Tower visualisation
stacktower render fastapi.json -t tower -o fastapi.svg

# Hand-drawn style with hover popups
stacktower render serde.json -t tower --style handdrawn --popups -o serde.svg

# Traditional node-link diagram
stacktower render yup.json -t nodelink -o yup.svg
```

### Tower options

| Flag | Description |
|---|---|
| `--style simple\|handdrawn` | Visual style |
| `--width`, `--height` | Frame dimensions (default: 800×600) |
| `--edges` | Show dependency edges |
| `--merge` | Merge subdivider blocks |
| `--randomize` | Randomise positions for a hand-drawn effect |
| `--ordering optimal\|barycentric` | Crossing-minimisation algorithm |
| `--ordering-timeout N` | Timeout for the optimal search, seconds (default: 60) |
| `--nebraska` | Show the "Nebraska guy" maintainer ranking |
| `--popups` | Hover popups with metadata |

`--ordering optimal` guarantees the minimum number of edge crossings and is exponential in the
worst case, which is what `--ordering-timeout` is for — it falls back rather than hanging.
`barycentric` is a heuristic: fast, usually good, not optimal.

### Node-link options

| Flag | Description |
|---|---|
| `--detailed` | Show all node metadata in labels |

### Global

| Flag | Description |
|---|---|
| `-v`, `--verbose` | Debug logging: search space size, timings |

## Included examples

The repository ships pre-parsed graphs so you can render without touching the network:

```bash
stacktower render examples/real/flask.json -t tower --style handdrawn --merge -o flask.svg
stacktower render examples/real/serde.json -t tower --popups -o serde.svg
stacktower render examples/real/express.json -t tower --ordering barycentric -o express.svg

stacktower render examples/test/diamond.json -t tower -o diamond.svg
```

`examples/test/` holds synthetic shapes — a diamond, a chain — that are useful for seeing what
the layout algorithms do without a hundred real packages in the way.

## The web server

```bash
stacktower server
```

Serves `blogpost/` on port 8080 with two API endpoints on top. Ports and paths are fixed; see
[API](./api.md) for the endpoints and [Configuration](./configuration.md) for what little is
configurable. Read [`internal/known-issues.md`](./internal/known-issues.md) before exposing it.

## Adding a new language

Three steps, all upstream's design:

1. **Registry client** in `pkg/integrations/<registry>/client.go` — parse the registry API,
   extract dependencies, use `integrations.BaseClient` for HTTP and caching.
2. **Source parser** in `pkg/source/<lang>/<lang>.go` — implement `source.PackageInfo`:
   `GetName`, `GetVersion`, `GetDependencies`, `ToMetadata`, `ToRepoInfo`.
3. **Wire into the CLI** in `internal/cli/parse.go`:

   ```go
   cmd.AddCommand(newParserCmd("<lang> <package>", "Parse <Lang> dependencies",
       func() (source.Parser, error) { return <lang>.NewParser(source.DefaultCacheTTL) }, &opts))
   ```

`source.Parse()` handles concurrent fetching, depth limits, and graph construction. Note that
the web server keeps its own registry map in `internal/cli/server.go`, so a language added this
way appears in the CLI but not in the server until that map is updated too.
