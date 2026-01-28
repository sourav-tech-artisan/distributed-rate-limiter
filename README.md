# Rate Limiter Service

A distributed rate limiter built with Go and Redis. Uses Token Bucket algorithm.

## Quick Start

```bash
# Make sure Redis is running
redis-server

# Run the service
go run ./cmd/server
```

Service starts at `http://localhost:8080`

## API

### Check rate limit

```bash
curl -X POST http://localhost:8080/api/v1/rate-limit/check \
  -H "Content-Type: application/json" \
  -d '{"key": "user:123", "profile": "default"}'
```

### List profiles

```bash
curl http://localhost:8080/api/v1/profiles
```

### Health check

```bash
curl http://localhost:8080/health
```

## Configuration

Edit `config.yaml` to add/modify rate limiting profiles. Each profile can have different limits.

See [DESIGN.md](./DESIGN.md) for details.
