# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Lambra is a microservices generator platform. Users define entities and endpoints via a web UI, and the platform generates complete Golang microservice code, pushes to GitLab, and deploys via Docker.

**Architecture**: Backend (Golang/Gin) + Frontend (React/Vite) + MySQL + Docker

## Development Commands

All commands run from the repository root:

```bash
# Start all services (MySQL, backend, frontend)
make up

# Stop services
make down

# View logs (all services or specific)
make logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Apply/rollback database migrations
make migrate-up
make migrate-down

# Run backend tests
make test

# Open container shells
make backend-sh
make frontend-sh
make mysql-sh

# Production
make prod-up
make prod-build
```

## Backend (Golang)

**Entry point**: `backend/cmd/server/main.go`
**Framework**: Gin
**Database**: MySQL with sqlx (not ORM)
**Go version**: 1.21

### Architecture Pattern

```
internal/
├── api/
│   ├── handlers/    # HTTP handlers (controllers)
│   ├── middleware/  # CORS, logging
│   └── router/      # Route definitions
├── config/          # Environment config
├── database/        # DB connection
├── generator/       # Code generation engine
├── models/          # Data models
├── repository/      # Database access layer
├── service/         # Business logic
└── utils/           # Shared utilities (strings, etc.)
```

### Dual Identifier Strategy

All entities use both internal BIGINT IDs and external UUIDs:
- `ID int64` (db:"id") - Internal, used for FK relationships, not exposed in API
- `UUID string` (db:"uuid" json:"id") - External, exposed to clients as "id"

IDs are derived from UUID v7 (first 8 bytes → int64). See `backend/internal/models/base.go` for `BaseEntity`.

### Key Patterns

**Repository methods**: Use `GetByUUID` for API lookups, `GetByID` for internal FK joins
**Soft deletes**: All deletions set `deleted_at` timestamp
**Response format**: Use `pkg/response/response.go` for consistent JSON responses

### Code Generator

The `internal/generator/` package generates Go code for microservices:
- `template_engine.go`: Go text/template with 30+ helper functions (case conversion, pluralization, type mapping)
- `code_generator.go`: Generates model, repository, service, handler, DTO, and migration files
- `templates.go`: Template strings for all code layers

## Frontend (React)

**Entry point**: `frontend/src/main.jsx`
**Framework**: React 18 with Vite
**Styling**: Tailwind CSS
**State**: React Query for server state
**HTTP**: Axios

### Structure

```
src/
├── api/         # Axios clients (projects.js, entities.js, endpoints.js, relations.js)
├── components/
│   ├── diagram/ # DatabaseDiagram, RelationModal, RelationEdge, EntityNode
│   ├── forms/   # EntityForm, EndpointForm
│   ├── layout/  # Layout, Sidebar
│   └── shared/  # StatusBadge, LoadingSpinner, ErrorAlert
├── hooks/       # Custom hooks (useProjects)
├── pages/       # Route components
└── lib/         # queryClient configuration
```

### Creating Relations

**Visual Relation Creation (New Method - Preferred):**
- Switch to **Diagram View** in ServiceDetail page
- Drag from one entity to another
- RelationModal opens automatically
- Configure relation type, ON DELETE/UPDATE behaviors
- Relations stored in `relations` table

**Field-based Relations (Legacy - Deprecated):**
- Old entities may have "relation" type fields
- Displayed as dashed lines in diagram
- Cannot be edited via diagram (view-only)
- Recommend migrating to new visual relations

**Relation Types:**
- **belongsTo** (pink): Source has FK to target (e.g., Post → User)
- **hasOne** (purple): Target has FK to source (e.g., User → Profile)
- **hasMany** (blue): Multiple targets have FK to source (e.g., User → Posts)
- **manyToMany** (orange): Junction table connects both (e.g., Posts ↔ Tags)

### API Base URL

Configured in `frontend/.env`:
```
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## Database

**Connection**: localhost:3306, database `lambra_db`, user `lambra`, password `lambra_secret`

**Migrations**: SQL files in `backend/migrations/` (001_initial_schema, 002_add_uuid_and_base_entity)

**Key tables**: projects, entities (with JSON fields), endpoints, relations, git_repositories, generation_snapshots, deployments

## API Routes

```
GET/POST   /api/v1/projects
GET/PUT/DEL /api/v1/projects/:id
POST       /api/v1/projects/:id/entities
GET        /api/v1/projects/:id/entities
GET/PUT/DEL /api/v1/entities/:id
GET        /api/v1/entities/:id/endpoints
GET        /api/v1/entities/:id/relations
POST       /api/v1/endpoints
GET/PUT/DEL /api/v1/endpoints/:id
POST       /api/v1/relations
GET/PUT/DEL /api/v1/relations/:id
POST       /api/v1/generate/entity
POST       /api/v1/generate/project
GET        /api/v1/generate/preview/:id
GET        /api/v1/generate/files/:id
```

## Docker

**Development**: `docker-compose.yml` with hot reload (Air for backend, Vite HMR for frontend)
**Production**: `docker-compose.prod.yml` with optimized builds

Services communicate on the `lambra-network` bridge network.

## Current Development Status

Phase 2.2 (Entity & Endpoint Management) is complete. Phase 2.3 (GitLab Integration) is next:
- Implement GitLab API client in `internal/generator/git_client.go`
- Add workspace manager for file generation
- Connect generate endpoint to push code to GitLab

### Recent Updates (March 2026)
- ✅ **Bug Fixes**: Fixed Relations feature bugs
  - Created `internal/utils/strings.go` with shared utility functions
  - Fixed `toSnakeCase()` and `pluralize()` function access
  - Fixed frontend API response handling in DatabaseDiagram.jsx
- ✅ **Documentation Cleanup**: Updated and organized all project documentation
  - Updated README.md with Visual Relations feature
  - Created DOCS_INDEX.md for documentation navigation
  - Updated PROGRESS.md with quick summary

See `PROGRESS.md` for detailed roadmap and completion status.
