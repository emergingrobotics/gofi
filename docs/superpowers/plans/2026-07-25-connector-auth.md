# Cloud-key + Connector Auth Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a second auth mode — a cloud API key (`X-API-KEY`) reaching a UDM through the Site Manager connector proxy — alongside the existing local username/password mode, without changing any `services/` call site.

**Architecture:** All change lives at the transport seam. `transport.Config` gains `APIKey` (injected as `X-API-KEY`) and `PathPrefix` (prepended to every path). `gofi.New()` picks the mode once: `ConsoleID` set → base URL `https://api.ui.com` + prefix `/v1/connector/consoles/{id}` + a no-op API-key auth manager; otherwise the current cookie flow. Services keep building the same local paths (`/proxy/network/...`); the transport rewrites the wire URL.

**Tech Stack:** Go, standard library `net/http`, `httptest`-based mock server, table-driven tests.

## Global Constraints

- Module import path is `github.com/unifi-go/gofi` — do NOT change it (separate task, touches every import).
- Every new function gets a test; every endpoint is exercised through the mock; `make test` must pass (CLAUDE.md rule).
- Use `gofi.NewValidationError(field, message)` for all construction-time validation.
- Do NOT add `SetAPIKey` (or any method) to the `transport.Transport` interface — the key is known at construction.
- API keys are read from the environment only in the CLIs — never add credential flags.
- Verified live 2026-07-25: classic v1 (`rest/networkconf`, `stat/sta`) and v2 (`static-dns`) paths return 200 through the connector on both consoles; site key is `default`.
- Work happens on branch `feat/connector-auth`. Commit after every task.

---

### Task 1: Transport — `X-API-KEY` injection and `PathPrefix`

**Files:**
- Modify: `transport/config.go` (add two fields to `Config`)
- Modify: `transport/transport.go` (store fields on `httpTransport`; use them in `Do`)
- Test: `transport/transport_test.go`

**Interfaces:**
- Produces: `transport.Config{ APIKey string; PathPrefix string }`; `Do()` sends header `X-API-KEY: <APIKey>` (unless a per-request `Headers["X-API-KEY"]` overrides it) and requests the path `PathPrefix + req.Path`.

- [ ] **Step 1: Write the failing tests**

Add to `transport/transport_test.go`:

```go
func TestDo_APIKeyHeader(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-KEY")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := DefaultConfig(srv.URL)
	cfg.APIKey = "secret-key"
	tr, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := tr.Do(context.Background(), NewRequest("GET", "/anything")); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if gotKey != "secret-key" {
		t.Errorf("X-API-KEY = %q, want %q", gotKey, "secret-key")
	}
}

func TestDo_NoAPIKeyHeaderWhenUnset(t *testing.T) {
	var present bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, present = r.Header["X-Api-Key"]
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr, _ := New(DefaultConfig(srv.URL))
	_, _ = tr.Do(context.Background(), NewRequest("GET", "/anything"))
	if present {
		t.Error("X-API-KEY header present but APIKey was not configured")
	}
}

func TestDo_PerRequestAPIKeyOverride(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-KEY")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := DefaultConfig(srv.URL)
	cfg.APIKey = "default-key"
	tr, _ := New(cfg)
	req := NewRequest("GET", "/anything").WithHeader("X-API-KEY", "override-key")
	_, _ = tr.Do(context.Background(), req)
	if gotKey != "override-key" {
		t.Errorf("X-API-KEY = %q, want per-request override %q", gotKey, "override-key")
	}
}

func TestDo_PathPrefix(t *testing.T) {
	cases := []struct {
		name, prefix, reqPath, wantPath string
	}{
		{"no prefix", "", "/proxy/network/x", "/proxy/network/x"},
		{"with prefix", "/v1/connector/consoles/abc", "/proxy/network/x", "/v1/connector/consoles/abc/proxy/network/x"},
		{"trailing slash prefix", "/v1/connector/consoles/abc/", "/proxy/network/x", "/v1/connector/consoles/abc/proxy/network/x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			cfg := DefaultConfig(srv.URL)
			cfg.PathPrefix = tc.prefix
			tr, _ := New(cfg)
			_, _ = tr.Do(context.Background(), NewRequest("GET", tc.reqPath))
			if gotPath != tc.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tc.wantPath)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./transport/ -run 'TestDo_APIKey|TestDo_NoAPIKey|TestDo_PerRequest|TestDo_PathPrefix' -v`
Expected: compile error / FAIL — `Config` has no field `APIKey` or `PathPrefix`.

- [ ] **Step 3: Add the config fields**

In `transport/config.go`, add to the `Config` struct (after `UserAgent`):

```go
	// APIKey, when set, is sent as the X-API-KEY header on every request.
	APIKey string

	// PathPrefix is prepended to every request path. Used to route through
	// the Site Manager connector, e.g. "/v1/connector/consoles/{id}".
	PathPrefix string
```

- [ ] **Step 4: Store and use the fields in the transport**

In `transport/transport.go`, add fields to `httpTransport` (after `userAgent`):

```go
	apiKey     string
	pathPrefix string
```

In `New()`, set them when building `t` (extend the struct literal):

```go
	t := &httpTransport{
		client:     client,
		baseURL:    baseURL,
		userAgent:  config.UserAgent,
		apiKey:     config.APIKey,
		pathPrefix: config.PathPrefix,
	}
```

In `Do()`, replace the URL-building line:

```go
	fullURL, err := t.baseURL.Parse(req.Path)
```

with prefix-aware joining:

```go
	reqPath := req.Path
	if t.pathPrefix != "" {
		reqPath = "/" + strings.Trim(t.pathPrefix, "/") + "/" + strings.TrimLeft(req.Path, "/")
	}
	fullURL, err := t.baseURL.Parse(reqPath)
```

Immediately after the `User-Agent` header block (before the CSRF block), inject the key so a per-request override in the `req.Headers` loop still wins:

```go
	if t.apiKey != "" {
		httpReq.Header.Set("X-API-KEY", t.apiKey)
	}
```

Add `"strings"` to the import block.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./transport/ -v`
Expected: PASS (all new tests plus the existing transport suite).

- [ ] **Step 6: Commit**

```bash
git add transport/config.go transport/transport.go transport/transport_test.go
git commit -m "feat(transport): inject X-API-KEY header and prepend PathPrefix

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: Auth — no-op API-key `Manager`

**Files:**
- Modify: `auth/auth.go` (add `apiKeyManager` + `NewAPIKey`)
- Test: `auth/auth_test.go`

**Interfaces:**
- Produces: `func NewAPIKey() Manager`. Its `Login`/`Logout`/`EnsureAuthenticated` return `nil`; `IsAuthenticated()` returns `true`; `Session()` returns `nil`.

- [ ] **Step 1: Write the failing test**

Add to `auth/auth_test.go`:

```go
func TestNewAPIKeyManager(t *testing.T) {
	m := NewAPIKey()
	ctx := context.Background()
	if err := m.Login(ctx); err != nil {
		t.Errorf("Login() = %v, want nil", err)
	}
	if err := m.Logout(ctx); err != nil {
		t.Errorf("Logout() = %v, want nil", err)
	}
	if err := m.EnsureAuthenticated(ctx); err != nil {
		t.Errorf("EnsureAuthenticated() = %v, want nil", err)
	}
	if !m.IsAuthenticated() {
		t.Error("IsAuthenticated() = false, want true")
	}
	if m.Session() != nil {
		t.Errorf("Session() = %v, want nil", m.Session())
	}
}
```

(If `context` is not already imported in `auth_test.go`, add it.)

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./auth/ -run TestNewAPIKeyManager -v`
Expected: FAIL — `undefined: NewAPIKey`.

- [ ] **Step 3: Implement the manager**

Append to `auth/auth.go`:

```go
// apiKeyManager is a Manager for X-API-KEY auth. There is no session to
// establish, so Login and Logout are no-ops and IsAuthenticated is always true.
type apiKeyManager struct{}

// NewAPIKey returns a Manager for API-key authentication.
func NewAPIKey() Manager { return &apiKeyManager{} }

func (m *apiKeyManager) Login(ctx context.Context) error               { return nil }
func (m *apiKeyManager) Logout(ctx context.Context) error              { return nil }
func (m *apiKeyManager) EnsureAuthenticated(ctx context.Context) error { return nil }
func (m *apiKeyManager) Session() *Session                             { return nil }
func (m *apiKeyManager) IsAuthenticated() bool                         { return true }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./auth/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add auth/auth.go auth/auth_test.go
git commit -m "feat(auth): add no-op NewAPIKey manager

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: Config fields and options

**Files:**
- Modify: `config.go` (add `APIKey`, `ConsoleID`, `BaseURL`)
- Modify: `options.go` (add `WithAPIKey`, `WithConnector`)
- Test: `client_test.go`

**Interfaces:**
- Produces: `gofi.Config{ APIKey string; ConsoleID string; BaseURL string }`; `func WithAPIKey(key string) Option`; `func WithConnector(consoleID string) Option`. `WithConnector` sets `ConsoleID` only; the base-URL/prefix derivation happens in `New()` (Task 4). `BaseURL`, when set, overrides URL derivation (used for connector production default and for pointing tests at the mock).

- [ ] **Step 1: Write the failing test**

Add to `client_test.go`:

```go
func TestOptions_APIKeyAndConnector(t *testing.T) {
	cfg := &Config{}
	WithAPIKey("k123")(cfg)
	WithConnector("console-abc")(cfg)
	if cfg.APIKey != "k123" {
		t.Errorf("APIKey = %q, want k123", cfg.APIKey)
	}
	if cfg.ConsoleID != "console-abc" {
		t.Errorf("ConsoleID = %q, want console-abc", cfg.ConsoleID)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test . -run TestOptions_APIKeyAndConnector -v`
Expected: FAIL — `undefined: WithAPIKey` / unknown field `APIKey`.

- [ ] **Step 3: Add config fields**

In `config.go`, add to the `Config` struct (after `Password`):

```go
	// APIKey authenticates via the X-API-KEY header. When set, Username and
	// Password are ignored and no login request is made. Requires ConsoleID
	// (connector mode).
	APIKey string

	// ConsoleID is the Site Manager console ID. When set, requests are routed
	// through the connector proxy at https://api.ui.com and Host/Port are unused.
	ConsoleID string

	// BaseURL overrides the derived base URL (host:port, or the connector's
	// https://api.ui.com). Mainly for testing against a mock server.
	BaseURL string
```

- [ ] **Step 4: Add options**

Append to `options.go`:

```go
// WithAPIKey authenticates via the X-API-KEY header instead of username/password.
func WithAPIKey(key string) Option {
	return func(c *Config) { c.APIKey = key }
}

// WithConnector routes requests through the Site Manager connector proxy for
// the given console ID. Combine with WithAPIKey.
func WithConnector(consoleID string) Option {
	return func(c *Config) { c.ConsoleID = consoleID }
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test . -run TestOptions_APIKeyAndConnector -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add config.go options.go client_test.go
git commit -m "feat: add APIKey/ConsoleID/BaseURL config fields and options

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: `New()` — mode selection, validation reorder, timeout fix

**Files:**
- Modify: `client_impl.go` (`New()`)
- Test: `client_test.go`

**Interfaces:**
- Consumes: Task 1 (`transport.Config.APIKey`, `.PathPrefix`), Task 2 (`auth.NewAPIKey`), Task 3 (config fields/options).
- Produces: `New()` accepting either `APIKey`+`ConsoleID` or `Username`+`Password`; resolved default `Timeout` is 30s; validation runs after options.

- [ ] **Step 1: Write the failing tests**

Add to `client_test.go`:

```go
func TestNew_APIKeyConnector(t *testing.T) {
	c, err := New(&Config{}, WithAPIKey("k"), WithConnector("abc"))
	if err != nil {
		t.Fatalf("New with API key+connector: %v", err)
	}
	if c == nil {
		t.Fatal("client is nil")
	}
}

func TestNew_NoCredentials(t *testing.T) {
	_, err := New(&Config{Host: "1.2.3.4"})
	if err == nil {
		t.Fatal("expected validation error with no credentials")
	}
}

func TestNew_BothCredentials(t *testing.T) {
	_, err := New(&Config{ConsoleID: "abc", APIKey: "k", Username: "u", Password: "p"})
	if err == nil {
		t.Fatal("expected validation error when both credential sets present")
	}
}

func TestNew_HostAndConsoleMutuallyExclusive(t *testing.T) {
	_, err := New(&Config{Host: "1.2.3.4", APIKey: "k", ConsoleID: "abc"})
	if err == nil {
		t.Fatal("expected error when Host and ConsoleID both set")
	}
}

func TestNew_APIKeyRequiresConnector(t *testing.T) {
	_, err := New(&Config{APIKey: "k"})
	if err == nil {
		t.Fatal("expected error: API key without ConsoleID")
	}
}

func TestNew_DefaultTimeoutIs30s(t *testing.T) {
	cfg := &Config{Host: "1.2.3.4", Username: "u", Password: "p"}
	if _, err := New(cfg); err != nil {
		t.Fatalf("New: %v", err)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("resolved Timeout = %v, want 30s", cfg.Timeout)
	}
}
```

(Ensure `time` is imported in `client_test.go`.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test . -run 'TestNew_' -v`
Expected: FAIL — `TestNew_APIKeyConnector` errors on premature validation; `TestNew_DefaultTimeoutIs30s` sees 900s; the mutual-exclusion/requires-connector cases pass no validation yet.

- [ ] **Step 3: Rewrite `New()`**

Replace the entire `New()` function in `client_impl.go` (from `func New(` through its closing `}`) with:

```go
func New(config *Config, opts ...Option) (Client, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	// Apply options first, so credential/connector options are visible to validation.
	for _, opt := range opts {
		opt(config)
	}

	// Validate.
	hasKey := config.APIKey != ""
	hasUserPass := config.Username != "" && config.Password != ""
	switch {
	case hasKey && hasUserPass:
		return nil, NewValidationError("APIKey", "set either APIKey or Username+Password, not both")
	case !hasKey && !hasUserPass:
		return nil, NewValidationError("APIKey", "one of APIKey (with ConsoleID) or Username+Password is required")
	}
	if config.Host != "" && config.ConsoleID != "" {
		return nil, NewValidationError("ConsoleID", "Host/Port and ConsoleID are mutually exclusive")
	}
	if hasKey && config.ConsoleID == "" && config.BaseURL == "" {
		return nil, NewValidationError("ConsoleID", "APIKey requires ConsoleID (connector mode)")
	}
	if !hasKey && config.Host == "" {
		return nil, NewValidationError("Host", "required")
	}

	// Defaults (guarded so options set above are preserved).
	if config.Port == 0 {
		config.Port = 443
	}
	if config.Site == "" {
		config.Site = "default"
	}
	if config.Timeout == 0 {
		config.Timeout = transport.DefaultConfig("").Timeout
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 10
	}

	// Resolve base URL.
	var baseURL string
	switch {
	case config.BaseURL != "":
		baseURL = config.BaseURL
	case config.ConsoleID != "":
		baseURL = "https://api.ui.com"
	default:
		u := &url.URL{Scheme: "https", Host: net.JoinHostPort(config.Host, strconv.Itoa(config.Port))}
		baseURL = u.String()
	}

	// Transport config.
	transportConfig := transport.DefaultConfig(baseURL)
	transportConfig.Timeout = config.Timeout
	transportConfig.MaxIdleConns = config.MaxIdleConns
	transportConfig.TLSConfig = config.TLSConfig
	transportConfig.APIKey = config.APIKey
	if config.ConsoleID != "" {
		transportConfig.PathPrefix = "/v1/connector/consoles/" + config.ConsoleID
	}

	if config.SkipTLSVerify {
		if transportConfig.TLSConfig == nil {
			transportConfig.TLSConfig = &tls.Config{}
		}
		transportConfig.TLSConfig.InsecureSkipVerify = true
	}

	baseTransport, err := transport.New(transportConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	var trans transport.Transport
	if config.RetryConfig != nil {
		retryConfig := &transport.RetryConfig{
			MaxRetries:     config.RetryConfig.MaxRetries,
			InitialBackoff: config.RetryConfig.InitialBackoff,
			MaxBackoff:     config.RetryConfig.MaxBackoff,
		}
		trans = transport.NewRetryTransport(baseTransport, retryConfig)
	} else {
		trans = baseTransport
	}

	// Select auth manager.
	var authMgr auth.Manager
	if config.APIKey != "" {
		authMgr = auth.NewAPIKey()
	} else {
		authMgr = auth.New(trans, config.Username, config.Password)
	}

	c := &client{
		config:    config,
		transport: trans,
		auth:      authMgr,
		logger:    config.Logger,
	}

	return c, nil
}
```

Leave the rest of `client_impl.go` unchanged.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test . -run 'TestNew_|TestOptions_' -v`
Expected: PASS. Then `go build ./...` to confirm no unused imports (all of `net`, `url`, `strconv`, `tls`, `fmt` are still used).

- [ ] **Step 5: Commit**

```bash
git add client_impl.go client_test.go
git commit -m "feat: select auth mode in New(); fix validation order and default timeout

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: Mock server — key auth, wrong-key 403, connector prefix strip

**Files:**
- Modify: `mock/server.go` (`Server` field; prefix strip + key auth in `ServeHTTP`)
- Modify: `mock/options.go` (`WithAPIKey`)
- Modify: `mock/response.go` (exact-body 403 helper)
- Test: `mock/server_test.go`

**Interfaces:**
- Produces: `mock.WithAPIKey(key string) Option`. A request bearing `X-API-KEY == key` is authenticated without a cookie and skips CSRF; a wrong/absent key (when a key is configured and no cookie) yields `403 {"code":"forbidden","httpStatusCode":403,"message":"insufficient permissions"}`. A `/v1/connector/consoles/{id}` path prefix is stripped before routing.

- [ ] **Step 1: Write the failing tests**

Add to `mock/server_test.go`:

```go
func TestServer_APIKeyAuth(t *testing.T) {
	srv := NewServer(WithAPIKey("good-key"))
	defer srv.Close()
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	// Correct key, no cookie -> authenticated (200).
	req, _ := http.NewRequest("GET", srv.URL()+"/proxy/network/api/s/default/rest/networkconf", nil)
	req.Header.Set("X-API-KEY", "good-key")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("correct key status = %d, want 200", resp.StatusCode)
	}

	// Wrong key -> 403 with exact body.
	req2, _ := http.NewRequest("GET", srv.URL()+"/proxy/network/api/s/default/rest/networkconf", nil)
	req2.Header.Set("X-API-KEY", "bad-key")
	resp2, _ := client.Do(req2)
	if resp2.StatusCode != http.StatusForbidden {
		t.Errorf("wrong key status = %d, want 403", resp2.StatusCode)
	}
	body, _ := io.ReadAll(resp2.Body)
	var parsed map[string]interface{}
	_ = json.Unmarshal(body, &parsed)
	if parsed["code"] != "forbidden" || parsed["message"] != "insufficient permissions" {
		t.Errorf("wrong-key body = %s, want forbidden/insufficient permissions", body)
	}
}

func TestServer_ConnectorPrefixStrip(t *testing.T) {
	srv := NewServer(WithAPIKey("good-key"))
	defer srv.Close()
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	req, _ := http.NewRequest("GET",
		srv.URL()+"/v1/connector/consoles/abc123/proxy/network/api/s/default/rest/networkconf", nil)
	req.Header.Set("X-API-KEY", "good-key")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("connector-prefixed status = %d, want 200 (prefix should be stripped)", resp.StatusCode)
	}
}
```

(Ensure `crypto/tls`, `encoding/json`, `io`, `net/http` are imported in `mock/server_test.go`.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./mock/ -run 'TestServer_APIKey|TestServer_ConnectorPrefix' -v`
Expected: FAIL — `undefined: WithAPIKey`; connector-prefixed path 404s.

- [ ] **Step 3: Add the mock option and server field**

In `mock/server.go`, add to the `Server` struct (after `scenario`):

```go
	apiKey string
```

In `mock/options.go`, append:

```go
// WithAPIKey configures the server to accept X-API-KEY authentication.
func WithAPIKey(key string) Option {
	return func(s *Server) { s.apiKey = key }
}
```

- [ ] **Step 4: Strip connector prefix and add key auth in `ServeHTTP`**

In `mock/server.go` `ServeHTTP`, immediately after the scenario block and before `path := r.URL.Path`, insert:

```go
	// Strip a Site Manager connector prefix, if present, so routing below is unchanged.
	if strings.HasPrefix(r.URL.Path, "/v1/connector/consoles/") {
		rest := strings.TrimPrefix(r.URL.Path, "/v1/connector/consoles/")
		if i := strings.Index(rest, "/"); i >= 0 {
			r.URL.Path = rest[i:]
		}
	}
```

Replace the existing auth+CSRF block:

```go
	// All other endpoints require authentication
	if s.requireAuth {
		if !s.isAuthenticated(r) {
			writeUnauthorized(w)
			return
		}
	}

	// Check CSRF token for non-GET requests
	if s.requireCSRF && r.Method != "GET" && r.Method != "HEAD" {
		if !s.validateCSRF(r) {
			writeForbidden(w, "Invalid CSRF token")
			return
		}
	}
```

with key-aware auth:

```go
	// All other endpoints require authentication.
	if s.requireAuth {
		if apiKey := r.Header.Get("X-API-KEY"); apiKey != "" {
			// Key-authenticated: validate the key, skip cookie and CSRF checks.
			if s.apiKey == "" || apiKey != s.apiKey {
				writeInsufficientPermissions(w)
				return
			}
		} else {
			if !s.isAuthenticated(r) {
				writeUnauthorized(w)
				return
			}
			if s.requireCSRF && r.Method != "GET" && r.Method != "HEAD" {
				if !s.validateCSRF(r) {
					writeForbidden(w, "Invalid CSRF token")
					return
				}
			}
		}
	}
```

- [ ] **Step 5: Add the exact-body 403 helper**

Append to `mock/response.go`:

```go
// writeInsufficientPermissions writes the exact 403 body a real console returns
// for a bad API key.
func writeInsufficientPermissions(w http.ResponseWriter) {
	writeJSON(w, http.StatusForbidden, map[string]interface{}{
		"code":           "forbidden",
		"httpStatusCode": 403,
		"message":        "insufficient permissions",
	})
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./mock/ -v`
Expected: PASS (new tests plus the existing mock suite — the cookie/CSRF path is unchanged when no `X-API-KEY` is present).

- [ ] **Step 7: Commit**

```bash
git add mock/server.go mock/options.go mock/response.go mock/server_test.go
git commit -m "feat(mock): X-API-KEY auth, connector prefix strip, exact 403 body

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: End-to-end — connector-mode client + `Connect()` fail-fast probe

**Files:**
- Modify: `client_impl.go` (`Connect()` adds a key-mode probe)
- Test: `client_integration_test.go`

**Interfaces:**
- Consumes: Tasks 4 and 5.
- Produces: in key mode, `Connect()` issues one GET to `/proxy/network/api/s/{site}/rest/networkconf` and returns an error if it is 401/403; a full service call succeeds through the connector-prefixed mock.

- [ ] **Step 1: Write the failing tests**

Add to `client_integration_test.go`:

```go
func TestConnectorMode_EndToEnd(t *testing.T) {
	srv := mock.NewServer(mock.WithAPIKey("test-key"))
	defer srv.Close()

	c, err := New(&Config{
		APIKey:        "test-key",
		ConsoleID:     "console-xyz",
		BaseURL:       srv.URL(),
		SkipTLSVerify: true,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect (connector mode): %v", err)
	}
	if !c.IsConnected() {
		t.Error("IsConnected() = false after connecting in key mode")
	}
	if _, err := c.Networks().List(ctx, "default"); err != nil {
		t.Errorf("Networks().List through connector: %v", err)
	}
}

func TestConnectorMode_BadKeyFailsConnect(t *testing.T) {
	srv := mock.NewServer(mock.WithAPIKey("good-key"))
	defer srv.Close()

	c, _ := New(&Config{
		APIKey:        "wrong-key",
		ConsoleID:     "console-xyz",
		BaseURL:       srv.URL(),
		SkipTLSVerify: true,
	})
	if err := c.Connect(context.Background()); err == nil {
		t.Fatal("expected Connect to fail with a bad API key")
	}
}
```

(Confirm the `mock` import and `Networks().List` signature match the existing integration tests; adjust the service call to whatever the existing tests use if `List` differs.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test . -run 'TestConnectorMode' -v`
Expected: `TestConnectorMode_BadKeyFailsConnect` FAILS — `Connect()` currently no-ops in key mode and never probes, so a bad key is not caught.

- [ ] **Step 3: Add the key-mode probe to `Connect()`**

In `client_impl.go` `Connect()`, after the `c.auth.Login(ctx)` block and before `c.connected.Store(true)`, insert:

```go
	// In key mode there is no login; issue one cheap read so a bad key or an
	// unreachable console fails here rather than on the first real call.
	if c.config.APIKey != "" {
		probe := transport.NewRequest("GET",
			fmt.Sprintf("/proxy/network/api/s/%s/rest/networkconf", c.config.Site))
		resp, err := c.transport.Do(ctx, probe)
		if err != nil {
			return fmt.Errorf("connect probe failed: %w", err)
		}
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return fmt.Errorf("authentication failed: status %d (check API key)", resp.StatusCode)
		}
	}
```

(`transport` and `fmt` are already imported in `client_impl.go`.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test . -run 'TestConnectorMode' -v`
Expected: PASS. Then `make test` for the whole suite.

- [ ] **Step 5: Commit**

```bash
git add client_impl.go client_integration_test.go
git commit -m "feat: connector end-to-end path + fail-fast key probe in Connect()

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 7: Shared CLI connection-resolution helper

**Files:**
- Create: `utilities/internal/conn/conn.go`
- Test: `utilities/internal/conn/conn_test.go`

**Interfaces:**
- Consumes: Tasks 3–4 (`gofi.Config`).
- Produces: package `conn` with exported env-name constants and
  `func ResolveConfig(w io.Writer, host string, port int, site string, secure bool) (*gofi.Config, error)`.
  It reads the environment, returns a ready `*gofi.Config` (connector mode when `UNIFI_API_KEY`+`UNIFI_CONSOLE_ID` are set, else local username/password), and writes a one-line note to `w` when a key overrides a present username/password. Import path: `github.com/unifi-go/gofi/utilities/internal/conn`.

- [ ] **Step 1: Write the failing tests**

Create `utilities/internal/conn/conn_test.go`:

```go
package conn

import (
	"io"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{EnvUsername, EnvPassword, EnvControllerIP, EnvAPIKey, EnvConsoleID} {
		t.Setenv(k, "")
	}
}

func TestResolveConfig_Connector(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvAPIKey, "k123")
	t.Setenv(EnvConsoleID, "console-abc")
	cfg, err := ResolveConfig(io.Discard, "", 443, "default", false)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.APIKey != "k123" || cfg.ConsoleID != "console-abc" {
		t.Errorf("got APIKey=%q ConsoleID=%q", cfg.APIKey, cfg.ConsoleID)
	}
	if cfg.Host != "" {
		t.Errorf("Host = %q, want empty in connector mode", cfg.Host)
	}
	if !cfg.SkipTLSVerify {
		t.Error("SkipTLSVerify = false, want true when secure=false")
	}
}

func TestResolveConfig_APIKeyRequiresConsole(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvAPIKey, "k123")
	if _, err := ResolveConfig(io.Discard, "", 443, "default", false); err == nil {
		t.Fatal("expected error: API key without console ID")
	}
}

func TestResolveConfig_Local(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvUsername, "admin")
	t.Setenv(EnvPassword, "pw")
	cfg, err := ResolveConfig(io.Discard, "10.0.0.1", 8443, "default", true)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Host != "10.0.0.1" || cfg.Port != 8443 || cfg.Username != "admin" {
		t.Errorf("got %+v", cfg)
	}
	if cfg.SkipTLSVerify {
		t.Error("SkipTLSVerify = true, want false when secure=true")
	}
}

func TestResolveConfig_LocalHostFromEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvUsername, "admin")
	t.Setenv(EnvPassword, "pw")
	t.Setenv(EnvControllerIP, "192.168.9.9")
	cfg, err := ResolveConfig(io.Discard, "", 443, "default", false)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Host != "192.168.9.9" {
		t.Errorf("Host = %q, want fallback from env", cfg.Host)
	}
}

func TestResolveConfig_NoCredentials(t *testing.T) {
	clearEnv(t)
	if _, err := ResolveConfig(io.Discard, "1.2.3.4", 443, "default", false); err == nil {
		t.Fatal("expected error when no credentials are set")
	}
}

func TestResolveConfig_KeyOverridesUserPassWithNote(t *testing.T) {
	clearEnv(t)
	t.Setenv(EnvAPIKey, "k123")
	t.Setenv(EnvConsoleID, "console-abc")
	t.Setenv(EnvUsername, "admin")
	t.Setenv(EnvPassword, "pw")
	var buf strings.Builder
	cfg, err := ResolveConfig(&buf, "", 443, "default", false)
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.APIKey != "k123" {
		t.Error("expected API key to win over username/password")
	}
	if !strings.Contains(buf.String(), EnvAPIKey) {
		t.Errorf("expected override note mentioning %s, got %q", EnvAPIKey, buf.String())
	}
}
```

Add `"strings"` to the test imports.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./utilities/internal/conn/ -v`
Expected: compile error — package `conn` / `ResolveConfig` / `Env*` do not exist yet.

- [ ] **Step 3: Implement the helper**

Create `utilities/internal/conn/conn.go`:

```go
// Package conn resolves UniFi connection configuration from flags and the
// environment, shared by the gofips/gofimac/gofinet CLIs.
package conn

import (
	"fmt"
	"io"
	"os"

	"github.com/unifi-go/gofi"
)

// Environment variable names used by all three CLIs.
const (
	EnvUsername     = "UNIFI_USERNAME"
	EnvPassword     = "UNIFI_PASSWORD"
	EnvControllerIP = "UNIFI_CONTROLLER_IP"
	EnvAPIKey       = "UNIFI_API_KEY"
	EnvConsoleID    = "UNIFI_CONSOLE_ID"
)

// ResolveConfig builds a *gofi.Config from the given flag values and the
// environment. When UNIFI_API_KEY is set it uses connector mode (requiring
// UNIFI_CONSOLE_ID) and ignores username/password, writing a note to w if
// those were also set. Otherwise it uses local username/password auth,
// falling back to UNIFI_CONTROLLER_IP for the host.
func ResolveConfig(w io.Writer, host string, port int, site string, secure bool) (*gofi.Config, error) {
	apiKey := os.Getenv(EnvAPIKey)
	consoleID := os.Getenv(EnvConsoleID)

	if apiKey != "" {
		if os.Getenv(EnvUsername) != "" || os.Getenv(EnvPassword) != "" {
			fmt.Fprintf(w, "Note: %s is set; ignoring %s/%s.\n", EnvAPIKey, EnvUsername, EnvPassword)
		}
		if consoleID == "" {
			return nil, fmt.Errorf("%s is required when %s is set (connector mode)", EnvConsoleID, EnvAPIKey)
		}
		return &gofi.Config{
			APIKey:        apiKey,
			ConsoleID:     consoleID,
			Site:          site,
			SkipTLSVerify: !secure,
		}, nil
	}

	username := os.Getenv(EnvUsername)
	password := os.Getenv(EnvPassword)
	if username == "" && password == "" {
		return nil, fmt.Errorf("no credentials: set %s (+ %s), or %s/%s", EnvAPIKey, EnvConsoleID, EnvUsername, EnvPassword)
	}
	if username == "" {
		return nil, fmt.Errorf("%s environment variable is required", EnvUsername)
	}
	if password == "" {
		return nil, fmt.Errorf("%s environment variable is required", EnvPassword)
	}
	if host == "" {
		host = os.Getenv(EnvControllerIP)
	}
	if host == "" {
		return nil, fmt.Errorf("--host is required (or set %s)", EnvControllerIP)
	}
	return &gofi.Config{
		Host:          host,
		Port:          port,
		Username:      username,
		Password:      password,
		Site:          site,
		SkipTLSVerify: !secure,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./utilities/internal/conn/ -v`
Expected: PASS (all six tests).

- [ ] **Step 5: Commit**

```bash
git add utilities/internal/conn/conn.go utilities/internal/conn/conn_test.go
git commit -m "feat(cli): shared connection-resolution helper (key/connector/local)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 8: Wire the shared helper into gofips, gofimac, gofinet

**Files:**
- Modify: `utilities/gofinet/main.go`, `utilities/gofimac/main.go`, `utilities/gofips/main.go`

**Interfaces:**
- Consumes: Task 7 (`conn.ResolveConfig`, `conn.Env*`), Tasks 4–6.
- Produces: each tool resolves its config via `conn.ResolveConfig` and connects in whichever mode the environment selects.

- [ ] **Step 1: Import the helper and drop the local env consts**

In each `main.go`, add the import `"github.com/unifi-go/gofi/utilities/internal/conn"`. Delete the tool's local `const ( envUsername = ... envPassword = ... envControllerIP = ... )` block — those names now live in `conn`.

- [ ] **Step 2: Replace the credential/host-resolution block with a call to the helper**

In each tool, replace the block that resolves the host and requires username/password (in `gofinet/main.go` that is lines 53–77, from `if *host == ""` through the `config := &gofi.Config{...}` literal, including the preceding `fmt.Fprintf(os.Stderr, "Connecting to %s...\n", *host)` status line) with:

```go
	config, err := conn.ResolveConfig(os.Stderr, *host, *port, *site, *secure)
	if err != nil {
		exitError(err.Error())
	}

	if config.ConsoleID != "" {
		fmt.Fprintf(os.Stderr, "Connecting via connector to console %s...\n", config.ConsoleID)
	} else {
		fmt.Fprintf(os.Stderr, "Connecting to %s...\n", config.Host)
	}
```

Then update the existing client-construction line to reuse `err` with `=` (not `:=`) if the compiler reports `err` already declared:

```go
	apiClient, err := gofi.New(config)
```

becomes

```go
	apiClient, err = gofi.New(config)
```

(Only if `err` is now in scope from `ResolveConfig`. If the tool declared `err` first at the `gofi.New` line, keep `:=`.)

- [ ] **Step 3: Update each `flag.Usage` "Environment Variables" section**

In each tool's usage function, replace the environment-variable lines so `UNIFI_API_KEY` leads, referencing the `conn` constants:

```go
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s\tAPI key (preferred; requires %s)\n", conn.EnvAPIKey, conn.EnvConsoleID)
		fmt.Fprintf(os.Stderr, "  %s\tSite Manager console ID (connector mode)\n", conn.EnvConsoleID)
		fmt.Fprintf(os.Stderr, "  %s\tUsername (required if no API key)\n", conn.EnvUsername)
		fmt.Fprintf(os.Stderr, "  %s\tPassword (required if no API key)\n", conn.EnvPassword)
		fmt.Fprintf(os.Stderr, "  %s\tController host (fallback for -H)\n\n", conn.EnvControllerIP)
```

Any other references to the old bare `envControllerIP`/`envUsername`/`envPassword` constants elsewhere in the file (e.g. the `-H` flag help text) must be changed to the `conn.Env*` form.

- [ ] **Step 4: Build and smoke-check**

Run: `go build ./utilities/...`
Expected: builds clean.

Run: `env -u UNIFI_USERNAME -u UNIFI_PASSWORD UNIFI_API_KEY=x go run ./utilities/gofinet 2>&1 | head -1`
Expected: `Error: UNIFI_CONSOLE_ID is required when UNIFI_API_KEY is set (connector mode)`

- [ ] **Step 5: Commit**

```bash
git add utilities/gofips/main.go utilities/gofimac/main.go utilities/gofinet/main.go
git commit -m "feat(cli): resolve config via shared conn helper in all three tools

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 9: Documentation

**Files:**
- Modify: `README.md`, `CLAUDE.md`, `doc.go`, `EXAMPLES.md` (and `examples/*/main.go` credential setup where present)

**Interfaces:** none (docs only).

- [ ] **Step 1: README quickstart**

Lead the quickstart with the API-key path; keep username/password as a documented fallback. Add a "Generating an API key" subsection:

```markdown
## Authentication

gofi supports two modes:

### Cloud API key via the Site Manager connector (recommended)

```bash
export UNIFI_API_KEY=...        # from unifi.ui.com; needs UniFi Applications -> Network scope
export UNIFI_CONSOLE_ID=...     # from GET https://api.ui.com/v1/hosts
```

```go
client, _ := gofi.New(&gofi.Config{}, gofi.WithAPIKey(os.Getenv("UNIFI_API_KEY")),
	gofi.WithConnector(os.Getenv("UNIFI_CONSOLE_ID")))
```

A key inherits the permissions of the account that created it — use Site Admin or Owner.
Console-issued keys (console UI -> profile icon -> API -> Create API Key) and cloud keys
(unifi.ui.com) are different credentials; connector access needs a cloud key.

### Local username/password (fallback)

```bash
export UNIFI_USERNAME=... UNIFI_PASSWORD=...
```

```go
client, _ := gofi.New(&gofi.Config{Host: "192.168.1.1", Username: u, Password: p})
```
```

- [ ] **Step 2: CLAUDE.md**

In "Key Technical Details", replace the single cookie-auth bullet with both mechanisms and note that key auth reaches classic v1/v2 paths through the connector (verified), while username/password uses the local cookie/CSRF flow.

- [ ] **Step 3: doc.go and EXAMPLES.md**

Switch the package example in `doc.go` and the credential setup in `EXAMPLES.md` (and any `examples/*/main.go`) to the `WithAPIKey`/`WithConnector` form, keeping one username/password example labelled as the fallback.

- [ ] **Step 4: Commit**

```bash
git add README.md CLAUDE.md doc.go EXAMPLES.md examples/
git commit -m "docs: document API-key + connector auth as the primary path

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Notes for the executor

- After Task 2, grep for callers that assume `IsAuthenticated() ⇒ non-nil Session()`:
  `grep -rn "IsAuthenticated" --include='*.go' .` The only production caller is
  `client.IsConnected()` (which never dereferences `Session()`); if any other caller
  dereferences `Session()` guarded only by `IsAuthenticated()`, fix it to nil-check.
  Task 6's `TestConnectorMode_EndToEnd` exercises `IsConnected()` in key mode.
- `client_test.go` may need `time` and `client_integration_test.go` may need `context`/`mock` imports — add them if the compiler complains.
- If the existing integration tests call the network service differently than `Networks().List(ctx, "default")`, mirror their exact call in Task 6.
- Do NOT touch `go.mod`'s module path or `client.Events()` — both are tracked as separate out-of-scope items in the design spec.
