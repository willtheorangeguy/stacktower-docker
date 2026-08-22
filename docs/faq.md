# Stacktower (Docker) — FAQ

## How is this different from matzehuels/stacktower?

It adds a `server` subcommand and a container that runs it. Everything about parsing and
rendering is upstream's, unchanged. If you only want the CLI, install upstream's —
`go install github.com/matzehuels/stacktower@latest`.

## Where should I file a bug?

Layout, ordering, registry parsing, SVG output:
[upstream](https://github.com/matzehuels/stacktower/issues). The web server, the Dockerfile, the
Compose file: the [fork issue tracker](https://github.com/willtheorangeguy/stacktower-docker/issues).

## Can I put the web interface on the internet?

No. There is no authentication, no request size limit, no concurrency limit, and no server
timeouts, and each render request spawns a subprocess that can run for a minute. Keep it on
localhost. [`internal/known-issues.md`](./internal/known-issues.md) has the detail.

## Why does the render endpoint return a 500?

`blogpost/tmp` does not exist and nothing creates it. See
[Troubleshooting](./troubleshooting.md) for the one-line workaround.

## Why does `source=python` not work on the API when `parse python` does?

The server keys its parsers by registry — `pypi`, `crates`, `npm`, `rubygems`, `packagist` —
while the CLI keys the same five by language. Two maps, two vocabularies. [API](./api.md) has
the mapping.

## Does it support Go, or Java?

No. Five registries: PyPI, crates.io, npm, Packagist, RubyGems. Adding one is a documented
three-step change — see [Usage](./usage.md) — plus a fourth step for the server's own map.

## Do I need a GitHub token?

Only for `parse --enrich`, which fetches stars, maintainers, and commit dates. Without it the
graph still renders; `--popups`, `--nebraska`, and brittle-package detection have nothing to
show. The web API never requests metadata, token or not.

## Why are the blocks different widths?

Width is proportional to how much rests on a package, not to the package's own size. That is the
whole idea: the widest block at the bottom is what everything above it depends on, which is
usually not the package anyone thinks about.

## What is `--nebraska`?

The XKCD reference. It ranks packages by how load-bearing they are against how few people
maintain them — the "some random person in Nebraska" of
[XKCD #2347](https://xkcd.com/2347/). It needs enriched metadata.

## Can I visualise something that is not a package graph?

Yes. The JSON format is any directed graph — nodes with ids, edges with `from` and `to`. Write
one by hand and render it. [API](./api.md) documents the schema.

## Why does rendering take so long?

`--ordering optimal` is an exponential search for the minimum number of edge crossings, capped
at 60 seconds by default. `--ordering barycentric` is a fast heuristic. The web API always uses
optimal and never overrides the cap.

## Is it offline?

Rendering is entirely offline. Parsing needs the registries. Registry responses are cached for
24 hours in `~/.cache/stacktower/`.

## What licence is this?

Apache-2.0, upstream's licence, retained. See the README's Attribution section and
[`LICENSE.md`](license.md).
