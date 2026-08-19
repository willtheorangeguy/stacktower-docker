# Stacktower (Docker) — Documentation

This repository is a fork of [matzehuels/stacktower](https://github.com/matzehuels/stacktower)
that adds a **web interface and a Docker runtime** on top of the upstream CLI. Stacktower
renders a package's dependency graph as a physical tower, in the spirit of
[XKCD #2347](https://xkcd.com/2347/): your application on top, everything it rests on below,
down to the one package maintained by some dude in Nebraska.

Upstream gives you `parse` and `render` subcommands. This fork adds a `server` subcommand and a
container that runs it, so the same visualisations are available from a browser.

```
docs/
├── README.md            this index
├── quickstart.md        container up, first tower rendered
├── installation.md      Docker, prebuilt binary, and from source
├── usage.md             the full CLI: parse, render, and every flag
├── configuration.md     environment variables, ports, caching
├── api.md               the two HTTP endpoints, and the graph JSON format
├── architecture.md      the pipeline, and how the server wraps the CLI
├── development.md       building, testing, linting, the release workflows
├── faq.md               questions the design raises
├── troubleshooting.md   concrete errors and their causes
├── roadmap.md           gaps and limitations of this fork
└── internal/
    └── known-issues.md  defects found while documenting (not fixed)
```

## Start here

- Just want a tower — [Quickstart](./quickstart.md)
- Using the CLI directly — [Usage](./usage.md) has every subcommand and flag
- Feeding it your own graph — [API](./api.md) documents the JSON format, which accepts any
  directed graph, not only package dependencies
- Running the web interface — read [`internal/known-issues.md`](./internal/known-issues.md)
  first: the render endpoint does not work as shipped, and the server has no authentication or
  request limits
- Working on the code — [Architecture](./architecture.md) and [Development](./development.md)

## Upstream and this fork

The dependency parsing, graph reduction, ordering, and SVG rendering are all upstream's work
under Apache-2.0. What this repository adds is `internal/cli/server.go`, the `Dockerfile`, and
`docker-compose.yml`.

Bugs in tower layout or registry parsing belong
[upstream](https://github.com/matzehuels/stacktower/issues). Bugs in the server or the container
belong [here](https://github.com/willtheorangeguy/stacktower-docker/issues).
