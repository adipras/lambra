# Lambra - Microservices Generator Platform

Lambra adalah platform untuk generate dan management microservices architecture. Dengan Lambra, Anda bisa membuat microservices baru hanya dengan mendefinisikan entities dan endpoints melalui UI yang user-friendly.

## Features

- **Service Generator**: Generate microservices dari template dengan konfigurasi entities & endpoints
- **Local Docker Deployment**: Semua service berjalan di local Docker
- **Git Integration**: Integrasi dengan GitLab untuk version control
- **Snapshot System**: Rollback ke versi sebelumnya dengan mudah
- **UI Dashboard**: Manage services, test endpoints, view metrics
- **OpenAPI Export**: Export API specification

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
│   │   ├── models/              # Data models
│   │   ├── repository/          # Data access layer
│   │   └── service/             # Business logic
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
│   │   │   ├── layout/          # Layout components
│   │   │   └── shared/          # Reusable components
│   │   ├── hooks/               # Custom React hooks
│   │   ├── pages/               # Page components
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

## Next Steps (Phase 2-4)

Berikut adalah fitur-fitur yang akan dikembangkan di phase selanjutnya:

### Phase 2: Core Generator Engine
- Template system dengan Handlebars
- Code generator service
- GitLab API integration
- Workspace management
- Generate endpoint implementation

### Phase 3: UI Dashboard Enhancement
- Service detail page
- Endpoint detail page
- Testing interface (like Swagger)
- Metrics & statistics
- Export OpenAPI specification

### Phase 4: Testing & Deployment
- Endpoint testing dengan Ambassador integration
- Snapshot system untuk rollback
- Deployment management
- Health monitoring
- Deployment logs viewer

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Your License Here]

---

**Lambra Platform** - Simplifying Microservices Development
