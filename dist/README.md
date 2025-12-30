# Lambra - Microservices Generator

Generate complete Golang microservices with a visual UI.

## Quick Start

### Prerequisites
- Docker & Docker Compose
- MySQL client (optional, for local DB access)

### 1. Setup

```bash
# Download distribution files
mkdir lambra && cd lambra
curl -O https://raw.githubusercontent.com/yourorg/lambra/main/dist/docker-compose.yml
curl -O https://raw.githubusercontent.com/yourorg/lambra/main/dist/.env.example

# Copy and customize environment
cp .env.example .env
```

### 2. Start Lambra

```bash
docker-compose up -d
```

### 3. Access

- **UI Dashboard**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **API Health**: http://localhost:8080/health

## Usage

1. **Create Project** - Define service name, namespace, and database config
2. **Add Entities** - Define your data models (fields, types, validation)
3. **Auto-Generate Endpoints** - CRUD endpoints generated automatically
4. **Deploy** - Click "Deploy" to run the service in Docker
5. **Test** - Use the built-in endpoint tester

## Database Configuration

When creating a project, configure the database for your generated service:

| Field | Value for Local MySQL |
|-------|----------------------|
| DB Host | `host.docker.internal` |
| DB Port | `3306` |
| DB User | Your MySQL user |
| DB Password | Your MySQL password |
| DB Name | Database name (will be created) |

> **Note**: `host.docker.internal` allows Docker containers to access MySQL on your host machine.

### MySQL on Host Machine Requirements

1. MySQL must listen on all interfaces:
   ```ini
   # /etc/mysql/mysql.conf.d/mysqld.cnf
   bind-address = 0.0.0.0
   ```

2. User must have remote access:
   ```sql
   CREATE USER 'youruser'@'%' IDENTIFIED BY 'yourpassword';
   GRANT ALL PRIVILEGES ON *.* TO 'youruser'@'%';
   FLUSH PRIVILEGES;
   ```

## Port Configuration

| Service | Default Port | Environment Variable |
|---------|-------------|---------------------|
| Frontend UI | 3000 | `FRONTEND_PORT` |
| Backend API | 8080 | `BACKEND_PORT` |
| MySQL | 3306 | `MYSQL_PORT` |
| Generated Services | 9000+ | Auto-assigned |

## Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Reset (remove data)
docker-compose down -v
```

## Troubleshooting

### Generated service can't connect to database
- Ensure MySQL is listening on `0.0.0.0`
- Ensure MySQL user has `%` host permission
- Check firewall settings

### Port already in use
- Change ports in `.env` file
- Or stop conflicting services

### Permission denied on Docker socket
```bash
sudo chmod 666 /var/run/docker.sock
# Or add user to docker group
sudo usermod -aG docker $USER
```

## Support

- Issues: https://github.com/yourorg/lambra/issues
- Documentation: https://lambra.dev/docs
