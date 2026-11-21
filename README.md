# Go REST API (SQLite v1)

Learning project: a **`/api/v1`** HTTP API in Go (SQLite) to replace [JSONPlaceholder](https://jsonplaceholder.typicode.com) for the Next.js learning app (`next_js_learning` in the same workspace) when ready.

**Status:** M0 complete — config, server, health, graceful shutdown. Implementation progress: [doc/MILESTONE_TRACKER.md](doc/MILESTONE_TRACKER.md).

## Prerequisites

- Go **1.22+** (module uses 1.27+)
- Run commands from this directory: `go_lang/rest_api`

## Configuration

Local settings live in **`config/local.yml`** (gitignored — create your own from team notes or copy fields below).

Example shape:

```yaml
env: "dev"
storage_path: "storage/storage.db"
http_server:
  address: "localhost:8080"
```

Optional overrides:

| Variable | Purpose |
|----------|---------|
| `CONFIG_PATH` | Path to YAML (default: `config/local.yml` via `-config` flag) |

```bash
go run ./cmd/post-api/main.go -config config/local.yml
```

## Run the server

```bash
go run ./cmd/post-api/main.go
```

You should see logs like `starting post-api` and `server started on port localhost:8080`.

Stop with **Ctrl+C** — the server shuts down gracefully.

## Health check

```bash
curl -s -i http://localhost:8080/api/v1/health
```

Expected: **200 OK**, `Content-Type: application/json`, body:

```json
{"status": "ok"}
```

Postman: import [doc/api/postman_collection.json](doc/api/postman_collection.json) (more routes as they are implemented).

## Project layout (early)

```text
cmd/post-api/main.go     # bootstrap: config, mux, server, signals
internal/config/         # YAML / env via cleanenv
internal/handler/        # HTTP handlers + RoutesRegister
constants/               # paths and HTTP methods
doc/                     # plan, milestones, OpenAPI, guides
storage/                 # SQLite file (when migrations are in place)
```

## Documentation

| Doc | Description |
|-----|-------------|
| [doc/MILESTONE_TRACKER.md](doc/MILESTONE_TRACKER.md) | What to build next (milestones M0–M10) |
| [doc/PLAN_SQLITE_API_V1.md](doc/PLAN_SQLITE_API_V1.md) | Decisions, API contract, schema |
| [doc/api/openapi.yaml](doc/api/openapi.yaml) | OpenAPI 3 spec |
| [doc/guides/route-registration.md](doc/guides/route-registration.md) | Central `ServeMux` registration |
| [doc/setup/config-and-server-startup-issues.md](doc/setup/config-and-server-startup-issues.md) | Config startup troubleshooting |

## Common commands

```bash
go mod tidy
go vet ./...
go build -o bin/post-api ./cmd/post-api
```

## Next milestone (M1 / M2)

- **M1:** JWT/CORS config fields, `go vet`, DB placeholder in `main`  
- **M2:** `modernc.org/sqlite`, migrations, `storage_path` — see milestone tracker

## Seed users (later)

When auth and seeds land (M3/M4), dev login passwords will be documented **here only** — not in committed config or hashes in docs.