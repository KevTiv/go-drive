# Go-Drive Microservices Architecture

This document describes the microservices architecture for the Go-Drive application.

## Architecture Overview

The application has been refactored into a microservices architecture with the following components:

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
                │  Supabase   │
                │ PostgreSQL  │
                └─────────────┘
```

## Services

### 1. API Gateway (HTTP/REST)
- **Purpose**: External-facing REST API
- **Port**: 8080
- **Protocol**: HTTP/REST
- **Responsibilities**:
  - Accept HTTP requests from clients
  - Route requests to appropriate gRPC services
  - Handle CORS
  - Response transformation

### 2. User Service (gRPC)
- **Purpose**: User management
- **Port**: 50051
- **Protocol**: gRPC (protobuf)
- **Responsibilities**:
  - User CRUD operations
  - Email verification
  - User listing with pagination

### 3. File Service (gRPC) - Future
- **Purpose**: File storage and management
- **Port**: 50052
- **Protocol**: gRPC (protobuf)
- **Responsibilities**:
  - File upload/download
  - File metadata management
  - Storage integration (S3/Supabase Storage)

## Technology Stack

- **Language**: Go 1.25.4
- **Inter-service Communication**: gRPC + Protocol Buffers
- **Database**: Supabase (PostgreSQL)
- **Container Orchestration**: Kubernetes
- **Container Runtime**: Docker

## Project Structure

```
go-drive/
├── proto/                          # Protocol buffer definitions
│   ├── user/                       # User service proto
│   │   └── user.proto
│   ├── file/                       # File service proto
│   │   └── file.proto
│   └── Makefile                    # Proto generation
├── services/                       # Microservices
│   ├── api-gateway/                # HTTP REST API Gateway
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── user-service/               # gRPC User Service
│   │   ├── main.go
│   │   ├── repository/             # Data access layer
│   │   │   └── postgres.go
│   │   ├── service/                # Business logic
│   │   │   └── user_service.go
│   │   └── Dockerfile
│   └── file-service/               # gRPC File Service (future)
├── k8s/                            # Kubernetes manifests
│   ├── base/                       # Base configurations
│   │   ├── namespace.yaml
│   │   ├── configmap.yaml
│   │   ├── secret.yaml
│   │   ├── user-service-deployment.yaml
│   │   ├── api-gateway-deployment.yaml
│   │   ├── ingress.yaml
│   │   └── kustomization.yaml
│   └── overlays/                   # Environment-specific configs
│       ├── dev/
│       └── prod/
├── scripts/                        # Database scripts
│   └── init.sql                    # Database initialization
└── docker-compose.microservices.yml
```

## Setup Instructions

### Prerequisites

1. **Install Protocol Buffer Compiler**:
   ```bash
   # macOS
   brew install protobuf

   # Linux
   apt-get install -y protobuf-compiler
   ```

2. **Install Go protobuf plugins**:
   ```bash
   cd proto
   make install-tools
   ```

3. **Generate protobuf code**:
   ```bash
   cd proto
   make all
   ```

### Development with Docker Compose

1. **Update environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your Supabase credentials
   ```

2. **Start services**:
   ```bash
   docker-compose -f docker-compose.microservices.yml up --build
   ```

3. **Test the API**:
   ```bash
   # Health check
   curl http://localhost:8080/health

   # Create user
   curl -X POST http://localhost:8080/api/v1/users \
     -H "Content-Type: application/json" \
     -d '{
       "first_name": "John",
       "surname": "Doe",
       "email": "john@example.com",
       "country": "USA"
     }'

   # List users
   curl http://localhost:8080/api/v1/users
   ```

### Kubernetes Deployment

1. **Update Kubernetes secrets**:
   ```bash
   # Edit k8s/base/secret.yaml with your Supabase credentials
   vim k8s/base/secret.yaml
   ```

2. **Apply Kubernetes manifests**:
   ```bash
   # Using kubectl
   kubectl apply -k k8s/base

   # Or using kustomize
   kustomize build k8s/base | kubectl apply -f -
   ```

3. **Check deployment status**:
   ```bash
   kubectl get pods -n go-drive
   kubectl get services -n go-drive
   ```

4. **Access the API**:
   ```bash
   # Port forward for local access
   kubectl port-forward -n go-drive service/api-gateway 8080:80

   # Or use the LoadBalancer IP
   kubectl get service api-gateway -n go-drive
   ```

## Supabase Configuration

### Database Setup

1. Create a Supabase project at https://supabase.com
2. Get your connection details from Project Settings > Database
3. Update the secrets in `k8s/base/secret.yaml`:
   ```yaml
   DB_HOST: "db.your-project.supabase.co"
   DB_PORT: "5432"
   DB_NAME: "postgres"
   DB_USER: "postgres"
   DB_PASSWORD: "your-database-password"
   ```

### Run Database Migrations

Execute the initialization script in your Supabase SQL Editor:
```bash
# Copy contents of scripts/init.sql and run in Supabase SQL Editor
```

Or use the Supabase CLI:
```bash
supabase db push
```

## API Endpoints

### User Service (via API Gateway)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/users` | Create a new user |
| GET | `/api/v1/users?id={id}` | Get user by ID |
| GET | `/api/v1/users` | List all users (paginated) |
| PUT | `/api/v1/users` | Update user |
| DELETE | `/api/v1/users?id={id}` | Delete user (soft delete) |
| GET | `/health` | Health check |

### Request Examples

**Create User**:
```json
POST /api/v1/users
{
  "first_name": "John",
  "surname": "Doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "country": "USA",
  "region": "California",
  "city": "San Francisco",
  "type": "premium"
}
```

**Update User**:
```json
PUT /api/v1/users
{
  "id": "user-uuid-here",
  "first_name": "Jane",
  "email": "jane@example.com"
}
```

## Monitoring and Debugging

### View Logs

```bash
# API Gateway logs
kubectl logs -f -n go-drive -l app=api-gateway

# User Service logs
kubectl logs -f -n go-drive -l app=user-service

# Docker Compose logs
docker-compose -f docker-compose.microservices.yml logs -f
```

### gRPC Debugging with grpcurl

```bash
# Install grpcurl
brew install grpcurl

# List services
grpcurl -plaintext localhost:50051 list

# Call CreateUser
grpcurl -plaintext -d '{
  "first_name": "Test",
  "surname": "User",
  "email": "test@example.com"
}' localhost:50051 user.UserService/CreateUser
```

## Scaling

### Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: user-service-hpa
  namespace: go-drive
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: user-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Security Considerations

1. **Use Kubernetes Secrets** for sensitive data
2. **Enable TLS** for gRPC communication in production
3. **Implement authentication** (JWT) in API Gateway
4. **Use Supabase Row Level Security (RLS)** for additional data protection
5. **Network Policies** to restrict inter-service communication

## Next Steps

1. Implement authentication service
2. Add file service for document storage
3. Implement event-driven architecture with message queue
4. Add distributed tracing (OpenTelemetry)
5. Implement circuit breakers and retries
6. Add comprehensive monitoring (Prometheus + Grafana)
