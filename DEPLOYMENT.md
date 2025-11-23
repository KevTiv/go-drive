# Go-Drive Microservices Deployment Guide

This guide provides step-by-step instructions for deploying the Go-Drive microservices application.

## Quick Start

### Option 1: Local Development with Docker Compose

```bash
# 1. Generate protobuf code
cd proto
make install-tools
make all
cd ..

# 2. Set up environment
cp .env.example .env
# Edit .env with your configuration

# 3. Start services
docker-compose -f docker-compose.microservices.yml up --build

# 4. Test the API
curl http://localhost:8080/health
```

### Option 2: Kubernetes Deployment

```bash
# 1. Configure Supabase credentials
vim k8s/base/secret.yaml

# 2. Build and push Docker images (adjust registry as needed)
docker build -t your-registry/user-service:latest -f services/user-service/Dockerfile .
docker build -t your-registry/api-gateway:latest -f services/api-gateway/Dockerfile .
docker push your-registry/user-service:latest
docker push your-registry/api-gateway:latest

# 3. Update image references in k8s manifests
vim k8s/base/user-service-deployment.yaml
vim k8s/base/api-gateway-deployment.yaml

# 4. Deploy to Kubernetes
kubectl apply -k k8s/base

# 5. Verify deployment
kubectl get pods -n go-drive
```

## Detailed Setup Guide

### Prerequisites

#### Required Tools

1. **Go 1.25.4 or later**
   ```bash
   go version
   ```

2. **Protocol Buffer Compiler**
   ```bash
   # macOS
   brew install protobuf

   # Ubuntu/Debian
   sudo apt-get install -y protobuf-compiler

   # Verify installation
   protoc --version
   ```

3. **Docker and Docker Compose**
   ```bash
   docker --version
   docker-compose --version
   ```

4. **Kubernetes cluster** (for production deployment)
   - Minikube (local)
   - GKE, EKS, or AKS (cloud)
   - k3s (lightweight)

5. **kubectl**
   ```bash
   kubectl version --client
   ```

6. **Supabase Account**
   - Sign up at https://supabase.com
   - Create a new project
   - Note down your database credentials

### Step 1: Supabase Setup

#### 1.1 Create Supabase Project

1. Go to https://supabase.com/dashboard
2. Click "New Project"
3. Fill in project details:
   - Name: go-drive
   - Database Password: (choose a strong password)
   - Region: (choose closest to your users)

#### 1.2 Get Database Connection Details

1. Navigate to Project Settings > Database
2. Note down:
   - Host: `db.xxxxx.supabase.co`
   - Port: `5432`
   - Database name: `postgres`
   - User: `postgres`
   - Password: (your chosen password)

#### 1.3 Initialize Database Schema

1. Open Supabase SQL Editor
2. Copy contents of `scripts/init.sql`
3. Execute the SQL script
4. Verify tables were created:
   ```sql
   SELECT tablename FROM pg_tables WHERE schemaname = 'public';
   ```

### Step 2: Generate Protobuf Code

```bash
# Install Go protobuf tools
cd proto
make install-tools

# Generate Go code from .proto files
make all

# Verify generated files
ls -la user/*.pb.go
ls -la file/*.pb.go
```

### Step 3: Local Development Setup

#### 3.1 Environment Configuration

Create `.env` file:
```bash
# Database (Supabase)
DB_HOST=db.xxxxx.supabase.co
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=your-password-here

# Supabase
SUPABASE_URL=https://xxxxx.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-key

# Service Configuration
GRPC_PORT=50051
PORT=8080
CORS_ORIGIN=http://localhost:5173
USER_SERVICE_ADDR=user-service:50051
```

#### 3.2 Build and Test Locally

```bash
# Download dependencies
go mod download

# Generate protobuf (if not done)
cd proto && make all && cd ..

# Build services
go build -o bin/user-service ./services/user-service
go build -o bin/api-gateway ./services/api-gateway

# Run user service
./bin/user-service &

# Run API gateway
./bin/api-gateway &

# Test
curl http://localhost:8080/health
```

### Step 4: Docker Compose Deployment

#### 4.1 Build Images

```bash
docker-compose -f docker-compose.microservices.yml build
```

#### 4.2 Start Services

```bash
# Start all services
docker-compose -f docker-compose.microservices.yml up -d

# View logs
docker-compose -f docker-compose.microservices.yml logs -f

# Check status
docker-compose -f docker-compose.microservices.yml ps
```

#### 4.3 Test Deployment

```bash
# Health check
curl http://localhost:8080/health

# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "surname": "Doe",
    "email": "john.doe@example.com",
    "country": "USA"
  }'

# List users
curl http://localhost:8080/api/v1/users
```

### Step 5: Kubernetes Deployment

#### 5.1 Prepare Kubernetes Cluster

**For Minikube (local testing)**:
```bash
# Start Minikube
minikube start --cpus=4 --memory=8192

# Enable ingress
minikube addons enable ingress

# Use Minikube's Docker daemon
eval $(minikube docker-env)
```

**For Cloud (GKE, EKS, AKS)**:
```bash
# Connect to your cluster
# GKE example:
gcloud container clusters get-credentials go-drive-cluster --region us-central1

# Verify connection
kubectl cluster-info
```

#### 5.2 Build and Push Docker Images

```bash
# Set your registry (Docker Hub, GCR, ECR, etc.)
export REGISTRY=your-dockerhub-username
# or
export REGISTRY=gcr.io/your-project-id

# Build images
docker build -t ${REGISTRY}/go-drive-user-service:latest -f services/user-service/Dockerfile .
docker build -t ${REGISTRY}/go-drive-api-gateway:latest -f services/api-gateway/Dockerfile .

# Push images
docker push ${REGISTRY}/go-drive-user-service:latest
docker push ${REGISTRY}/go-drive-api-gateway:latest
```

#### 5.3 Update Kubernetes Manifests

Update image references in deployment files:

**k8s/base/user-service-deployment.yaml**:
```yaml
spec:
  template:
    spec:
      containers:
      - name: user-service
        image: your-registry/go-drive-user-service:latest
```

**k8s/base/api-gateway-deployment.yaml**:
```yaml
spec:
  template:
    spec:
      containers:
      - name: api-gateway
        image: your-registry/go-drive-api-gateway:latest
```

#### 5.4 Configure Secrets

Edit `k8s/base/secret.yaml` with your Supabase credentials:
```yaml
stringData:
  DB_HOST: "db.xxxxx.supabase.co"
  DB_PORT: "5432"
  DB_NAME: "postgres"
  DB_USER: "postgres"
  DB_PASSWORD: "your-actual-password"
  SUPABASE_URL: "https://xxxxx.supabase.co"
  SUPABASE_ANON_KEY: "your-anon-key"
  SUPABASE_SERVICE_KEY: "your-service-key"
```

**Security Note**: For production, use external secret management:
```bash
# Using kubectl create secret
kubectl create secret generic go-drive-secrets \
  --from-literal=DB_PASSWORD='your-password' \
  --from-literal=DB_HOST='db.xxxxx.supabase.co' \
  -n go-drive
```

#### 5.5 Deploy to Kubernetes

```bash
# Apply all manifests
kubectl apply -k k8s/base

# Or apply individually
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/base/configmap.yaml
kubectl apply -f k8s/base/secret.yaml
kubectl apply -f k8s/base/user-service-deployment.yaml
kubectl apply -f k8s/base/api-gateway-deployment.yaml
kubectl apply -f k8s/base/ingress.yaml
```

#### 5.6 Verify Deployment

```bash
# Check namespace
kubectl get namespaces

# Check pods
kubectl get pods -n go-drive

# Check services
kubectl get services -n go-drive

# Check deployments
kubectl get deployments -n go-drive

# View pod logs
kubectl logs -f -n go-drive -l app=user-service
kubectl logs -f -n go-drive -l app=api-gateway
```

#### 5.7 Access the Application

**Option 1: Port Forward (Local Testing)**:
```bash
kubectl port-forward -n go-drive service/api-gateway 8080:80

# Test
curl http://localhost:8080/health
```

**Option 2: LoadBalancer (Cloud)**:
```bash
# Get external IP
kubectl get service api-gateway -n go-drive

# Use the EXTERNAL-IP
curl http://<EXTERNAL-IP>/health
```

**Option 3: Ingress (with domain)**:
```bash
# Get ingress IP
kubectl get ingress -n go-drive

# Add to /etc/hosts
echo "<INGRESS-IP> go-drive.local" | sudo tee -a /etc/hosts

# Access
curl http://go-drive.local/health
```

### Step 6: Monitoring and Maintenance

#### 6.1 View Logs

```bash
# Real-time logs
kubectl logs -f -n go-drive -l app=user-service
kubectl logs -f -n go-drive -l app=api-gateway

# Logs from all pods
kubectl logs -n go-drive --all-containers=true --selector app=user-service
```

#### 6.2 Scale Services

```bash
# Scale user service
kubectl scale deployment user-service --replicas=3 -n go-drive

# Scale API gateway
kubectl scale deployment api-gateway --replicas=3 -n go-drive

# Verify
kubectl get pods -n go-drive
```

#### 6.3 Update Deployment

```bash
# Update image
kubectl set image deployment/user-service \
  user-service=your-registry/go-drive-user-service:v2.0 \
  -n go-drive

# Check rollout status
kubectl rollout status deployment/user-service -n go-drive

# Rollback if needed
kubectl rollout undo deployment/user-service -n go-drive
```

## Troubleshooting

### Common Issues

#### 1. Pods Not Starting

```bash
# Describe pod
kubectl describe pod <pod-name> -n go-drive

# Check events
kubectl get events -n go-drive --sort-by='.lastTimestamp'
```

#### 2. Database Connection Issues

```bash
# Test from pod
kubectl exec -it <pod-name> -n go-drive -- /bin/sh
nc -zv db.xxxxx.supabase.co 5432
```

#### 3. gRPC Connection Issues

```bash
# Check if user-service is accessible
kubectl exec -it <api-gateway-pod> -n go-drive -- /bin/sh
nc -zv user-service 50051
```

#### 4. Image Pull Errors

```bash
# For private registries, create imagePullSecret
kubectl create secret docker-registry regcred \
  --docker-server=<your-registry> \
  --docker-username=<username> \
  --docker-password=<password> \
  -n go-drive

# Add to deployment spec
spec:
  imagePullSecrets:
  - name: regcred
```

## Production Considerations

### 1. Security

- Use TLS for gRPC communication
- Implement network policies
- Enable Pod Security Policies
- Use external secret management (HashiCorp Vault, AWS Secrets Manager)
- Enable Supabase Row Level Security (RLS)

### 2. High Availability

- Run multiple replicas (minimum 2 per service)
- Use Pod Disruption Budgets
- Distribute across availability zones
- Implement health checks and readiness probes

### 3. Monitoring

- Set up Prometheus for metrics
- Configure Grafana dashboards
- Implement distributed tracing (Jaeger, Zipkin)
- Set up log aggregation (ELK, Loki)

### 4. Performance

- Configure resource requests and limits
- Enable horizontal pod autoscaling
- Use connection pooling for database
- Implement caching (Redis)

### 5. Backup and Disaster Recovery

- Regular database backups (Supabase handles this)
- Test restore procedures
- Document runbooks
- Set up alerting

## Next Steps

1. Implement authentication (JWT)
2. Add file service for document storage
3. Set up CI/CD pipeline
4. Implement monitoring and alerting
5. Add rate limiting
6. Set up automated testing in CI
7. Document API with OpenAPI/Swagger

## Support

For issues or questions:
- Check logs: `kubectl logs -n go-drive <pod-name>`
- Review Kubernetes events: `kubectl get events -n go-drive`
- Consult README.microservices.md for architecture details
