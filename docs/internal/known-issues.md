# Known Issues — stacktower-docker

Concrete defects and gaps found while writing this repository's documentation in
August 2026. **Nothing here was changed** — each one needs a code, configuration, or
licensing decision rather than a documentation one.

Ordered by severity. See [`docs/roadmap.md`](../roadmap.md) for the narrative version,
which also covers deliberate non-goals.


**7 open:** 1 high, 4 medium, 2 low.

## 1. The render endpoint cannot work: the directory it writes to is never created

**Severity:** High  
**Where:** `internal/cli/server.go` -> `renderHandler`; `Dockerfile`; `.gitignore`

**What:** `renderHandler` begins with `os.CreateTemp("blogpost/tmp", "render-*.json")`. Nothing creates `blogpost/tmp`: there is no `MkdirAll` anywhere in the fork's code (the only one in the tree is in `pkg/httputil/cache.go`, for the HTTP cache), the `Dockerfile` copies `blogpost` as it exists in the repository, and `.gitignore` contains `tmp/`, so the directory cannot be committed even as an empty placeholder. `os.CreateTemp` does not create parent directories. Every `POST /api/render` therefore fails at its first step and returns `500 Error creating temp file`.

**Why it matters:** This is the one endpoint that produces the pictures, in a fork whose entire reason for existing is putting those pictures in a browser. It fails on every fresh clone and in every container built from this repository -- not intermittently and not under load, but always and from the first request. The failure is also opaque from the outside: the response is a bare 500, the real cause goes to the server's stdout, and the page shows an empty frame, so a user has no way to tell a missing directory from a broken renderer.

It presumably worked during development because `blogpost/tmp` existed on that machine as a side effect of an earlier run, and being gitignored it never travelled.

**Suggested fix:** `os.MkdirAll("blogpost/tmp", 0o755)` at server startup, which also fixes the container without touching the Dockerfile. `os.MkdirTemp` in the system temp directory would be better still -- the files are read back and deleted within the request, so there is no reason for them to live under the document root where they are also briefly web-accessible.

## 2. The web server has no authentication, no request limits, and no timeouts

**Severity:** Medium  
**Where:** `internal/cli/server.go` -> `runServer`, `renderHandler`, `dependenciesHandler`

**What:** `http.ListenAndServe(":8080", nil)` uses a zero-value server, so `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` are all unset -- connections have no deadline. Neither handler checks any credential. `renderHandler` does `io.ReadAll(r.Body)` with no cap, writes the result to disk, and runs the binary as a subprocess with `--ordering optimal` and no `--ordering-timeout`, so each request may occupy a core for the built-in 60 seconds. There is no concurrency limit on either handler. `dependenciesHandler` parses with `maxNodes: 500`, five times the CLI default, fetching from a package registry on demand. `docker-compose.yml` publishes the port.

**Why it matters:** Any client that can reach the port can spawn unbounded subprocesses, each potentially minute-long, and hold connections open indefinitely -- so a handful of concurrent requests will exhaust CPU or process limits on the host, not merely on the container. The uncapped body read means a single request can also consume as much memory as the sender cares to provide. None of this requires authentication because there is none.

For a tool run on localhost this is entirely reasonable, and that is very likely the intent. The problem was that nothing said so: the repository ships a Compose file that publishes 8080 and a README that did not mention the server at all, so an operator had no signal that this was a localhost-only prototype rather than a deployable service.

**Suggested fix:** Documented in this pass -- the README, `docs/api.md`, and `docs/faq.md` now say plainly not to expose it. For code: use an explicit `http.Server` with read and write timeouts, wrap the body in `http.MaxBytesReader`, pass `--ordering-timeout` well below 60s, and bound concurrency with a semaphore. Binding to `127.0.0.1:8080` in the Compose file would be a sensible default given the intent.

## 3. The Pages workflow deploys a CNAME for a domain this repository does not control

**Severity:** Medium  
**Where:** `.github/workflows/pages.yml`, `blogpost/CNAME`

**What:** `pages.yml` runs on every push to `main` touching `blogpost/**` and uploads `blogpost` as the GitHub Pages artifact. `blogpost/CNAME` contains `stacktower.io`, upstream's domain. A CNAME file in a Pages artifact sets the site's custom domain. Both the workflow and the file were inherited from upstream unchanged. `blogpost/robots.txt` likewise opens with '# Robots.txt for stacktower.io'.

**Why it matters:** The deployment cannot serve the fork's site: GitHub will either reject the custom domain because the DNS does not point here, or accept the configuration and leave the site unreachable at a domain that resolves elsewhere. Either way the workflow runs on every relevant push and produces nothing usable, which is the kind of failure that gets ignored rather than fixed. Claiming another project's domain in your own Pages configuration is also poor form regardless of whether it succeeds, and if upstream's DNS ever pointed at GitHub without their own verification, it stops being merely untidy.

**Suggested fix:** Delete `blogpost/CNAME` in this fork so Pages uses the default `willtheorangeguy.github.io/stacktower-docker` domain, or delete `pages.yml` if the fork does not want a site at all. The second is probably right -- the fork's contribution is the server and the container, and the blog post is upstream's writing.

## 4. The Homebrew workflow bumps a formula in upstream's tap

**Severity:** Medium  
**Where:** `.github/workflows/homebrew.yml`

**What:** On every published release of this repository the workflow runs `dawidd6/action-homebrew-bump-formula` with `tap: matzehuels/homebrew-tap` and `formula: stacktower`, authenticated with `secrets.HOMEBREW_TAP_TOKEN`. Inherited from upstream unchanged. `continue-on-error: true` is set.

**Why it matters:** If that secret is ever populated with a token that can write to upstream's tap, cutting a release here would push a formula bump into someone else's repository pointing Homebrew users at this fork's version. That is a supply-chain-shaped mistake made by accident rather than intent, and the `continue-on-error: true` means the workflow reports success either way, so nothing draws attention to it. Today the secret is presumably unset and the step fails quietly, which is the harmless case -- but the configuration is one repository-secret away from the harmful one, and secrets get added for unrelated reasons.

**Suggested fix:** Delete the workflow. A fork that adds a Dockerfile has no reason to publish a Homebrew formula, and the CLI it would install is upstream's anyway.

## 5. The server keeps a second, differently-keyed registry map

**Severity:** Medium  
**Where:** `internal/cli/server.go` -> `parserFactories`; `internal/cli/parse.go`

**What:** `server.go` declares its own map of parser constructors keyed `pypi`, `crates`, `npm`, `rubygems`, `packagist`, while `parse.go` registers the same five parsers as CLI subcommands keyed `python`, `rust`, `javascript`, `ruby`, `php`. `runParseForServer` is likewise described in its own comment as 'an adaptation of runParse from parse.go', and drops the logger and all metadata providers.

**Why it matters:** Two consequences, and the second is quieter. First, the API and the CLI take different names for the same thing, so `?source=python` returns `400 Source 'python' not supported` -- a confusing answer, since `parse python` is the documented spelling everywhere else. Second, the upstream procedure for adding a language, which this repository's own documentation sets out in three steps, updates only `parse.go`; the new language appears in the CLI and silently does not appear in the server, with nothing to signal the omission.

Passing no metadata providers also means graphs from `/api/dependencies` never carry `repo_*` fields, so `--popups` and `--nebraska` have nothing to show for anything fetched through the web interface.

**Suggested fix:** Derive the server's map from the same registration the CLI uses, and accept both the language and registry spellings as aliases. That is a small refactor in `parse.go` to expose the parser table, and it removes the whole class of drift.

## 6. The published image is named for the wrong repository and only ever tagged latest

**Severity:** Low  
**Where:** `.github/workflows/docker-image.yml`

**What:** On every push to `main` the workflow builds and pushes `ghcr.io/${{ github.repository_owner }}/stacktower:latest`. The repository is `stacktower-docker`; the image is `stacktower`. `:latest` is the only tag -- no version, no commit SHA.

**Why it matters:** Nothing can be pinned. An operator running this image has no way to state which build they are on, no way to roll back after a bad push to `main`, and no way to tell two images apart. The name mismatch is minor on its own but compounds it: `ghcr.io/willtheorangeguy/stacktower` suggests a mirror of upstream's CLI rather than this fork's server image, which is the opposite of what it contains.

**Suggested fix:** Add `${{ github.sha }}` and a semver tag alongside `latest`, and name the image after the repository. `docker/metadata-action` does both in a few lines.

## 7. The final image stage is unpinned

**Severity:** Low  
**Where:** `Dockerfile`

**What:** The build stage pins `golang:1.25.5-alpine3.23`; the final stage is `FROM alpine:latest`.

**Why it matters:** Two builds of the same commit weeks apart can land on different Alpine releases, so a failure that appears in one and not the other cannot be attributed to the source. It is a small surface -- the final stage carries only a static `CGO_ENABLED=0` binary and the `blogpost` directory -- which is why this is Low; but the build stage right above it is pinned to the patch level, so the inconsistency looks like an oversight rather than a decision.

**Suggested fix:** Pin it to a release, `alpine:3.23` or narrower. `scratch` would also work given the binary is static, if nothing needs a shell for debugging.


---

## Also, across every repository

**`.bandit` is present on disk but untracked in git.** Verified in PyWorkout, treklogger,
skyscanner-cli, booking-cli, piggy, and aibot — the config file exists locally in each but
`git ls-files` does not know about it, so none of it reached GitHub.

The August 2026 security sweep therefore looks complete locally and landed nowhere. Worth
checking across all 44 repositories it covered.
