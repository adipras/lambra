# Lambra - Setup Guide

Panduan lengkap untuk setup dan menjalankan Lambra platform di local environment.

## Prerequisites

### Required Software

1. **Docker Desktop** (version 20.10 atau lebih baru)
   - Download: https://www.docker.com/products/docker-desktop
   - Pastikan Docker daemon running

2. **Docker Compose** (version 2.0 atau lebih baru)
   - Biasanya sudah include di Docker Desktop
   - Verify: `docker-compose --version`

3. **Git**
   - Download: https://git-scm.com/downloads
   - Verify: `git --version`

### Optional (untuk development)

- **Go 1.21+** (jika ingin run backend tanpa Docker)
- **Node.js 18+** (jika ingin run frontend tanpa Docker)
- **MySQL Client** (untuk connect ke database langsung)
- **Make** (untuk menjalankan Makefile commands)

## Step-by-Step Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd lambra
```

### 2. Verify Docker Installation

```bash
# Check Docker
docker --version
# Output: Docker version 20.10.x, build xxxxx

# Check Docker Compose
docker-compose --version
# Output: Docker Compose version v2.x.x

# Check Docker is running
docker ps
# Should not show any error
```

### 3. Setup Environment Files

**Backend Environment:**

```bash
cd backend
cp .env.example .env
```

Edit `backend/.env` jika perlu (untuk development, default values sudah OK):

```env
PORT=8080
ENV=development
GIN_MODE=debug

DB_HOST=mysql
DB_PORT=3306
DB_USER=lambra
DB_PASSWORD=lambra_secret
DB_NAME=lambra_db

WORKSPACE_PATH=/workspace
```

**Frontend Environment:**

```bash
cd ../frontend
cp .env.example .env
```

Default value sudah sesuai untuk local development:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### 4. Start Services

Kembali ke root directory:

```bash
cd ..
```

**Option A: Using Makefile (Recommended)**

```bash
make up
```

**Option B: Using Docker Compose**

```bash
docker-compose up -d
```

Ini akan:
1. Pull MySQL 8.0 image
2. Build backend image (dengan Air for hot reload)
3. Start MySQL container
4. Start backend container (wait for MySQL healthy)
5. Start frontend container

**Check Status:**

```bash
docker-compose ps

# Expected output:
NAME                IMAGE               STATUS              PORTS
lambra-backend      ...                 Up                  0.0.0.0:8080->8080/tcp
lambra-frontend     ...                 Up                  0.0.0.0:5173->5173/tcp
lambra-mysql        mysql:8.0           Up (healthy)        0.0.0.0:3306->3306/tcp
```

### 5. Apply Database Migrations

```bash
# Wait a few seconds for MySQL to be ready, then:
make migrate-up

# Or manually:
docker-compose exec backend sh -c "mysql -hmysql -ulambra -plambra_secret lambra_db < /root/migrations/001_initial_schema.up.sql"
```

Jika sukses, akan muncul message: "Migrations applied successfully!"

### 6. Verify Services

**Health Check Backend:**

```bash
curl http://localhost:8080/health

# Expected response:
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "healthy",
    "database": "connected"
  }
}
```

**Access Frontend:**

Open browser: http://localhost:5173

Anda akan melihat Lambra Dashboard.

## Accessing Services

### Frontend Dashboard

- **URL**: http://localhost:5173
- **Pages**:
  - Dashboard: http://localhost:5173/
  - Services List: http://localhost:5173/services

### Backend API

- **Base URL**: http://localhost:8080
- **API Base**: http://localhost:8080/api/v1
- **Health Check**: http://localhost:8080/health
- **Readiness**: http://localhost:8080/ready

### MySQL Database

**Connection Info:**
- Host: `localhost`
- Port: `3306`
- Database: `lambra_db`
- User: `lambra`
- Password: `lambra_secret`

**Connect via Docker:**

```bash
make mysql-sh

# Or manually:
docker-compose exec mysql mysql -ulambra -plambra_secret lambra_db
```

**Connect via MySQL Client:**

```bash
mysql -h127.0.0.1 -P3306 -ulambra -plambra_secret lambra_db
```

## Development Workflow

### Hot Reload

**Backend (Air)**:
- Edit any `.go` file di `backend/`
- Air akan auto-detect changes dan rebuild
- Service auto-restart dengan binary baru

**Frontend (Vite)**:
- Edit any file di `frontend/src/`
- Vite HMR akan langsung update di browser
- No need to refresh

### View Logs

```bash
# All services
make logs

# Backend only
docker-compose logs -f backend

# Frontend only
docker-compose logs -f frontend

# MySQL only
docker-compose logs -f mysql
```

### Access Container Shell

```bash
# Backend container
make backend-sh
# atau: docker-compose exec backend sh

# Frontend container
make frontend-sh
# atau: docker-compose exec frontend sh

# MySQL shell
make mysql-sh
# atau: docker-compose exec mysql mysql -ulambra -plambra_secret lambra_db
```

### Install Go Dependencies (Backend)

Jika menambah dependency baru di `go.mod`:

```bash
# From host
cd backend
go mod download
go mod tidy

# Or from container
docker-compose exec backend sh
go mod download
go mod tidy
exit
```

Lalu restart backend:

```bash
docker-compose restart backend
```

### Install NPM Dependencies (Frontend)

Jika menambah dependency baru di `package.json`:

```bash
# From host (if you have Node.js)
cd frontend
npm install

# Or let container auto-install on next start
docker-compose restart frontend
```

## Testing

### Test API Endpoints

**Using cURL:**

```bash
# Health check
curl http://localhost:8080/health

# Get all projects
curl http://localhost:8080/api/v1/projects

# Create project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Service",
    "description": "My test service",
    "namespace": "test-ns"
  }'
```

**Using Frontend UI:**
- Open http://localhost:5173
- Navigate to Services
- Click "New Service"
- Fill the form and submit

### Run Go Tests

```bash
make test

# Or manually:
docker-compose exec backend go test ./... -v
```

## Troubleshooting

### Issue: Port sudah digunakan

**Error:**
```
Error starting userland proxy: listen tcp4 0.0.0.0:8080: bind: address already in use
```

**Solution:**

Cek process yang menggunakan port:

```bash
# macOS/Linux
lsof -i :8080
kill -9 <PID>

# Or change port in docker-compose.yml
ports:
  - "8081:8080"  # Change host port
```

### Issue: MySQL tidak ready

**Error:**
```
Error: Can't connect to MySQL server on 'mysql'
```

**Solution:**

Tunggu sampai MySQL container healthy:

```bash
docker-compose ps

# Wait until mysql shows "Up (healthy)"
# Then restart backend:
docker-compose restart backend
```

### Issue: Backend build error

**Error:**
```
go: module ... not found
```

**Solution:**

```bash
# Clean and rebuild
docker-compose down
docker-compose build --no-cache backend
docker-compose up -d
```

### Issue: Frontend tidak load

**Error:** Blank page atau VITE errors

**Solution:**

```bash
# Check logs
docker-compose logs frontend

# Restart frontend
docker-compose restart frontend

# If still error, rebuild
docker-compose down
docker-compose up -d frontend
```

### Issue: Database migration fails

**Solution:**

```bash
# Check if MySQL is ready
docker-compose exec mysql mysqladmin ping -h localhost

# Check if database exists
docker-compose exec mysql mysql -ulambra -plambra_secret -e "SHOW DATABASES;"

# Manually apply migration
docker-compose exec mysql mysql -ulambra -plambra_secret lambra_db < backend/migrations/001_initial_schema.up.sql
```

## Stopping & Cleaning

### Stop Services

```bash
make down
# Or: docker-compose down
```

### Stop and Remove All Data

```bash
make clean
# Or: docker-compose down -v

# Warning: This will delete all data in MySQL!
```

### Remove Docker Images

```bash
docker-compose down --rmi all
```

## Next Steps

Setelah setup berhasil:

1. **Explore the Dashboard** - Buka http://localhost:5173
2. **Create a Project** - Click "New Service" dan isi form
3. **Test API Endpoints** - Use curl atau Postman
4. **Read Phase 2 Docs** - Implement generator engine
5. **Customize Templates** - Edit templates di `backend/templates/`

## Getting Help

Jika ada masalah:
1. Check logs: `make logs`
2. Verify all containers running: `docker-compose ps`
3. Check this troubleshooting guide
4. Check GitHub Issues
5. Contact team

---

Setup complete! Happy coding with Lambra! 🚀
