# Rate Limiter Service

A distributed rate limiter service built with Go, Redis, and Gin. Provides flexible algorithm selection with Token Bucket as the default implementation.

## Features

- **Algorithm Selection**: Configurable rate limiting algorithms (Token Bucket by default)
- **Distributed**: Redis-backed for multi-instance deployments
- **RESTful API**: HTTP/JSON API for rate limit checks
- **Configuration**: YAML configuration with environment variable overrides
- **Circuit Breaker**: Built-in circuit breaker for Redis connection protection
- **Structured Logging**: JSON-formatted logs with configurable levels

## Status

🚧 **Under Development** - Currently in initial setup phase. Core rate limiting functionality coming soon.

## Prerequisites

- Go 1.21 or higher
- Redis (for distributed rate limiting)

## Installation

### From Source

```bash
git clone <repository-url>
cd distributed-rate-limiter
make deps
make build
```

### Using Go Install

```bash
go install github.com/souravkumar/distributed-rate-limiter/cmd/server@latest
```

## Configuration

The service uses a `config.yaml` file for configuration. You can also override settings using environment variables.

### Example Configuration

See `config.yaml` for the default configuration structure.

### Environment Variables

All configuration values can be overridden via environment variables (uppercase with underscores):

- `SERVER_HOST`
- `SERVER_PORT`
- `REDIS_ADDR`
- `RATE_LIMITER_DEFAULT_ALGORITHM`
- etc.

## Usage

### Run the Service

```bash
# Using make
make run

# Or directly
go run ./cmd/server

# Or using the binary
./ratelimiter
```

The service will start on `http://localhost:8080` by default.

### Endpoints

- `GET /health` - Health check endpoint

More endpoints coming soon (rate limit check endpoint).

## Development

### Project Structure

```
.
├── cmd/
│   └── server/          # Server entry point
├── internal/
│   ├── api/             # HTTP handlers and routes
│   ├── config/          # Configuration management
│   └── logger/          # Logging setup
├── config.yaml          # Default configuration
├── Makefile             # Build automation
└── README.md
```

### Make Commands

- `make deps` - Download dependencies
- `make build` - Build the binary
- `make run` - Run the service
- `make test` - Run tests
- `make clean` - Clean build artifacts

## Design Documentation

- [DESIGN.md](./DESIGN.md) - Detailed design and architecture
- [FRAMEWORK.md](./FRAMEWORK.md) - Framework and library selection
- [FUTURE.md](./FUTURE.md) - Planned improvements and features

## License

MIT

