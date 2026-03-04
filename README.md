# Lambra - Microservices Generator Platform

Lambra adalah platform untuk generate dan management microservices architecture. Dengan Lambra, Anda bisa membuat microservices baru hanya dengan mendefinisikan entities dan endpoints melalui UI yang user-friendly.

> **📚 Documentation:** See [DOCS_INDEX.md](DOCS_INDEX.md) for a complete guide to all documentation.

**Quick Links:**
- [Setup Guide](SETUP.md) - Langkah-langkah instalasi
- [Feature Documentation](FEATURE_VISUAL_RELATIONS.md) - Visual Relations feature
- [Testing Checklist](TESTING_CHECKLIST.md) - Panduan testing
- [Progress Tracker](PROGRESS.md) - Roadmap dan status development

## Features

### Core Features
- **Service Generator**: Generate microservices dari template dengan konfigurasi entities & endpoints
- **Visual Database Diagram**: Interactive diagram untuk memvisualisasikan entity dan relasi
- **Visual Relations**: Drag-and-drop relation creation (belongsTo, hasOne, hasMany, manyToMany)
- **Code Generation**: Auto-generate Golang models, repositories, services, handlers, dan migrations
- **Local Docker Deployment**: Semua service berjalan di local Docker dengan hot reload
- **Snapshot System**: Rollback ke versi sebelumnya dengan mudah
- **UI Dashboard**: Manage services, test endpoints, view metrics

### Relations Feature (🎉 NEW!)
- **Drag-and-Drop Interface**: Tarik koneksi dari entity ke entity
- **4 Tipe Relasi**: belongsTo, hasOne, hasMany, manyToMany
- **Smart Generation**: Auto-generate FK constraints dan junction tables
- **Visual Feedback**: Color-coded edges dengan tooltips
- **Full CRUD**: Create, edit, delete relasi melalui diagram
- **Interactive**: Hover untuk melihat detail, klik untuk edit/delete

## Tech Stack

### Backend
- Golang 1.21
- Gin Framework
- MySQL Database
- sqlx (SQL toolkit)
- Docker & Docker Compose

### Frontend
- React 18
- Vite
- Tailwind CSS
- React Query (data fetching)
- React Router (routing)
- Axios (HTTP client)

## Prerequisites

Sebelum menjalankan Lambra, pastikan Anda sudah install:

- **Docker** (version 20.10+)
- **Docker Compose** (version 2.0+)
- **Git**

## Quick Start

### 1. Clone Repository

```bash
git clone <repository-url>
cd lambra
```

### 2. Setup Environment Variables

**Backend:**
```bash
cd backend
cp .env.example .env
# Edit .env sesuai kebutuhan (opsional untuk development)
```

**Frontend:**
```bash
cd frontend
cp .env.example .env
# Default sudah sesuai untuk local development
```

### 3. Start Services

```bash
# Dari root directory lambra/
docker-compose up -d
```

Perintah ini akan:
- Start MySQL database
- Start backend API (dengan hot reload)
- Start frontend development server

### 4. Apply Database Migrations

```bash
# Apply migrations
docker-compose exec backend sh -c "mysql -hmysql -ulambra -plambra_secret lambra_db < /root/migrations/001_initial_schema.up.sql"
```

### 5. Access Application

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **API Health Check**: http://localhost:8080/health
- **MySQL**: localhost:3306

## Development

### Hot Reload

Kedua backend dan frontend sudah configured dengan hot reload:

**Backend** menggunakan Air:
- Setiap perubahan di `backend/` akan otomatis rebuild dan restart
- Konfigurasi di `backend/.air.toml`

**Frontend** menggunakan Vite:
- Setiap perubahan di `frontend/src/` akan langsung terlihat
- Hot Module Replacement (HMR) enabled

### View Logs

```bash
# All services
docker-compose logs -f

# Backend only
docker-compose logs -f backend

# Frontend only
docker-compose logs -f frontend

# MySQL only
docker-compose logs -f mysql
```

### Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (will delete database data)
docker-compose down -v
```

## Project Structure

```
lambra/
├── backend/                      # Backend Golang application
│   ├── cmd/server/              # Application entry point
│   ├── internal/                # Private application code
│   │   ├── api/                 # HTTP handlers, middleware, router
│   │   ├── config/              # Configuration management
│   │   ├── database/            # Database connection
│   │   ├── generator/           # Code generation engine
│   │   ├── models/              # Data models
│   │   ├── repository/          # Data access layer
│   │   ├── service/             # Business logic
│   │   └── utils/               # Shared utilities (strings, etc.)
│   ├── migrations/              # Database migrations
│   ├── templates/               # Code generation templates
│   │   └── docker/              # Docker templates for generated services
│   ├── pkg/                     # Public packages
│   ├── Dockerfile               # Production dockerfile
│   ├── Dockerfile.dev           # Development dockerfile with hot reload
│   └── .air.toml                # Air configuration for hot reload
│
├── frontend/                     # Frontend React application
│   ├── src/
│   │   ├── api/                 # API client & endpoints
│   │   ├── components/          # React components
│   │   │   ├── diagram/         # Visual diagram components
│   │   │   ├── forms/           # EntityForm, EndpointForm
│   │   │   ├── layout/          # Layout components
│   │   │   └── shared/          # Reusable components
│   │   ├── hooks/               # Custom React hooks
│   │   ├── pages/               # Page components
│   │   ├── stores/              # Zustand state stores
│   │   ├── lib/                 # Utilities & configurations
│   │   ├── App.jsx              # Main App component
│   │   └── main.jsx             # Entry point
│   ├── Dockerfile               # Production dockerfile
│   └── nginx.conf               # Nginx configuration for production
│
├── deploy/                       # Kubernetes deployment examples
│   ├── dev-deployment.yaml
│   ├── stag-deployment.yaml
│   └── prod-deployment.yaml
│
├── docker-compose.yml            # Development compose file
├── docker-compose.prod.yml       # Production compose file
└── README.md
```

## API Endpoints

### Health Checks
- `GET /health` - Health check
- `GET /ready` - Readiness check

### Projects (Services)
- `GET /api/v1/projects` - Get all projects
- `GET /api/v1/projects/:id` - Get project by ID
- `POST /api/v1/projects` - Create new project
- `PUT /api/v1/projects/:id` - Update project
- `DELETE /api/v1/projects/:id` - Delete project

### Entities & Relations
- `GET /api/v1/entities/:id` - Get entity detail
- `GET /api/v1/entities/:id/endpoints` - Get entity endpoints
- `GET /api/v1/entities/:id/relations` - Get entity relations
- `POST /api/v1/relations` - Create new relation (visual)
- `PUT /api/v1/relations/:id` - Update relation
- `DELETE /api/v1/relations/:id` - Delete relation

## Cara Membuat Relations (Visual)

### Metode Baru - Drag-and-Drop (Recommended) ✨

1. Buka **Service Detail** page
2. Switch ke **Diagram View**
3. **Drag dari satu entity ke entity lain**
4. RelationModal akan terbuka
5. Pilih tipe relasi:
   - **Belongs To** (Pink) - Source punya FK ke target
   - **Has One** (Purple) - Target punya FK ke source
   - **Has Many** (Blue) - Target entities punya FK ke source
   - **Many to Many** (Orange) - Junction table menghubungkan kedua entity
6. Configure FK field name, ON DELETE/UPDATE behavior
7. Klik **Create Relation**

Relasi akan muncul sebagai garis berwarna di diagram!

### Metode Lama - Field-based (Deprecated)

⚠️ Metode ini sudah tidak direkomendasikan. Gunakan **Diagram View** untuk membuat relasi.

Entity yang sudah ada dengan relasi field akan tetap ditampilkan sebagai garis putus-putus (dashed).

## Generated Services

Services yang di-generate oleh Lambra akan memiliki struktur yang sama dan siap dijalankan di local Docker.

### Template Variables

Setiap service yang di-generate akan menggunakan template dengan variables:
- `ServiceName`: Nama service
- `Port`: Port untuk service
- `DatabaseName`, `DatabaseUser`, `DatabasePassword`: Database credentials
- `Environment`: dev/staging/production
- `Endpoints`: List of endpoints dengan detail

### Running Generated Services

Service yang di-generate akan memiliki:
- `docker-compose.yml` - Configuration untuk local Docker
- `Dockerfile` - Docker image configuration
- `Makefile` - Commands untuk manage service
- `README.md` - Documentation

Untuk menjalankan generated service:

```bash
cd /path/to/generated/service
make up           # Start service
make logs         # View logs
make down         # Stop service
make migrate-up   # Apply migrations
make migrate-down # Rollback migrations
```

## Database

### MySQL Connection

**Host**: localhost
**Port**: 3306
**Database**: lambra_db
**User**: lambra
**Password**: lambra_secret

### Connecting with MySQL Client

```bash
# Using docker-compose
docker-compose exec mysql mysql -ulambra -plambra_secret lambra_db

# Using local MySQL client
mysql -h127.0.0.1 -P3306 -ulambra -plambra_secret lambra_db
```

### Migrations

**Apply migrations:**
```bash
docker-compose exec backend sh -c "mysql -hmysql -ulambra -plambra_secret lambra_db < /root/migrations/001_initial_schema.up.sql"
```

**Rollback migrations:**
```bash
docker-compose exec backend sh -c "mysql -hmysql -ulambra -plambra_secret lambra_db < /root/migrations/001_initial_schema.down.sql"
```

## Production Deployment

Untuk production, gunakan `docker-compose.prod.yml`:

```bash
# Set environment variables
export MYSQL_ROOT_PASSWORD=secure_root_password
export MYSQL_PASSWORD=secure_password

# Start production services
docker-compose -f docker-compose.prod.yml up -d
```

Perbedaan production vs development:
- Production menggunakan production Dockerfile (optimized)
- Frontend di-serve melalui Nginx
- GIN_MODE = release
- No hot reload
- Smaller image sizes

## Troubleshooting

### Backend tidak bisa connect ke database

Pastikan MySQL sudah ready:
```bash
docker-compose ps
docker-compose logs mysql
```

Tunggu sampai MySQL healthy, kemudian restart backend:
```bash
docker-compose restart backend
```

### Frontend tidak bisa call API

Cek VITE_API_BASE_URL di `frontend/.env`:
```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### Port sudah digunakan

Ubah port di `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"  # Change 8081 to available port
```

## ✅ Fitur yang Sudah Implementasi

### ✅ Core Features
- Project & Service Management (CRUD)
- Entity Management dengan custom fields
- Endpoint Management (auto-generated CRUD)
- Visual Database Diagram dengan ReactFlow
- **Visual Relations** - Drag-and-drop relation creation
- Code Generation untuk Golang microservices
- Docker-based deployment
- Snapshot & Rollback system

### 🎉 Visual Relations Feature (Terbaru!)
Fitur terbaru yang memungkinkan Anda membuat relasi antar entity secara visual:
- **Drag-and-drop interface** - Tarik dari satu entity ke entity lain
- **4 tipe relasi**: belongsTo, hasOne, hasMany, manyToMany
- **Visual feedback** - Warna berbeda untuk setiap tipe relasi
- **Interactive diagram** - Hover, edit, dan delete relasi
- **Smart generation** - Auto-generate FK constraints dan junction tables
- **Full CRUD** - Create, read, update, delete relasi
- **Snapshot support** - Relasi termasuk dalam snapshot/rollback

Lihat `FEATURE_VISUAL_RELATIONS.md` untuk detail implementasi.

## 🔨 Sedang Dalam Pengembangan

### Phase 2.3: GitLab Integration (Optional)
- GitLab API client untuk push code
- Workspace management untuk generated files
- Automated version control

### Phase 3+: Future Enhancements
- Endpoint testing interface (Swagger-like)
- Metrics & statistics dashboard
- OpenAPI specification export
- Ambassador integration untuk edge testing
- Deployment health monitoring

Lihat `PROGRESS.md` untuk roadmap lengkap dan status terbaru.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Your License Here]

---

**Lambra Platform** - Simplifying Microservices Development
