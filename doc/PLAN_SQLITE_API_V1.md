# Plan: SQLite REST API (`/api/v1`) + Next.js integration

**Repository:** `go_lang/rest_api`  
**Consumer:** `next_js_learning` (JSON over HTTP; today mimics JSONPlaceholder paths under a configurable base URL)  
**Status:** **P0–P1 complete; P3 complete; P5 comment GET + interim POST/DELETE complete** (`userId` in body/query until M4). Auth (P2), DELETE **403/404** polish, post writes, CORS, Next wiring pending — **[MILESTONE_TRACKER.md](./MILESTONE_TRACKER.md)**.  
**Owner intent:** Learn by building CRUD in Go with SQLite; replace JSONPlaceholder when the API is ready.

---

## 1. Decisions (locked for v1)

| # | Topic | Decision |
|---|--------|----------|
| 1 | Database | **SQLite** — single file, no external DB service; use existing `storage_path` in config (e.g. `storage/storage.db`). |
| 2 | API versioning | All HTTP routes under **`/api/v1`** prefix. Next.js uses one **base URL** env var; service paths stay relative (`/posts`, `/posts/:id`, …). |
| 3 | Users | **`users` table**; **migration seeds exactly 5 users**. Posts reference `user_id` (FK). **Registration + authentication** are **in scope** for v1 (see §4.5). |
| 4 | CRUD scope (v1) | **Posts:** create, list all, get by id, **delete (API — required)**. **Comments:** create, list by post, get by id, delete. Post/comment **update (PATCH)** optional stretch. |
| 5 | Deployment | **Local dev first** (`localhost`); free host **not decided** until the app works end-to-end locally (§8 deferred). |
| 6 | Go layering | **Handler flow for v1:** `handler` → `store` (no separate `service` package yet). Handlers own HTTP + light orchestration (auth context, status codes); `internal/store` owns SQL. Revisit `internal/service` only if handlers grow unwieldy. |
| 7 | CORS | **Enabled** for browser calls from Next.js (especially client-side comment form via `NEXT_PUBLIC_API_BASE_URL`). Must allow **`Authorization`** header when the front end sends tokens later. |
| 8 | JSON contract | **Request/response bodies and relative paths must match [JSONPlaceholder](https://jsonplaceholder.typicode.com/guide/)** for every endpoint Next calls today (see **§3.6**). Go may add `/api/v1` + auth routes; must not break shapes or paths under the base URL. |
| 9 | Comment **create** | **`Authorization` required** (JWT). **Any logged-in user** may create on any post (includes post owner). Request body still `{ name, email, body }`; response JSON unchanged. |
| 10 | Seed post count | **Locked: 50 posts** (ids 1–50), users 1–5 — see **§5.2.1**. |
| 18 | Comment **delete** | **Comment author only** (user who created it, via JWT `sub`). Others → **403**. Store `author_user_id` on comment (not exposed in JSONPlaceholder response). |
| 19 | Client JWT usage | When logged in, send **`Authorization: Bearer`** on **every API call where possible** (Next: extend `lib/api/http.ts` + store token after login). Public GETs still work **without** token; server ignores optional Bearer on reads in v1. |
| 11 | Health | **`GET /api/v1/health`** only (liveness JSON or plain `ok`). |
| 12 | `POST /posts` | **JWT required.** Body **`{ "title", "body" }` only** — **ignore `userId`** in body if sent; set `userId` from token. Response full **Post** JSON. |
| 13 | SQLite driver | **`modernc.org/sqlite`** (pure Go, `database/sql`) — see **§6.6**. |
| 14 | `DELETE` responses | **`204 No Content`**, empty body (posts and comments). |
| 15 | Dev seed passwords | Dummy password documented **only in project `README.md`** (never in plan/tracker/commits of hash). |
| 16 | Refresh tokens | **Out of scope** for v1. |
| 17 | Logout | **In scope:** `POST /api/v1/auth/logout` with valid Bearer → **`204`**; client discards token (stateless JWT; no server session store in v1). |
| 20 | API docs | **OpenAPI 3** (`doc/api/openapi.yaml`) + **Postman** (`doc/api/postman_collection.json`); see `doc/api/README.md`. Update when implementation ships. |

### Scope clarifications (2026-11-21)

| Topic | In v1? | Notes |
|--------|--------|--------|
| User **registration** | ✅ Yes | Public endpoint; store password hash only. |
| User **authentication** | ✅ Yes | Login → token; middleware on protected routes. |
| **Post DELETE** (API) | ✅ Yes | Full server flow (auth + ownership rules + cascade comments). Verify with curl/Postman. |
| **Post DELETE** (Next.js UI) | ❌ No | Ignored for now; no delete button in Next. |
| **Post CREATE** (Next UI) | ❌ No | Same as before — API/Postman first unless added later. |

---

## 2. Goals and non-goals

### Goals

- Provide a **stable JSON API** that lets the Next.js app replace JSONPlaceholder with minimal changes (base URL + `/api/v1` prefix) for **read** flows and **comment create** (unchanged UX).
- Implement **v1 CRUD** for posts and comments plus **user register/login** and **protected mutations** where required (especially **post delete**).
- Use **versioned SQL migrations** including **seed of 5 users**, then posts/comments seed tied to those users.
- Keep local dev **simple:** `CONFIG_PATH`, SQLite file, `go run cmd/post-api/main.go`.
- Document **deployment options** and **SQLite-on-PaaS** pitfalls before choosing a host.

### Non-goals (v1)

- PostgreSQL migration (future).
- Next.js UI for **delete post** (or create post).
- Comment **update** unless added as v1.1.
- Post **update** (PATCH/PUT) unless added as v1.1.
- OAuth / social login / email verification / password reset.
- **Refresh tokens** (access JWT only; short-ish TTL + re-login).
- Route handlers inside Next.js (`app/api/...`) — data stays in Go only.

---

## 3. Next.js integration contract

### 3.1 Environment variables (consumer app)

Next reads a **single base** (no path logic in `lib/api/http.ts`):

| Variable | Used by | Example (local) |
|----------|---------|-----------------|
| `API_BASE_URL` | Server Components, build-time `generateStaticParams` | `http://localhost:8080/api/v1` |
| `NEXT_PUBLIC_API_BASE_URL` | Client Components (e.g. add comment form) | `http://localhost:8080/api/v1` |

**Important:** Base URL must **include** `/api/v1` (no trailing slash required if paths in Next start with `/posts`).

**Auth and Next (v1):** **GET** posts/comments stay public (no token required). **Mutations require JWT** on Go. **When the user is logged in**, Next should attach the Bearer token on **all** `fetch` calls through `lib/api/http.ts` (and auth helpers) where a token exists — see decision **#19**. Login UI + token storage is a Next follow-up; until then use curl/Postman for protected routes.

Files that define usage today:

- `next_js_learning/lib/api/config.ts` — `apiBaseUrl` from env.
- `next_js_learning/lib/api/posts.ts` — `GET /posts`, `GET /posts/:id`.
- `next_js_learning/lib/api/comments.ts` — `GET /posts/:id/comments`, `GET /comments/:id`, `POST /posts/:id/comments`.
- `next_js_learning/lib/api/http.ts` — GET: **404 → `null`**; other errors throw. POST: non-OK throws.

### 3.2 JSON shapes (must match `lib/api/types.ts`)

**Post** (unchanged for Next)

```json
{
  "id": 1,
  "userId": 1,
  "title": "string",
  "body": "string"
}
```

**Comment**

```json
{
  "id": 1,
  "postId": 1,
  "name": "string",
  "email": "string",
  "body": "string"
}
```

**User** (API-only for auth; not used by current Next types)

```json
{
  "id": 1,
  "email": "user1@example.com",
  "displayName": "User One"
}
```

Never return `password` or `password_hash` in JSON.

Go structs should use JSON tags `userId`, `postId`, `displayName` (camelCase).

### 3.3 Next.js routes vs API (reference)

| Next route (UI) | API calls today | v1 API after migration |
|-----------------|-----------------|-------------------------|
| `/posts` | `GET /posts` | `GET /api/v1/posts` (public) |
| `/posts/[postsId]` | `GET /posts/:id` | `GET /api/v1/posts/:id` (public) |
| `/posts/.../comments` | `GET /posts/:id/comments` | `GET /api/v1/posts/:id/comments` (public) |
| `/posts/.../comments/[commentId]` | `GET /comments/:id` | `GET /api/v1/comments/:id` (public) |
| Add comment (form/modal) | `POST /posts/:id/comments` | `POST /api/v1/posts/:id/comments` (**JWT required**) |

\*Go API requires login; Next must be updated later to attach Bearer token on comment create (or test via curl/Postman until then).

**Not wired in Next v1:** `POST /posts`, `DELETE /posts/{id}`, auth endpoints — test with curl/Postman.

### 3.4 CORS requirements

- **Methods:** at least `GET`, `POST`, `DELETE`, `OPTIONS`.
- **Headers:** `Content-Type`, `Accept`, **`Authorization`** (for API clients and future Next auth).
- **Origins:** configurable allowlist (dev: `http://localhost:3000`; production: Vercel + GitHub Pages origins when known).
- Preflight must succeed for client-side `createComment`.

### 3.5 Static export / GitHub Pages

- Build-time fetch needs API reachable during `next build` (or reduced `generateStaticParams`).
- GHP runtime uses **browser** `fetch` for POST → public HTTPS API + CORS.

### 3.6 JSONPlaceholder compatibility (Next.js contract)

**Source of truth (public API):** [JSONPlaceholder Guide](https://jsonplaceholder.typicode.com/guide/)  
**Verified live examples (2026-11-21):** `GET /posts/1`, `GET /posts/1/comments`, `GET /comments/1` on `jsonplaceholder.typicode.com`.

**How Next maps URLs today**

| Layer | Value |
|--------|--------|
| JSONPlaceholder | `https://jsonplaceholder.typicode.com` + path |
| Go API (planned) | `https://<host>/api/v1` + **same path** |
| Next env | `API_BASE_URL` = full base including `/api/v1` |
| Next HTTP paths | `/posts`, `/posts/{id}`, … (unchanged in `lib/api/*.ts`) |

So: `fetch(apiBaseUrl + '/posts/2')` must behave like `fetch('https://jsonplaceholder.typicode.com/posts/2')` for status codes and JSON shape.

#### 3.6.1 Endpoints the Next app actually calls (must match)

| Method | Path (after base URL) | Request body | Success response | Notes |
|--------|------------------------|--------------|------------------|--------|
| `GET` | `/posts` | — | **`Post[]`** (raw JSON array, no wrapper) | Same fields as guide; 100 items is fine; empty `[]` only if you intend no posts |
| `GET` | `/posts/{id}` | — | **`Post`** object | Missing id → **`404`** (Next `apiGet` treats as `null` → `notFound()`) |
| `GET` | `/posts/{postId}/comments` | — | **`Comment[]`** raw array | Each item includes **`postId`** (usually equals URL segment). Next does not require 404 if post missing here—it checks post first via `GET /posts/{id}` |
| `GET` | `/comments/{id}` | — | **`Comment`** object | Missing → **`404`**. Comment detail page checks **`commentData.postId === postsId` from URL |
| `POST` | `/posts/{postId}/comments` | **`{ "name", "email", "body" }` only** | **`Comment`** with **`id`**, **`postId`**, `name`, `email`, `body` | `Content-Type: application/json`. **`postId` in response must match URL** (not only body fields). New `id` from DB. |

**Not called by Next today** (still align with JSONPlaceholder when you implement CRUD / Postman):

| Method | Path | JSONPlaceholder request | JSONPlaceholder response |
|--------|------|-------------------------|---------------------------|
| `POST` | `/posts` | `{ "title", "body", "userId" }` | Full **`Post`** including new **`id`** (guide shows `id: 101`) |
| `DELETE` | `/posts/{id}` | — | Guide shows **`{}`** and fake success; Go may use **`204`** empty body (Next ignores this until UI exists) |
| `PUT` / `PATCH` | `/posts/{id}` | per guide | Full or partial **Post** — optional for v1 |

#### 3.6.2 Resource JSON (field names and types)

Must match `next_js_learning/lib/api/types.ts` and JSONPlaceholder **camelCase** keys:

**Post**

| Field | Type | Required on read |
|-------|------|------------------|
| `id` | number | yes |
| `userId` | number | yes |
| `title` | string | yes |
| `body` | string | yes |

Example (live JSONPlaceholder; key order irrelevant):

```json
{
  "userId": 1,
  "id": 1,
  "title": "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
  "body": "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\n..."
}
```

**Comment**

| Field | Type | Required on read |
|-------|------|------------------|
| `id` | number | yes |
| `postId` | number | yes |
| `name` | string | yes |
| `email` | string | yes |
| `body` | string | yes |

Example (single comment from JSONPlaceholder):

```json
{
  "postId": 1,
  "id": 1,
  "name": "id labore ex et quam laborum",
  "email": "Eliseo@gardner.biz",
  "body": "laudantium enim quasi est quidem magnam voluptate ipsam eos\n..."
}
```

**Do not**

- Wrap lists as `{ "data": [...] }` or `{ "posts": [...] }`.
- Use snake_case in JSON (`user_id`, `post_id`) unless you add a translation layer (plan: **no layer** — emit camelCase from Go `json` tags).
- Omit `userId` on posts or `postId` on comments.
- Return strings for numeric ids (`"1"` breaks strict TS and comparisons like `commentData.postId !== Number(postsId)`).

#### 3.6.3 HTTP semantics Next depends on

| Situation | Expected behavior |
|-----------|---------------------|
| `GET` missing post/comment | **404** (not `200` + `null`) |
| `GET` list endpoints | **200** + array |
| `POST` comment create success | **2xx** + JSON **Comment** (JSONPlaceholder fakes **201**; any `response.ok` works with current `apiPost`) |
| `POST` validation failure | non-OK status → Next throws |
| Response `Content-Type` | `application/json` for JSON bodies |

#### 3.6.4 Nested comments vs JSONPlaceholder docs

The [guide](https://jsonplaceholder.typicode.com/guide/) documents `GET /posts/1/comments` (equivalent to `GET /comments?postId=1`).  
**Next uses only the nested form** for list + create:

- `GET /posts/{postId}/comments`
- `POST /posts/{postId}/comments`

You do **not** need `GET /comments?postId=` for Next v1 unless you want parity with JSONPlaceholder extras.

Official examples often show `POST /comments` with `postId` in the body; **your API must still implement nested `POST`** with body **`{ name, email, body }`** and set **`postId` from the URL**.

#### 3.6.5 Seeds and `userId`

JSONPlaceholder posts use `userId` **1–10** (10 posts per user). This plan seeds **5 users** (ids **1–5**). That is compatible as long as every seeded post returns numeric **`userId`** in **1..5**. Next renders `userId` only indirectly (type exists); list/detail use `title` and `body`.

#### 3.6.6 Extensions that must not break Next

| Feature | Rule |
|---------|------|
| `/api/v1` prefix | Only in Go base URL env, not in Next path strings |
| `/auth/register`, `/auth/login`, … | New paths; different JSON — OK |
| `POST /posts` with auth | Prefer JSONPlaceholder-shaped **request/response** when body is `{ title, body, userId }`; if auth sets `userId` from token, still return full **Post** JSON |
| `DELETE /posts/{id}` | Protected; response body may differ from JSONPlaceholder `{}` |

---

## 4. API specification (`/api/v1`)

Base path: **`/api/v1`**. All bodies and responses: **`application/json`** unless noted.

### 4.1 Posts

| Method | Path | Auth | Description | Success | Errors |
|--------|------|------|-------------|---------|--------|
| `GET` | `/posts` | Public | List all posts | `200` + `Post[]` | `500` |
| `GET` | `/posts/{id}` | Public | Get one post | `200` + `Post` | `404`, `400` |
| `POST` | `/posts` | **Required** | Create post for logged-in user | `201` + `Post` | `400`, `401` |
| `DELETE` | `/posts/{id}` | **Required** | Delete post if allowed (see §4.5.3) | `204` | `401`, `403`, `404`, `400` |

**Create post body (v1):**

```json
{
  "title": "required",
  "body": "required"
}
```

- **JWT required.** **`userId` is never taken from the body** (ignore if client sends JSONPlaceholder-style `userId`). Set from token `sub`.
- **Response:** `201` + full **Post** (`id`, `userId`, `title`, `body`).
- Unauthenticated → **`401`**.

**Post delete flow (v1 — backend only):**

1. Client sends `DELETE /api/v1/posts/{id}` with `Authorization: Bearer <token>`.
2. Middleware validates token → `userId`.
3. Load post; if missing → `404`.
4. If `post.user_id != authenticated userId` → `403` (v1 rule: **only author may delete**).
5. Delete post; SQLite **ON DELETE CASCADE** removes comments.
6. Return `204 No Content`.
7. Verify: subsequent `GET /posts/{id}` → `404`; `GET /posts/{id}/comments` → `404` for missing post.

Next.js does **not** call this endpoint in v1.

### 4.2 Comments

| Method | Path | Auth | Description | Success | Errors |
|--------|------|------|-------------|---------|--------|
| `GET` | `/posts/{postId}/comments` | Public | List comments for post | `200` + `Comment[]` | `404` post, `400` |
| `GET` | `/comments/{id}` | Public | Get one comment | `200` + `Comment` | `404`, `400` |
| `POST` | `/posts/{postId}/comments` | **Required** | Create comment | `201` + `Comment` | `401`, `404`, `400` |
| `DELETE` | `/comments/{id}` | **Required** | Delete own comment only | `204` empty | `401`, `404`, `403` |

**Comment create:** any **authenticated** user may `POST` on any existing post; persist **`author_user_id`** from JWT (see §5.1).

**Comment delete:** only the user who **created** that comment (`author_user_id == token sub`). Post owner **cannot** delete someone else’s comment unless they wrote it.

**Create comment body:**

```json
{
  "name": "required",
  "email": "required",
  "body": "required"
}
```

### 4.3 Health / meta

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | **`/api/v1/health`** | Liveness (e.g. `{ "status": "ok" }` or plain text) — **only** health URL in v1 |
| `GET` | `/` | Optional welcome string; not used for probes |

### 4.4 HTTP semantics aligned with Next

- **GET** missing resource → **`404`** (Next treats as `null` on GET).
- **POST** validation → **`400`** with `{ "error": "..." }`.
- **DELETE** missing → **`404`**; forbidden → **`403`**; unauthenticated → **`401`**.

### 4.5 Authentication & registration

**Approach (v1 recommendation):** Stateless **JWT** (or signed session token) issued at login; `Authorization: Bearer <token>` on protected routes. Passwords stored with **bcrypt** (or `argon2id` if you prefer one extra dependency).

| Method | Path | Auth | Body | Success |
|--------|------|------|------|---------|
| `POST` | `/auth/register` | Public | `email`, `password`, `displayName` | `201` + `User` (no secrets) |
| `POST` | `/auth/login` | Public | `email`, `password` | `200` + `{ "token": "...", "user": User }` |
| `POST` | `/auth/logout` | Required | Bearer token | **`204`** empty (client deletes token; no refresh/revocation list in v1) |
| `GET` | `/auth/me` | Required | — | `200` + `User` |

**Validation (register):**

- Unique `email` (case-insensitive compare recommended).
- Password minimum length (e.g. 8) — document in API errors.
- On duplicate email → `409 Conflict`.

**Login failures:** `401` with generic message (do not reveal whether email exists).

**JWT claims (minimum):** `sub` = user id, `exp`, `iat`. Sign with server secret from config/env (`JWT_SECRET` — never commit).

#### 4.5.1 Protected vs public routes (summary)

| Route | v1 access |
|-------|-----------|
| `GET` posts, comments | Public |
| `POST` comments | **Authenticated** (any user) |
| `POST` posts | Authenticated; `userId` from token only |
| `DELETE` posts | Authenticated + **author only** → **204** |
| `DELETE` comments | Authenticated + **comment author only** → **204** |
| `POST` register / login / logout | Register & login public; logout requires Bearer |

#### 4.5.4 Comment ownership (create / delete)

- **Create:** set `comments.author_user_id` from JWT; JSON response remains JSONPlaceholder shape (no extra field).
- **Delete:** load comment; if `author_user_id != authenticated user` → **403**; else delete → **204**.
- **Seeded comments:** assign each seed comment an `author_user_id` (1–5) for realistic delete tests.

#### 4.5.5 Client: JWT on every request when logged in

| Layer | Behavior |
|-------|----------|
| **Next `lib/api/http.ts`** (planned) | Read token from cookie/localStorage/context; if present, set `Authorization: Bearer <token>` on **every** `apiGet` / `apiPost` call. |
| **Go public GET** | No token required; optional Bearer may be sent and **ignored** in v1. |
| **Go protected routes** | Missing/invalid token → **401**. |

#### 4.5.2 Seeded users vs registered users

- Migration **`001_...`** (or combined schema migration) inserts **5 users** with known dev passwords (document only in this repo’s dev docs / README, not in production).
- `POST /auth/register` adds **additional** users at runtime (ids ≥ 6 unless you reserve 1–5 for seed-only — **recommendation:** allow register for any unique email; seed users 1–5 are normal rows).

#### 4.5.3 Post delete authorization

**v1 rule:** authenticated user must own the post (`posts.user_id == token user id`). Admin override is **out of scope**.

---

## 5. SQLite data model

### 5.1 Tables

**`users`**

| Column | Type | Notes |
|--------|------|--------|
| `id` | INTEGER | PRIMARY KEY AUTOINCREMENT |
| `email` | TEXT | NOT NULL UNIQUE |
| `password_hash` | TEXT | NOT NULL |
| `display_name` | TEXT | NOT NULL |
| `created_at` | TEXT | Optional ISO8601 |

**`posts`**

| Column | Type | Notes |
|--------|------|--------|
| `id` | INTEGER | PRIMARY KEY AUTOINCREMENT |
| `user_id` | INTEGER | NOT NULL, REFERENCES `users(id)` ON DELETE RESTRICT* |
| `title` | TEXT | NOT NULL |
| `body` | TEXT | NOT NULL |
| `created_at` | TEXT | Optional ISO8601 |

\* **RESTRICT** prevents deleting a user who still has posts (or use CASCADE if you want user delete to wipe posts — v1: **no user DELETE endpoint**; RESTRICT is safer).

**`comments`**

| Column | Type | Notes |
|--------|------|--------|
| `id` | INTEGER | PRIMARY KEY AUTOINCREMENT |
| `post_id` | INTEGER | NOT NULL, REFERENCES `posts(id)` ON DELETE CASCADE |
| `name` | TEXT | NOT NULL |
| `email` | TEXT | NOT NULL |
| `body` | TEXT | NOT NULL |
| `author_user_id` | INTEGER | NOT NULL, REFERENCES `users(id)` — **who wrote it** (DELETE auth); **not** in public JSON |
| `created_at` | TEXT | Optional |

**`schema_migrations`** (recommended)

| Column | Type | Notes |
|--------|------|--------|
| `version` | TEXT | PRIMARY KEY — applied migration id |
| `applied_at` | TEXT | ISO8601 |

Enable **foreign keys** on each connection: `PRAGMA foreign_keys = ON;`.

### 5.2 Migrations & seeding

Use **ordered migration files** (e.g. `migrations/0001_users.up.sql`, `0002_posts_comments.up.sql`, `0003_seed.up.sql`) or embedded SQL with version tracking.

**Migration content (required):**

1. Create `users`, `posts`, `comments`, `schema_migrations`.
2. **Seed 5 users** (ids 1–5):

   | id | email (example) | display_name |
   |----|-----------------|--------------|
   | 1 | `alice@example.com` | Alice |
   | 2 | `bob@example.com` | Bob |
   | 3 | `carol@example.com` | Carol |
   | 4 | `dave@example.com` | Dave |
   | 5 | `eve@example.com` | Eve |

   - Passwords: single shared **dummy dev password** → bcrypt in migration/seed code; **plain password documented only in `README.md`** (not in this plan).

3. **Seed posts** distributed across `user_id` 1–5 — **default count: 50** (see §5.2.1).
4. **Seed comments** for at least posts `1–5` (multiple per post; enough for comment detail static paths in Next).

**Idempotency:** migrations run once per version; seed migration should use `INSERT OR IGNORE` / check counts or run only when `schema_migrations` first applied — avoid duplicate users on restart.

#### 5.2.1 Seed post count: realism vs Next `generateStaticParams`

**What Next does at build time**

In `next_js_learning`, several `page.tsx` files export `generateStaticParams()` that call **`getPosts()`** (and sometimes **`getComments()`** per post). During `next build`, Next **pre-renders** one static HTML page per returned id (e.g. `/posts/1` … `/posts/N`, plus comment routes for configured posts).

| Factor | Many posts (e.g. **100**, JSONPlaceholder parity) | Fewer posts (e.g. **10–50**) |
|--------|---------------------------------------------------|--------------------------------|
| **Realism** | Feels like the real demo; paging/list UI looks full | Smaller list; still valid for learning |
| **Build time** | More API calls + more static pages = **slower `next build`** | Faster builds; less load on Go API during CI |
| **Go seed/migration** | More INSERTs; slightly larger SQLite file | Quick to seed by hand |
| **Static export (GHP)** | Larger `out/` (tracker noted 338 routes with JSONPlaceholder-scale data) | Smaller export |

**Locked:** seed exactly **50 posts** (ids **1–50**), spread across users **1–5**. (Raise to 100 only via explicit plan change.)

**Operational note:** While developing Go **locally**, run the API before `next build` so `generateStaticParams` can reach `API_BASE_URL`. If the API is down, build fails or returns fewer pages.

### 5.3 Config (existing + auth)

- `config/local.yml`: `storage_path: "storage/storage.db"`.
- Add when implementing: `jwt_secret`, `jwt_ttl`, `cors_allowed_origins` (YAML or env).
- Deploy: set `STORAGE_PATH`, `JWT_SECRET`, mount volume path (see §8).

---

## 6. Go application plan

### 6.1 Layering (decided)

**Call flow:**

```text
HTTP request
  → middleware (CORS, optional JWT) 
  → internal/handler/*     # decode/encode JSON, params, status codes, simple rules (e.g. post owner delete)
  → internal/store/*       # SQLite CRUD only
  → models
```

**No `internal/service` in v1.** Auth helpers live in `internal/auth` (hash, JWT, middleware); handlers call `auth` + `store` directly.

**Guidelines**

| Layer | Do | Avoid |
|-------|-----|--------|
| **Handler** | `json.Decode`, `writeJSON(w, status, ...)`, call store, map `sql.ErrNoRows` → 404 | Long SQL strings |
| **Store** | Parameterized queries, return `(T, error)` or domain errors | `http.ResponseWriter`, JWT parsing |
| **Auth** | `HashPassword`, `IssueToken`, `Middleware(requireAuth)` | Post/comment CRUD SQL |

If a handler exceeds ~40–50 lines or duplicates rules, extract a function in the same package or introduce `internal/service` later (document in changelog).

### 6.2 Suggested packages (when coding starts)

```text
cmd/post-api/main.go           # wire: config, db, migrations, mux, handlers, server
internal/config/               # CORS, JWT secret, storage path
internal/database/             # open SQLite, run migrations, pragmas
internal/models/               # User, Post, Comment + JSON tags (camelCase)
internal/auth/                 # hash, JWT, middleware (not a “service” layer)
internal/handler/              # posts, comments, auth HTTP — calls store
internal/store/                # SQL (CRUD); imported only by handlers (+ auth store if needed)
migrations/                    # versioned SQL + seed (5 users)
doc/
```

### 6.3 Router

- Prefer **Go 1.22+** `ServeMux` patterns: `GET /api/v1/posts`, `POST /api/v1/auth/login`, etc.
- **Single mount:** business routes under `/api/v1`.
- Auth middleware wraps only protected handlers (composition, not global block on GET).

### 6.4 Config startup (known issue)

Before feature work, fix `internal/config/config.go` `MustLoad` so local default path works without `CONFIG_PATH` (see `doc/setup/config-and-server-startup-issues.md`). Document run command in README when fixed.

**Progress (2026-11-21 — Step 1):** The API **starts and listens** on `localhost:8080` when run from the project root with:

```bash
CONFIG_PATH=config/local.yml go run cmd/post-api/main.go
```

`cleanenv` loads `config/local.yml` without requiring `CONFIG_PATH` (default path / `-config` flag). **`GET /api/v1/health`** verified (200 JSON). Graceful shutdown + `handler.RoutesRegister`. **M0 / P0 complete** — see root `README.md` and **MILESTONE_TRACKER.md** M0.

**Progress (2026-11-22):** `modernc.org/sqlite`, migrations, schema + seed (**10** posts, 5 users, comments). **Posts:** `GET /api/v1/posts`, `GET /api/v1/posts/{id}` (M5). **Comments:** `GET /api/v1/posts/{postId}/comments`, `GET /api/v1/comments/{commentId}` (M7 GETs, Postman). Route wiring: `RoutesRegisterHealthCheck` / `Posts` / `Comments`. Dev password in `README.md`; Postman login uses `password`.

### 6.5 Dependencies

- SQLite driver: **`modernc.org/sqlite`** — **in use** (see `go.mod`, `internal/database/open.go`).
- Passwords: `golang.org/x/crypto/bcrypt`.
- JWT: `github.com/golang-jwt/jwt/v5`.

### 6.6 What the SQLite driver is (decided: `modernc.org/sqlite`)

**SQLite** is the database engine (single `.db` file). Go does not ship SQLite built in. A **driver** is a package that implements `database/sql`’s driver interface so your code can run:

```go
import (
  "database/sql"
  _ "modernc.org/sqlite" // registers driver name "sqlite"
)
db, err := sql.Open("sqlite", "file:storage/storage.db")
```

| Option | Pros | Cons |
|--------|------|------|
| **`modernc.org/sqlite`** ✅ v1 | Pure Go; **no CGO**; `go build` works everywhere with default toolchain | Slightly different perf than CGO for huge workloads (irrelevant for this app) |
| **`github.com/mattn/go-sqlite3`** | Very common in tutorials; fast | Requires **CGO** + C toolchain on the machine; cross-compile harder |

**You still write:** standard `database/sql` + SQL strings in `internal/store`. The driver is only the connector under `sql.Open`.

**README:** dummy seed password for users 1–5 documented in root `README.md`; optional `sqlite3` inspect examples there.

---

## 7. Implementation phases (recommended order)

> **Day-to-day checklist:** use **[MILESTONE_TRACKER.md](./MILESTONE_TRACKER.md)** (smaller tasks, effort, done/partial status). Phases below map 1:1: P0≈M0–M1, P1≈M2–M3, P2≈M4, etc.

| Phase | Deliverable | Verification |
|-------|-------------|--------------|
| **P0** | Config load + listen on configured address | **Done** (M0, 2026-11-21): config, health, shutdown, README |
| **P1** | Migrations: schema + **5 seeded users** + posts/comments seed | **Done** (M2–M3, 2026-11-22): `sqlite3` → 5 users, 10 posts, FK intact |
| **P2** | Auth: register, login, `GET /me`, JWT middleware | **Deferred** (M4); interim creates use body `userId` per tracker |
| **P3** | `GET /posts`, `GET /posts/{id}` (public) | **Done** (M5, 2026-11-22): Postman/curl; JSON matches Next shape |
| **P4** | `POST /posts` (auth), **`DELETE /posts/{id}` (auth + author)** | curl: owner deletes; other user → `403`; cascade comments |
| **P5** | Comment GET (public), POST (**JWT**), DELETE (auth) → **204** | **GET + interim POST/DELETE done** (2026-11-22); **403/404** delete polish + JWT (M4) pending |
| **P6** | CORS + `Authorization`; **logout** `204` | Preflight; login → logout |
| **P7** | Next `.env` → local Go; reads work | Add-comment needs Next JWT work later |
| **P8** | **Deferred** — local complete first | Free host TBD (§8) |

**Stretch (post-v1):** post PATCH, **refresh tokens**, rate limits, Next login + comment JWT, OpenAPI, deploy, Postgres. **Comment PATCH** implemented locally (2026-11-22): `PATCH /api/v1/comments/{id}?userId=` — partial `name`/`body`; not used by Next v1.

---

## 8. Free / low-cost deployment options (Go + SQLite)

SQLite needs a **writable, persistent filesystem** (or accept **re-seed on every deploy** for demos only).

| Platform | Fit for Go API | SQLite persistence | Notes |
|----------|----------------|--------------------|--------|
| **[Fly.io](https://fly.io)** | Good (Dockerfile / fly.toml) | **Volumes** attachable — suitable for SQLite file | Free allowance; set `JWT_SECRET` as secret |
| **[Render](https://render.com)** | Good free web service | **Ephemeral** on free tier — DB file lost on redeploy unless paid disk | Re-run migrations + seed on boot |
| **[Railway](https://railway.app)** | Good | Ephemeral unless volume add-on | Trial/monthly credits |
| **[Koyeb](https://www.koyeb.com)** | Good | Ephemeral on free | Similar to Render |
| **[Google Cloud Run](https://cloud.google.com/run)** | Good (container) | **Ephemeral** | Generous free tier |
| **[Oracle Cloud Always Free](https://www.oracle.com/cloud/free/)** | VM + run binary | **Full disk** | Best SQLite persistence for $0; more ops |

### Recommendation (when you finish local v1)

1. **Learning + persistence:** Fly.io + volume, or Oracle free VM.
2. **Fastest deploy:** Render/Railway with **migrate + seed on boot** (data resets on redeploy).

**Until then:** decision **#5** = **local only**. **Deploy checklist (later):** `JWT_SECRET`, `STORAGE_PATH`, probe **`GET /api/v1/health`**, CORS.

---

## 9. Security and operations (v1 baseline)

- **HTTPS** in production; never send passwords or JWT over plain HTTP except localhost.
- **bcrypt** (cost factor ≥ 10) for `password_hash`; never log passwords or tokens.
- **JWT secret** from env; rotate procedure documented when deploying.
- **Post DELETE** requires auth — reduces anonymous abuse; still add rate limits on login/register when public on internet (stretch).
- Validate and bound string lengths (title, body, email, displayName).
- Parameterized SQL only.
- Do not commit `JWT_SECRET` or production `.env`.

---

## 10. Testing checklist (manual → automated later)

### Data & migrations

- [ ] Fresh DB: migrations apply; exactly **5 seeded users** exist.
- [ ] Seeded users can **login** and receive JWT.
- [ ] Posts reference valid `user_id` (1–5).

### Auth

- [ ] Register new user → `201`; duplicate email → `409`.
- [ ] Login wrong password → `401`.
- [ ] `GET /auth/me` without token → `401`.
- [ ] `POST /auth/logout` with token → `204`; without → `401`.

### Posts

- [ ] Public list/detail work without token.
- [ ] `POST /posts` without token → `401`; with token → `userId` matches token user.
- [ ] **Delete post:** owner + token → `204`; GET post → `404`; comments gone.
- [ ] **Delete post:** non-owner → `403`; missing id → `404`.
- [ ] Get post `404` for missing id (Next `notFound` path).

### Comments

- [ ] List/get comments without auth.
- [ ] Create comment **with** JWT (any user); without → `401`.
- [ ] Delete **own** comment → `204`; another user’s comment → `403`.

### Integration

- [ ] CORS: OPTIONS + POST comment from Next origin.
- [ ] Next: `/posts`, detail, comments, add comment — **no** delete post UI.

### Example: post delete (curl)

```bash
# Login as alice (seed user 1)
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password"}' | jq -r .token)

curl -i -X DELETE http://localhost:8080/api/v1/posts/1 \
  -H "Authorization: Bearer $TOKEN"
```

---

## 11. Open items

| Item | Owner | Notes |
|------|--------|--------|
| Extract `internal/service` | Deferred | Handler → store locked for v1 |
| Free host | Deferred | Local first — §8 |
| Next: login UI + token storage | Later | Attach JWT on all `http.ts` calls when logged in (#19) |
| Keep OpenAPI/Postman in sync | Ongoing | `doc/api/*` when handlers ship |

---

## 12. Related documents

| Document | Location |
|----------|----------|
| OpenAPI + Postman | `doc/api/openapi.yaml`, `doc/api/postman_collection.json`, `doc/api/README.md` |
| Config startup issues | `doc/setup/config-and-server-startup-issues.md` |
| Next learning tracker | `../next_js_learning/docs/LEARNING_TRACKER.md` |
| Next API layer | `../next_js_learning/lib/api/` |

---

## 13. Changelog

| Date | Change |
|------|--------|
| 2026-11-21 | Initial plan: decisions 1–7, API v1 spec, SQLite schema, phases, deployment |
| 2026-11-21 | **Users/auth in scope**; migration **seeds 5 users**; **post DELETE API flow** (auth + author); **Next delete-post UI out of scope**; JWT/CORS `Authorization`; updated schema, phases P1–P8, tests |
| 2026-11-21 | **§3.6 JSONPlaceholder compatibility** verified against official guide + live API; decision **#8** lock request/response for Next-integrated routes |
| 2026-11-21 | **Decision #6:** v1 **handler → store** flow; §6.1 layering guidelines; no `internal/service` initially |
| 2026-11-21 | Decisions **#9–#17:** local-first deploy; JWT comment create; **50** seed posts (§5.2.1); `/api/v1/health`; POST posts ignore body `userId`; **modernc** sqlite; DELETE **204**; README dev password; no refresh; **logout** in scope |
| 2026-11-21 | **#10 locked** 50 posts; **#18** comment delete by **author_user_id**; **#19** JWT on all client calls when logged in; **`modernc.org/sqlite`** confirmed |
| 2026-11-21 | **#20** `doc/api/openapi.yaml` + Postman collection (contract-first) |
| 2026-11-21 | **Step 1 / P0:** Server starts locally with `CONFIG_PATH=config/local.yml`; setup log updated (`doc/setup/config-and-server-startup-issues.md`) |
| 2026-11-21 | **`GET /api/v1/health`** verified (200 + JSON) |
| 2026-11-21 | **M0 / P0 complete** — [README.md](../README.md), `RoutesRegister`, milestone tracker updated |
| 2026-11-22 | **P1 complete** (M2–M3): migrations, split SQL, schema + seed; **P3 complete** (M5): public post GETs |
| 2026-11-22 | README dev password; Postman/plan auth examples use seed plaintext `password` (not `password123`) |
| 2026-11-22 | **M7 comment GETs** — nested + by-id routes; aligns with `next_js_learning/lib/api/comments.ts`; P5 GET slice done |
| 2026-11-22 | **M7 comment POST** — interim body **`userId`** → `author_user_id`; JSON omits author; P5 POST slice done |
| 2026-11-22 | **M7 comment DELETE** — interim query **`userId`**, **204** verified Postman; P5 DELETE slice (403/404 polish deferred) |
| 2026-11-22 | **Comment PATCH (stretch)** — partial `name`/`body`, author **403**, **200** full `Comment`; Postman |

---

*When implementation continues, update this doc: mark phases complete, record chosen host, env examples, and final comment-delete auth rule.*