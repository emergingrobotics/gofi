# UDM Go Client Library - Implementation Plan

## Document Information

| Field | Value |
|-------|-------|
| Version | 1.0.0 |
| Status | Planning |
| Created | 2025-01-19 |
| Target | AI Implementation |

---

## Overview

This document provides a detailed, phased implementation plan for building the UDM Pro Go client library. **Each phase MUST be fully completed and tested before proceeding to the next phase.**

### Critical Rules

1. **No phase advancement without 100% test coverage** for that phase
2. **Every public function requires a corresponding test**
3. **Every API endpoint requires mock server support**
4. **Run `make test` and ensure all tests pass before marking any task complete**
5. **Commit after each completed phase**

---

## Progress Tracking Legend

- `[ ]` - Not started
- `[~]` - In progress
- `[x]` - Completed and tested
- `[!]` - Blocked or needs attention

---

## Phase 0: Project Scaffolding

**Goal:** Set up the project structure, Go module, and build infrastructure.

**Prerequisites:** None

**Exit Criteria:**
- `go build ./...` succeeds
- `make test` runs (even with no tests)
- All directories exist

### Tasks

#### 0.1 Initialize Go Module
- [ ] Create `go.mod` with module path `github.com/[org]/udm`
- [ ] Set Go version to 1.21+
- [ ] Create `go.sum` (empty initially)

#### 0.2 Create Directory Structure
- [ ] Create `auth/` directory
- [ ] Create `transport/` directory
- [ ] Create `types/` directory
- [ ] Create `api/v1/` directory
- [ ] Create `api/v2/` directory
- [ ] Create `services/` directory
- [ ] Create `websocket/` directory
- [ ] Create `mock/` directory
- [ ] Create `mock/fixtures/` directory
- [ ] Create `mock/scenarios/` directory
- [ ] Create `internal/` directory
- [ ] Create `examples/` directory

#### 0.3 Create Makefile
- [ ] Create `Makefile` with targets:
  ```makefile
  .PHONY: all build test lint clean

  all: lint test build

  build:
  	go build ./...

  test:
  	go test -v -race -cover ./...

  lint:
  	golangci-lint run ./...

  clean:
  	go clean ./...

  coverage:
  	go test -coverprofile=coverage.out ./...
  	go tool cover -html=coverage.out -o coverage.html
  ```

#### 0.4 Create Package Documentation Files
- [ ] Create `doc.go` with package documentation
- [ ] Create `auth/doc.go`
- [ ] Create `transport/doc.go`
- [ ] Create `types/doc.go`
- [ ] Create `services/doc.go`
- [ ] Create `websocket/doc.go`
- [ ] Create `mock/doc.go`
- [ ] Create `internal/doc.go`

#### 0.5 Create Placeholder Files
- [ ] Create `client.go` (empty Client interface)
- [ ] Create `config.go` (empty Config struct)
- [ ] Create `errors.go` (empty error definitions)
- [ ] Create `options.go` (empty Option type)

#### 0.6 Verification
- [ ] Run `go build ./...` - must succeed
- [ ] Run `go mod tidy` - must succeed
- [ ] Run `make test` - must succeed (no tests yet is OK)

---

## Phase 1: Core Types

**Goal:** Define all data types used throughout the library.

**Prerequisites:** Phase 0 complete

**Exit Criteria:**
- All types compile
- JSON marshaling/unmarshaling tests pass for all types
- `make test` passes

### Tasks

#### 1.1 Flexible JSON Types (`types/flex.go`)
- [ ] Implement `FlexInt` struct
  - [ ] `Val float64` field
  - [ ] `Txt string` field
  - [ ] `UnmarshalJSON([]byte) error` method
  - [ ] `MarshalJSON() ([]byte, error)` method
  - [ ] `Int() int` method
  - [ ] `Int64() int64` method
  - [ ] `Float64() float64` method
  - [ ] `String() string` method
- [ ] Implement `FlexBool` struct
  - [ ] `Val bool` field
  - [ ] `Txt string` field
  - [ ] `UnmarshalJSON([]byte) error` method
  - [ ] `MarshalJSON() ([]byte, error)` method
  - [ ] `Bool() bool` method
  - [ ] `String() string` method
- [ ] Implement `FlexString` struct (string or []string)
  - [ ] `Val string` field
  - [ ] `Arr []string` field
  - [ ] `UnmarshalJSON([]byte) error` method
  - [ ] `MarshalJSON() ([]byte, error)` method
- [ ] Write tests: `types/flex_test.go`
  - [ ] Test FlexInt with numeric JSON
  - [ ] Test FlexInt with string JSON
  - [ ] Test FlexBool with bool JSON
  - [ ] Test FlexBool with string JSON ("true", "false")
  - [ ] Test FlexBool with numeric JSON (0, 1)
  - [ ] Test FlexString with string JSON
  - [ ] Test FlexString with array JSON

#### 1.2 Common Types (`types/common.go`)
- [ ] Define `APIResponse[T any]` generic struct
  - [ ] `Meta` struct with `RC`, `Message`, `Count`
  - [ ] `Data []T` field
- [ ] Define `CommandRequest` struct
  - [ ] `Cmd string` field
  - [ ] Additional fields for various commands
- [ ] Define `MAC` type (string alias with validation)
- [ ] Define `DeviceState` enum (int with String method)
- [ ] Write tests: `types/common_test.go`
  - [ ] Test APIResponse JSON parsing
  - [ ] Test MAC validation
  - [ ] Test DeviceState String()

#### 1.3 Site Types (`types/site.go`)
- [ ] Define `Site` struct with all fields from API docs
- [ ] Define `HealthData` struct
- [ ] Define `SysInfo` struct
- [ ] Write tests: `types/site_test.go`
  - [ ] Test Site JSON marshaling/unmarshaling
  - [ ] Test HealthData JSON marshaling/unmarshaling
  - [ ] Test SysInfo JSON marshaling/unmarshaling

#### 1.4 Device Types (`types/device.go`)
- [ ] Define `Device` struct with all fields
- [ ] Define `DeviceBasic` struct
- [ ] Define `DeviceUplink` struct
- [ ] Define `DeviceConfigNetwork` struct
- [ ] Define `SystemStats` struct
- [ ] Define `RadioTable` struct
- [ ] Define `RadioTableStats` struct
- [ ] Define `VAPTable` struct
- [ ] Define `PortTable` struct
- [ ] Define `Temperature` struct
- [ ] Define `Storage` struct
- [ ] Write tests: `types/device_test.go`
  - [ ] Test Device JSON with AP data
  - [ ] Test Device JSON with Switch data
  - [ ] Test Device JSON with UDM data
  - [ ] Test DeviceState constants

#### 1.5 Network Types (`types/network.go`)
- [ ] Define `Network` struct with all fields
- [ ] Define `NetworkPurpose` constants
- [ ] Write tests: `types/network_test.go`
  - [ ] Test Network JSON marshaling/unmarshaling
  - [ ] Test with VLAN enabled/disabled
  - [ ] Test with DHCP enabled/disabled

#### 1.6 WLAN Types (`types/wlan.go`)
- [ ] Define `WLAN` struct with all fields
- [ ] Define `WLANGroup` struct
- [ ] Define `SecurityType` constants
- [ ] Define `WPAMode` constants
- [ ] Define `MACFilterPolicy` constants
- [ ] Write tests: `types/wlan_test.go`
  - [ ] Test WLAN JSON marshaling/unmarshaling
  - [ ] Test WLANGroup JSON marshaling/unmarshaling
  - [ ] Test various security configurations

#### 1.7 Firewall Types (`types/firewall.go`)
- [ ] Define `FirewallRule` struct with all fields
- [ ] Define `FirewallGroup` struct
- [ ] Define `FirewallRuleIndexUpdate` struct
- [ ] Define `Ruleset` constants
- [ ] Define `Action` constants
- [ ] Define `GroupType` constants
- [ ] Write tests: `types/firewall_test.go`
  - [ ] Test FirewallRule JSON marshaling/unmarshaling
  - [ ] Test FirewallGroup JSON marshaling/unmarshaling
  - [ ] Test with various rulesets

#### 1.8 Traffic Rule Types (`types/trafficrule.go`)
- [ ] Define `TrafficRule` struct (v2 API)
- [ ] Define `TargetDevice` struct
- [ ] Define `Schedule` struct
- [ ] Define `Bandwidth` struct
- [ ] Define `IPRange` struct
- [ ] Define `TrafficAction` constants
- [ ] Define `MatchingTarget` constants
- [ ] Write tests: `types/trafficrule_test.go`
  - [ ] Test TrafficRule JSON marshaling/unmarshaling
  - [ ] Test with various schedules
  - [ ] Test with various targets

#### 1.9 Client/Station Types (`types/client.go`)
- [ ] Define `Client` struct with all fields
- [ ] Write tests: `types/client_test.go`
  - [ ] Test Client JSON with wireless client
  - [ ] Test Client JSON with wired client
  - [ ] Test Client JSON with guest client

#### 1.10 User Types (`types/user.go`)
- [ ] Define `User` struct (known client)
- [ ] Define `UserGroup` struct
- [ ] Write tests: `types/user_test.go`
  - [ ] Test User JSON marshaling/unmarshaling
  - [ ] Test UserGroup JSON marshaling/unmarshaling

#### 1.11 Routing Types (`types/routing.go`)
- [ ] Define `Route` struct
- [ ] Write tests: `types/routing_test.go`
  - [ ] Test Route JSON marshaling/unmarshaling

#### 1.12 Port Types (`types/port.go`)
- [ ] Define `PortForward` struct
- [ ] Define `PortProfile` struct
- [ ] Define `Protocol` constants
- [ ] Write tests: `types/port_test.go`
  - [ ] Test PortForward JSON marshaling/unmarshaling
  - [ ] Test PortProfile JSON marshaling/unmarshaling

#### 1.13 Setting Types (`types/setting.go`)
- [ ] Define `Setting` struct (base)
- [ ] Define `SettingMgmt` struct
- [ ] Define `SettingConnectivity` struct
- [ ] Define `SettingCountry` struct
- [ ] Define `SettingGuestAccess` struct
- [ ] Define `SettingDPI` struct
- [ ] Define `SettingIPS` struct
- [ ] Define `SettingNTP` struct
- [ ] Define `SettingSNMP` struct
- [ ] Define `SettingRsyslog` struct
- [ ] Define `SettingRadius` struct
- [ ] Define `RADIUSProfile` struct
- [ ] Define `DynamicDNS` struct
- [ ] Write tests: `types/setting_test.go`
  - [ ] Test each setting type JSON marshaling/unmarshaling

#### 1.14 Event Types (`types/event.go`)
- [ ] Define `Event` struct
- [ ] Define `EventType` constants
- [ ] Define `Alarm` struct
- [ ] Write tests: `types/event_test.go`
  - [ ] Test Event JSON marshaling/unmarshaling
  - [ ] Test Alarm JSON marshaling/unmarshaling

#### 1.15 System Types (`types/system.go`)
- [ ] Define `Status` struct
- [ ] Define `AdminUser` struct
- [ ] Define `Backup` struct
- [ ] Define `SpeedTestStatus` struct
- [ ] Write tests: `types/system_test.go`
  - [ ] Test each type JSON marshaling/unmarshaling

#### 1.16 Phase 1 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors
- [ ] Verify test coverage > 90% for types package

---

## Phase 2: Internal Utilities

**Goal:** Implement internal helper functions.

**Prerequisites:** Phase 1 complete

**Exit Criteria:**
- All utility functions tested
- `make test` passes

### Tasks

#### 2.1 MAC Address Utilities (`internal/mac.go`)
- [ ] Implement `NormalizeMAC(string) string` - lowercase, no separators
- [ ] Implement `FormatMAC(string) string` - colon-separated
- [ ] Implement `ValidateMAC(string) bool`
- [ ] Write tests: `internal/mac_test.go`
  - [ ] Test NormalizeMAC with various formats
  - [ ] Test FormatMAC with various formats
  - [ ] Test ValidateMAC with valid/invalid MACs

#### 2.2 URL Building (`internal/url.go`)
- [ ] Implement `BuildAPIPath(site, endpoint string) string`
- [ ] Implement `BuildV2APIPath(site, endpoint string) string`
- [ ] Implement `BuildRESTPath(site, resource, id string) string`
- [ ] Implement `BuildCmdPath(site, manager string) string`
- [ ] Write tests: `internal/url_test.go`
  - [ ] Test each path builder

#### 2.3 JSON Utilities (`internal/json.go`)
- [ ] Implement `ParseAPIResponse[T]([]byte) (*types.APIResponse[T], error)`
- [ ] Implement `IsErrorResponse([]byte) bool`
- [ ] Implement `ExtractErrorMessage([]byte) string`
- [ ] Write tests: `internal/json_test.go`
  - [ ] Test parsing success responses
  - [ ] Test parsing error responses
  - [ ] Test with malformed JSON

#### 2.4 Phase 2 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 3: Error Handling

**Goal:** Implement comprehensive error types and handling.

**Prerequisites:** Phase 2 complete

**Exit Criteria:**
- All error types defined
- Error wrapping/unwrapping works correctly
- `make test` passes

### Tasks

#### 3.1 Sentinel Errors (`errors.go`)
- [ ] Define `ErrNotConnected`
- [ ] Define `ErrAlreadyConnected`
- [ ] Define `ErrAuthenticationFailed`
- [ ] Define `ErrSessionExpired`
- [ ] Define `ErrInvalidCSRFToken`
- [ ] Define `ErrNotFound`
- [ ] Define `ErrPermissionDenied`
- [ ] Define `ErrAlreadyExists`
- [ ] Define `ErrInvalidRequest`
- [ ] Define `ErrRateLimited`
- [ ] Define `ErrServerError`

#### 3.2 APIError Type (`errors.go`)
- [ ] Define `APIError` struct
  - [ ] `StatusCode int`
  - [ ] `RC string`
  - [ ] `Message string`
  - [ ] `Endpoint string`
- [ ] Implement `Error() string` method
- [ ] Implement `Is(error) bool` method
- [ ] Implement `Unwrap() error` method

#### 3.3 ValidationError Type (`errors.go`)
- [ ] Define `ValidationError` struct
  - [ ] `Field string`
  - [ ] `Message string`
- [ ] Implement `Error() string` method

#### 3.4 Error Tests (`errors_test.go`)
- [ ] Test APIError.Error() output format
- [ ] Test APIError.Is() with each sentinel error
- [ ] Test APIError.Unwrap() returns correct sentinel
- [ ] Test errors.Is() works through wrapping
- [ ] Test ValidationError.Error() output format

#### 3.5 Phase 3 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 4: Transport Layer

**Goal:** Implement HTTP transport with authentication support.

**Prerequisites:** Phase 3 complete

**Exit Criteria:**
- HTTP client can make requests
- Cookie handling works
- CSRF token handling works
- `make test` passes

### Tasks

#### 4.1 Transport Configuration (`transport/config.go`)
- [ ] Define `Config` struct for transport layer
- [ ] Define `Option` functional options type
- [ ] Implement `WithTimeout(time.Duration) Option`
- [ ] Implement `WithTLSConfig(*tls.Config) Option`
- [ ] Write tests: `transport/config_test.go`

#### 4.2 Request Builder (`transport/request.go`)
- [ ] Define `Request` struct
  - [ ] `Method string`
  - [ ] `Path string`
  - [ ] `Body interface{}`
  - [ ] `Headers map[string]string`
- [ ] Implement `NewRequest(method, path string) *Request`
- [ ] Implement `(*Request) WithBody(interface{}) *Request`
- [ ] Implement `(*Request) WithHeader(key, value string) *Request`
- [ ] Write tests: `transport/request_test.go`

#### 4.3 Response Parser (`transport/response.go`)
- [ ] Define `Response` struct
  - [ ] `StatusCode int`
  - [ ] `Body []byte`
  - [ ] `Headers http.Header`
- [ ] Implement `(*Response) Parse(v interface{}) error`
- [ ] Implement `(*Response) IsSuccess() bool`
- [ ] Implement `(*Response) Error() error`
- [ ] Write tests: `transport/response_test.go`

#### 4.4 HTTP Transport (`transport/transport.go`)
- [ ] Define `Transport` interface
  ```go
  type Transport interface {
      Do(ctx context.Context, req *Request) (*Response, error)
      SetCSRFToken(token string)
      GetCSRFToken() string
  }
  ```
- [ ] Implement `httpTransport` struct
  - [ ] `client *http.Client`
  - [ ] `baseURL *url.URL`
  - [ ] `csrfToken atomic.Value`
- [ ] Implement `New(baseURL string, opts ...Option) (Transport, error)`
- [ ] Implement `Do(ctx, *Request) (*Response, error)` method
  - [ ] Build full URL
  - [ ] Serialize body if present
  - [ ] Add headers (Content-Type, Accept, CSRF)
  - [ ] Execute request
  - [ ] Handle response
- [ ] Implement `SetCSRFToken(string)` method
- [ ] Implement `GetCSRFToken() string` method
- [ ] Write tests: `transport/transport_test.go`
  - [ ] Test successful GET request
  - [ ] Test successful POST request
  - [ ] Test request with body
  - [ ] Test CSRF token inclusion
  - [ ] Test error handling

#### 4.5 Retry Handler (`transport/retry.go`)
- [ ] Define `RetryConfig` struct
- [ ] Implement `RetryTransport` wrapper
- [ ] Implement retry logic with exponential backoff
- [ ] Write tests: `transport/retry_test.go`
  - [ ] Test retry on 5xx errors
  - [ ] Test no retry on 4xx errors
  - [ ] Test max retries limit
  - [ ] Test backoff timing

#### 4.6 Phase 4 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 5: Authentication

**Goal:** Implement authentication management.

**Prerequisites:** Phase 4 complete

**Exit Criteria:**
- Login/logout works
- Session refresh works
- CSRF token is managed correctly
- `make test` passes

### Tasks

#### 5.1 Session Management (`auth/session.go`)
- [ ] Define `Session` struct
  - [ ] `Token string`
  - [ ] `CSRFToken string`
  - [ ] `ExpiresAt time.Time`
  - [ ] `Username string`
- [ ] Implement `(*Session) IsValid() bool`
- [ ] Implement `(*Session) NeedsRefresh() bool`
- [ ] Write tests: `auth/session_test.go`

#### 5.2 Auth Manager (`auth/auth.go`)
- [ ] Define `Manager` interface
  ```go
  type Manager interface {
      Login(ctx context.Context) error
      Logout(ctx context.Context) error
      EnsureAuthenticated(ctx context.Context) error
      Session() *Session
  }
  ```
- [ ] Implement `manager` struct
  - [ ] `transport transport.Transport`
  - [ ] `username string`
  - [ ] `password string`
  - [ ] `session *Session`
  - [ ] `mu sync.RWMutex`
  - [ ] `refreshing bool`
  - [ ] `refreshCh chan struct{}`
- [ ] Implement `New(transport, username, password string) Manager`
- [ ] Implement `Login(ctx) error` method
  - [ ] POST to `/api/auth/login`
  - [ ] Parse response
  - [ ] Extract CSRF token from headers
  - [ ] Store session
- [ ] Implement `Logout(ctx) error` method
  - [ ] POST to `/api/logout`
  - [ ] Clear session
- [ ] Implement `EnsureAuthenticated(ctx) error` method
  - [ ] Check if session valid
  - [ ] Handle concurrent refresh
  - [ ] Refresh if needed
- [ ] Implement `Session() *Session` method
- [ ] Write tests: `auth/auth_test.go`
  - [ ] Test successful login
  - [ ] Test failed login
  - [ ] Test logout
  - [ ] Test session refresh
  - [ ] Test concurrent refresh handling

#### 5.3 CSRF Handler (`auth/csrf.go`)
- [ ] Implement `CSRFHandler` struct with atomic token storage
- [ ] Implement `Get() string`
- [ ] Implement `Set(string)`
- [ ] Implement `UpdateFromResponse(*http.Response)`
- [ ] Write tests: `auth/csrf_test.go`
  - [ ] Test concurrent access
  - [ ] Test update from response

#### 5.4 Phase 5 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 6: Mock Server Foundation

**Goal:** Implement the mock server infrastructure.

**Prerequisites:** Phase 5 complete

**Exit Criteria:**
- Mock server can start/stop
- Authentication endpoints work
- Basic request routing works
- `make test` passes

### Tasks

#### 6.1 Mock State (`mock/state.go`)
- [ ] Define `State` struct with all data stores
- [ ] Implement `NewState() *State`
- [ ] Implement thread-safe accessors for each store
- [ ] Implement `Reset()` method
- [ ] Write tests: `mock/state_test.go`

#### 6.2 Mock Server Core (`mock/server.go`)
- [ ] Define `Server` struct
- [ ] Implement `NewServer(opts ...Option) *Server`
- [ ] Implement `Close()`
- [ ] Implement `URL() string`
- [ ] Implement `Host() string`
- [ ] Implement route registration
- [ ] Write tests: `mock/server_test.go`
  - [ ] Test server starts
  - [ ] Test server stops
  - [ ] Test TLS works

#### 6.3 Mock Options (`mock/options.go`)
- [ ] Implement `WithoutAuth() Option`
- [ ] Implement `WithoutCSRF() Option`
- [ ] Implement `WithFixtures(*Fixtures) Option`
- [ ] Implement `WithScenario(Scenario) Option`
- [ ] Write tests: `mock/options_test.go`

#### 6.4 Auth Handlers (`mock/handlers_auth.go`)
- [ ] Implement `handleLogin(w, r)`
- [ ] Implement `handleLogout(w, r)`
- [ ] Implement `handleStatus(w, r)`
- [ ] Implement `handleSelf(w, r)`
- [ ] Implement authentication verification helpers
- [ ] Implement CSRF verification helpers
- [ ] Write tests: `mock/handlers_auth_test.go`
  - [ ] Test login success
  - [ ] Test login failure
  - [ ] Test logout
  - [ ] Test auth verification
  - [ ] Test CSRF verification

#### 6.5 Response Helpers (`mock/response.go`)
- [ ] Implement `writeJSON(w, status, data)`
- [ ] Implement `writeAPIResponse(w, data)`
- [ ] Implement `writeAPIError(w, status, msg)`
- [ ] Write tests: `mock/response_test.go`

#### 6.6 Fixtures (`mock/fixtures.go`)
- [ ] Define `Fixtures` struct
- [ ] Implement `DefaultFixtures() *Fixtures`
- [ ] Implement `LoadFixtures(dir string) (*Fixtures, error)`
- [ ] Implement `(*State) LoadFixtures(*Fixtures)`
- [ ] Create `mock/fixtures/devices.json`
- [ ] Create `mock/fixtures/networks.json`
- [ ] Create `mock/fixtures/wlans.json`
- [ ] Create `mock/fixtures/clients.json`
- [ ] Write tests: `mock/fixtures_test.go`

#### 6.7 Scenarios (`mock/scenarios.go`)
- [ ] Define `Scenario` interface
- [ ] Implement `ErrorScenario` struct
- [ ] Implement predefined scenarios:
  - [ ] `ScenarioSessionExpired`
  - [ ] `ScenarioCSRFFailure`
  - [ ] `ScenarioRateLimit`
  - [ ] `ScenarioServerError`
- [ ] Write tests: `mock/scenarios_test.go`

#### 6.8 Phase 6 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 7: Client Core

**Goal:** Implement the main Client type and connection management.

**Prerequisites:** Phase 6 complete

**Exit Criteria:**
- Client can connect/disconnect
- Client integrates with transport and auth
- `make test` passes with mock server

### Tasks

#### 7.1 Configuration (`config.go`)
- [ ] Define `Config` struct with all fields
- [ ] Implement config validation
- [ ] Implement default values
- [ ] Write tests: `config_test.go`

#### 7.2 Options (`options.go`)
- [ ] Define `Option` type
- [ ] Implement `WithTimeout(time.Duration) Option`
- [ ] Implement `WithRetry(maxRetries int, backoff time.Duration) Option`
- [ ] Implement `WithLogger(Logger) Option`
- [ ] Implement `WithTLSConfig(*tls.Config) Option`
- [ ] Implement `WithHTTPClient(*http.Client) Option`
- [ ] Implement `WithUserAgent(string) Option`
- [ ] Write tests: `options_test.go`

#### 7.3 Client Interface (`client.go`)
- [ ] Define `Client` interface with all methods
- [ ] Define `Logger` interface
- [ ] Write documentation for interface

#### 7.4 Client Implementation (`client_impl.go`)
- [ ] Implement `client` struct
- [ ] Implement `New(config *Config, opts ...Option) (Client, error)`
- [ ] Implement `Connect(ctx context.Context) error`
- [ ] Implement `Disconnect(ctx context.Context) error`
- [ ] Implement `IsConnected() bool`
- [ ] Implement `Do(ctx, *Request) (*Response, error)` for raw access
- [ ] Write tests: `client_test.go`
  - [ ] Test New with valid config
  - [ ] Test New with invalid config
  - [ ] Test Connect success
  - [ ] Test Connect failure
  - [ ] Test Disconnect
  - [ ] Test IsConnected states

#### 7.5 Integration Test
- [ ] Write `client_integration_test.go`
  - [ ] Test full connect/disconnect cycle with mock
  - [ ] Test reconnection on session expiry
  - [ ] Test CSRF refresh

#### 7.6 Phase 7 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 8: Site Service

**Goal:** Implement site management operations.

**Prerequisites:** Phase 7 complete

**Exit Criteria:**
- All SiteService methods implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 8.1 Mock Handlers for Sites (`mock/handlers_site.go`)
- [ ] Implement `handleListSites(w, r)`
- [ ] Implement `handleGetSite(w, r, id)`
- [ ] Implement `handleCreateSite(w, r)`
- [ ] Implement `handleUpdateSite(w, r, id)`
- [ ] Implement `handleDeleteSite(w, r, id)`
- [ ] Implement `handleHealth(w, r, site)`
- [ ] Implement `handleSysInfo(w, r, site)`
- [ ] Write tests: `mock/handlers_site_test.go`

#### 8.2 Site Service Interface (`services/site.go`)
- [ ] Define `SiteService` interface
- [ ] Implement `siteService` struct
- [ ] Implement `List(ctx) ([]types.Site, error)`
- [ ] Implement `Get(ctx, id) (*types.Site, error)`
- [ ] Implement `Create(ctx, name, desc) (*types.Site, error)`
- [ ] Implement `Update(ctx, *types.Site) (*types.Site, error)`
- [ ] Implement `Delete(ctx, id) error`
- [ ] Implement `Health(ctx, site) ([]types.HealthData, error)`
- [ ] Implement `SysInfo(ctx, site) (*types.SysInfo, error)`
- [ ] Write tests: `services/site_test.go`
  - [ ] Test List success
  - [ ] Test List empty
  - [ ] Test Get success
  - [ ] Test Get not found
  - [ ] Test Create success
  - [ ] Test Update success
  - [ ] Test Delete success
  - [ ] Test Health success
  - [ ] Test SysInfo success

#### 8.3 Client Integration
- [ ] Add `Sites() SiteService` method to client
- [ ] Implement lazy initialization
- [ ] Write integration tests

#### 8.4 Phase 8 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors
- [ ] Verify 100% test coverage for site service

---

## Phase 9: Device Service

**Goal:** Implement device management operations.

**Prerequisites:** Phase 8 complete

**Exit Criteria:**
- All DeviceService methods implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 9.1 Mock Handlers for Devices (`mock/handlers_device.go`)
- [ ] Implement `handleDeviceStat(w, r, site, endpoint)`
- [ ] Implement `handleDeviceBasicStat(w, r, site)`
- [ ] Implement `handleDeviceUpdate(w, r, site, id)`
- [ ] Implement `handleDeviceCommand(w, r, site)` with commands:
  - [ ] adopt
  - [ ] restart
  - [ ] force-provision
  - [ ] upgrade
  - [ ] upgrade-external
  - [ ] set-locate
  - [ ] unset-locate
  - [ ] power-cycle
  - [ ] spectrum-scan
- [ ] Write tests: `mock/handlers_device_test.go`

#### 9.2 Device Service Interface (`services/device.go`)
- [ ] Define `DeviceService` interface
- [ ] Implement `deviceService` struct
- [ ] Implement `List(ctx, site) ([]types.Device, error)`
- [ ] Implement `ListBasic(ctx, site) ([]types.DeviceBasic, error)`
- [ ] Implement `Get(ctx, site, id) (*types.Device, error)`
- [ ] Implement `GetByMAC(ctx, site, mac) (*types.Device, error)`
- [ ] Implement `Update(ctx, site, *types.Device) (*types.Device, error)`
- [ ] Implement `Adopt(ctx, site, mac) error`
- [ ] Implement `Forget(ctx, site, mac) error`
- [ ] Implement `Restart(ctx, site, mac) error`
- [ ] Implement `ForceProvision(ctx, site, mac) error`
- [ ] Implement `Upgrade(ctx, site, mac) error`
- [ ] Implement `UpgradeExternal(ctx, site, mac, url) error`
- [ ] Implement `Locate(ctx, site, mac) error`
- [ ] Implement `Unlocate(ctx, site, mac) error`
- [ ] Implement `PowerCyclePort(ctx, site, switchMAC, portIdx) error`
- [ ] Implement `SetLEDOverride(ctx, site, mac, mode) error`
- [ ] Implement `SpectrumScan(ctx, site, mac) error`
- [ ] Write tests: `services/device_test.go`
  - [ ] Test each method success case
  - [ ] Test each method error case
  - [ ] Test MAC normalization

#### 9.3 Client Integration
- [ ] Add `Devices() DeviceService` method to client
- [ ] Write integration tests

#### 9.4 Phase 9 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors
- [ ] Verify 100% test coverage for device service

---

## Phase 10: Network Service

**Goal:** Implement network management operations.

**Prerequisites:** Phase 9 complete

**Exit Criteria:**
- All NetworkService methods implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 10.1 Mock Handlers for Networks (`mock/handlers_network.go`)
- [ ] Implement `handleNetworkREST(w, r, site, endpoint)`
- [ ] Handle GET (list/get)
- [ ] Handle POST (create)
- [ ] Handle PUT (update)
- [ ] Handle DELETE
- [ ] Write tests: `mock/handlers_network_test.go`

#### 10.2 Network Service Interface (`services/network.go`)
- [ ] Define `NetworkService` interface
- [ ] Implement `networkService` struct
- [ ] Implement `List(ctx, site) ([]types.Network, error)`
- [ ] Implement `Get(ctx, site, id) (*types.Network, error)`
- [ ] Implement `Create(ctx, site, *types.Network) (*types.Network, error)`
- [ ] Implement `Update(ctx, site, *types.Network) (*types.Network, error)`
- [ ] Implement `Delete(ctx, site, id) error`
- [ ] Write tests: `services/network_test.go`
  - [ ] Test List success
  - [ ] Test Get success
  - [ ] Test Get not found
  - [ ] Test Create success
  - [ ] Test Create validation error
  - [ ] Test Update success
  - [ ] Test Delete success

#### 10.3 Client Integration
- [ ] Add `Networks() NetworkService` method to client
- [ ] Write integration tests

#### 10.4 Phase 10 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 11: WLAN Service

**Goal:** Implement WLAN management operations.

**Prerequisites:** Phase 10 complete

**Exit Criteria:**
- All WLANService methods implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 11.1 Mock Handlers for WLANs (`mock/handlers_wlan.go`)
- [ ] Implement `handleWLANREST(w, r, site, endpoint)`
- [ ] Implement `handleWLANGroupREST(w, r, site, endpoint)`
- [ ] Write tests: `mock/handlers_wlan_test.go`

#### 11.2 WLAN Service Interface (`services/wlan.go`)
- [ ] Define `WLANService` interface
- [ ] Implement `wlanService` struct
- [ ] Implement all WLAN methods:
  - [ ] `List(ctx, site) ([]types.WLAN, error)`
  - [ ] `Get(ctx, site, id) (*types.WLAN, error)`
  - [ ] `Create(ctx, site, *types.WLAN) (*types.WLAN, error)`
  - [ ] `Update(ctx, site, *types.WLAN) (*types.WLAN, error)`
  - [ ] `Delete(ctx, site, id) error`
  - [ ] `Enable(ctx, site, id) error`
  - [ ] `Disable(ctx, site, id) error`
  - [ ] `SetMACFilter(ctx, site, id, policy, macs) error`
- [ ] Implement all WLAN Group methods:
  - [ ] `ListGroups(ctx, site) ([]types.WLANGroup, error)`
  - [ ] `GetGroup(ctx, site, id) (*types.WLANGroup, error)`
  - [ ] `CreateGroup(ctx, site, *types.WLANGroup) (*types.WLANGroup, error)`
  - [ ] `UpdateGroup(ctx, site, *types.WLANGroup) (*types.WLANGroup, error)`
  - [ ] `DeleteGroup(ctx, site, id) error`
- [ ] Write tests: `services/wlan_test.go`

#### 11.3 Client Integration
- [ ] Add `WLANs() WLANService` method to client
- [ ] Write integration tests

#### 11.4 Phase 11 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 12: Firewall Service

**Goal:** Implement firewall management operations.

**Prerequisites:** Phase 11 complete

**Exit Criteria:**
- All FirewallService methods implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 12.1 Mock Handlers for Firewall (`mock/handlers_firewall.go`)
- [ ] Implement `handleFirewallRuleREST(w, r, site, endpoint)`
- [ ] Implement `handleFirewallGroupREST(w, r, site, endpoint)`
- [ ] Implement `handleFirewallReorder(w, r, site)`
- [ ] Write tests: `mock/handlers_firewall_test.go`

#### 12.2 Mock Handlers for Traffic Rules (`mock/handlers_traffic.go`)
- [ ] Implement `handleTrafficRulesV2(w, r, site, endpoint)`
- [ ] Handle GET, POST, PUT, DELETE
- [ ] Note: PUT returns 201, not 200
- [ ] Write tests: `mock/handlers_traffic_test.go`

#### 12.3 Firewall Service Interface (`services/firewall.go`)
- [ ] Define `FirewallService` interface
- [ ] Implement `firewallService` struct
- [ ] Implement all Firewall Rule methods:
  - [ ] `ListRules(ctx, site) ([]types.FirewallRule, error)`
  - [ ] `GetRule(ctx, site, id) (*types.FirewallRule, error)`
  - [ ] `CreateRule(ctx, site, *types.FirewallRule) (*types.FirewallRule, error)`
  - [ ] `UpdateRule(ctx, site, *types.FirewallRule) (*types.FirewallRule, error)`
  - [ ] `DeleteRule(ctx, site, id) error`
  - [ ] `EnableRule(ctx, site, id) error`
  - [ ] `DisableRule(ctx, site, id) error`
  - [ ] `ReorderRules(ctx, site, ruleset, []types.FirewallRuleIndexUpdate) error`
- [ ] Implement all Firewall Group methods:
  - [ ] `ListGroups(ctx, site) ([]types.FirewallGroup, error)`
  - [ ] `GetGroup(ctx, site, id) (*types.FirewallGroup, error)`
  - [ ] `CreateGroup(ctx, site, *types.FirewallGroup) (*types.FirewallGroup, error)`
  - [ ] `UpdateGroup(ctx, site, *types.FirewallGroup) (*types.FirewallGroup, error)`
  - [ ] `DeleteGroup(ctx, site, id) error`
- [ ] Implement all Traffic Rule methods (v2 API):
  - [ ] `ListTrafficRules(ctx, site) ([]types.TrafficRule, error)`
  - [ ] `GetTrafficRule(ctx, site, id) (*types.TrafficRule, error)`
  - [ ] `CreateTrafficRule(ctx, site, *types.TrafficRule) (*types.TrafficRule, error)`
  - [ ] `UpdateTrafficRule(ctx, site, *types.TrafficRule) (*types.TrafficRule, error)`
  - [ ] `DeleteTrafficRule(ctx, site, id) error`
- [ ] Write tests: `services/firewall_test.go`

#### 12.4 Client Integration
- [ ] Add `Firewall() FirewallService` method to client
- [ ] Write integration tests

#### 12.5 Phase 12 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 13: Client Service

**Goal:** Implement connected client/station operations.

**Prerequisites:** Phase 12 complete

**Exit Criteria:**
- All ClientService methods implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 13.1 Mock Handlers for Clients (`mock/handlers_client.go`)
- [ ] Implement `handleClientStat(w, r, site)`
- [ ] Implement `handleAllUserStat(w, r, site)`
- [ ] Implement `handleClientCommand(w, r, site)` with commands:
  - [ ] block-sta
  - [ ] unblock-sta
  - [ ] kick-sta
  - [ ] forget-sta
  - [ ] authorize-guest
  - [ ] unauthorize-guest
- [ ] Write tests: `mock/handlers_client_test.go`

#### 13.2 Client Service Interface (`services/client.go`)
- [ ] Define `ClientService` interface
- [ ] Define `ClientListOption` type
- [ ] Define `GuestAuthOption` type
- [ ] Implement `WithinHours(int) ClientListOption`
- [ ] Implement `WithDuration(int) GuestAuthOption`
- [ ] Implement `WithUploadLimit(int) GuestAuthOption`
- [ ] Implement `WithDownloadLimit(int) GuestAuthOption`
- [ ] Implement `WithDataLimit(int64) GuestAuthOption`
- [ ] Implement `WithAPMAC(string) GuestAuthOption`
- [ ] Implement `clientService` struct
- [ ] Implement all methods:
  - [ ] `ListActive(ctx, site) ([]types.Client, error)`
  - [ ] `ListAll(ctx, site, ...ClientListOption) ([]types.Client, error)`
  - [ ] `Get(ctx, site, mac) (*types.Client, error)`
  - [ ] `Block(ctx, site, mac) error`
  - [ ] `Unblock(ctx, site, mac) error`
  - [ ] `Kick(ctx, site, mac) error`
  - [ ] `AuthorizeGuest(ctx, site, mac, ...GuestAuthOption) error`
  - [ ] `UnauthorizeGuest(ctx, site, mac) error`
  - [ ] `Forget(ctx, site, mac) error`
  - [ ] `SetFingerprint(ctx, site, mac, devID) error`
- [ ] Write tests: `services/client_test.go`

#### 13.3 Client Integration
- [ ] Add `Clients() ClientService` method to client
- [ ] Write integration tests

#### 13.4 Phase 13 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 14: User Service

**Goal:** Implement known client/user operations.

**Prerequisites:** Phase 13 complete

**Exit Criteria:**
- All UserService methods implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 14.1 Mock Handlers for Users (`mock/handlers_user.go`)
- [ ] Implement `handleUserREST(w, r, site, endpoint)`
- [ ] Implement `handleUserGroupREST(w, r, site, endpoint)`
- [ ] Write tests: `mock/handlers_user_test.go`

#### 14.2 User Service Interface (`services/user.go`)
- [ ] Define `UserService` interface
- [ ] Implement `userService` struct
- [ ] Implement all User methods:
  - [ ] `List(ctx, site) ([]types.User, error)`
  - [ ] `Get(ctx, site, id) (*types.User, error)`
  - [ ] `GetByMAC(ctx, site, mac) (*types.User, error)`
  - [ ] `Create(ctx, site, *types.User) (*types.User, error)`
  - [ ] `Update(ctx, site, *types.User) (*types.User, error)`
  - [ ] `Delete(ctx, site, id) error`
  - [ ] `DeleteByMAC(ctx, site, mac) error`
  - [ ] `SetFixedIP(ctx, site, mac, ip, networkID) error`
  - [ ] `ClearFixedIP(ctx, site, mac) error`
- [ ] Implement all User Group methods:
  - [ ] `ListGroups(ctx, site) ([]types.UserGroup, error)`
  - [ ] `GetGroup(ctx, site, id) (*types.UserGroup, error)`
  - [ ] `CreateGroup(ctx, site, *types.UserGroup) (*types.UserGroup, error)`
  - [ ] `UpdateGroup(ctx, site, *types.UserGroup) (*types.UserGroup, error)`
  - [ ] `DeleteGroup(ctx, site, id) error`
- [ ] Write tests: `services/user_test.go`

#### 14.3 Client Integration
- [ ] Add `Users() UserService` method to client
- [ ] Write integration tests

#### 14.4 Phase 14 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 15: Routing & Port Services

**Goal:** Implement routing, port forwarding, and port profile operations.

**Prerequisites:** Phase 14 complete

**Exit Criteria:**
- All routing and port services implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 15.1 Mock Handlers (`mock/handlers_routing.go`, `mock/handlers_port.go`)
- [ ] Implement `handleRoutingREST(w, r, site, endpoint)`
- [ ] Implement `handlePortForwardREST(w, r, site, endpoint)`
- [ ] Implement `handlePortProfileREST(w, r, site, endpoint)`
- [ ] Write tests for each handler

#### 15.2 Routing Service (`services/routing.go`)
- [ ] Define `RoutingService` interface
- [ ] Implement all methods:
  - [ ] `List(ctx, site) ([]types.Route, error)`
  - [ ] `Get(ctx, site, id) (*types.Route, error)`
  - [ ] `Create(ctx, site, *types.Route) (*types.Route, error)`
  - [ ] `Update(ctx, site, *types.Route) (*types.Route, error)`
  - [ ] `Delete(ctx, site, id) error`
  - [ ] `Enable(ctx, site, id) error`
  - [ ] `Disable(ctx, site, id) error`
- [ ] Write tests: `services/routing_test.go`

#### 15.3 Port Forward Service (`services/portforward.go`)
- [ ] Define `PortForwardService` interface
- [ ] Implement all methods:
  - [ ] `List(ctx, site) ([]types.PortForward, error)`
  - [ ] `Get(ctx, site, id) (*types.PortForward, error)`
  - [ ] `Create(ctx, site, *types.PortForward) (*types.PortForward, error)`
  - [ ] `Update(ctx, site, *types.PortForward) (*types.PortForward, error)`
  - [ ] `Delete(ctx, site, id) error`
  - [ ] `Enable(ctx, site, id) error`
  - [ ] `Disable(ctx, site, id) error`
- [ ] Write tests: `services/portforward_test.go`

#### 15.4 Port Profile Service (`services/portprofile.go`)
- [ ] Define `PortProfileService` interface
- [ ] Implement all methods:
  - [ ] `List(ctx, site) ([]types.PortProfile, error)`
  - [ ] `Get(ctx, site, id) (*types.PortProfile, error)`
  - [ ] `Create(ctx, site, *types.PortProfile) (*types.PortProfile, error)`
  - [ ] `Update(ctx, site, *types.PortProfile) (*types.PortProfile, error)`
  - [ ] `Delete(ctx, site, id) error`
- [ ] Write tests: `services/portprofile_test.go`

#### 15.5 Client Integration
- [ ] Add `Routing() RoutingService` method to client
- [ ] Add `PortForwards() PortForwardService` method to client
- [ ] Add `PortProfiles() PortProfileService` method to client
- [ ] Write integration tests

#### 15.6 Phase 15 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 16: Settings Service

**Goal:** Implement settings management operations.

**Prerequisites:** Phase 15 complete

**Exit Criteria:**
- All SettingService methods implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 16.1 Mock Handlers for Settings (`mock/handlers_setting.go`)
- [ ] Implement `handleSettingREST(w, r, site, endpoint)`
- [ ] Implement `handleRadiusProfileREST(w, r, site, endpoint)`
- [ ] Implement `handleDynamicDNSREST(w, r, site, endpoint)`
- [ ] Write tests: `mock/handlers_setting_test.go`

#### 16.2 Setting Service Interface (`services/setting.go`)
- [ ] Define `SettingService` interface
- [ ] Implement `settingService` struct
- [ ] Implement base setting methods:
  - [ ] `Get(ctx, site, key) (*types.Setting, error)`
  - [ ] `Update(ctx, site, *types.Setting) (*types.Setting, error)`
- [ ] Implement typed setting accessors (GetMgmt, UpdateMgmt, etc. for all setting types)
- [ ] Implement RADIUS profile methods
- [ ] Implement Dynamic DNS methods
- [ ] Write tests: `services/setting_test.go`

#### 16.3 Client Integration
- [ ] Add `Settings() SettingService` method to client
- [ ] Write integration tests

#### 16.4 Phase 16 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 17: System Service

**Goal:** Implement system-level operations.

**Prerequisites:** Phase 16 complete

**Exit Criteria:**
- All SystemService methods implemented
- All methods tested with mock server
- `make test` passes

### Tasks

#### 17.1 Mock Handlers for System (`mock/handlers_system.go`)
- [ ] Implement `handleStatus(w, r)` (no auth)
- [ ] Implement `handleSelf(w, r)`
- [ ] Implement `handleReboot(w, r)` (requires CSRF + super admin)
- [ ] Implement `handleSpeedTest(w, r, site)`
- [ ] Implement `handleSpeedTestStatus(w, r, site)`
- [ ] Implement `handleBackupList(w, r)`
- [ ] Implement `handleBackupCreate(w, r)`
- [ ] Implement `handleBackupDelete(w, r)`
- [ ] Implement `handleAdminList(w, r)`
- [ ] Write tests: `mock/handlers_system_test.go`

#### 17.2 System Service Interface (`services/system.go`)
- [ ] Define `SystemService` interface
- [ ] Implement `systemService` struct
- [ ] Implement all methods:
  - [ ] `Status(ctx) (*types.Status, error)`
  - [ ] `Self(ctx) (*types.AdminUser, error)`
  - [ ] `Reboot(ctx) error`
  - [ ] `SpeedTest(ctx, site) error`
  - [ ] `SpeedTestStatus(ctx, site) (*types.SpeedTestStatus, error)`
  - [ ] `ListBackups(ctx) ([]types.Backup, error)`
  - [ ] `CreateBackup(ctx) error`
  - [ ] `DeleteBackup(ctx, filename) error`
  - [ ] `DownloadBackup(ctx, filename) (io.ReadCloser, error)`
  - [ ] `ListAdmins(ctx) ([]types.AdminUser, error)`
- [ ] Write tests: `services/system_test.go`

#### 17.3 Client Integration
- [ ] Add `System() SystemService` method to client
- [ ] Write integration tests

#### 17.4 Phase 17 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 18: WebSocket Support

**Goal:** Implement WebSocket event streaming.

**Prerequisites:** Phase 17 complete

**Exit Criteria:**
- WebSocket connection works
- Event streaming works
- `make test` passes

### Tasks

#### 18.1 WebSocket Dependencies
- [ ] Add `github.com/gorilla/websocket` to go.mod
- [ ] Run `go mod tidy`

#### 18.2 Mock WebSocket Handler (`mock/handlers_websocket.go`)
- [ ] Implement `handleWebSocket(w, r)`
- [ ] Implement `(*Server) BroadcastEvent(event)`
- [ ] Implement simulation helpers:
  - [ ] `SimulateClientConnect(site, client)`
  - [ ] `SimulateClientDisconnect(site, mac)`
  - [ ] `SimulateDeviceUpdate(site, device)`
  - [ ] `SimulateAlarm(site, alarm)`
- [ ] Write tests: `mock/handlers_websocket_test.go`

#### 18.3 WebSocket Client (`websocket/client.go`)
- [ ] Define `Client` struct
- [ ] Implement `New(url string, opts ...Option) (*Client, error)`
- [ ] Implement `Connect(ctx context.Context, headers http.Header) error`
- [ ] Implement `Close() error`
- [ ] Implement `ReadMessage() ([]byte, error)`
- [ ] Write tests: `websocket/client_test.go`

#### 18.4 Event Service Interface (`services/events.go`)
- [ ] Define `EventService` interface
- [ ] Define `EventType` constants
- [ ] Define `Event` struct
- [ ] Implement `eventService` struct
- [ ] Implement `Subscribe(ctx, site) (<-chan Event, error)`
- [ ] Implement `SubscribeFiltered(ctx, site, ...EventType) (<-chan Event, error)`
- [ ] Implement `Close() error`
- [ ] Implement reconnection with backoff
- [ ] Write tests: `services/events_test.go`
  - [ ] Test Subscribe success
  - [ ] Test event delivery
  - [ ] Test filtered subscription
  - [ ] Test reconnection
  - [ ] Test context cancellation

#### 18.5 Client Integration
- [ ] Add `Events() EventService` method to client
- [ ] Write integration tests

#### 18.6 Phase 18 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 19: Concurrency & Batch Operations

**Goal:** Implement concurrency utilities and batch operations.

**Prerequisites:** Phase 18 complete

**Exit Criteria:**
- Batch operations work correctly
- Concurrent requests are handled safely
- `make test` passes

### Tasks

#### 19.1 Batch Operations (`batch.go`)
- [ ] Define `BatchResult[T]` struct
- [ ] Implement `BatchGet[T](ctx, client, site, ids, getter) []BatchResult[T]`
- [ ] Implement `BatchCreate[T](ctx, client, site, items, creator) []BatchResult[T]`
- [ ] Implement `BatchDelete(ctx, client, site, ids, deleter) []error`
- [ ] Write tests: `batch_test.go`
  - [ ] Test BatchGet success
  - [ ] Test BatchGet with partial failures
  - [ ] Test BatchCreate success
  - [ ] Test BatchDelete success

#### 19.2 Concurrent Request Handling
- [ ] Implement request semaphore in client
- [ ] Test concurrent request limiting
- [ ] Test no deadlocks under load

#### 19.3 Session Refresh Under Concurrency
- [ ] Test concurrent requests during session refresh
- [ ] Test CSRF token update propagation
- [ ] Test no race conditions (run with -race)

#### 19.4 Phase 19 Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `go test -race ./...` - no races detected
- [ ] Run `make lint` - no errors

---

## Phase 20: Examples & Documentation

**Goal:** Create usage examples and complete documentation.

**Prerequisites:** Phase 19 complete

**Exit Criteria:**
- All examples compile and run
- README is complete
- GoDoc is complete
- `make test` passes

### Tasks

#### 20.1 Basic Example (`examples/basic/main.go`)
- [ ] Demonstrate connection
- [ ] Demonstrate listing devices
- [ ] Demonstrate listing networks
- [ ] Add comments explaining each step

#### 20.2 CRUD Example (`examples/crud/main.go`)
- [ ] Demonstrate creating a network
- [ ] Demonstrate creating a WLAN
- [ ] Demonstrate updating configuration
- [ ] Demonstrate deleting resources

#### 20.3 Concurrent Example (`examples/concurrent/main.go`)
- [ ] Demonstrate batch operations
- [ ] Demonstrate concurrent device management
- [ ] Show proper error handling

#### 20.4 WebSocket Example (`examples/websocket/main.go`)
- [ ] Demonstrate event subscription
- [ ] Demonstrate event filtering
- [ ] Show graceful shutdown

#### 20.5 Error Handling Example (`examples/errors/main.go`)
- [ ] Demonstrate error type checking
- [ ] Demonstrate retry patterns
- [ ] Show reconnection handling

#### 20.6 README.md
- [ ] Write project overview
- [ ] Write installation instructions
- [ ] Write quick start guide
- [ ] Write feature list
- [ ] Write API overview
- [ ] Write contribution guide
- [ ] Add badges (build, coverage, go report)

#### 20.7 GoDoc
- [ ] Review all public type documentation
- [ ] Review all public function documentation
- [ ] Add package-level examples
- [ ] Verify `go doc` output

#### 20.8 Phase 20 Verification
- [ ] Run `go build ./examples/...` - all compile
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no errors

---

## Phase 21: Final Testing & Polish

**Goal:** Comprehensive testing and code quality.

**Prerequisites:** Phase 20 complete

**Exit Criteria:**
- Test coverage > 85%
- No lint errors
- No race conditions
- All examples work

### Tasks

#### 21.1 Test Coverage Analysis
- [ ] Run `make coverage`
- [ ] Identify untested code paths
- [ ] Add tests for any coverage gaps
- [ ] Achieve > 85% coverage

#### 21.2 Integration Test Suite
- [ ] Write comprehensive integration tests
- [ ] Test all error scenarios
- [ ] Test all edge cases

#### 21.3 Race Condition Testing
- [ ] Run all tests with `-race` flag
- [ ] Fix any detected races

#### 21.4 Lint & Code Quality
- [ ] Run `golangci-lint run`
- [ ] Fix all warnings
- [ ] Run `go vet ./...`
- [ ] Run `staticcheck ./...`

#### 21.5 API Compatibility Check
- [ ] Verify all documented endpoints are implemented
- [ ] Verify all types match API documentation
- [ ] Cross-reference with UNIFI_UDM_PRO_API_DOCUMENTATION.md

#### 21.6 Performance Testing
- [ ] Benchmark critical paths
- [ ] Test under concurrent load
- [ ] Profile memory usage

#### 21.7 Final Verification
- [ ] Run `make all` (lint, test, build)
- [ ] All tests pass
- [ ] No lint errors
- [ ] No race conditions
- [ ] Coverage > 85%
- [ ] All examples compile and run

---

## Summary Checklist

### Phase Completion Tracking

| Phase | Name | Status | Tests Pass | Coverage |
|-------|------|--------|------------|----------|
| 0 | Project Scaffolding | [ ] | N/A | N/A |
| 1 | Core Types | [ ] | [ ] | ___% |
| 2 | Internal Utilities | [ ] | [ ] | ___% |
| 3 | Error Handling | [ ] | [ ] | ___% |
| 4 | Transport Layer | [ ] | [ ] | ___% |
| 5 | Authentication | [ ] | [ ] | ___% |
| 6 | Mock Server Foundation | [ ] | [ ] | ___% |
| 7 | Client Core | [ ] | [ ] | ___% |
| 8 | Site Service | [ ] | [ ] | ___% |
| 9 | Device Service | [ ] | [ ] | ___% |
| 10 | Network Service | [ ] | [ ] | ___% |
| 11 | WLAN Service | [ ] | [ ] | ___% |
| 12 | Firewall Service | [ ] | [ ] | ___% |
| 13 | Client Service | [ ] | [ ] | ___% |
| 14 | User Service | [ ] | [ ] | ___% |
| 15 | Routing & Port Services | [ ] | [ ] | ___% |
| 16 | Settings Service | [ ] | [ ] | ___% |
| 17 | System Service | [ ] | [ ] | ___% |
| 18 | WebSocket Support | [ ] | [ ] | ___% |
| 19 | Concurrency & Batch | [ ] | [ ] | ___% |
| 20 | Examples & Documentation | [ ] | [ ] | ___% |
| 21 | Final Testing & Polish | [ ] | [ ] | ___% |

### Final Sign-Off Criteria

- [ ] All 21 phases marked complete
- [ ] Overall test coverage > 85%
- [ ] `make all` passes without errors
- [ ] `go test -race ./...` passes without warnings
- [ ] All examples compile and run successfully
- [ ] README is complete and accurate
- [ ] All public APIs have GoDoc comments

---

## Notes for AI Implementation

1. **Never skip tests** - Every function must have corresponding tests
2. **Use the mock server** - All integration tests use the mock, not real UDM
3. **Run tests frequently** - After implementing each function, run `make test`
4. **Check coverage** - Use `make coverage` to verify no gaps
5. **Follow the order** - Complete phases in sequence, don't jump ahead
6. **Mark progress** - Update checkboxes as tasks complete
7. **Commit often** - Commit after each phase with message "Phase N: [Name] complete"
