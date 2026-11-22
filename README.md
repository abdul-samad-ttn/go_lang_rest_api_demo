# Go REST API (SQLite v1)

Learning project: a **`/api/v1`** HTTP API in Go (SQLite) to replace [JSONPlaceholder](https://jsonplaceholder.typicode.com) for the Next.js learning app (`next_js_learning` in the same workspace) when ready.

**Status:** Migrations, seed, public **posts** and **comment GETs** are in place. Implementation progress: [doc/MILESTONE_TRACKER.md](doc/MILESTONE_TRACKER.md).

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

Postman: import [doc/api/postman_collection.json](doc/api/postman_collection.json). **Working today:** Health, **GET posts** (list + by id), **GET comments** (by post + by id). Auth/comment **POST** pending M7/M4.

```bash
curl -s http://localhost:8080/api/v1/posts | head -c 200
curl -s http://localhost:8080/api/v1/posts/1
curl -s http://localhost:8080/api/v1/posts/1/comments
curl -s http://localhost:8080/api/v1/comments/1
```

Migrations run automatically on startup (`database.Migrate`); no separate migrate CLI.

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

## What's next

See [doc/MILESTONE_TRACKER.md](doc/MILESTONE_TRACKER.md). Suggested order:

- **M7 POST** — `POST /posts/{postId}/comments` (interim **`userId` in body** for `author_user_id`)
- **M9** + **M8** — point `next_js_learning` at `http://localhost:8080/api/v1`; CORS for `localhost:3000`
- **M6** — `POST /posts` with body `userId` when you want create-from-API

## Dev seed users (local)

`migrations/0003_seed.up.sql` inserts **5 users** (ids 1–5). Each row stores a **bcrypt** hash in `password_hash`; the plaintext is for local development only.

| Field | Value |
|-------|--------|
| **Dev password (all seed users)** | `password` |

| id | email | display name |
|----|-------|----------------|
| 1 | `alice@example.com` | Alice |
| 2 | `bob@example.com` | Bob |
| 3 | `carol@example.com` | Carol |
| 4 | `dave@example.com` | Dave |
| 5 | `eve@example.com` | Eve |

Auth endpoints are not wired yet — use these when login/register land (M4). Document the plaintext **only here** (not in plan/tracker commits or YAML). Do not log or return `password_hash` from the API.

Optional: inspect the DB after migrate:

```bash
sqlite3 storage/storage.db "SELECT id, email, display_name FROM users ORDER BY id;"
```