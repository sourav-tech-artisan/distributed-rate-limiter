# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added - Initial Setup (First Commit)

- **Project Structure**: Set up basic Go project structure with `cmd/server` entry point and `internal` packages
- **Configuration Management**: 
  - Viper integration for YAML config file + environment variable overrides
  - Configuration structure for server, Redis, rate limiter, circuit breaker, and logging
  - `config.yaml` template with default values
- **HTTP Server**:
  - Gin framework setup
  - Basic router with health check endpoint (`GET /health`)
  - Placeholder for rate limit check endpoint (`POST /api/v1/rate-limit/check`)
  - Graceful shutdown support
- **Logging**:
  - Zerolog integration for structured JSON logging
  - Configurable log levels and format (JSON/console)
  - Request logging middleware (basic setup)
- **Build System**:
  - Makefile with common commands (build, run, test, clean, deps)
  - `.gitignore` for build artifacts

### Framework Stack

- **HTTP Framework**: Gin v1.10.0
- **Configuration**: Viper v1.19.0
- **Logging**: Zerolog v1.32.0
- **Redis Client**: (To be added in next commit)
- **Circuit Breaker**: (To be added in next commit)

### Next Steps

- Redis client integration
- Circuit breaker implementation
- Token bucket rate limiter algorithm
- Rate limit check endpoint implementation

