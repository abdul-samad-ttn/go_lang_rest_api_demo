# Milestone tracker — SQLite REST API v1

**Repository:** `go_lang/rest_api`  
**Companion specs:** [PLAN_SQLITE_API_V1.md](./PLAN_SQLITE_API_V1.md) (decisions & API contract), [doc/api/openapi.yaml](./api/openapi.yaml), [setup log](./setup/config-and-server-startup-issues.md)  
**Last updated:** 2026-11-22 (M7 comment **GET + POST + DELETE** verified Postman; **403/404 delete polish** deferred)  
**How you use this doc:** Check off tasks **as you** complete them. Effort is approximate **solo learning time** (not calendar deadlines). When stuck, read the linked plan section, try curl/Postman, then ask for hints — not full implementations.

### Owner working decisions (override plan until changed)

| Topic | Choice | Notes |
|--------|--------|--------|
| **Post seed count** | **10 posts** (ids **1–10**) in SQL migration | Plan §1 decision **#10** still says 50 for full Next/static build parity; use **10** for now in `migrations/` seed SQL. Bump later via a new migration or seed amendment if you want 50. |
| **Migration files** | **Split DDL vs seed** | `0001_smoke.up.sql` (runner test), `0002_schema.up.sql` (tables), `0003_seed.up.sql` (INSERTs). Runner loops sorted `migrations/*.sql` + `schema_migrations` skip/apply. |
| **Run / build** | From repo root | `go run ./cmd/post-api` or `go build -o bin/post-api ./cmd/post-api` — not bare `go build` or `go run cmd/post-api` (no `./`). Start from `rest_api/` so `migrations/` resolves. |
| **Layering & logs** | handler → store → models | **Store** returns `error` only (no logger). **Handler** maps errors to HTTP + safe client messages; optional logger on handlers later. |
| **Route registration** | Split by area | `RoutesRegisterHealthCheck`, `RoutesRegisterPosts`, `RoutesRegisterComments` in `internal/handler/routes.go` (wired from `main`). |
| **Create post/comment (interim)** | **Body `userId`** — **no JWT yet** | Until **M4**, implement `POST` post (M6) and `POST` comment (M7) using **`userId` from JSON body** (JSONPlaceholder-shaped). Plan §4.5 still targets token-only `userId` later — swap in M4 without changing Next GET shapes. |
| **Delete comment (interim)** | Query **`userId`** (camelCase) — **no JWT yet** | `DELETE /api/v1/comments/{commentId}?userId=2`. **Not** in `ServeMux` pattern. Owner match via `author_user_id` in SQL. **Follow-up:** distinct **403** vs **404** (`GetCommentAuthorUserId` + handler mapping). |
| **404 bodies** | Plain text for now | Missing post → **404** with message like `Post not found` (not `sql.ErrNoRows` text). JSON `{"error":"..."}` polish in **M8**. |

### Frontend target (`next_js_learning` on Vercel)

The hobby Next app on Vercel still calls **JSONPlaceholder** today (`API_BASE_URL` / `NEXT_PUBLIC_API_BASE_URL` → `https://jsonplaceholder.typicode.com`; see `lib/api/config.ts`, `.github/workflows/deploy-pages.yml`). Your Go API replaces that backend under **`/api/v1`** with the same **path shapes** Next already uses (`/posts`, `/posts/{id}`, comment routes).

| Vercel UI (your screenshots) | Next fetch today | Go v1 endpoint (when built) |
|------------------------------|------------------|----------------------------|
| Posts list (numbered titles) | `GET /posts` | `GET /api/v1/posts` |
| Post detail (title + body) | `GET /posts/1` | `GET /api/v1/posts/1` |
| Comments for post | `GET /posts/1/comments` | `GET /api/v1/posts/1/comments` |
| Single comment | `GET /comments/1` | `GET /api/v1/comments/1` |
| Add comment (form) | `POST /posts/1/comments` | `POST /api/v1/posts/1/comments` (**body `userId` interim**; JWT in M4) |

**Swap happens in M9:** point env to `http://localhost:8080/api/v1` locally first; Vercel needs a **deployed** Go API + CORS (M8) + public URL — until then production keeps JSONPlaceholder.

### Implementation log (what you have done — agent-maintained)

| When | What | Where |
|------|------|--------|
| 2026-11-21 | Config loads `config/local.yml` without required `CONFIG_PATH`; health + graceful shutdown | M0 |
| 2026-11-21 | Central routes: `handler.RoutesRegister` | `internal/handler/routes.go` |
| 2026-11-21 | Only route registered so far: **`GET /api/v1/health`** | `constants` + `health.go` |
| 2026-11-21 | SQLite open: `mkdir`, `sql.Open`, `Ping`, `PRAGMA foreign_keys` | `internal/database/open.go` |
| 2026-11-21 | DB wired in startup **before** server listen; `defer db.Close()` | `cmd/post-api/main.go` |
| 2026-11-21 | Migration runner: `schema_migrations`, transactions, skip/applied logs | `internal/database/migrate.go` |
| 2026-11-21 | `database.Migrate` called from `main` after `Open` | `cmd/post-api/main.go` |
| 2026-11-22 | Schema migration `users`, `posts`, `comments` | `migrations/0002_schema.up.sql` |
| 2026-11-22 | Seed: 5 users, **10** posts, comments on posts 1–5 | `migrations/0003_seed.up.sql` |
| 2026-11-22 | Post **model** (`json:"userId"`, etc.) | `internal/models/post.go` |
| 2026-11-22 | Post **store** `ListPosts`, `GetPostById` | `internal/store/post.go` |
| 2026-11-22 | Post **handlers** + `RoutesRegister(router, postsHandler)` | `internal/handler/posts.go`, `routes.go` |
| 2026-11-22 | **`GET /api/v1/posts`**, **`GET /api/v1/posts/{id}`** verified (Postman) | `constants` + handlers |
| 2026-11-22 | Dev password documented (plaintext `password` for seed users) | [README.md](../README.md) |
| 2026-11-22 | Postman + plan curl: login/register examples use seed password `password` | `doc/api/postman_collection.json`, `PLAN` §10 |
| 2026-11-22 | Comment **model/store/handler**; split route registrars | `internal/models/comment.go`, `store/comment.go`, `handler/comments.go`, `routes.go` |
| 2026-11-22 | **`GET /api/v1/posts/{postId}/comments`**, **`GET /api/v1/comments/{commentId}`** (Postman) | `constants` + handlers |
| 2026-11-22 | Comment GET polish: `postId` **int64**, empty **`[]`**, 404 **Comment not found** | model/store/handler + constants |
| 2026-11-22 | **`POST /api/v1/posts/{postId}/comments`** (interim body **`userId`** → `author_user_id`) | `store/comment.go`, `handler/comments.go`, `routes.go` |
| 2026-11-22 | **`DELETE /api/v1/comments/{commentId}?userId=`** (interim); `ON DELETE CASCADE` on `comments.post_id` | `handler/routes.go`, `store` `DELETE … author_user_id` |
| 2026-11-22 | DELETE verified Postman (e.g. comment **9**, `userId=1` → **204**); POST **201** id **13**; list post **1** updates | Postman + SQLite |
| — | DELETE **403/404** polish, post POST/DELETE, auth, CORS, Next env swap | **Follow-up** |

---

## Progress at a glance

| Milestone | Name | Status | Tasks done | Est. total |
|-----------|------|--------|------------|------------|
| **M0** | HTTP server & local run | **Done** | 10 / 10 | ~4–6 h |
| **M1** | Config polish & app wiring | **In progress** | 4 / 7 | ~2–3 h |
| **M2** | SQLite connect & migration runner | **Done** | 8 / 8 | ~4–6 h |
| **M3** | Schema + seed data | **Done** | 8 / 9 | ~5–8 h |
| **M4** | Auth (register, login, JWT, me, logout) | **Deferred** | 0 / 12 | ~8–12 h |
| **M5** | Posts — public reads | **Done** | 6 / 7 | ~4–6 h |
| **M6** | Posts — create & delete | Not started | 0 / 8 | ~5–7 h |
| **M7** | Comments — full CRUD | **In progress** | 9 / 10 | ~6–9 h |
| **M8** | CORS, errors, handler layout | Not started | 0 / 7 | ~4–6 h |
| **M9** | Wire Next.js to local API | Not started | 0 / 6 | ~2–4 h |
| **M10** | README, Postman sync, deploy prep | **In progress** | 3 / 5 | ~3–5 h |

**Rough total to v1 (local):** ~47–76 hours of hands-on work, spread over milestones.

### Contract & planning (already done — not counted in milestones)

- [x] Architecture decisions locked — `PLAN_SQLITE_API_V1.md` §1  
- [x] OpenAPI 3 spec — `doc/api/openapi.yaml`  
- [x] Postman collection — `doc/api/postman_collection.json`  
- [x] JSONPlaceholder compatibility notes — plan §3.6  

---

## Legend

| Symbol | Meaning |
|--------|---------|
| `[x]` | **Done** — you verified it (curl, logs, or sqlite) |
| `[~]` | **Partial** — works with workaround or missing polish |
| `[ ]` | **Not started** — your next learning slot |
| **Effort** | `30m` … `3h` = focused time for one learner pass (code + verify) |

**Verify column:** minimum check before you tick the box.

---

## M0 — HTTP server & local run

**Goal:** Run the API on `localhost:8080`, health probe, graceful stop. **You learn:** Go `ServeMux` (1.22+), config load, signals.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [x] | Load `config/local.yml` via `cleanenv` | 1h | Already in `internal/config`; understand struct tags | Log shows `localhost:8080`, `storage_path` |
| [x] | Run without extra env vars | 30m | Fix `log.Fatal` in `MustLoad` when `CONFIG_PATH` empty (plan setup #6) | `go run ./cmd/post-api` works with no export |
| [x] | `GET /api/v1/health` | 45m | Route + JSON `{ "status": "ok" }`, `Content-Type: application/json` | `curl -s -i localhost:8080/api/v1/health` → 200 |
| [x] | Route constants | 30m | `constants/constants.go` for method + path | Pattern is `GET /api/v1/health` |
| [x] | Graceful shutdown | 1h | Goroutine + `<-done` + `Shutdown(ctx)` | Ctrl+C → “shutdown gracefully”, no orphan bind |
| [~] | Error logging style | 20m | Prefer `logger.Fatalf("...: %v", err)` | Readable single-line fatals (deferred polish) |
| [x] | Remove debug `fmt.Println(cfg)` | 10m | Or gate behind `cfg.Env == "dev"` | Startup logs clean |
| [-] | Optional `GET /` welcome | 20m | Plain text OK (not required for Next) | **Skipped** for v1 |
| [x] | Document run command in `README.md` | 30m | Run, health curl, doc links | [README.md](../README.md) |
| [x] | Mark M0 complete in this file | 5m | Update table above → **Done** | M0 status **Done** (2026-11-21) |

**Milestone exit criteria:** Server starts, health 200, graceful stop; README stub with run instructions. **Met 2026-11-21.**

**References:** `doc/setup/config-and-server-startup-issues.md`, plan §6.4, §4.3.

**Agent note:** Owner implements code; see [AGENTS.md](../AGENTS.md).

---

## M1 — Config polish & app wiring

**Goal:** Extend config for auth/CORS later; thin `main.go`. **You learn:** env/YAML fields, composing `http.Handler`.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [ ] | Add `jwt_secret`, `jwt_ttl` to config | 45m | YAML + struct tags; **never commit real secrets** | Load succeeds; zero value fails fast or uses dev default in `local.yml` only (gitignored) |
| [ ] | Add `cors_allowed_origins` (slice or string) | 30m | Plan §2 / §6 | Config prints in dev log |
| [x] | Central route registration | 45m | `handler.RoutesRegister(mux, postsHandler)` — see [route-registration guide](./guides/route-registration.md) | Routes live in `RoutesRegister`, not scattered in `main` |
| [~] | Create `internal/app` or `internal/server` package | 1–2h | Optional later: `New(cfg) http.Handler` + CORS (M8) | Defer until middleware |
| [x] | Move health handler to `internal/handler/health.go` | 45m | Handler func `HealthCheckHandler` | Health still 200 via curl |
| [x] | Wire DB in `main` | 30m | `database.Open(cfg.StoragePath, logger)` + `defer db.Close()` | Startup logs “database opened successfully” |
| [~] | `go vet ./...` and `go build ./...` in CI habit | 20m | Use `go build ./...` from repo root (not bare `go build`) | Exit 0 |

**Milestone exit criteria:** `main.go` < ~40 lines of wiring; health lives under `internal/handler`.

**References:** plan §5.3, §6.1.

---

## M2 — SQLite connect & migration runner

**Goal:** Open DB file, enforce FKs, apply versioned SQL. **You learn:** `database/sql`, `modernc.org/sqlite`, migration patterns.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [x] | `go get modernc.org/sqlite` | 15m | Blank import in open helper | `go mod tidy` |
| [x] | `internal/database/open.go` | 1h | `Open(path, logger) (*sql.DB, error)` + `PRAGMA foreign_keys = ON` | Startup: “database opened successfully” |
| [~] | DSN format for SQLite file | 30m | Using `file:` + path (no query params yet) | Works for `storage/storage.db`; tune later if needed |
| [x] | `schema_migrations` table + runner | 2–3h | Read `migrations/*.sql`, record version, skip applied | Second start: skip logs, no duplicate errors |
| [x] | Versioned SQL files | 30m | `0001_smoke`, `0002_schema`, `0003_seed` | Three rows in `schema_migrations` on fresh DB |
| [x] | Call migrations from `main` on startup | 45m | Fail fast on migration error | Log: applied / already applied |
| [x] | Close DB on shutdown | 30m | `defer db.Close()` in `main` | Clean restart |
| [x] | Document inspecting DB | 20m | `sqlite3 storage/storage.db ".tables"` | [README.md](../README.md) seed users + query example |

**Milestone exit criteria:** App starts, migrations table exists, FK pragma on. **Met 2026-11-22.**

**References:** plan §5.2, §6.6, decision #13.

---

## M3 — Schema + seed data

**Goal:** Tables + 5 users, **10 posts** (see owner table above), comments on posts 1–5. **You learn:** SQL DDL, FKs, bcrypt in seed, idempotent seeds.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [x] | Migration: `users` table | 45m | Plan §5.1 | `.schema users` |
| [x] | Migration: `posts` + FK to users | 45m | FK to `users(id)` in `0002_schema` | Tables present |
| [x] | Migration: `comments` + FKs | 45m | Include `author_user_id` | `comments` table present |
| [x] | Seed 5 users (ids 1–5) | 1–2h | bcrypt hashes; emails per plan §5.2 | `SELECT id,email FROM users` → 5 rows |
| [x] | Document dev password in `README.md` only | 15m | Decision #15 | Not in plan/tracker commits |
| [x] | Seed **10** posts (ids 1–10, `user_id` 1–5) | 45m–1h | Owner decision 2026-11-21 (not plan #50 yet) | `SELECT COUNT(*) FROM posts` → **10** |
| [x] | Seed comments for posts 1–5 | 1h | Enough ids for Next comment routes | `SELECT COUNT(*) FROM comments` > 0 |
| [x] | Idempotent seed strategy | 1h | Runner applies each file once via `schema_migrations` | Restart skips `0003_seed` |
| [ ] | Optional: health checks DB | 45m | e.g. `SELECT 1` in health (stretch) | Health 200 when DB up |

**Milestone exit criteria:** Fresh `storage.db` from one startup has full seed; FKs enforced. **Met 2026-11-22.**

**References:** plan §5.1–5.2.1.

---

## M4 — Auth (register, login, JWT, me, logout)

**Goal:** Users can register/login; middleware validates Bearer token. **You learn:** bcrypt, JWT claims, middleware wrapping handlers.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [ ] | `internal/store/user.go` — CRUD by email/id | 1–2h | SQL only | Unit test or sqlite manual |
| [ ] | `POST /api/v1/auth/register` | 1–2h | 201 User JSON; 409 duplicate | curl / Postman |
| [ ] | Password validation (min length) | 30m | 400 on weak password | curl |
| [ ] | `POST /api/v1/auth/login` | 1–2h | 200 `{ token, user }` | Postman saves token |
| [ ] | JWT: sign with `jwt_secret`, `sub`, `exp` | 1–2h | `golang-jwt/jwt/v5` | Decode at jwt.io (dev only) |
| [ ] | Middleware `RequireAuth` | 1–2h | 401 if missing/invalid Bearer | Protected route test |
| [ ] | `GET /api/v1/auth/me` | 45m | 200 User | curl with token |
| [ ] | `POST /api/v1/auth/logout` | 30m | **204** empty (stateless) | Postman |
| [ ] | Handler package `internal/handler/auth.go` | 1h | Thin HTTP layer | No SQL in handler |
| [ ] | JSON errors `{ "error": "..." }` | 45m | Plan §4.4 | 401/400 bodies |
| [ ] | Update OpenAPI if shapes drift | 30m | Keep contract-first | Diff yaml |
| [ ] | Postman: register + login flows | 30m | Collection vars | End-to-end in Postman |

**Milestone exit criteria:** Alice (seed) logs in; token works on `/auth/me`; register creates user 6+.

**References:** plan §4.5, openapi `/auth/*`.

---

## M5 — Posts — public reads

**Goal:** Next can list and load posts. **You learn:** JSON encoding, camelCase mapping, 404 semantics.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [x] | `internal/models/post.go` + `internal/store/post.go` | 1–2h | Map `user_id` → JSON `userId` | List returns 10 rows |
| [x] | `GET /api/v1/posts` | 1h | 200 raw array `Post[]` | Postman / curl; JSONPlaceholder shape |
| [x] | `GET /api/v1/posts/{id}` | 45m | 200 or 404 | Postman id 3; 999 → 404 |
| [x] | Invalid id → 400 | 30m | Non-numeric id | `Invalid id` message |
| [x] | `internal/handler/posts.go` | 1h | Register on mux via `RoutesRegister` | — |
| [-] | Optional auth middleware pass-through | 30m | Public GET; ignore optional Bearer | **Deferred** with M4 |
| [x] | Postman smoke | 30m | Two GETs | List + by id verified 2026-11-22 |
| [~] | `errors.Is(sql.ErrNoRows)` → 404 only | 15m | Other store errors → 500 | 404 body is safe text; 500 mapping optional polish |

**Milestone exit criteria:** Responses match `next_js_learning/lib/api/types.ts` Post type. **Met 2026-11-22.**

**References:** plan §3.6, §4.1, §4.2 (GET).

---

## M6 — Posts — create & delete

**Goal:** Authenticated create; author-only delete with 204. **You learn:** auth context, ownership, status 201/204.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [ ] | `store` Create post | 45m | `user_id` from caller | SQL insert |
| [ ] | `store` Delete post by id | 45m | CASCADE comments | Comments gone after delete |
| [ ] | `POST /api/v1/posts` (interim: no JWT) | 1–2h | Body `{title,body,userId}` per owner table; **M4** adds token | 201 + full Post |
| [ ] | `DELETE /api/v1/posts/{id}` | 1–2h | Author only → 204; other user → 403 | curl with alice vs bob token |
| [ ] | Unauthenticated → 401 | 30m | Both routes | curl |
| [ ] | Missing post delete → 404 | 30m | | curl |
| [ ] | Postman protected post flows | 45m | Create + delete | Collection |
| [ ] | Manual cascade check | 30m | Delete post with comments | GET comments empty / 404 |

**Milestone exit criteria:** Plan §4.1 delete flow steps 1–7 satisfied via curl/Postman.

**References:** plan §4.1, §4.5.3.

---

## M7 — Comments — full CRUD

**Goal:** Match Next comment routes + auth on write/delete. **You learn:** nested routes, hidden `author_user_id` in JSON.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [x] | `store` ListByPostID, GetByID, Create, Delete | 2h | CRUD store methods; interim author on DELETE in SQL | Postman 2026-11-22 |
| [x] | `GET /api/v1/posts/{postId}/comments` | 1h | 200 array; 404 if post missing | Postman post 1 → 2 comments (2026-11-22) |
| [x] | `GET /api/v1/comments/{id}` | 45m | 200 / 404 | Postman **100** → 404; id **1** → 200 (2026-11-22) |
| [x] | `POST /api/v1/posts/{postId}/comments` (interim: no JWT) | 1–2h | Body + **`userId`** for `author_user_id`; public JSON without `author_user_id` | Postman post **1** → **201**; list shows new id (2026-11-22) |
| [x] | `DELETE /api/v1/comments/{id}` (interim `?userId=`) | 1–2h | Route, author match in SQL, **204** empty | Postman comment **9**, `userId=1` → **204** (2026-11-22) |
| [x] | JSON `postId` camelCase | 30m | `int64` + `json:"postId"`; list `[]` not `null` | Postman post **20** → `[]` |
| [x] | Postman comment suite | 45m | GET list/by-id, POST create, DELETE with **`userId`** query | Collection updated |
| [ ] | Cross-user delete denied | 30m | Bob cannot delete Alice’s comment | **403** (deferred — may **500** until author SELECT split) |
| [ ] | DELETE missing comment → **404** | 30m | `GetCommentAuthorUserId` + `sql.ErrNoRows` in handler (not `errors.New` in store) | curl |
| [ ] | List after post delete | 30m | Post gone → nested GET 404 | curl |
| [ ] | Update openapi if needed | 30m | | yaml matches behavior |

**Milestone exit criteria:** All comment routes in plan §4.2 work; shapes match `IComment`.

**References:** plan §4.2, decision #9, #18.

---

## M8 — CORS, errors, handler layout

**Goal:** Browser calls from `localhost:3000` work; consistent errors. **You learn:** CORS preflight, middleware chains.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [ ] | CORS middleware | 1–2h | Allow origin `http://localhost:3000`, headers incl. `Authorization` | Browser preflight or curl `-X OPTIONS` |
| [ ] | Wrap mux with middleware | 45m | CORS outermost | — |
| [ ] | Centralize JSON error helper | 1h | 400/401/403/404/409/500 | One handler test |
| [ ] | `Content-Type` on all JSON responses | 30m | `Set` not duplicate `Add` | curl -i |
| [ ] | 204 responses: no body | 30m | DELETE, logout | curl -i |
| [ ] | Review handler → store boundary | 1h | No SQL in handlers | Quick grep |
| [ ] | `go test ./...` (at least store tests) | 1–2h | Start with one store test | CI green locally |

**Milestone exit criteria:** Next dev server can `fetch` GET from browser (if you test); mutations ready for token later.

**References:** plan §2 goals (CORS), decision #7.

---

## M9 — Wire Next.js to local API

**Goal:** Replace JSONPlaceholder for **read** paths (+ comment create when token ready). **You learn:** env vars, build-time API dependency.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [ ] | `next_js_learning/.env.local` | 15m | `API_BASE_URL=http://localhost:8080/api/v1` | — |
| [ ] | `NEXT_PUBLIC_API_BASE_URL` same base | 15m | Client fetch | — |
| [ ] | Run Go API, then `npm run dev` | 30m | Posts list/detail | Browser |
| [ ] | `next build` with API up | 45m | `generateStaticParams` | Build succeeds; **10** post pages (current seed) |
| [ ] | Comment pages for seeded data | 30m | List/detail | Browser |
| [ ] | Document “start API before build” | 20m | README in either repo | Teammate note |

**Milestone exit criteria:** Next reads from your API locally; no JSONPlaceholder for integrated routes.

**References:** plan §3.1, §5.2.1.

---

## M10 — README, Postman sync, deploy prep (local complete)

**Goal:** Another dev (or future you) can run the stack; deploy deferred. **You learn:** docs as part of shipping.

| Done | Task | Effort | What you do | Verify |
|:----:|------|--------|-------------|--------|
| [x] | Root `README.md`: run, seed users, health, migrate | 1–2h | Dev password, `CONFIG_PATH`, migrations on startup | Fresh clone walkthrough |
| [~] | Postman matches implemented routes | 45m | Auth still spec-only | **Health**, **Posts**, **Comments** GET + POST + DELETE (interim) verified |
| [x] | Mark phases in `PLAN_SQLITE_API_V1.md` changelog | 20m | One line per milestone done | P1 + P3 done 2026-11-22 |
| [x] | `.gitignore` check (`local.yml`, `.db`) | 15m | No secrets committed | `storage/`, `config/local.yml` ignored |
| [ ] | Deploy host decision (optional note) | 30m | plan §8 when ready | Paragraph in README |

**Milestone exit criteria:** **v1 local** declared done in this tracker; deploy is explicit follow-up.

---

## Suggested learning order (one path)

Do milestones **in order** (M0 → M10). Within a milestone, do tasks **top to bottom** — later tasks often depend on earlier ones.

```text
M0 → M1 → M2 → M3 → M5 → M7 (comment GETs) → M6/M7 POSTs (body userId) → M8 → M9 → M4 (JWT) → M6/M7 auth hardening → M10
```

**Current position (2026-11-22):** M2, M3, M5 **done**; **M7** comment CRUD **interim** verified (GET/POST/DELETE Postman). **Next:** M7 **403/404** delete polish (optional), **M8** CORS + **M9** Next reads, or **M6** post POST. **M4 JWT** later.

---

## How to work with your assistant (learning mode)

1. **Pick one unchecked task** in the current milestone.  
2. **Try it yourself** (docs, plan §, driver readme).  
3. Ask: “I'm on M2 task X — I'm stuck at …” or “review my approach for …” — not “implement M4 for me.”  
4. After verify passes, **change `[ ]` → `[x]`** in this file and update the progress table counts.  
5. When a milestone exit criteria is met, set milestone **Status** to **Done** in the table at the top.

---

## Changelog (this document)

| Date | Change |
|------|--------|
| 2026-11-21 | Initial tracker: M0–M10, M0 partial completion from local verification |
| 2026-11-21 | Health confirmed working (`GET /api/v1/health` → 200 JSON) |
| 2026-11-21 | **M0 Done** — root [README.md](../README.md); central `RoutesRegister`; M1 route task checked |
| 2026-11-22 | **M2 Done** — migration loop, `schema_migrations`, split SQL files; **M3 Done** — schema + 10-post seed |
| 2026-11-22 | **M5 Done** — model/store/handler; GET posts; owner: interim **body `userId`** on create (JWT later) |
| 2026-11-22 | README dev password; Postman/plan login examples → `password`; **M10** partial (3/5); P1/P3 marked in plan |
| 2026-11-22 | **M7** — comment GET routes + Postman; route registrars split; **M7** 5/10 |
| 2026-11-22 | **M7** GET polish verified (Postman); tracker **7/10** — POST comment next |
| 2026-11-22 | **M7 POST comment** verified (Postman); interim **`userId`** in body; tracker **8/10** |
| 2026-11-22 | **M7 DELETE comment** — route + interim **`?userId=`**; **403/404** polish deferred in tracker |
| 2026-11-22 | **M7 DELETE** Postman verified (**204**); tracker **9/10**; Postman collection DELETE `userId` query |