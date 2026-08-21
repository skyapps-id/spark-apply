# Spark Apply

HTTP server untuk deploy docker-compose services dengan update image tag.

## Fitur

- Deploy service via HTTP POST
- Update image tag di docker-compose.yaml
- Validasi service name dan tag untuk keamanan
- Support environment variables untuk konfigurasi

## Environment Variables

- `PORT`: Port server (default: 8080)
- `BASE_FOLDER`: Path ke folder services (default: /Users/ajii/Documents/spark-apply/services)

## API Endpoint

### POST /deploy

Deploy service dengan update image tag.

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

## Contoh Usage

```bash
curl -X POST http://spark-apply-api:8080/deploy \
  -H "Content-Type: application/json" \
  -d '{"service_name": "cctv-platform-be", "tag": "5f69a27c5e437b278e7d0a2da6a9b91dfadd935e", "compose_filename": "docker-compose.yml"}'
```

## Security

- Service name validation: hanya alphanumeric, underscore, dan dash
- Tag validation: hanya alphanumeric, dot, underscore, dan dash
- Docker socket access terbatas untuk command yang diperlukan saja

## Development

```bash
# Build
go build

# Run
./spark-apply

# Docker
docker compose up -d
```
