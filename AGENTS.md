# MetroTracker Backend Agent Guide

## Project Overview

- Go module: `metry.uk/provext4/MetroTrackerBackend`
- API entry point: `cmd/api/main.go`
- Application code belongs under `internal/`.
- `internal/mapas/` is organized by domain, application, infrastructure, and transport layers.
- Database migrations belong in `migrations/`; tests belong next to the code they cover or in `test/` when they exercise multiple packages.

## Technology Stack

- Go for backend services.
- PostgreSQL with PostGIS for persistent and geospatial data.
- Redis for caching, ephemeral state, and real-time coordination.
- WebSockets for live client updates.
- REST for the HTTP API.

## Architecture

- Build the service as a Modular Monolith: bounded contexts are deployed together but remain independently organized and loosely coupled.
- Apply Hexagonal/Clean Architecture within each module: domain rules at the center, application use cases around them, and infrastructure and transport adapters at the edge.
- Dependencies point inward. Domain and application packages must not depend on HTTP, WebSocket, PostgreSQL/PostGIS, Redis, or other infrastructure packages.
- Define ports as interfaces in the domain or application layer; implement them through inbound adapters (REST and WebSockets) and outbound adapters (PostgreSQL/PostGIS and Redis).
- Integrate modules through explicit application contracts, not direct access to another module's infrastructure or persistence models.

### Request and Data Flow

```text
                       CLIENTS
              +-----------+-----------+
              |                       |
            REST                  WebSocket
              |                       |
              +-----------+-----------+
                          |
                   +------+------+
                   |  Transport  |
                   |  HTTP / WS  |
                   +------+------+
                          |
                   +------+------+
                   | Application |
                   |  Use Cases  |
                   +------+------+
                          |
                   +------+------+
                   |    Domain    |
                   | Business     |
                   |    Rules     |
                   +------+------+
                          |
              +-----------+-----------+
              |                       |
       +------+-------+       +-------+------+
       | PostgreSQL   |       |    Redis     |
       |  + PostGIS   |       | Cache/Event  |
       +--------------+       +--------------+
```

- REST and WebSocket handlers delegate to application use cases; they do not contain business rules.
- The application layer invokes domain logic and accesses persistence, cache, and event capabilities only through outbound ports.

## Development Commands

Run these from the repository root:

```sh
go run ./cmd/api
go test ./...
go vet ./...
gofmt -w .
```

Use `gofmt` on every changed Go file before finishing. Run `go mod tidy` only when dependency declarations actually change, and include the resulting `go.mod` and `go.sum` updates.

## Implementation Conventions

- Keep `cmd/` thin: compose dependencies and start the process there.
- Keep domain logic independent of HTTP, WebSocket, database, cache, and framework concerns.
- Define interfaces at the consuming domain or application layer; infrastructure implements those interfaces.
- Keep REST and WebSocket handlers as inbound adapters: validate and map requests, invoke a use case, and map its result to a response.
- Keep PostgreSQL/PostGIS and Redis code in outbound adapters; do not expose driver-specific types beyond the infrastructure layer.
- Do not import packages from another bounded context unless the dependency is intentional and minimal.
- Return contextual errors with `%w`; do not discard errors.
- Use explicit request/response DTOs in transport code rather than exposing domain entities directly.
- Prefer table-driven tests for business rules and edge cases.

## Change Safety

- Do not modify generated files, migrations, or public API contracts without an explicit reason.
- Make migrations forward-only and provide both an `up` and `down` migration when the migration tooling supports them.
- Keep configuration and secrets out of source control; use environment variables for runtime credentials.
- Before handing off a change, run the relevant formatter and the broadest practical test command.
