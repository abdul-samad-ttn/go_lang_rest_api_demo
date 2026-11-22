# Agent instructions — `go_lang/rest_api`

**Read this before editing anything in this repository.**

## Owner preference: learn by doing

The owner implements application code themselves. Agents **coach, review, and document** — they do **not** take over implementation unless explicitly asked.

## Hard rule: no code edits without explicit permission

Do **not** create, modify, or delete **application or project source files** (e.g. `*.go`, `go.mod`, `config/*.yml`, migrations, handler/store code) unless the owner **explicitly** says in that chat to make the change (e.g. “apply this”, “edit the file”, “implement X for me”).

This applies in **every** agent chat and window for this project until the owner revokes or changes it.

**Allowed without a special ask:**

- Answer questions, explain concepts, review diffs or pasted code
- Propose snippets in the chat for the owner to copy
- Create or update **documentation only** (`doc/**`, this file, guides) when the owner asks for docs
- Create or update **Cursor rules** when the owner asks agents to remember a constraint

**When the owner asks for guidance only** (e.g. “tell me how”, “what should I do”):

- Do not use Write/StrReplace/edit tools on Go or config files
- Point to file paths, patterns, and verification steps

## Related docs

| Doc | Purpose |
|-----|---------|
| [doc/MILESTONE_TRACKER.md](doc/MILESTONE_TRACKER.md) | What to build, in order |
| [doc/PLAN_SQLITE_API_V1.md](doc/PLAN_SQLITE_API_V1.md) | API contract and decisions |
| [doc/guides/route-registration.md](doc/guides/route-registration.md) | How to wire `ServeMux` routes |

## Stack reminders

- Go 1.22+ `ServeMux` patterns: `GET /api/v1/...`
- Layering: `handler` → `store` (no `service` in v1)
- SQLite: `modernc.org/sqlite`, migrations under plan §5

---

*Last updated: 2026-11-21 — owner requested persistent “no code unless I say so” rule.*