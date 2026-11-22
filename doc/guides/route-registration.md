# Guide: central route registration (Go 1.22+ `ServeMux`)

**Milestone:** M1 — config polish & app wiring  
**You implement this** — follow the steps and verify with Postman/curl.  
**Current state:** Health is registered directly in `cmd/post-api/main.go` line ~44.

---

## Why register routes in one place?

- `main.go` stays a **bootstrap** file: config, db (later), server, signals
- New endpoints get added in **one** package (`internal/handler` or `internal/server`)
- Matches plan layering: handlers own HTTP wiring; `main` does not grow one `HandleFunc` per feature

---

## Recommended layout (v1)

```text
internal/handler/
  health.go       # HealthCheckHandler
  routes.go       # Register(mux *http.ServeMux)  ← you add this
  (later) post.go, comment.go, auth.go ...
cmd/post-api/main.go   # router := ...; handler.Register(router)
```

Keep path strings in `constants/` (you already have `HealthCheckPath` and HTTP methods).

---

## Step-by-step (your hands)

### 1. Create `internal/handler/routes.go`

- Package name: `handler` (same as `health.go`)
- Export a function, for example: `Register(mux *http.ServeMux)`
- Inside `Register`, move the registration you have in `main` today:

  - Pattern: `constants.GET + " " + constants.HealthCheckPath`
  - Handler: `HealthCheckHandler`

**Mux pattern rule:** Method must be uppercase **`GET`**, not `Get` (Go 1.22+).

### 2. Slim down `main.go`

- After `router := http.NewServeMux()`, call **`handler.Register(router)`** instead of a single `HandleFunc` for health
- Remove the duplicate `HandleFunc` line for health (only one registration)

### 3. Verify

```bash
go run ./cmd/post-api/main.go
curl -s -i http://localhost:8080/api/v1/health
```

Expect: **200**, `Content-Type: application/json`, `{"status": "ok"}`.

Postman: same as before.

### 4. Tick M1 in [MILESTONE_TRACKER.md](../MILESTONE_TRACKER.md)

- Partial credit toward “Create `internal/app` or move mux registration” — you don’t need a full `internal/app` yet; `handler.Register` is enough for now.

---

## Pattern to grow later (when you have a DB)

When M2 gives you `*sql.DB`, upgrade `Register` to take dependencies instead of only `mux`:

```go
type API struct {
    // DB *sql.DB   // M2+
}

func (api *API) Register(mux *http.ServeMux) {
    mux.HandleFunc(constants.GET+" "+constants.HealthCheckPath, HealthCheckHandler)
    // mux.HandleFunc(constants.GET+" "+constants.PostsPath, api.listPosts)
}
```

In `main`:

1. Load config  
2. Open DB (M2)  
3. `api := &handler.API{ /* DB: db */ }`  
4. `mux := http.NewServeMux()`  
5. `api.Register(mux)`  

Health can stay stateless with no DB; posts/comments handlers will use `API` or separate structs with `store` injected.

---

## Optional helpers (only if you want)

| Idea | Purpose |
|------|---------|
| `constants.Route(method, path string) string` | Avoid repeating `GET + " " + path` |
| `internal/server/server.go` with `New(cfg) (http.Handler, error)` | Moves mux + Register + middleware (CORS in M8) out of `main` |

Start with **`handler.Register(mux)`** only; refactor to `server.New` when CORS or middleware wraps the mux.

---

## Common mistakes

| Mistake | Fix |
|---------|-----|
| `Get /path` in pattern | Use `GET` |
| Register same route in `main` and `Register` | One place only |
| Handler in wrong package | HTTP handlers stay in `internal/handler`, not `main` |
| Import cycle | `routes.go` imports `constants`; handlers do not import `main` |

---

## What you are not doing yet

- Auth middleware on `Register` — M4  
- Post/comment routes — M5–M7  
- CORS wrapper — M8  

Only consolidate **health** first; add routes to `Register` as you build each milestone.