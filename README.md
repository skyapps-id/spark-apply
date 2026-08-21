# Spark Apply

HTTP server for deploying docker-compose services with image tag updates. Part of the SkyApps ID CI/CD pipeline ecosystem.

## Features

- Deploy service via HTTP POST
- Update image tag in docker-compose.yaml
- Service name and tag validation for security
- Environment variables support for configuration

## CI/CD Pipeline Integration

This project is part of the SkyApps ID CI/CD pipeline ecosystem and works together with [github-runner](https://github.com/skyapps-id/github-runner):

1. **github-runner**: Self-hosted GitHub Actions runner with BuildKit for building container images
2. **spark-apply**: HTTP server that handles service deployment with image tag updates

**Workflow:**
- Developer pushes code to GitHub
- github-runner builds container image using BuildKit
- Image is pushed to container registry
- spark-apply updates image tag in docker-compose.yaml and redeploys service

## Environment Variables

- `PORT`: Server port (default: 8080)
- `BASE_FOLDER`: Path to services folder (default: /Users/ajii/Documents/spark-apply/services)

## API Endpoints

### POST /deploy

Deploy service with image tag update.

**Request Body:**
```json
{
  "service_name": "cctv-platform-be",
  "tag": "5f69a27c5e437b278e7d0a2da6a9b91dfadd935e",
  "compose_filename": "docker-compose.yml"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Service deployed",
  "output": "..."
}
```

## Network Setup

Before deploying services, create the required Docker network manually:

```bash
docker network create pipeline-net
```

This network should be created once and will be used by all services.

## Complete CI/CD Example

Here's how to use spark-apply together with github-runner in a complete CI/CD pipeline:

**1. Setup github-runner:**
```bash
git clone https://github.com/skyapps-id/github-runner.git
cd github-runner
cp .env.example .env
# Edit .env with your runner token
docker network create pipeline-net
docker compose up -d
```

**2. Configure GitHub Actions workflow:**
```yaml
name: Build and Deploy

on:
  push:
    branches: [ 'main' ]

jobs:
  build:
    runs-on: [self-hosted, linux, buildctl]
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Build and push image
        run: |
          buildctl build \
            --frontend dockerfile.v0 \
            --local context=. \
            --local dockerfile=. \
            --output type=image,name=docker.io/YOUR_ORG/YOUR_SERVICE:${{ github.sha }},push=true
      
      - name: Deploy service
        run: |
          curl -X POST http://spark-apply-api:8080/deploy \
            -H "Content-Type: application/json" \
            -d '{"service_name": "your-service", "tag": "${{ github.sha }}", "compose_filename": "docker-compose.yml"}'
```

**3. Setup spark-apply:**
```bash
git clone https://github.com/skyapps-id/spark-apply.git
cd spark-apply
go build
./spark-apply
```

## Usage Examples

```bash
# Create network (one-time setup)
docker network create pipeline-net

# Deploy service
curl -X POST http://spark-apply-api:8080/deploy \
  -H "Content-Type: application/json" \
  -d '{"service_name": "cctv-platform-be", "tag": "5f69a27c5e437b278e7d0a2da6a9b91dfadd935e", "compose_filename": "docker-compose.yml"}'
```

## Security

- Service name validation: only alphanumeric, underscore, and dash allowed
- Tag validation: only alphanumeric, dot, underscore, and dash allowed
- Docker socket access limited to required commands only

## Development

```bash
# Build
go build

# Run
./spark-apply

# Docker
docker compose up -d
```