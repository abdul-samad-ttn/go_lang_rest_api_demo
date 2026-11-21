# Setup Log: Config Loading and Server Startup Issues

Tracks the errors hit while setting up config loading and the HTTP server, the root cause of each, and the fix status.

- **Date:** 2026-11-21
- **Branch:** `basic_setup`
- **Files involved:**
  - `internal/config/config.go`
  - `cmd/post-api/main.go`
  - `config/local.yml`

## Step 1 verified — server starts locally (2026-11-21)

Config loading and HTTP listen on the address from `config/local.yml` work when `CONFIG_PATH` points at that file. From the repo root (`go_lang/rest_api`):

```bash
CONFIG_PATH=config/local.yml go run cmd/post-api/main.go
```

**Observed startup (success):**

```
&{dev storage/storage.db {localhost:8080}}
starting post-api
server started on port localhost:8080
```

The process stays running on **`http://localhost:8080`** until stopped (Ctrl+C). Values loaded: `env=dev`, `storage_path=storage/storage.db`, `http_server.address=localhost:8080`.

**Health (2026-11-21):** `GET /api/v1/health` verified — **200**, `Content-Type: application/json`, body `{"status": "ok"}` (`constants.GET` + `HealthCheckPath` in `cmd/post-api/main.go`).

**Still open for Step 1 / P0 polish:** issue **#6** (`log.Fatal` without `CONFIG_PATH`); optional `logger.Fatalf` formatting (**#8**).

## Summary

| # | Issue | Where | Status |
|---|-------|-------|--------|
| 1 | `go tidy` is not a Go command | terminal | Resolved (use `go mod tidy`) |
| 2 | Malformed struct tags (missing closing quote, comment inside the tag) | `internal/config/config.go` | Fixed |
| 3 | `storagePath` was unexported, so it could never be loaded | `internal/config/config.go` | Fixed (renamed to `StoragePath`) |
| 4 | `MustLoad` had no return value and never read the config | `internal/config/config.go` | Fixed (`cleanenv.ReadConfig` + `return &cfg`) |
| 5 | Missing config file returned `nil` instead of stopping | `internal/config/config.go` | Fixed (`log.Fatalf`) |
| 6 | `log.Fatal` in the default-path branch exits the program (`exit status 1`) | `internal/config/config.go` | **Fixed** (2026-11-21: `log.Printf`; default `-config` / `config/local.yml` without `CONFIG_PATH`) |
| 7 | `HTTPServer.Addr` has no `yaml:"address"` tag, so the address is empty and the server binds to `:80` (`permission denied`) | `internal/config/config.go` | **Fixed** (verified: binds `localhost:8080` with `CONFIG_PATH`) |
| 8 | `logger.Fatal("failed to start server", err)` prints without a separator | `cmd/post-api/main.go` | **Open** |
| 9 | "server started" message was printed after the blocking `ListenAndServe` | `cmd/post-api/main.go` | Fixed (moved before `ListenAndServe`) |

## Error 1: `go tidy: unknown command`

**Command:** `go tidy`

**Cause:** `tidy` is a subcommand of `go mod`, not of `go`.

**Fix:** run `go mod tidy`. It also keeps `github.com/ilyakaznacheev/cleanenv` correctly recorded in `go.mod` and `go.sum`.

## Error 2: `exit status 1` right after "using default path"

**Command:**

```bash
go run cmd/post-api/main.go -config config/local.yml
```

**Output:**

```
config path is not set, using default path: config/local.yml
exit status 1
```

**Cause:** in `MustLoad()` (`internal/config/config.go`), the `CONFIG_PATH` environment variable is empty, so the fallback branch runs:

```go
if configPath == "" {
    flags := flag.String("config", "config/local.yml", "path to config file")
    flag.Parse()
    configPath = *flags
    log.Fatal("config path is not set, using default path: config/local.yml")
}
```

- `flag.Parse()` reads `-config` correctly.
- `log.Fatal` then prints the message and calls `os.Exit(1)`.
- The program exits before the file check and `cleanenv.ReadConfig` run.
- Passing `-config` does not avoid this. The branch depends on `CONFIG_PATH`, not on the flag.

**Fix (pending):** change `log.Fatal` to `log.Printf` or remove the line. The message describes a fallback, not an error.

**Workaround (no code change):**

```bash
CONFIG_PATH=config/local.yml go run cmd/post-api/main.go
```

**Verified (2026-11-21):** the workaround above starts the server successfully; config prints `localhost:8080` and `ListenAndServe` runs without the previous `:80` bind error.

## Error 3: `listen tcp :80: bind: permission denied`

**Command:**

```bash
CONFIG_PATH=config/local.yml go run cmd/post-api/main.go
```

**Output:**

```
&{dev storage/storage.db {}}
starting post-api
server started on port
failed to start serverlisten tcp :80: bind: permission denied
exit status 1
```

**Cause:**

- `config/local.yml` uses `address: "localhost:8080"` under `http_server`.
- The struct field is `Addr string` with no tag, so the YAML key `address` is not mapped.
- The loaded config shows an empty `HTTPServer` (`{}`), and the "server started on port" line prints an empty address.
- With an empty `Addr`, `http.Server` listens on `:http`, which is port 80.
- Ports below 1024 need elevated privileges on Linux, so the bind fails with `permission denied`.

**Fix (pending):** add a tag to the field in `internal/config/config.go`:

```go
type HTTPServer struct {
    Addr string `yaml:"address" env:"ADDRESS" env-default:"localhost:8080"`
}
```

Do not run the server as root or use `sudo` to get around this.

**Resolution (2026-11-21):** `HTTPServer.Addr` now has `yaml:"address"`. With `CONFIG_PATH=config/local.yml`, startup loads `{localhost:8080}` and the server listens on port **8080** (see **Step 1 verified** at the top of this document).

## Error 4: Missing separator in the fatal log

**Output:** `failed to start serverlisten tcp :80: ...`

**Cause:** `logger.Fatal("failed to start server", err)` uses `Print`-style joining. No space is added when one operand is a string.

**Fix (pending):** in `cmd/post-api/main.go`, use `logger.Fatalf("failed to start server: %v", err)`.

## Changes already applied

`internal/config/config.go`:

- Imported `github.com/ilyakaznacheev/cleanenv`.
- Fixed the closing quote on the `env-required:"true"` tags.
- Renamed `storagePath` to `StoragePath` so it is exported.
- Replaced the missing return with `cleanenv.ReadConfig(configPath, &cfg)` and `return &cfg`.
- Changed the missing-file case from `return nil` to `log.Fatalf`.

`cmd/post-api/main.go`:

- Moved the "server started" log before `server.ListenAndServe()`.
- Switched it to `logger.Println`.

## Remaining to-do

- [ ] Replace `log.Fatal` in the default-path branch of `MustLoad` with `log.Printf` (or remove it).
- [x] Add `yaml:"address"` (and env/default) to `HTTPServer.Addr`.
- [ ] Use `logger.Fatalf("failed to start server: %v", err)` in `main.go`.
- [ ] Clean up the stray text in the `Config.Env` tag comment, and keep either `env-required` or `env-default` for `Env`, not both.
- [ ] Run `go mod tidy`, `go vet ./...`, and `go build ./...`.
- [x] Verify server starts with `CONFIG_PATH=config/local.yml go run cmd/post-api/main.go` (listens on `localhost:8080`).
- [ ] After fixing `log.Fatal` (#6), verify `go run cmd/post-api/main.go -config config/local.yml` (no env var).
- [ ] `curl http://localhost:8080/api/v1/health` → `{"status":"ok"}` (route must use pattern `GET /api/v1/health` in Go 1.22+).
- [ ] Add a test for config loading (valid file, missing file, missing required field).

## How to verify

**Current (Step 1 — works today):**

```bash
cd /path/to/go_lang/rest_api
export CONFIG_PATH=config/local.yml   # optional: set once per shell
go run cmd/post-api/main.go
# expect: config dump, "starting post-api", "server started on port localhost:8080"
```

**In another terminal (after health route uses `GET`):**

```bash
curl -s http://localhost:8080/api/v1/health
# expected: {"status": "ok"}
```

**Target after issue #6 is fixed (no `CONFIG_PATH` required):**

```bash
go mod tidy
go vet ./...
go run cmd/post-api/main.go -config config/local.yml
# or: go run cmd/post-api/main.go   # default config/local.yml
```
