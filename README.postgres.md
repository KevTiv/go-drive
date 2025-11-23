# PostgreSQL 16 Setup Guide

This guide covers the PostgreSQL 16 setup for the go-drive microservices project, including Row Level Security (RLS) configuration.

## Architecture Overview

The project uses **PostgreSQL 16** with **Row Level Security (RLS)** to enforce data isolation between microservices:

- Each microservice has its own database role with specific permissions
- RLS policies control which data each service can access
- Shared database utilities in `internal/database` package
- Migrations managed via `internal/database/migration.go`

## Database Roles

### Service Roles (with RLS)

1. **user_service** - Full access to `users` table, read-only on `files` and `folders`
2. **file_service** - Full access to `files` and `folders`, read-only on `users`
3. **analytics_reader** - Read-only access to all tables (only non-deleted records)

### Admin Role

- **postgres** - Superuser for database administration and migrations

## Local Development Setup

### Using Docker Compose

1. **Start PostgreSQL 16:**
   ```bash
   docker-compose up -d postgres
   ```

2. **Verify database is running:**
   ```bash
   docker-compose ps
   docker-compose logs postgres
   ```

3. **Connect to database:**
   ```bash
   # Using psql
   docker-compose exec postgres psql -U postgres -d postgres

   # Or from host (if psql installed)
   psql -h localhost -U postgres -d postgres
   ```

4. **Start all services:**
   ```bash
   docker-compose up -d
   ```

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=postgres

# Service-specific Database Users
USER_SERVICE_DB_USER=user_service
USER_SERVICE_DB_PASSWORD=user_service_password
```

## Kubernetes Setup

### 1. Create Namespace

```bash
kubectl apply -f k8s/base/namespace.yaml
```

### 2. Update Secrets

Edit `k8s/base/secret.yaml` and change the default passwords:

```yaml
stringData:
  DB_PASSWORD: "your-strong-postgres-password"
  USER_SERVICE_DB_PASSWORD: "your-strong-user-service-password"
  FILE_SERVICE_DB_PASSWORD: "your-strong-file-service-password"
  ANALYTICS_DB_PASSWORD: "your-strong-analytics-password"
```

### 3. Deploy PostgreSQL

```bash
kubectl apply -k k8s/base
```

### 4. Verify PostgreSQL is Running

```bash
# Check pod status
kubectl get pods -n go-drive

# Check StatefulSet
kubectl get statefulset -n go-drive

# View logs
kubectl logs -n go-drive postgres-0

# Check PVC
kubectl get pvc -n go-drive
```

### 5. Connect to Database

```bash
# Port forward to access locally
kubectl port-forward -n go-drive postgres-0 5432:5432

# Connect using psql
psql -h localhost -U postgres -d postgres
```

## Database Schema

### Tables

1. **users** - User accounts with RLS
2. **files** - File metadata with RLS
3. **folders** - Folder hierarchy with RLS
4. **schema_migrations** - Migration tracking

### Row Level Security Policies

#### Users Table
- `user_service_all` - user_service can CRUD all users
- `file_service_read_users` - file_service can read users
- `analytics_read_users` - analytics_reader can read non-deleted users

#### Files Table
- `file_service_all` - file_service can CRUD all files
- `user_service_read_files` - user_service can read files
- `analytics_read_files` - analytics_reader can read non-deleted files

#### Folders Table
- `file_service_all` - file_service can CRUD all folders
- `user_service_read_folders` - user_service can read folders
- `analytics_read_folders` - analytics_reader can read non-deleted folders

## Migrations

### Running Migrations

Migrations are automatically applied on initial database setup via init scripts. For additional migrations:

1. **Create migration file:**
   ```bash
   # Create new migration in scripts/migrations/
   touch scripts/migrations/003_your_migration_name.sql
   ```

2. **Write migration SQL:**
   ```sql
   -- Migration: Your migration description
   -- Version: 003_your_migration_name

   -- Your migration SQL here

   -- Record migration
   INSERT INTO schema_migrations (version, description)
   VALUES ('003_your_migration_name', 'Your migration description')
   ON CONFLICT (version) DO NOTHING;
   ```

3. **Apply migration:**
   ```bash
   # In Kubernetes
   kubectl exec -n go-drive postgres-0 -- psql -U postgres -d postgres -f /path/to/migration.sql

   # In Docker Compose
   docker-compose exec postgres psql -U postgres -d postgres -f /path/to/migration.sql
   ```

### Migration Best Practices

- Always use transactions for DDL changes
- Test migrations in development first
- Use `IF EXISTS` / `IF NOT EXISTS` for idempotency
- Record all migrations in `schema_migrations` table
- Include rollback procedure in migration comments

## Testing RLS Policies

### Connect as Different Roles

```sql
-- Connect as user_service
SET ROLE user_service;

-- Test creating a user (should work)
INSERT INTO users (firstname, surname, email)
VALUES ('Test', 'User', 'test@example.com');

-- Test reading users (should work)
SELECT * FROM users;

-- Connect as file_service
SET ROLE file_service;

-- Test reading users (should work - read-only)
SELECT * FROM users;

-- Test creating user (should fail - no permission)
INSERT INTO users (firstname, surname, email)
VALUES ('Test', 'User', 'fail@example.com');

-- Reset to superuser
RESET ROLE;
```

## Shared Internal Packages

### internal/database

Connection management, pooling, and health checks:

```go
import "go-drive/internal/database"

cfg := database.Config{
    Host:     "postgres",
    Port:     "5432",
    User:     "user_service",
    Password: "user_service_password",
    DBName:   "postgres",
}

conn, err := database.NewConnection(cfg)
```

### internal/postgres

RLS policy management:

```go
import "go-drive/internal/postgres"

mgr := postgres.NewRLSManager(db)

// Enable RLS on a table
err := mgr.EnableRLS(ctx, "users")

// Create a policy
policy := postgres.PolicyConfig{
    Name:    "user_service_all",
    Table:   "users",
    Command: "ALL",
    Role:    "user_service",
    Using:   "true",
}
err = mgr.CreatePolicy(ctx, policy)
```

### internal/domain

Shared domain models:

```go
import "go-drive/internal/domain"

user := domain.User{
    FirstName: "John",
    Surname:   "Doe",
    Email:     "john@example.com",
    Type:      domain.UserTypeStandard,
}
```

## Backup and Restore

### Backup

```bash
# Kubernetes
kubectl exec -n go-drive postgres-0 -- pg_dump -U postgres postgres > backup.sql

# Docker Compose
docker-compose exec postgres pg_dump -U postgres postgres > backup.sql
```

### Restore

```bash
# Kubernetes
kubectl exec -i -n go-drive postgres-0 -- psql -U postgres postgres < backup.sql

# Docker Compose
docker-compose exec -T postgres psql -U postgres postgres < backup.sql
```

## Monitoring

### Check Connection Pool Stats

```sql
SELECT
    datname,
    numbackends,
    xact_commit,
    xact_rollback,
    blks_read,
    blks_hit
FROM pg_stat_database
WHERE datname = 'postgres';
```

### Check Active Connections

```sql
SELECT
    usename,
    application_name,
    client_addr,
    state,
    query
FROM pg_stat_activity
WHERE datname = 'postgres';
```

### Check Table Sizes

```sql
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## Troubleshooting

### Connection Refused

```bash
# Check if PostgreSQL is running
kubectl get pods -n go-drive
docker-compose ps

# Check PostgreSQL logs
kubectl logs -n go-drive postgres-0
docker-compose logs postgres

# Verify service is accessible
kubectl port-forward -n go-drive postgres-0 5432:5432
psql -h localhost -U postgres -d postgres
```

### Permission Denied Errors

```sql
-- Check current role
SELECT current_user, session_user;

-- Check table permissions
SELECT grantee, privilege_type
FROM information_schema.role_table_grants
WHERE table_name='users';

-- Check RLS policies
SELECT * FROM pg_policies WHERE tablename = 'users';
```

### Migration Issues

```sql
-- Check applied migrations
SELECT * FROM schema_migrations ORDER BY applied_at DESC;

-- Check for pending migrations
-- Compare with files in scripts/migrations/
```

## Production Considerations

1. **Use strong passwords** - Change default passwords in secrets
2. **Enable SSL/TLS** - Set `DB_SSLMODE=require`
3. **Use persistent volumes** - Ensure PVCs are backed by reliable storage
4. **Set up replication** - For high availability
5. **Configure backups** - Automated daily backups
6. **Monitor performance** - Use pg_stat_statements extension
7. **Tune connection pools** - Adjust based on load
8. **Set resource limits** - Memory and CPU limits for pods

## Additional Resources

- [PostgreSQL 16 Documentation](https://www.postgresql.org/docs/16/)
- [Row Level Security](https://www.postgresql.org/docs/16/ddl-rowsecurity.html)
- [Kubernetes StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
