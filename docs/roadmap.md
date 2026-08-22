# Stacktower (Docker) — Roadmap

This is a fork. Its scope is the web interface and the container; the visualisation itself is
upstream's and deliberately untouched. Defects are in
[`internal/known-issues.md`](./internal/known-issues.md).

## Where it is

`stacktower server` serves the static `blogpost/` site with two endpoints behind it: one that
parses a package from a registry, one that renders graph JSON to SVG. A two-stage Dockerfile and
a Compose file run it on port 8080. That is the whole of the fork — about 200 lines of Go and
two container files.

## The immediate gap: the server is a prototype

It works the way a first version works. In rough order of how much they matter:

**The temp directory it writes to does not exist**, so the render endpoint fails on any fresh
checkout or container. One `os.MkdirAll` fixes it.

**No timeouts, no body limit, no concurrency limit, no authentication.** `http.ListenAndServe`
with a zero-value server has no read or write deadline, the render handler reads the whole
request body into memory, and each request can spawn a subprocess that runs for a minute.

**Render options are hardcoded into an argument list.** Style, dimensions, ordering, and merge
are baked in, so the API can produce exactly one kind of picture.

**A second registry map.** The server keys parsers by registry name; the CLI keys the same five
by language name. Adding a language to the CLI silently does not add it to the server.

**No tests.** The rest of the repository is tested carefully, including an end-to-end shell
suite. `server.go` has none.

## Considered

**Calling the render package instead of shelling out.** Removes the process spawn, the two disk
round-trips, and the temp directory that does not exist. It needs the render path exposed as a
library function, which is an upstream change.

**Per-request render options.** Query parameters or a JSON envelope for style, dimensions, and
ordering. Mostly blocked on the same refactor.

**Pinning `alpine:latest`.** The final stage is unpinned, so builds are not reproducible across
time.

**Versioned images.** `docker-image.yml` publishes only `:latest`, so nothing can be pinned to a
known build.

**Fixing the inherited workflows.** `docs.yml` deploys a CNAME for a domain this repository does
not control, and `homebrew.yml` targets upstream's tap. Both should be adjusted or removed.

## Non-goals

**Diverging from upstream on the visualisation.** Parsing, reduction, layering, ordering, and
rendering stay upstream's. A fix belongs in
[matzehuels/stacktower](https://github.com/matzehuels/stacktower), not here — a fork that starts
improving the layout stops being mergeable.

**A hosted service.** The server has no authentication and spawns subprocesses per request.
Making it safe to host for other people is a different project from making it convenient to run
locally.

**Replacing the CLI.** `parse` and `render` remain the primary interface. The server is a
convenience layer over them, not a successor.

## Contributing

Server and container changes are welcome here; visualisation changes upstream. The missing
`os.MkdirAll` and a `http.Server` with timeouts are both small, self-contained places to start.
