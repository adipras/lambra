# Lambra System Analysis

> Evidence-based system analysis produced with REA (Reverse Engineering Assistant) tooling.
> Each finding cites its artifact SHA-256 and REA Evidence ID so claims are verifiable.
> Date of analysis: 14 November 2026. Analyzed build: `main` @ 63da8f8.

---

## 1. System Identity

Lambra is a self-contained microservice generator platform: a web UI defines entities,
relations, and endpoints; a Go backend generates complete CRUD microservice source code
(Gin + sqlx + MySQL), builds it as a Docker image, and runs it as an isolated container
with its own database.

| Component | Technology | Role |
|---|---|---|
| `lambra-frontend` | React 18 + Vite (node:18-alpine) | Visual modeling UI, login |
| `lambra-backend` | Go 1.21 (Gin, sqlx, JWT) | API, code generator, deploy orchestrator |
| `lambra-mysql` | MySQL 8.0 | Platform metadata + per-service databases |
| Generated services | Go 1.21 (Gin, sqlx), Docker | User-defined CRUD APIs |

## 2. Runtime Topology (observed)

Source: REA process capture of `docker compose ps`
(Evidence `ev_50a849fa7d755162d2a1930a0c0ccdef71687206ebfc594347608763e70fceed`,
controlled replay, 2026-11-14T02:44:48Z).

| Container | Image | Status observed | Host port | Internal port |
|---|---|---|---|---|
| `lambra-backend` | `lambra-backend` (custom, Air hot-reload) | Up | 8080 | 8080 |
| `lambra-frontend` | `node:18-alpine` | Up | 5173 | 5173 |
| `lambra-mysql` | `mysql:8.0` | Up (healthy) | 3307 | 3306 |

Notes:
- Generated services (e.g. `svc-test`) run **outside** the compose project, attached to
  the shared external network `lambra_lambra-network`, each with an allocator-assigned
  host port (observed range 9006–9284).
- Network bridge subnet: `172.20.0.0/16` (docker-compose.yml ipam).
- Process-tree sampling is partial (REA limitation); short-lived descendants like
  per-deploy `docker compose up` invocations are not retained in this capture.

## 3. Request Lifecycle (platform)

1. Browser authenticates: `POST /api/v1/auth/login` → JWT (HS256, 24 h, claims
   `uuid`/`username`/`role`) — open endpoint; everything else under `/api/v1` requires
   `Authorization: Bearer`.
2. Frontend attaches the token via axios interceptor; a 401 clears the session and
   redirects to `/login` (the login call itself is exempt to allow inline errors).
3. Service/project/entity/relation CRUD is stored in MySQL (`lambra_db`).
4. `POST /projects/:id/deploy` executes the generation pipeline:

   ```
   entities/relations (DB) → template engine (30+ helpers)
     → Go source tree (model/repository/service/handler/DTO/migration/main)
     → docker-compose.yml + multi-stage Dockerfile
     → docker compose up -d --build  (via mounted docker.sock)
     → health check → deployment + snapshot + per-step logs recorded
   ```

5. Every deploy writes a `generation_snapshot`; rollback
   (`POST /snapshots/:id/rollback`) restores entity definitions and redeploys.
   Attribution: `created_by` is taken from the JWT (via request context), falling back
   to `system` for unauthenticated internal calls.

## 4. Data Model (platform)

Tables observed in migrations 001–005:
`projects`, `git_repositories`, `entities` (JSON fields), `endpoints`, `relations`,
`generation_snapshots`, `deployments`, `deployment_logs` (with `step`),
`templates`, `users` (auth, seeded admin, bcrypt).

Dual identifier strategy throughout: internal `BIGINT id` (FK joins, never exposed by
APIs) and external `CHAR(36) uuid` (all API surfaces). Soft deletes via `deleted_at`.

## 5. API Surface (from source, router.go)

| Group | Endpoints |
|---|---|
| Auth (open) | `POST /auth/login` |
| Auth (protected) | `GET /auth/me` |
| Projects | CRUD + `/entities`, `/endpoints`, `/relations`, `/deploy`, `/start`, `/stop`, `/redeploy`, `/destroy`, `/status`, `/deployments`, `/container-logs` |
| Entities / Endpoints / Relations | CRUD (+ `/endpoints/:id/test`) |
| Generation | `POST /generate/entity`, `POST /generate/project`, `GET /generate/preview/:id`, `GET /generate/files/:id` |
| Snapshots | list/create/get/metadata/`POST /:id/rollback`/delete |
| Deployments | get, `/logs`, `/logs/stream` (SSE) |
| Export | `/export/openapi`, `/export/postman` |
| Health (open) | `/health`, `/ready` |

## 6. Security Posture (summary)

Implemented: JWT on all mutating/read APIs, bcrypt password storage, per-service DB
isolation, UUID-only external identifiers, DB password excluded from JSON responses
(`models.Project.DBPassword json:"-"` + custom marshaler), soft-delete audit fields.

Known gaps (documented for the submission): single seeded admin (RBAC on roadmap),
`lambra` MySQL user holds broad privileges in dev (`docker/mysql-init/01-grants.sql`),
dev JWT secret default (`lambra-dev-secret-change-me`) — production requires
`JWT_SECRET`.

## 7. Companion Reports

- [BINARY_ANALYSIS.md](BINARY_ANALYSIS.md) — REA build-metadata and ELF analysis of the
  platform backend binary and a generated service binary, including a dependency
  version-gap finding.
- [FRONTEND_ANALYSIS.md](FRONTEND_ANALYSIS.md) — REA JavaScript application-graph
  analysis of the React source tree (module structure, semantic-graph statistics,
  route candidates).

## 8. Evidence Register

| Evidence ID | Operation | Subject |
|---|---|---|
| `ev_4293e88d...cfbc` | inspect_go_binary | backend binary `7d054778...f9c28` |
| `ev_446c9503...1d02` | inspect_go_binary | svc-test binary `9922e578...f5445` |
| `ev_85edfa3a...b114` | inspect_elf_notes | backend binary |
| `ev_492820d8...e3bb` | inspect_elf_notes | svc-test binary |
| `ev_16d763c6...ab43` | analyze_javascript_application | frontend tree `fa4d9079...d4dd5` |
| `ev_50a849fa...fceed` | capture_process_scenario | `docker compose ps` topology |
