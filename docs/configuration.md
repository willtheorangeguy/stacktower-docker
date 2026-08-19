# Stacktower (Docker) — Configuration

There is no configuration file. Two environment variables, a cache directory, and a fixed port.

## Environment variables

| Variable | Used by | Description |
|---|---|---|
| `GITHUB_TOKEN` | `parse --enrich` | GitHub API token for repository metadata |
| `GITLAB_TOKEN` | `parse --enrich` | GitLab API token for repository metadata |

Without a token, `--enrich` cannot fetch stars, maintainers, or commit dates, so `--popups`,
`--nebraska`, and brittle-package detection have nothing to display. A token needs no scopes
beyond public repository read.

## Caching

HTTP responses from package registries are cached in `~/.cache/stacktower/` with a 24-hour TTL.
`--refresh` bypasses it for a single run.

Registry data changes slowly and the graphs are large, so the cache is the difference between a
re-parse taking a second and taking a minute. Inside the container the cache lives in the
container's filesystem and is lost when it is recreated — mount a volume at
`/root/.cache/stacktower` if that matters to you.

## Server

The `server` subcommand takes no flags and reads no environment variables.

| Setting | Value | Configurable |
|---|---|---|
| Listen address | `:8080` | No — hardcoded |
| Static file root | `blogpost/` | No — relative to the working directory |
| Render temp directory | `blogpost/tmp` | No — hardcoded, and not created automatically |
| Render options | `-t tower --style handdrawn --width 982 --height 500 --ordering optimal --merge --randomize` | No |
| Parse limits | `maxDepth 10`, `maxNodes 500` | No |
| Read/write timeouts | none | No |
| Authentication | none | No |

The port can be remapped on the host side in `docker-compose.yml`:

```yaml
ports:
  - "9000:8080"
```

Because the static root is relative to the working directory, `stacktower server` only works
when run from the repository root. The container's `WORKDIR /app` with `blogpost` copied
alongside the binary satisfies this.

The absent temp directory is why the render endpoint fails as shipped —
[Troubleshooting](./troubleshooting.md) has the workaround,
[`internal/known-issues.md`](./internal/known-issues.md) the detail.

## Docker

```yaml
services:
  stacktower:
    build: .
    ports:
      - "8080:8080"
```

The image is a two-stage build: `golang:1.25.5-alpine3.23` compiles a static binary with
`CGO_ENABLED=0`, and the final `alpine:latest` stage carries only that binary and `blogpost/`.
Note that `alpine:latest` is unpinned, so two builds a month apart can differ in their base
layer.

No volumes are declared. Nothing the server writes is meant to survive the container, which is
consistent with the render endpoint's temp files — and is why the missing `blogpost/tmp` cannot
be worked around from outside the image without a mount.
