#!/bin/bash

# Production Deployment Script
# File: scripts/deploy.sh

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DEPLOYMENT_TYPE=${1:-docker-compose}
ENVIRONMENT=${2:-production}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

show_help() {
    cat << EOF
Barbershop Production Deployment Script

USAGE:
    ./scripts/deploy.sh [DEPLOYMENT_TYPE] [ENVIRONMENT]

DEPLOYMENT TYPES:
    docker-compose    Deploy using Docker Compose (default)
    kubernetes        Deploy to Kubernetes cluster
    aws-ecs          Deploy to AWS ECS
    gcp-run          Deploy to Google Cloud Run
    azure-container   Deploy to Azure Container Instances
    heroku           Deploy to Heroku

ENVIRONMENTS:
    production       Production environment (default)
    staging          Staging environment
    development      Development environment

EXAMPLES:
    ./scripts/deploy.sh docker-compose production
    ./scripts/deploy.sh kubernetes staging
    ./scripts/deploy.sh aws-ecs production

PREREQUISITES:
    - Docker and Docker Compose installed
    - Environment variables configured
    - SSL certificates ready (for production)
    - Domain DNS configured

EOF
}

# Validate prerequisites
validate_prerequisites() {
    log_info "Validating prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    # Check deployment type specific requirements
    case $DEPLOYMENT_TYPE in
        docker-compose)
            if ! command -v docker-compose &> /dev/null; then
                log_error "Docker Compose is not installed"
                exit 1
            fi
            ;;
        kubernetes)
            if ! command -v kubectl &> /dev/null; then
                log_error "kubectl is not installed"
                exit 1
            fi
            ;;
        aws-ecs)
            if ! command -v aws &> /dev/null; then
                log_error "AWS CLI is not installed"
                exit 1
            fi
            ;;
    esac
    
    # Check environment file
    ENV_FILE="$PROJECT_ROOT/.env.$ENVIRONMENT"
    if [[ ! -f "$ENV_FILE" ]]; then
        log_error "Environment file not found: $ENV_FILE"
        exit 1
    fi
    
    log_success "Prerequisites validated"
}

# Load environment variables
load_environment() {
    log_info "Loading environment configuration..."
    
    ENV_FILE="$PROJECT_ROOT/.env.$ENVIRONMENT"
    set -a
    source "$ENV_FILE"
    set +a
    
    # Validate required variables
    REQUIRED_VARS=(
        "DATABASE_URL"
        "JWT_SECRET"
        "CORS_ALLOWED_ORIGINS"
    )
    
    for var in "${REQUIRED_VARS[@]}"; do
        if [[ -z "${!var}" ]]; then
            log_error "Required environment variable $var is not set"
            exit 1
        fi
    done
    
    log_success "Environment loaded: $ENVIRONMENT"
}

# Build application
build_application() {
    log_info "Building application..."
    
    cd "$PROJECT_ROOT"
    
    # Build Docker image
    IMAGE_TAG="${DOCKER_REGISTRY:-barbershop}/barbershop-api:${VERSION:-latest}"
    
    docker build \
        --tag "$IMAGE_TAG" \
        --build-arg VERSION="${VERSION:-latest}" \
        --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
        --build-arg VCS_REF="${GIT_COMMIT:-$(git rev-parse HEAD)}" \
        .
    
    # Push to registry if specified
    if [[ -n "$DOCKER_REGISTRY" ]]; then
        log_info "Pushing image to registry..."
        docker push "$IMAGE_TAG"
    fi
    
    log_success "Application built: $IMAGE_TAG"
}

# Deploy with Docker Compose
deploy_docker_compose() {
    log_info "Deploying with Docker Compose..."
    
    cd "$PROJECT_ROOT"
    
    # Create necessary directories
    mkdir -p backups logs ssl
    
    # Set up SSL certificates
    setup_ssl_certificates
    
    # Deploy services
    docker-compose -f docker-compose.prod.yml down
    docker-compose -f docker-compose.prod.yml up -d
    
    # Wait for services to be healthy
    wait_for_services
    
    # Run database migrations
    run_migrations
    
    log_success "Docker Compose deployment completed"
}

# Deploy to Kubernetes
deploy_kubernetes() {
    log_info "Deploying to Kubernetes..."
    
    cd "$PROJECT_ROOT"
    
    # Apply Kubernetes configurations
    kubectl apply -f k8s/namespace.yaml
    kubectl apply -f k8s/configmap.yaml
    kubectl apply -f k8s/secrets.yaml
    kubectl apply -f k8s/pvc.yaml
    kubectl apply -f k8s/postgres.yaml
    kubectl apply -f k8s/redis.yaml
    kubectl apply -f k8s/barbershop-api.yaml
    kubectl apply -f k8s/barbershop-service.yaml
    kubectl apply -f k8s/ingress.yaml
    kubectl apply -f k8s/hpa.yaml
    kubectl apply -f k8s/network-policy.yaml
    
    # Wait for deployment
    kubectl wait --for=condition=available --timeout=300s deployment/barbershop-api -n barbershop-prod
    
    # Run database migrations
    run_k8s_migrations
    
    log_success "Kubernetes deployment completed"
}

# Deploy to AWS ECS
deploy_aws_ecs() {
    log_info "Deploying to AWS ECS..."
    
    # Register task definition
    aws ecs register-task-definition \
        --family barbershop-api \
        --cli-input-json file://aws/task-definition.json
    
    # Update service
    aws ecs update-service \
        --cluster barbershop-cluster \
        --service barbershop-api \
        --task-definition barbershop-api
    
    # Wait for deployment
    aws ecs wait services-stable \
        --cluster barbershop-cluster \
        --services barbershop-api
    
    log_success "AWS ECS deployment completed"
}

# Deploy to Google Cloud Run
deploy_gcp_run() {
    log_info "Deploying to Google Cloud Run..."
    
    # Build and push to GCR
    IMAGE_URL="gcr.io/${GCP_PROJECT_ID}/barbershop-api:${VERSION:-latest}"
    
    docker tag barbershop-api:latest "$IMAGE_URL"
    docker push "$IMAGE_URL"
    
    # Deploy to Cloud Run
    gcloud run deploy barbershop-api \
        --image="$IMAGE_URL" \
        --platform=managed \
        --region="${GCP_REGION:-us-central1}" \
        --allow-unauthenticated \
        --memory=512Mi \
        --cpu=1 \
        --max-instances=10 \
        --set-env-vars="ENVIRONMENT=production" \
        --set-secrets="DATABASE_URL=database-url:latest,JWT_SECRET=jwt-secret:latest"
    
    log_success "Google Cloud Run deployment completed"
}

# Setup SSL certificates
setup_ssl_certificates() {
    log_info "Setting up SSL certificates..."
    
    SSL_DIR="$PROJECT_ROOT/ssl"
    
    if [[ "$ENVIRONMENT" == "production" ]]; then
        # Production: Use Let's Encrypt
        if [[ ! -f "$SSL_DIR/fullchain.pem" ]]; then
            log_info "Obtaining SSL certificate with Let's Encrypt..."
            
            docker run --rm \
                -v "$SSL_DIR:/etc/letsencrypt" \
                -v "$PROJECT_ROOT/www:/var/www/certbot" \
                certbot/certbot certonly \
                --webroot \
                --webroot-path=/var/www/certbot \
                --email "$SSL_EMAIL" \
                --agree-tos \
                --no-eff-email \
                -d "$DOMAIN_NAME"
        fi
    else
        # Development/Staging: Use self-signed certificates
        if [[ ! -f "$SSL_DIR/fullchain.pem" ]]; then
            log_info "Generating self-signed SSL certificate..."
            
            openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                -keyout "$SSL_DIR/privkey.pem" \
                -out "$SSL_DIR/fullchain.pem" \
                -subj "/C=US/ST=State/L=City/O=Organization/CN=$DOMAIN_NAME"
        fi
    fi
    
    log_success "SSL certificates ready"
}

# Wait for services to be healthy
wait_for_services() {
    log_info "Waiting for services to be healthy..."
    
    local max_attempts=30
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        if curl -f http://localhost:${API_PORT:-8080}/health &> /dev/null; then
            log_success "Services are healthy"
            return 0
        fi
        
        attempt=$((attempt + 1))
        log_info "Attempt $attempt/$max_attempts - waiting for services..."
        sleep 10
    done
    
    log_error "Services failed to become healthy"
    exit 1
}

# Run database migrations
run_migrations() {
    log_info "Running database migrations..."
    
    # Wait for database to be ready
    while ! docker-compose -f docker-compose.prod.yml exec -T postgres pg_isready -U "${DB_USER}" &> /dev/null; do
        log_info "Waiting for database..."
        sleep 5
    done
    
    # Run migrations
    docker-compose -f docker-compose.prod.yml exec -T barbershop-api ./scripts/migrate.sh
    
    log_success "Database migrations completed"
}

# Run migrations in Kubernetes
run_k8s_migrations() {
    log_info "Running database migrations in Kubernetes..."
    
    # Create migration job
    kubectl create job migrate-$(date +%s) \
        --from=deployment/barbershop-api \
        -n barbershop-prod \
        -- ./scripts/migrate.sh
    
    log_success "Migration job created"
}

# Health check after deployment
health_check() {
    log_info "Performing post-deployment health check..."
    
    local api_url
    case $DEPLOYMENT_TYPE in
        docker-compose)
            api_url="http://localhost:${API_PORT:-8080}"
            ;;
        kubernetes)
            api_url="https://${DOMAIN_NAME}"
            ;;
        *)
            api_url="$API_URL"
            ;;
    esac
    
    # Test health endpoint
    if curl -f "$api_url/health" &> /dev/null; then
        log_success "Health check passed"
    else
        log_error "Health check failed"
        exit 1
    fi
    
    # Test API endpoints
    local endpoints=(
        "/api/v1/barbers"
        "/api/v1/barbers/featured"
    )
    
    for endpoint in "${endpoints[@]}"; do
        if curl -f "$api_url$endpoint" &> /dev/null; then
            log_success "✓ $endpoint"
        else
            log_warning "✗ $endpoint (may require authentication)"
        fi
    done
}

# Rollback function
rollback() {
    log_info "Rolling back deployment..."
    
    case $DEPLOYMENT_TYPE in
        docker-compose)
            docker-compose -f docker-compose.prod.yml down
            # Restore from backup if needed
            ;;
        kubernetes)
            kubectl rollout undo deployment/barbershop-api -n barbershop-prod
            ;;
        aws-ecs)
            # Rollback to previous task definition
            aws ecs update-service \
                --cluster barbershop-cluster \
                --service barbershop-api \
                --task-definition barbershop-api:$((REVISION - 1))
            ;;
    esac
    
    log_success "Rollback completed"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    
    # Remove temporary files
    rm -f /tmp/barbershop-*
    
    # Clean up old Docker images
    docker image prune -f
    
    log_success "Cleanup completed"
}

# Main deployment function
main() {
    log_info "🚀 Starting Barbershop Production Deployment"
    log_info "============================================="
    log_info "Deployment Type: $DEPLOYMENT_TYPE"
    log_info "Environment: $ENVIRONMENT"
    log_info "============================================="
    
    # Handle help
    if [[ "$1" == "--help" || "$1" == "-h" ]]; then
        show_help
        exit 0
    fi
    
    # Validate and prepare
    validate_prerequisites
    load_environment
    
    # Build application
    build_application
    
    # Deploy based on type
    case $DEPLOYMENT_TYPE in
        docker-compose)
            deploy_docker_compose
            ;;
        kubernetes)
            deploy_kubernetes
            ;;
        aws-ecs)
            deploy_aws_ecs
            ;;
        gcp-run)
            deploy_gcp_run
            ;;
        azure-container)
            deploy_azure_container
            ;;
        heroku)
            deploy_heroku
            ;;
        *)
            log_error "Unknown deployment type: $DEPLOYMENT_TYPE"
            show_help
            exit 1
            ;;
    esac
    
    # Post-deployment tasks
    health_check
    cleanup
    
    log_success "🎉 Deployment completed successfully!"
    log_info "API URL: $API_URL"
    log_info "Health Check: $API_URL/health"
    log_info "API Docs: $API_URL/api/docs"
}

# Error handling
trap 'log_error "Deployment failed"; exit 1' ERR

# Run main function
main "$@"

---
# GitHub Actions CI/CD Pipeline
# File: .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]
    tags: ['v*']
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy to'
        required: true
        default: 'production'
        type: choice
        options:
          - production
          - staging

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: barbershop_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: |
        go mod download
        go test -v ./...
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/barbershop_test?sslmode=disable

  build:
    needs: test
    runs-on: ubuntu-latest
    outputs:
      image: ${{ steps.image.outputs.image }}
      digest: ${{ steps.build.outputs.digest }}
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3
    
    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=tag
          type=sha,prefix={{branch}}-
    
    - name: Build and push
      id: build
      uses: docker/build-push-action@v5
      with:
        context: .
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
    
    - name: Output image
      id: image
      run: |
        echo "image=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}" >> $GITHUB_OUTPUT

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: 
      name: ${{ github.event.inputs.environment || 'production' }}
      url: ${{ vars.APP_URL }}
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Deploy to production
      run: |
        chmod +x scripts/deploy.sh
        ./scripts/deploy.sh docker-compose production
      env:
        IMAGE: ${{ needs.build.outputs.image }}
        DATABASE_URL: ${{ secrets.DATABASE_URL }}
        JWT_SECRET: ${{ secrets.JWT_SECRET }}
        CORS_ALLOWED_ORIGINS: ${{ vars.CORS_ALLOWED_ORIGINS }}
        DOMAIN_NAME: ${{ vars.DOMAIN_NAME }}
        SSL_EMAIL: ${{ vars.SSL_EMAIL }}