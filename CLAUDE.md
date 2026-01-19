# gofi - Go UniFi Controller

A Go module for programmatic control of UniFi UDM Pro devices (v10+).

## Project Resources

- **API Reference**: `UNIFI_UDM_PRO_API_DOCUMENTATION.md` - Complete endpoint documentation
- **Architecture**: `docs/DESIGN.md` - Design with mermaid diagrams, service interfaces, type definitions
- **Implementation Plan**: `docs/plan.md` - Phased plan with progress tracking

## Critical Rules

1. **Every function MUST have a test** - No exceptions. Run `make test` to verify.
2. **Every endpoint MUST be supported in the mock server** - Tests use the mock, not real hardware.
3. **No phase advancement without 100% test coverage** - Complete and test each phase before moving on.
4. **Phases are sequential** - Follow `docs/plan.md` in order.

## Concurrency Requirements

- **CSRF tokens**: Use `atomic.Value` for thread-safe storage and updates
- **Session refresh**: Use `sync.Mutex` to prevent concurrent refresh races
- **Rate limiting**: Implement semaphore-based request limiting
- **Connection pooling**: Configure `http.Transport` appropriately

## Architecture Overview

```
gofi/
├── client.go          # Main client, authentication, request handling
├── types.go           # All domain types (Device, Network, WLAN, etc.)
├── errors.go          # Sentinel errors, APIError type
├── services/          # Service implementations
│   ├── site.go
│   ├── device.go
│   ├── network.go
│   ├── wlan.go
│   ├── firewall.go
│   ├── client.go
│   ├── user.go
│   ├── routing.go
│   └── ...
├── mock/              # Mock server for testing
│   ├── server.go
│   ├── handlers.go
│   ├── fixtures/
│   └── scenarios/
└── examples/
```

## Key Technical Details

- **Auth**: Cookie-based session via `POST /api/auth/login`
- **CSRF**: Extract from cookie, send as `X-CSRF-Token` header
- **Base Path**: UDM Pro uses `/proxy/network` prefix
- **API Versions**: v1 (`/api/s/{site}/...`) and v2 (`/v2/api/site/{site}/...`)
- **WebSocket**: Events at `wss://{host}/proxy/network/wss/s/{site}/events`

## Type Patterns

Use flexible types for UniFi's inconsistent JSON:

```go
type FlexInt int64    // Handles "123" or 123
type FlexBool bool    // Handles "true", true, 0, 1
```

## Mock Server

The mock server must:
- Support all endpoints with realistic responses
- Allow fixture loading for consistent test data
- Support error scenarios (auth failures, rate limits, not found)
- Simulate WebSocket events

## Commands

```bash
make test      # Run all tests
make lint      # Run linter
make build     # Build the module
make coverage  # Generate coverage report
```

## Reference Implementations

Study these for patterns:
- `github.com/paultyng/go-unifi` - Terraform provider, CRUD patterns
- `github.com/unpoller/unifi` - FlexInt/FlexBool types
- `github.com/thib3113/unifi-client` - TypeScript, comprehensive types
