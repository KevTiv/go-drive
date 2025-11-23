# Go-Drive - Cloud Storage Microservices

A production-ready microservices application for cloud storage and file management, built with Go, gRPC, Kubernetes, and Supabase.

## 🏗️ Architecture

This project follows a microservices architecture with:
- **gRPC** for efficient inter-service communication
- **Protocol Buffers** for type-safe contracts
- **Kubernetes** for orchestration and scaling
- **PostgreSQL 16** with Row Level Security (RLS) for data persistence
- **Docker** for containerization
- **Shared internal packages** for common business logic

```
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway (HTTP/REST)                   │
│                         Port: 8080                           │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ gRPC
                       │
           ┌───────────┴───────────┐
           │                       │
    ┌──────▼──────┐        ┌──────▼──────┐
    │ User Service│        │ File Service│
    │  (gRPC)     │        │  (gRPC)     │
    │ Port: 50051 │        │ Port: 50052 │
    └──────┬──────┘        └──────┬──────┘
           │                       │
           └───────────┬───────────┘
                       │
                ┌──────▼──────┐
                │ PostgreSQL  │
                │  16 + RLS   │
                └─────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.25.4+
- Docker & Docker Compose
- Protocol Buffer Compiler (`protoc`)
- Kubernetes cluster (Minikube, k3s, or cloud)
- PostgreSQL 16 (or use provided Docker setup)

### Installation

1. **Clone and setup:**
   ```bash
   git clone <your-repo>
   cd go-drive
   ```

2. **Generate protobuf code:**
   ```bash
   # Ensure Go bin is in PATH
   export PATH=$PATH:$(go env GOPATH)/bin

   # Install protobuf tools
   cd proto
   make install-tools
   make all
   cd ..
   ```

3. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials (defaults work for local development)
   ```

4. **Start with Docker Compose:**
   ```bash
   # Start PostgreSQL 16 and all services
   docker-compose up --build

   # Or start just the database
   docker-compose up -d postgres
   ```

5. **Test the API:**
   ```bash
   # Health check
   curl http://localhost:8080/health

   # Create a user
   curl -X POST http://localhost:8080/api/v1/users \
     -H "Content-Type: application/json" \
     -d '{
       "first_name": "John",
       "surname": "Doe",
       "email": "john@example.com",
       "country": "USA"
     }'
   ```

## 📁 Project Structure

```
go-drive/
├── proto/                          # Protocol buffer definitions
│   ├── user/user.proto            # User service contract
│   ├── file/file.proto            # File service contract
│   └── Makefile                    # Proto generation
├── services/                       # Microservices
│   ├── user-service/              # gRPC User Service
│   │   ├── main.go
│   │   ├── repository/            # Data access layer
│   │   ├── service/               # Business logic
│   │   └── Dockerfile
│   ├── api-gateway/               # HTTP REST API Gateway
│   │   ├── main.go
│   │   └── Dockerfile
│   └── file-service/              # File Service (future)
├── k8s/                           # Kubernetes manifests
│   └── base/                      # Base configurations
│       ├── namespace.yaml
│       ├── configmap.yaml
│       ├── secret.yaml
│       ├── user-service-deployment.yaml
│       ├── api-gateway-deployment.yaml
│       └── ingress.yaml
├── scripts/                       # Database scripts
│   └── init.sql                   # Schema initialization
├── docker-compose.microservices.yml
├── Makefile.microservices         # Build automation
├── .env.example                   # Environment template
├── README.md                      # This file
├── README.microservices.md        # Detailed architecture docs
├── DEPLOYMENT.md                  # Deployment guide
└── TESTING.md                     # Testing guide
```

## 🎯 Services

### API Gateway (Port 8080)
External-facing REST API that routes requests to internal gRPC services.

**Endpoints:**
- `POST /api/v1/users` - Create user
- `GET /api/v1/users?id={id}` - Get user by ID
- `GET /api/v1/users` - List users
- `PUT /api/v1/users` - Update user
- `DELETE /api/v1/users?id={id}` - Delete user
- `GET /health` - Health check

### User Service (Port 50051)
gRPC service for user management with full CRUD operations.

**Methods:**
- `CreateUser` - Register new user
- `GetUser` - Retrieve user details
- `UpdateUser` - Update user information
- `DeleteUser` - Soft delete user
- `ListUsers` - Paginated user listing
- `VerifyEmail` - Email verification

### File Service (Port 50052) - Future
Planned service for file storage and management.

## 🧪 Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
go test -v ./services/...

# Run with coverage
go test -cover ./services/user-service/...

# Generate coverage report
go test -coverprofile=coverage.out ./services/user-service/...
go tool cover -html=coverage.out
```

**Test Coverage:**
- Repository Layer: 80.6%
- Service Layer: 100%
- Total: 42+ automated tests

See [TESTING.md](TESTING.md) for detailed testing guide.

## 🐳 Docker Deployment

### Local Development

```bash
# Start all services
docker-compose -f docker-compose.microservices.yml up -d

# View logs
docker-compose -f docker-compose.microservices.yml logs -f

# Stop services
docker-compose -f docker-compose.microservices.yml down
```

### Build Images

```bash
# Using Makefile
make -f Makefile.microservices docker-build

# Or manually
docker build -t go-drive-user-service -f services/user-service/Dockerfile .
docker build -t go-drive-api-gateway -f services/api-gateway/Dockerfile .
```

## ☸️ Kubernetes Deployment

### Quick Deploy

```bash
# Update secrets with your Supabase credentials
vim k8s/base/secret.yaml

# Deploy to Kubernetes
kubectl apply -k k8s/base

# Check status
kubectl get pods -n go-drive

# Access API Gateway
kubectl port-forward -n go-drive service/api-gateway 8080:80
```

### Production Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for comprehensive deployment guide including:
- Supabase configuration
- Building and pushing images
- Kubernetes setup
- Monitoring and scaling
- Troubleshooting

## 🗄️ Database

The application uses Supabase (PostgreSQL) for data persistence.

### Setup Database

1. Create a Supabase project at https://supabase.com
2. Get connection details from Project Settings > Database
3. Run the initialization script:
   ```sql
   -- Copy and execute scripts/init.sql in Supabase SQL Editor
   ```

### Schema

- **users** - User profiles and authentication
- **files** - File metadata (future)
- **folders** - Folder hierarchy (future)

## 🛠️ Development

### Using the Makefile

```bash
# Generate protobuf code
make -f Makefile.microservices proto

# Build services locally
make -f Makefile.microservices build

# Run tests
make -f Makefile.microservices test

# Quick start
make -f Makefile.microservices quick-start

# Clean build artifacts
make -f Makefile.microservices clean
```

### Adding New Services

1. Define protobuf contract in `proto/`
2. Generate code: `cd proto && make`
3. Create service directory in `services/`
4. Implement repository and service layers
5. Add Dockerfile
6. Create Kubernetes manifests
7. Write tests

## 📚 Documentation

- [README.microservices.md](README.microservices.md) - Detailed architecture and API docs
- [DEPLOYMENT.md](DEPLOYMENT.md) - Complete deployment guide
- [TESTING.md](TESTING.md) - Testing strategy and guide
- [.env.example](.env.example) - Environment variables reference

## 🔧 Configuration

Key environment variables:

```env
# Database
DB_HOST=db.xxxxx.supabase.co
DB_PASSWORD=your-password

# Services
GRPC_PORT=50051
PORT=8080
USER_SERVICE_ADDR=user-service:50051

# CORS
CORS_ORIGIN=http://localhost:5173
```

## 🏃 Performance

- **gRPC** provides 7-10x better performance than REST for inter-service communication
- **Connection pooling** for database connections
- **Horizontal scaling** via Kubernetes
- **Caching** ready (Redis integration planned)

## 🔐 Security

- Supabase Row Level Security (RLS) enabled
- Secrets managed via Kubernetes Secrets
- TLS/SSL for production gRPC communication
- Input validation at service layer
- SQL injection protection via prepared statements

## 📊 Monitoring (Planned)

- Prometheus metrics
- Grafana dashboards
- Distributed tracing (Jaeger/Zipkin)
- Structured logging

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Ensure all tests pass
5. Submit a pull request

## 📝 License

[Your License]

## 🙏 Acknowledgments

- Built with Go and gRPC
- Database by Supabase
- Orchestrated with Kubernetes
- Inspired by microservices best practices

## 📞 Support

For issues or questions:
- Check [DEPLOYMENT.md](DEPLOYMENT.md) for troubleshooting
- Review [TESTING.md](TESTING.md) for test failures
- Check logs: `kubectl logs -n go-drive <pod-name>`

## 🗺️ Roadmap

- [x] User service with gRPC
- [x] API Gateway
- [x] Kubernetes deployment
- [x] Comprehensive testing
- [ ] File service implementation
- [ ] Authentication service
- [ ] Event-driven architecture
- [ ] Caching layer (Redis)
- [ ] Monitoring and observability
- [ ] CI/CD pipeline

---

**Legacy monolith code backed up to:** `.archive/legacy-monolith/`
