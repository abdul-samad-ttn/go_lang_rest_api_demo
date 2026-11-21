# API documentation (contract-first)

Machine-readable specs for the **planned / in-progress** v1 API. Update these files when routes or shapes change (keep in sync with [PLAN_SQLITE_API_V1.md](../PLAN_SQLITE_API_V1.md)).

| File | Use |
|------|-----|
| [openapi.yaml](./openapi.yaml) | **Swagger / OpenAPI 3** — import into Swagger UI, Stoplight, Redocly, IDE plugins |
| [postman_collection.json](./postman_collection.json) | **Postman** or **Insomnia** (Import → Postman v2.1) |

## Swagger UI (local, no Go code required)

1. Open [Swagger Editor](https://editor.swagger.io/) → **File → Import file** → `openapi.yaml`.
2. Or run Swagger UI via Docker:

   ```bash
   docker run -p 8081:8080 \
     -e SWAGGER_JSON=/spec/openapi.yaml \
     -v "$(pwd)/doc/api/openapi.yaml:/spec/openapi.yaml" \
     swaggerapi/swagger-ui
   ```

   Then open http://localhost:8081 (spec only; calls hit your Go server when it is running on port 8080).

## Postman

1. **Import** `postman_collection.json`.
2. Collection variables:
   - `baseUrl` — default `http://localhost:8080/api/v1`
   - `token` — set automatically by **Auth → Login (seed alice)** test script
   - `postId`, `commentId` — adjust as needed
3. Run **Login** first, then protected requests (collection uses Bearer `{{token}}`).

## Optional: generate Postman from OpenAPI

In Postman: **Import → Link/File → openapi.yaml** to refresh requests when the YAML changes (you may still want the hand-tuned collection for login scripts).

## Serving docs from the Go app (later)

When the API is implemented, optional additions:

- Embed `openapi.yaml` and serve `GET /api/v1/openapi.yaml`
- Mount Swagger UI behind a dev-only flag

Not required for v1 local development.