# Rate Limiter Service

Multi-tenant rate limiter as a service. Built with Go, PostgreSQL, and Redis.

## Quick Start

```bash
# Start PostgreSQL and Redis
docker-compose up -d

# Run the service (auto-migrates database on startup)
go run ./cmd/server
```

## Usage

### 1. Create an account

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "you@example.com", "password": "secret123"}'
```

### 2. Create a rate limiting profile

```bash
curl -X POST http://localhost:8080/api/v1/profiles \
  -H "X-API-Key: rl_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"name": "api", "algorithm": "token_bucket", "limit": 100, "window": "1m"}'
```

### 3. Check rate limit

```bash
curl -X POST http://localhost:8080/api/v1/rate-limit/check \
  -H "X-API-Key: rl_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"key": "user:123", "profile": "api"}'
```

