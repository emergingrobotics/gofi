# UniFi UDM Pro API Documentation (Version 10+)

## Table of Contents

1. [Overview](#overview)
2. [API Types](#api-types)
3. [Authentication](#authentication)
4. [Base URLs and Path Prefixes](#base-urls-and-path-prefixes)
5. [Request/Response Format](#requestresponse-format)
6. [Error Handling](#error-handling)
7. [Rate Limiting](#rate-limiting)
8. [Local Controller API Endpoints](#local-controller-api-endpoints)
9. [Site Manager (Cloud) API Endpoints](#site-manager-cloud-api-endpoints)
10. [WebSocket API](#websocket-api)
11. [Go Implementation Guide](#go-implementation-guide)
12. [Reference Implementations](#reference-implementations)
13. [References](#references)

---

## Overview

The Ubiquiti UniFi UDM Pro provides two primary API access methods:

1. **Local Controller API** - Direct access to the device's UniFi Network Application
2. **Site Manager API** - Cloud-based access via `api.ui.com` for managing multiple sites

This documentation focuses on UniFi OS version 4.x/5.x and UniFi Network Application version 10.x, though most endpoints are backward compatible with version 8.x+.

### Key Differences from Legacy Controllers

| Aspect | Legacy Controller | UDM Pro / UniFi OS |
|--------|-------------------|---------------------|
| Port | 8443 | 443 |
| Login Endpoint | `/api/login` | `/api/auth/login` |
| API Prefix | None | `/proxy/network` |
| CSRF Token | Not required | Required for some operations |
| Local Admin | Optional | **Required** for API access (MFA bypass) |

---

## API Types

### Local Controller API (v1)
- **Base Path**: `/proxy/network/api/s/{site}/`
- **Purpose**: Full device and network management
- **Access**: Cookie-based session authentication
- **Capabilities**: Read/Write

### Local Controller API (v2)
- **Base Path**: `/proxy/network/v2/api/site/{site}/`
- **Purpose**: Newer endpoints for traffic rules, notifications
- **Access**: Cookie-based session authentication
- **Capabilities**: Read/Write

### Integrations API (v1)
- **Base Path**: `/proxy/network/integrations/v1/`
- **Purpose**: Integration-specific endpoints
- **Access**: Cookie-based session authentication
- **Capabilities**: Read/Write

### Site Manager API (Cloud)
- **Base URL**: `https://api.ui.com/v1/`
- **Purpose**: Multi-site management via cloud
- **Access**: API Key authentication
- **Capabilities**: Read-only (currently)

---

## Authentication

### Method 1: Local Admin Session (Recommended for Local API)

Local admin accounts bypass UI.com MFA requirements and are the recommended approach for API automation.

#### Creating a Local Admin Account

1. In UniFi Network: Navigate to **Settings > Admins**
2. Click **Create New Admin**
3. Select **Local Access Only** (disable Remote/Cloud access)
4. Assign appropriate role (Site Admin or View Only)
5. Save credentials securely

#### Login Request

```http
POST /api/auth/login HTTP/1.1
Host: {udm-ip}
Content-Type: application/json
Accept: application/json

{
    "username": "local_admin",
    "password": "secure_password",
    "rememberMe": true,
    "token": ""
}
```

#### Login Response

**Success (HTTP 200)**:
```json
{
    "unique_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "first_name": "Admin",
    "last_name": "User",
    "full_name": "Admin User",
    "email": "admin@local",
    "email_status": "UNVERIFIED",
    "status": "ACTIVE",
    "sso_account": "",
    "sso_uuid": "",
    "sso_username": "",
    "sso_picture": "",
    "uid_sso_id": "",
    "uid_sso_account": "",
    "username": "local_admin",
    "local_account_exist": true,
    "groups": [],
    "roles": [
        {
            "name": "Super Admin",
            "system_role": true,
            "unique_id": "role-id-here",
            "system_key": "super_admin"
        }
    ],
    "permissions": {},
    "scopes": ["*:*:*"],
    "cloud_access_granted": false,
    "update_time": "2025-01-15T10:00:00Z",
    "avatar_relative_path": "",
    "avatar": "",
    "nfc_cards": [],
    "id": "user-id-here",
    "isOwner": true,
    "isSuperAdmin": true
}
```

#### Important Headers from Login Response

| Header | Description |
|--------|-------------|
| `Set-Cookie` | Session cookie (TOKEN) - must be included in all subsequent requests |
| `X-CSRF-Token` | CSRF token - required for certain write operations |

#### Go Implementation - Login

```go
package unifi

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/cookiejar"
)

type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    CSRFToken  string
    Site       string
}

type LoginRequest struct {
    Username   string `json:"username"`
    Password   string `json:"password"`
    RememberMe bool   `json:"rememberMe"`
    Token      string `json:"token"`
}

type LoginResponse struct {
    UniqueID   string `json:"unique_id"`
    Username   string `json:"username"`
    FullName   string `json:"full_name"`
    IsSuperAdmin bool `json:"isSuperAdmin"`
}

func NewClient(baseURL, site string, skipTLSVerify bool) (*Client, error) {
    jar, err := cookiejar.New(nil)
    if err != nil {
        return nil, err
    }

    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: skipTLSVerify,
        },
    }

    return &Client{
        BaseURL: baseURL,
        Site:    site,
        HTTPClient: &http.Client{
            Jar:       jar,
            Transport: transport,
        },
    }, nil
}

func (c *Client) Login(username, password string) (*LoginResponse, error) {
    loginReq := LoginRequest{
        Username:   username,
        Password:   password,
        RememberMe: true,
        Token:      "",
    }

    body, err := json.Marshal(loginReq)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", c.BaseURL+"/api/auth/login", bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("login failed with status: %d", resp.StatusCode)
    }

    // Store CSRF token for later use
    if csrfToken := resp.Header.Get("X-CSRF-Token"); csrfToken != "" {
        c.CSRFToken = csrfToken
    }

    var loginResp LoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
        return nil, err
    }

    return &loginResp, nil
}
```

### Method 2: API Key (Site Manager / Cloud API)

#### Obtaining an API Key

1. Sign in to **UniFi Site Manager** at `https://unifi.ui.com`
2. Navigate to **Settings > API Keys** (or **API** section in GA)
3. Click **Create New API Key**
4. Copy and store the key securely (shown only once)

#### Using the API Key

```http
GET /v1/hosts HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key_here
Accept: application/json
```

#### Go Implementation - API Key Auth

```go
package unifi

type CloudClient struct {
    APIKey     string
    HTTPClient *http.Client
}

func NewCloudClient(apiKey string) *CloudClient {
    return &CloudClient{
        APIKey:     apiKey,
        HTTPClient: &http.Client{},
    }
}

func (c *CloudClient) doRequest(method, endpoint string, body []byte) (*http.Response, error) {
    req, err := http.NewRequest(method, "https://api.ui.com"+endpoint, bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }

    req.Header.Set("X-API-KEY", c.APIKey)
    req.Header.Set("Accept", "application/json")
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    return c.HTTPClient.Do(req)
}
```

---

## Base URLs and Path Prefixes

### Local API URL Construction

```
https://{udm-ip}:443/proxy/network/api/s/{site}/{endpoint}
```

| Component | Description | Example |
|-----------|-------------|---------|
| `{udm-ip}` | IP address or hostname of UDM Pro | `192.168.1.1` |
| `/proxy/network` | **Required** prefix for all Network API calls | - |
| `/api/s/` | API path prefix | - |
| `{site}` | Site identifier (usually `default`) | `default` |
| `{endpoint}` | Specific API endpoint | `stat/device` |

### API Version Prefixes

| API Version | Path Pattern |
|-------------|--------------|
| v1 (Legacy) | `/proxy/network/api/s/{site}/{endpoint}` |
| v2 (Modern) | `/proxy/network/v2/api/site/{site}/{endpoint}` |
| Integrations | `/proxy/network/integrations/v1/{endpoint}` |

### Site Identifier

The site identifier is typically:
- `default` - For single-site installations
- 8-character alphanumeric string - For multi-site controllers

Retrieve available sites via:
```http
GET /api/self/sites
```

---

## Request/Response Format

### Standard Response Structure

All API responses follow this JSON structure:

```json
{
    "meta": {
        "rc": "ok",
        "msg": "optional message"
    },
    "data": [
        { /* object 1 */ },
        { /* object 2 */ }
    ]
}
```

### Response Codes

| `meta.rc` | Meaning |
|-----------|---------|
| `ok` | Success |
| `error` | Error occurred (see `meta.msg`) |

### Common Error Messages

| `meta.msg` | Description |
|------------|-------------|
| `api.err.LoginRequired` | Session expired or not authenticated |
| `api.err.NoPermission` | Insufficient permissions |
| `api.err.Invalid` | Invalid request parameters |
| `api.err.InvalidObject` | Object not found |
| `api.err.AlreadyExists` | Duplicate object |

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created (some PUT operations) |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden / Invalid CSRF Token |
| 404 | Not Found |
| 429 | Rate Limited |
| 500 | Internal Server Error |

### Go Implementation - Response Handling

```go
package unifi

type APIResponse[T any] struct {
    Meta struct {
        RC      string `json:"rc"`
        Message string `json:"msg,omitempty"`
        Count   int    `json:"count,omitempty"`
    } `json:"meta"`
    Data []T `json:"data"`
}

type ErrorResponse struct {
    Meta struct {
        RC      string `json:"rc"`
        Message string `json:"msg"`
    } `json:"meta"`
    Data []interface{} `json:"data"`
}

func (c *Client) doAPIRequest(method, endpoint string, body interface{}) (*http.Response, error) {
    var reqBody []byte
    var err error

    if body != nil {
        reqBody, err = json.Marshal(body)
        if err != nil {
            return nil, err
        }
    }

    url := fmt.Sprintf("%s/proxy/network/api/s/%s/%s", c.BaseURL, c.Site, endpoint)
    req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    // Include CSRF token for write operations
    if c.CSRFToken != "" && (method == "POST" || method == "PUT" || method == "DELETE") {
        req.Header.Set("X-CSRF-Token", c.CSRFToken)
    }

    return c.HTTPClient.Do(req)
}
```

---

## Error Handling

### CSRF Token Errors

For write operations on UDM Pro, you may receive:

```json
{
    "message": "Invalid CSRF Token"
}
```

**Solution**: Include the `X-CSRF-Token` header obtained from the login response.

### Session Expiration

```json
{
    "meta": {
        "rc": "error",
        "msg": "api.err.LoginRequired"
    },
    "data": []
}
```

**Solution**: Re-authenticate and retry the request.

### Go Implementation - Error Types

```go
package unifi

import "errors"

var (
    ErrLoginRequired    = errors.New("authentication required")
    ErrNoPermission     = errors.New("insufficient permissions")
    ErrInvalidCSRFToken = errors.New("invalid CSRF token")
    ErrNotFound         = errors.New("resource not found")
    ErrRateLimited      = errors.New("rate limited")
)

func parseAPIError(resp *http.Response, body []byte) error {
    switch resp.StatusCode {
    case 401:
        return ErrLoginRequired
    case 403:
        if bytes.Contains(body, []byte("CSRF")) {
            return ErrInvalidCSRFToken
        }
        return ErrNoPermission
    case 404:
        return ErrNotFound
    case 429:
        return ErrRateLimited
    default:
        return fmt.Errorf("API error: %s", string(body))
    }
}
```

---

## Rate Limiting

### Site Manager API (Cloud)

| Environment | Rate Limit |
|-------------|------------|
| Early Access (EA) | 100 requests/minute |
| v1 Stable | 10,000 requests/minute |

When rate limited, you'll receive:
- HTTP Status: `429 Too Many Requests`
- Header: `Retry-After: {seconds}`

### Local API

The local API does not have documented rate limits, but aggressive polling can impact controller performance. Recommended practices:
- Polling interval: 30+ seconds for stats
- Use WebSocket API for real-time events
- Batch requests where possible

### Go Implementation - Rate Limiter

```go
package unifi

import (
    "sync"
    "time"
)

type RateLimiter struct {
    requests  int
    limit     int
    window    time.Duration
    lastReset time.Time
    mu        sync.Mutex
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        limit:     limit,
        window:    window,
        lastReset: time.Now(),
    }
}

func (r *RateLimiter) Wait() {
    r.mu.Lock()
    defer r.mu.Unlock()

    now := time.Now()
    if now.Sub(r.lastReset) >= r.window {
        r.requests = 0
        r.lastReset = now
    }

    if r.requests >= r.limit {
        sleepDuration := r.window - now.Sub(r.lastReset)
        time.Sleep(sleepDuration)
        r.requests = 0
        r.lastReset = time.Now()
    }

    r.requests++
}
```

---

## Local Controller API Endpoints

### Controller-Level Endpoints (No Site Context)

These endpoints don't require a site parameter:

#### GET /status (No Auth Required)

Returns basic server information.

```http
GET /status HTTP/1.1
Host: {udm-ip}
```

**Response**:
```json
{
    "meta": {
        "rc": "ok"
    },
    "data": []
}
```

---

#### GET /api/self

Returns current authenticated user information.

```http
GET /proxy/network/api/self HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [{
        "name": "admin",
        "email": "admin@example.com",
        "is_super": true,
        "site_role": "admin"
    }]
}
```

---

#### GET /api/self/sites

Returns all sites accessible to the current user.

```http
GET /proxy/network/api/self/sites HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "site_id_here",
            "name": "default",
            "desc": "Default",
            "role": "admin",
            "role_hotspot": false,
            "attr_hidden_id": "default",
            "attr_no_delete": true,
            "health": [
                {"subsystem": "wlan", "status": "ok", "num_ap": 2},
                {"subsystem": "wan", "status": "ok"},
                {"subsystem": "lan", "status": "ok"}
            ]
        }
    ]
}
```

---

#### POST /api/logout

Terminates the current session.

```http
POST /proxy/network/api/logout HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

---

#### POST /api/system/reboot

Reboots the controller. **Requires X-CSRF-Token header and Super Admin rights.**

```http
POST /api/system/reboot HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json
```

---

### Site-Level Statistics Endpoints

Base path: `/proxy/network/api/s/{site}/`

#### GET stat/health

Returns site health status for all subsystems.

```http
GET /proxy/network/api/s/default/stat/health HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "subsystem": "wlan",
            "status": "ok",
            "num_ap": 2,
            "num_adopted": 2,
            "num_pending": 0,
            "num_user": 15,
            "num_guest": 3,
            "rx_bytes-r": 1234567,
            "tx_bytes-r": 2345678
        },
        {
            "subsystem": "wan",
            "status": "ok",
            "wan_ip": "203.0.113.1",
            "gateways": ["192.168.1.1"],
            "latency": 15,
            "uptime": 1234567,
            "drops": 0,
            "xput_down": 100.5,
            "xput_up": 20.3,
            "speedtest_status": "Idle",
            "speedtest_lastrun": 1642345678,
            "speedtest_ping": 15
        },
        {
            "subsystem": "lan",
            "status": "ok",
            "num_sw": 3,
            "num_adopted": 3,
            "num_pending": 0,
            "num_user": 45
        },
        {
            "subsystem": "vpn",
            "status": "ok",
            "remote_user_num_active": 0,
            "remote_user_num_inactive": 2,
            "site_to_site_num_active": 1,
            "site_to_site_num_inactive": 0
        }
    ]
}
```

**Go Struct**:
```go
type HealthData struct {
    Subsystem          string  `json:"subsystem"`
    Status             string  `json:"status"`
    NumAP              int     `json:"num_ap,omitempty"`
    NumAdopted         int     `json:"num_adopted,omitempty"`
    NumPending         int     `json:"num_pending,omitempty"`
    NumUser            int     `json:"num_user,omitempty"`
    NumGuest           int     `json:"num_guest,omitempty"`
    NumSW              int     `json:"num_sw,omitempty"`
    WanIP              string  `json:"wan_ip,omitempty"`
    Latency            int     `json:"latency,omitempty"`
    Uptime             int64   `json:"uptime,omitempty"`
    XputDown           float64 `json:"xput_down,omitempty"`
    XputUp             float64 `json:"xput_up,omitempty"`
    SpeedtestStatus    string  `json:"speedtest_status,omitempty"`
    SpeedtestLastRun   int64   `json:"speedtest_lastrun,omitempty"`
    SpeedtestPing      int     `json:"speedtest_ping,omitempty"`
    RxBytesR           int64   `json:"rx_bytes-r,omitempty"`
    TxBytesR           int64   `json:"tx_bytes-r,omitempty"`
}
```

---

#### GET stat/sysinfo

Returns detailed system information.

```http
GET /proxy/network/api/s/default/stat/sysinfo HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [{
        "timezone": "America/New_York",
        "autobackup": true,
        "build": "atag_10.0.1_12345",
        "version": "10.0.1",
        "hostname": "UniFi-UDM-Pro",
        "name": "UDM Pro",
        "uptime": 1234567,
        "ip_addrs": ["192.168.1.1"],
        "inform_port": 8080,
        "update_available": false,
        "update_downloaded": false
    }]
}
```

**Go Struct**:
```go
type SysInfo struct {
    Timezone          string   `json:"timezone"`
    Autobackup        bool     `json:"autobackup"`
    Build             string   `json:"build"`
    Version           string   `json:"version"`
    Hostname          string   `json:"hostname"`
    Name              string   `json:"name"`
    Uptime            int64    `json:"uptime"`
    IPAddrs           []string `json:"ip_addrs"`
    InformPort        int      `json:"inform_port"`
    UpdateAvailable   bool     `json:"update_available"`
    UpdateDownloaded  bool     `json:"update_downloaded"`
}
```

---

#### GET stat/sta

Returns all currently connected clients (stations).

```http
GET /proxy/network/api/s/default/stat/sta HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "client_id_1",
            "mac": "aa:bb:cc:dd:ee:ff",
            "site_id": "site_id",
            "oui": "Apple",
            "is_guest": false,
            "first_seen": 1642345678,
            "last_seen": 1642567890,
            "is_wired": false,
            "hostname": "iPhone",
            "name": "John's iPhone",
            "usergroup_id": "",
            "network_id": "network_id_1",
            "ip": "192.168.1.100",
            "essid": "MyWiFi",
            "bssid": "00:11:22:33:44:55",
            "channel": 36,
            "radio": "na",
            "radio_proto": "ax",
            "vlan": 0,
            "signal": -45,
            "noise": -95,
            "rssi": 50,
            "tx_rate": 1200000,
            "rx_rate": 1200000,
            "tx_bytes": 123456789,
            "rx_bytes": 987654321,
            "tx_packets": 12345,
            "rx_packets": 54321,
            "uptime": 3600,
            "satisfaction": 98,
            "ap_mac": "00:11:22:33:44:55",
            "_uptime_by_uap": 3600,
            "_last_seen_by_uap": 1642567890,
            "_is_guest_by_uap": false
        }
    ]
}
```

**Go Struct**:
```go
type Client struct {
    ID              string `json:"_id"`
    MAC             string `json:"mac"`
    SiteID          string `json:"site_id"`
    OUI             string `json:"oui"`
    IsGuest         bool   `json:"is_guest"`
    FirstSeen       int64  `json:"first_seen"`
    LastSeen        int64  `json:"last_seen"`
    IsWired         bool   `json:"is_wired"`
    Hostname        string `json:"hostname"`
    Name            string `json:"name"`
    UsergroupID     string `json:"usergroup_id"`
    NetworkID       string `json:"network_id"`
    IP              string `json:"ip"`
    ESSID           string `json:"essid"`
    BSSID           string `json:"bssid"`
    Channel         int    `json:"channel"`
    Radio           string `json:"radio"`
    RadioProto      string `json:"radio_proto"`
    VLAN            int    `json:"vlan"`
    Signal          int    `json:"signal"`
    Noise           int    `json:"noise"`
    RSSI            int    `json:"rssi"`
    TxRate          int64  `json:"tx_rate"`
    RxRate          int64  `json:"rx_rate"`
    TxBytes         int64  `json:"tx_bytes"`
    RxBytes         int64  `json:"rx_bytes"`
    TxPackets       int64  `json:"tx_packets"`
    RxPackets       int64  `json:"rx_packets"`
    Uptime          int64  `json:"uptime"`
    Satisfaction    int    `json:"satisfaction"`
    APMAC           string `json:"ap_mac"`
    SwitchMAC       string `json:"sw_mac,omitempty"`
    SwitchPort      int    `json:"sw_port,omitempty"`
}

func (c *Client) ListActiveClients() ([]Client, error) {
    resp, err := c.doAPIRequest("GET", "stat/sta", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var apiResp APIResponse[Client]
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }

    if apiResp.Meta.RC != "ok" {
        return nil, fmt.Errorf("API error: %s", apiResp.Meta.Message)
    }

    return apiResp.Data, nil
}
```

---

#### GET stat/alluser

Returns all known clients (including historical/offline).

```http
GET /proxy/network/api/s/default/stat/alluser HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `within` | int | Hours to look back (optional) |
| `type` | string | Filter: `all`, `guest`, `user` |

---

#### GET stat/device

Returns all adopted network devices.

```http
GET /proxy/network/api/s/default/stat/device HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Optional POST filter**:
```json
{
    "macs": ["00:11:22:33:44:55", "aa:bb:cc:dd:ee:ff"]
}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "device_id_1",
            "mac": "00:11:22:33:44:55",
            "model": "UDMPRO",
            "model_in_lts": false,
            "model_in_eol": false,
            "type": "udm",
            "name": "UDM Pro",
            "serial": "SERIAL123456",
            "version": "10.0.1.12345",
            "adopted": true,
            "site_id": "site_id",
            "state": 1,
            "inform_url": "http://192.168.1.1:8080/inform",
            "inform_ip": "192.168.1.1",
            "last_seen": 1642567890,
            "uptime": 1234567,
            "upgradable": false,
            "adoptable_when_upgraded": false,
            "cfgversion": "abc123",
            "config_network": {
                "type": "dhcp",
                "ip": "192.168.1.1",
                "bonding_enabled": false
            },
            "license_state": "registered",
            "two_phase_adopt": false,
            "connected_at": 1641234567,
            "provisioned_at": 1641234567,
            "led_override": "default",
            "led_override_color": "#0000ff",
            "led_override_color_brightness": 100,
            "outdoor_mode_override": "default",
            "system-stats": {
                "cpu": "5.2",
                "mem": "45.3",
                "uptime": "1234567"
            },
            "internet": true,
            "uplink": {
                "full_duplex": true,
                "ip": "192.168.1.1",
                "mac": "00:11:22:33:44:55",
                "name": "eth0",
                "netmask": "255.255.255.0",
                "num_port": 1,
                "rx_bytes": 123456789,
                "rx_dropped": 0,
                "rx_errors": 0,
                "rx_multicast": 0,
                "rx_packets": 1234567,
                "speed": 1000,
                "tx_bytes": 987654321,
                "tx_dropped": 0,
                "tx_errors": 0,
                "tx_packets": 7654321,
                "type": "wire",
                "up": true,
                "uplink_mac": "aa:bb:cc:dd:ee:ff",
                "uplink_remote_port": 1
            }
        }
    ]
}
```

**Go Struct**:
```go
type Device struct {
    ID                      string                 `json:"_id"`
    MAC                     string                 `json:"mac"`
    Model                   string                 `json:"model"`
    ModelInLTS              bool                   `json:"model_in_lts"`
    ModelInEOL              bool                   `json:"model_in_eol"`
    Type                    string                 `json:"type"`
    Name                    string                 `json:"name"`
    Serial                  string                 `json:"serial"`
    Version                 string                 `json:"version"`
    Adopted                 bool                   `json:"adopted"`
    SiteID                  string                 `json:"site_id"`
    State                   int                    `json:"state"`
    InformURL               string                 `json:"inform_url"`
    InformIP                string                 `json:"inform_ip"`
    LastSeen                int64                  `json:"last_seen"`
    Uptime                  int64                  `json:"uptime"`
    Upgradable              bool                   `json:"upgradable"`
    ConfigVersion           string                 `json:"cfgversion"`
    LicenseState            string                 `json:"license_state"`
    ConnectedAt             int64                  `json:"connected_at"`
    ProvisionedAt           int64                  `json:"provisioned_at"`
    LEDOverride             string                 `json:"led_override"`
    Internet                bool                   `json:"internet"`
    SystemStats             map[string]string      `json:"system-stats"`
    Uplink                  *DeviceUplink          `json:"uplink,omitempty"`
    ConfigNetwork           *DeviceConfigNetwork   `json:"config_network,omitempty"`

    // AP-specific fields
    RadioTable              []RadioTableEntry      `json:"radio_table,omitempty"`
    RadioTableStats         []RadioTableStats      `json:"radio_table_stats,omitempty"`
    VAPTable                []VAPEntry             `json:"vap_table,omitempty"`
    NumSTA                  int                    `json:"num_sta,omitempty"`
    UserNumSTA              int                    `json:"user-num_sta,omitempty"`
    GuestNumSTA             int                    `json:"guest-num_sta,omitempty"`

    // Switch-specific fields
    PortTable               []PortEntry            `json:"port_table,omitempty"`
    PortOverrides           []PortOverride         `json:"port_overrides,omitempty"`
    TotalMaxPower           int                    `json:"total_max_power,omitempty"`

    // Gateway-specific fields
    WANType                 string                 `json:"wan_type,omitempty"`
    SpeedtestStatus         string                 `json:"speedtest_status,omitempty"`
    SpeedtestStatusSaved    bool                   `json:"speedtest-status-saved,omitempty"`
}

type DeviceUplink struct {
    FullDuplex     bool   `json:"full_duplex"`
    IP             string `json:"ip"`
    MAC            string `json:"mac"`
    Name           string `json:"name"`
    Netmask        string `json:"netmask"`
    NumPort        int    `json:"num_port"`
    RxBytes        int64  `json:"rx_bytes"`
    TxBytes        int64  `json:"tx_bytes"`
    Speed          int    `json:"speed"`
    Type           string `json:"type"`
    Up             bool   `json:"up"`
    UplinkMAC      string `json:"uplink_mac"`
    UplinkRemotePort int  `json:"uplink_remote_port"`
}

type DeviceConfigNetwork struct {
    Type            string `json:"type"`
    IP              string `json:"ip"`
    BondingEnabled  bool   `json:"bonding_enabled"`
}
```

#### Device States

| State | Meaning |
|-------|---------|
| 0 | Disconnected |
| 1 | Connected |
| 2 | Pending Adoption |
| 4 | Upgrading |
| 5 | Provisioning |
| 6 | Heartbeat Missed |
| 7 | Adopting |
| 9 | Adoption Failed |
| 10 | Isolated |
| 11 | Adopting (wireless) |

---

#### GET stat/device-basic

Returns minimal device information (faster for large deployments).

```http
GET /proxy/network/api/s/default/stat/device-basic HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "mac": "00:11:22:33:44:55",
            "type": "udm",
            "model": "UDMPRO"
        }
    ]
}
```

---

#### GET stat/event

Returns site events (limited to 3000 results).

```http
GET /proxy/network/api/s/default/stat/event HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `within` | int | Hours to look back |
| `start` | int | Pagination start |
| `end` | int | Pagination end |

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "event_id_1",
            "time": 1642567890000,
            "datetime": "2025-01-15T10:00:00Z",
            "key": "EVT_AP_Connected",
            "msg": "AP[00:11:22:33:44:55] was connected",
            "site_id": "site_id",
            "subsystem": "wlan",
            "ap": "00:11:22:33:44:55",
            "ap_name": "Office AP",
            "is_admin": false
        }
    ]
}
```

**Go Struct**:
```go
type Event struct {
    ID          string `json:"_id"`
    Time        int64  `json:"time"`
    Datetime    string `json:"datetime"`
    Key         string `json:"key"`
    Message     string `json:"msg"`
    SiteID      string `json:"site_id"`
    Subsystem   string `json:"subsystem"`
    AP          string `json:"ap,omitempty"`
    APName      string `json:"ap_name,omitempty"`
    Client      string `json:"client,omitempty"`
    User        string `json:"user,omitempty"`
    Hostname    string `json:"hostname,omitempty"`
    SSID        string `json:"ssid,omitempty"`
    IsAdmin     bool   `json:"is_admin"`
}
```

#### Common Event Keys

| Key | Description |
|-----|-------------|
| `EVT_AP_Connected` | AP connected to controller |
| `EVT_AP_Disconnected` | AP disconnected |
| `EVT_AP_Restarted` | AP restarted |
| `EVT_AP_Upgraded` | AP firmware upgraded |
| `EVT_WU_Connected` | Wireless client connected |
| `EVT_WU_Disconnected` | Wireless client disconnected |
| `EVT_WU_Roam` | Client roamed to different AP |
| `EVT_LU_Connected` | Wired client connected |
| `EVT_LU_Disconnected` | Wired client disconnected |
| `EVT_SW_Connected` | Switch connected |
| `EVT_SW_Disconnected` | Switch disconnected |
| `EVT_GW_Connected` | Gateway connected |
| `EVT_GW_WANTransition` | WAN failover occurred |
| `EVT_IPS_Alert` | IPS/IDS alert triggered |
| `EVT_AD_Login` | Admin login |

---

#### GET stat/alarm

Returns site alarms.

```http
GET /proxy/network/api/s/default/stat/alarm HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `archived` | bool | Include archived alarms (`true`/`false`) |

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "alarm_id_1",
            "time": 1642567890000,
            "datetime": "2025-01-15T10:00:00Z",
            "key": "alarm_type",
            "msg": "Alarm message",
            "site_id": "site_id",
            "archived": false
        }
    ]
}
```

---

#### GET stat/report/{interval}.{type}

Returns historical statistics reports.

**Intervals**: `5minutes`, `hourly`, `daily`
**Types**: `site`, `user`, `ap`

```http
GET /proxy/network/api/s/default/stat/report/hourly.site HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `within` | int | Hours to look back (default varies by interval) |
| `attrs` | string | Comma-separated attributes to return |

**5-minute defaults**: Past 12 hours
**Hourly defaults**: Past 7 days
**Daily defaults**: Past 52 weeks

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "time": 1642564800000,
            "o": "site",
            "oid": "site_id",
            "bytes": 123456789,
            "num_sta": 15,
            "lan-num_sta": 10,
            "wlan-num_sta": 5,
            "wan-rx_bytes": 100000000,
            "wan-tx_bytes": 50000000,
            "wlan_bytes": 12345678,
            "wan-rx_packets": 100000,
            "wan-tx_packets": 50000
        }
    ]
}
```

---

#### GET stat/dpi

Returns Deep Packet Inspection statistics.

```http
POST /proxy/network/api/s/default/stat/dpi HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
Content-Type: application/json

{
    "type": "by_app"
}
```

**Request Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `type` | string | `by_app` or `by_cat` (required) |

---

#### GET stat/rogueap

Returns detected neighboring/rogue access points.

```http
GET /proxy/network/api/s/default/stat/rogueap HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Optional POST parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `within` | int | Hours to look back |

---

### REST Endpoints (CRUD Operations)

Base path: `/proxy/network/api/s/{site}/rest/`

All REST endpoints support:
- **GET** - List all / Get by ID
- **POST** - Create new
- **PUT** - Update by ID
- **DELETE** - Delete by ID

---

#### rest/networkconf - Network Configuration

Manages LANs, VLANs, and VPN networks.

**GET all networks**:
```http
GET /proxy/network/api/s/default/rest/networkconf HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**GET specific network**:
```http
GET /proxy/network/api/s/default/rest/networkconf/{network_id} HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "network_id_1",
            "site_id": "site_id",
            "name": "Default",
            "purpose": "corporate",
            "vlan_enabled": false,
            "vlan": 1,
            "ip_subnet": "192.168.1.0/24",
            "dhcpd_enabled": true,
            "dhcpd_start": "192.168.1.100",
            "dhcpd_stop": "192.168.1.200",
            "dhcpd_leasetime": 86400,
            "dhcpguard_enabled": false,
            "dhcpd_dns_enabled": true,
            "dhcpd_dns_1": "1.1.1.1",
            "dhcpd_dns_2": "1.0.0.1",
            "dhcpd_gateway_enabled": true,
            "dhcpd_gateway": "192.168.1.1",
            "domain_name": "localdomain",
            "enabled": true,
            "is_nat": true,
            "networkgroup": "LAN"
        }
    ]
}
```

**Go Struct**:
```go
type Network struct {
    ID                  string `json:"_id,omitempty"`
    SiteID              string `json:"site_id,omitempty"`
    Name                string `json:"name"`
    Purpose             string `json:"purpose"`
    VLANEnabled         bool   `json:"vlan_enabled"`
    VLAN                int    `json:"vlan,omitempty"`
    IPSubnet            string `json:"ip_subnet"`
    DHCPDEnabled        bool   `json:"dhcpd_enabled"`
    DHCPDStart          string `json:"dhcpd_start,omitempty"`
    DHCPDStop           string `json:"dhcpd_stop,omitempty"`
    DHCPDLeaseTime      int    `json:"dhcpd_leasetime,omitempty"`
    DHCPDDNSEnabled     bool   `json:"dhcpd_dns_enabled"`
    DHCPDDNS1           string `json:"dhcpd_dns_1,omitempty"`
    DHCPDDNS2           string `json:"dhcpd_dns_2,omitempty"`
    DHCPDGatewayEnabled bool   `json:"dhcpd_gateway_enabled"`
    DHCPDGateway        string `json:"dhcpd_gateway,omitempty"`
    DomainName          string `json:"domain_name,omitempty"`
    Enabled             bool   `json:"enabled"`
    IsNAT               bool   `json:"is_nat"`
    NetworkGroup        string `json:"networkgroup"`
    IGMPSnooping        bool   `json:"igmp_snooping,omitempty"`
    DHCPGuardEnabled    bool   `json:"dhcpguard_enabled"`
}

// Network purposes
const (
    NetworkPurposeCorporate = "corporate"
    NetworkPurposeGuest     = "guest"
    NetworkPurposeWAN       = "wan"
    NetworkPurposeVPN       = "vpn"
    NetworkPurposeVLAN      = "vlan-only"
)
```

**Create network**:
```http
POST /proxy/network/api/s/default/rest/networkconf HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "name": "IoT Network",
    "purpose": "corporate",
    "vlan_enabled": true,
    "vlan": 20,
    "ip_subnet": "192.168.20.0/24",
    "dhcpd_enabled": true,
    "dhcpd_start": "192.168.20.100",
    "dhcpd_stop": "192.168.20.200",
    "dhcpd_leasetime": 86400,
    "enabled": true,
    "is_nat": true,
    "networkgroup": "LAN"
}
```

**Update network**:
```http
PUT /proxy/network/api/s/default/rest/networkconf/{network_id} HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "_id": "network_id_1",
    "name": "Updated IoT Network",
    "dhcpd_stop": "192.168.20.250"
}
```

**Delete network**:
```http
DELETE /proxy/network/api/s/default/rest/networkconf/{network_id} HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
```

---

#### rest/wlanconf - Wireless Network Configuration

Manages WiFi networks (SSIDs).

**GET all WLANs**:
```http
GET /proxy/network/api/s/default/rest/wlanconf HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "wlan_id_1",
            "site_id": "site_id",
            "name": "MyWiFi",
            "enabled": true,
            "security": "wpapsk",
            "wpa_mode": "wpa2",
            "wpa_enc": "ccmp",
            "x_passphrase": "wifi_password",
            "hide_ssid": false,
            "is_guest": false,
            "networkconf_id": "network_id_1",
            "usergroup_id": "usergroup_id_1",
            "ap_group_ids": ["ap_group_id_1"],
            "wlan_bands": ["2g", "5g"],
            "wpa3_support": false,
            "wpa3_transition": false,
            "pmf_mode": "disabled",
            "fast_roaming_enabled": true,
            "uapsd_enabled": true,
            "minrate_ng_enabled": false,
            "minrate_na_enabled": false,
            "mac_filter_enabled": false,
            "mac_filter_policy": "allow",
            "dtim_mode": "default",
            "dtim_ng": 1,
            "dtim_na": 3,
            "schedule_enabled": false,
            "iapp_enabled": true,
            "l2_isolation": false,
            "group_rekey": 3600,
            "radius_mac_auth_enabled": false,
            "radius_das_enabled": false,
            "no2ghz_oui": false
        }
    ]
}
```

**Go Struct**:
```go
type WLAN struct {
    ID                  string   `json:"_id,omitempty"`
    SiteID              string   `json:"site_id,omitempty"`
    Name                string   `json:"name"`
    Enabled             bool     `json:"enabled"`
    Security            string   `json:"security"`
    WPAMode             string   `json:"wpa_mode,omitempty"`
    WPAEnc              string   `json:"wpa_enc,omitempty"`
    Passphrase          string   `json:"x_passphrase,omitempty"`
    HideSSID            bool     `json:"hide_ssid"`
    IsGuest             bool     `json:"is_guest"`
    NetworkConfID       string   `json:"networkconf_id,omitempty"`
    UsergroupID         string   `json:"usergroup_id,omitempty"`
    APGroupIDs          []string `json:"ap_group_ids,omitempty"`
    WLANBands           []string `json:"wlan_bands,omitempty"`
    WPA3Support         bool     `json:"wpa3_support"`
    WPA3Transition      bool     `json:"wpa3_transition"`
    PMFMode             string   `json:"pmf_mode,omitempty"`
    FastRoamingEnabled  bool     `json:"fast_roaming_enabled"`
    UAPSDEnabled        bool     `json:"uapsd_enabled"`
    MACFilterEnabled    bool     `json:"mac_filter_enabled"`
    MACFilterPolicy     string   `json:"mac_filter_policy,omitempty"`
    MACFilterList       []string `json:"mac_filter_list,omitempty"`
    ScheduleEnabled     bool     `json:"schedule_enabled"`
    Schedule            []string `json:"schedule,omitempty"`
    L2Isolation         bool     `json:"l2_isolation"`
    GroupRekey          int      `json:"group_rekey,omitempty"`

    // RADIUS settings
    RADIUSMACAuthEnabled bool   `json:"radius_mac_auth_enabled"`
    RADIUSProfileID      string `json:"radius_profile_id,omitempty"`
}

// Security types
const (
    SecurityOpen    = "open"
    SecurityWPAPSK  = "wpapsk"
    SecurityWPA2PSK = "wpapsk"  // Use wpa_mode to differentiate
    SecurityWPA3    = "wpa3"
    SecurityWPAEAP  = "wpaeap"  // Enterprise
)

// WPA modes
const (
    WPAModeWPA1 = "wpa1"
    WPAModeWPA2 = "wpa2"
    WPAModeWPA3 = "wpa3"
    WPAModeBoth = "both"  // WPA1+WPA2
)
```

**Create WLAN**:
```http
POST /proxy/network/api/s/default/rest/wlanconf HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "name": "Guest WiFi",
    "enabled": true,
    "security": "wpapsk",
    "wpa_mode": "wpa2",
    "wpa_enc": "ccmp",
    "x_passphrase": "GuestPassword123!",
    "hide_ssid": false,
    "is_guest": true,
    "networkconf_id": "guest_network_id",
    "ap_group_ids": ["default_ap_group_id"],
    "wlan_bands": ["2g", "5g"]
}
```

---

#### rest/firewallrule - Firewall Rules

Manages user-defined firewall rules (does not include auto-generated rules).

**GET all firewall rules**:
```http
GET /proxy/network/api/s/default/rest/firewallrule HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "rule_id_1",
            "site_id": "site_id",
            "name": "Block IoT to Main LAN",
            "enabled": true,
            "ruleset": "LAN_IN",
            "rule_index": 2000,
            "action": "drop",
            "protocol": "all",
            "protocol_match_excepted": false,
            "logging": false,
            "state_new": false,
            "state_established": false,
            "state_invalid": false,
            "state_related": false,
            "src_firewallgroup_ids": ["firewall_group_iot"],
            "dst_firewallgroup_ids": ["firewall_group_main"],
            "src_mac_address": "",
            "dst_address": "",
            "src_address": "",
            "dst_port": "",
            "src_port": "",
            "icmp_typename": ""
        }
    ]
}
```

**Go Struct**:
```go
type FirewallRule struct {
    ID                      string   `json:"_id,omitempty"`
    SiteID                  string   `json:"site_id,omitempty"`
    Name                    string   `json:"name"`
    Enabled                 bool     `json:"enabled"`
    Ruleset                 string   `json:"ruleset"`
    RuleIndex               int      `json:"rule_index"`
    Action                  string   `json:"action"`
    Protocol                string   `json:"protocol"`
    ProtocolMatchExcepted   bool     `json:"protocol_match_excepted"`
    Logging                 bool     `json:"logging"`
    StateNew                bool     `json:"state_new"`
    StateEstablished        bool     `json:"state_established"`
    StateInvalid            bool     `json:"state_invalid"`
    StateRelated            bool     `json:"state_related"`
    SrcFirewallGroupIDs     []string `json:"src_firewallgroup_ids,omitempty"`
    DstFirewallGroupIDs     []string `json:"dst_firewallgroup_ids,omitempty"`
    SrcMACAddress           string   `json:"src_mac_address,omitempty"`
    DstAddress              string   `json:"dst_address,omitempty"`
    SrcAddress              string   `json:"src_address,omitempty"`
    DstPort                 string   `json:"dst_port,omitempty"`
    SrcPort                 string   `json:"src_port,omitempty"`
    ICMPTypename            string   `json:"icmp_typename,omitempty"`
}

// Rulesets
const (
    RulesetWANIn    = "WAN_IN"
    RulesetWANOut   = "WAN_OUT"
    RulesetWANLocal = "WAN_LOCAL"
    RulesetLANIn    = "LAN_IN"
    RulesetLANOut   = "LAN_OUT"
    RulesetLANLocal = "LAN_LOCAL"
    RulesetGuestIn  = "GUEST_IN"
    RulesetGuestOut = "GUEST_OUT"
)

// Actions
const (
    ActionAccept = "accept"
    ActionDrop   = "drop"
    ActionReject = "reject"
)
```

**Enable/Disable a firewall rule**:
```http
PUT /proxy/network/api/s/default/rest/firewallrule/{rule_id} HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "enabled": false
}
```

---

#### rest/firewallgroup - Firewall Groups

Manages firewall groups (IP groups, port groups).

**GET all firewall groups**:
```http
GET /proxy/network/api/s/default/rest/firewallgroup HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "group_id_1",
            "site_id": "site_id",
            "name": "Blocked IPs",
            "group_type": "address-group",
            "group_members": [
                "10.0.0.0/8",
                "192.168.100.0/24",
                "203.0.113.50"
            ]
        },
        {
            "_id": "group_id_2",
            "site_id": "site_id",
            "name": "Web Ports",
            "group_type": "port-group",
            "group_members": [
                "80",
                "443",
                "8080"
            ]
        }
    ]
}
```

**Go Struct**:
```go
type FirewallGroup struct {
    ID           string   `json:"_id,omitempty"`
    SiteID       string   `json:"site_id,omitempty"`
    Name         string   `json:"name"`
    GroupType    string   `json:"group_type"`
    GroupMembers []string `json:"group_members"`
}

// Group types
const (
    GroupTypeAddress     = "address-group"
    GroupTypeIPv6Address = "ipv6-address-group"
    GroupTypePort        = "port-group"
)
```

---

#### rest/portforward - Port Forwarding

Manages port forwarding rules.

**GET all port forwards**:
```http
GET /proxy/network/api/s/default/rest/portforward HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "portforward_id_1",
            "site_id": "site_id",
            "name": "SSH Server",
            "enabled": true,
            "pfwd_interface": "wan",
            "src": "any",
            "dst_port": "22",
            "fwd": "192.168.1.50",
            "fwd_port": "22",
            "proto": "tcp_udp",
            "log": false
        }
    ]
}
```

**Go Struct**:
```go
type PortForward struct {
    ID            string `json:"_id,omitempty"`
    SiteID        string `json:"site_id,omitempty"`
    Name          string `json:"name"`
    Enabled       bool   `json:"enabled"`
    PfwdInterface string `json:"pfwd_interface"`
    Src           string `json:"src"`
    DstPort       string `json:"dst_port"`
    Fwd           string `json:"fwd"`
    FwdPort       string `json:"fwd_port"`
    Proto         string `json:"proto"`
    Log           bool   `json:"log"`
}

// Protocols
const (
    ProtoTCP    = "tcp"
    ProtoUDP    = "udp"
    ProtoTCPUDP = "tcp_udp"
)
```

**Create port forward**:
```http
POST /proxy/network/api/s/default/rest/portforward HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "name": "Web Server",
    "enabled": true,
    "pfwd_interface": "wan",
    "src": "any",
    "dst_port": "8443",
    "fwd": "192.168.1.100",
    "fwd_port": "443",
    "proto": "tcp",
    "log": true
}
```

---

#### rest/portconf - Switch Port Profiles

Manages switch port profiles.

**GET all port profiles**:
```http
GET /proxy/network/api/s/default/rest/portconf HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "portconf_id_1",
            "site_id": "site_id",
            "name": "All",
            "forward": "all",
            "native_networkconf_id": "",
            "dot1x_ctrl": "force_authorized",
            "dot1x_idle_timeout": 300,
            "egress_rate_limit_kbps_enabled": false,
            "poe_mode": "auto",
            "stormctrl_enabled": false,
            "autoneg": true,
            "isolation": false,
            "lldpmed_enabled": true,
            "stp_port_mode": true
        }
    ]
}
```

**Go Struct**:
```go
type PortProfile struct {
    ID                           string   `json:"_id,omitempty"`
    SiteID                       string   `json:"site_id,omitempty"`
    Name                         string   `json:"name"`
    Forward                      string   `json:"forward"`
    NativeNetworkConfID          string   `json:"native_networkconf_id,omitempty"`
    TaggedNetworkConfIDs         []string `json:"tagged_networkconf_ids,omitempty"`
    VoiceNetworkConfID           string   `json:"voice_networkconf_id,omitempty"`
    Dot1xCtrl                    string   `json:"dot1x_ctrl"`
    Dot1xIdleTimeout             int      `json:"dot1x_idle_timeout"`
    EgressRateLimitKbpsEnabled   bool     `json:"egress_rate_limit_kbps_enabled"`
    EgressRateLimitKbps          int      `json:"egress_rate_limit_kbps,omitempty"`
    POEMode                      string   `json:"poe_mode"`
    StormCtrlEnabled             bool     `json:"stormctrl_enabled"`
    StormCtrlBcastEnabled        bool     `json:"stormctrl_bcast_enabled,omitempty"`
    StormCtrlBcastRate           int      `json:"stormctrl_bcast_rate,omitempty"`
    StormCtrlMcastEnabled        bool     `json:"stormctrl_mcast_enabled,omitempty"`
    StormCtrlMcastRate           int      `json:"stormctrl_mcast_rate,omitempty"`
    StormCtrlUcastEnabled        bool     `json:"stormctrl_ucast_enabled,omitempty"`
    StormCtrlUcastRate           int      `json:"stormctrl_ucast_rate,omitempty"`
    Autoneg                      bool     `json:"autoneg"`
    Isolation                    bool     `json:"isolation"`
    LLDPMEDEnabled               bool     `json:"lldpmed_enabled"`
    STPPortMode                  bool     `json:"stp_port_mode"`
}
```

---

#### rest/routing - Static Routes

Manages user-defined static routes.

**GET all routes**:
```http
GET /proxy/network/api/s/default/rest/routing HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "route_id_1",
            "site_id": "site_id",
            "name": "VPN Route",
            "enabled": true,
            "type": "static",
            "static-route_network": "10.0.0.0/8",
            "static-route_nexthop": "192.168.1.254",
            "static-route_distance": 1,
            "static-route_interface": "LAN"
        }
    ]
}
```

**Go Struct**:
```go
type Route struct {
    ID                    string `json:"_id,omitempty"`
    SiteID                string `json:"site_id,omitempty"`
    Name                  string `json:"name"`
    Enabled               bool   `json:"enabled"`
    Type                  string `json:"type"`
    StaticRouteNetwork    string `json:"static-route_network,omitempty"`
    StaticRouteNexthop    string `json:"static-route_nexthop,omitempty"`
    StaticRouteDistance   int    `json:"static-route_distance,omitempty"`
    StaticRouteInterface  string `json:"static-route_interface,omitempty"`
}
```

---

#### rest/user - Known Clients

Manages known/configured client entries.

**GET all known clients**:
```http
GET /proxy/network/api/s/default/rest/user HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "user_id_1",
            "site_id": "site_id",
            "mac": "aa:bb:cc:dd:ee:ff",
            "name": "John's Laptop",
            "hostname": "johns-laptop",
            "note": "Primary work device",
            "use_fixedip": true,
            "fixed_ip": "192.168.1.50",
            "network_id": "network_id_1",
            "usergroup_id": "",
            "blocked": false,
            "last_seen": 1642567890,
            "first_seen": 1641234567,
            "oui": "Apple",
            "fingerprint_source": 0,
            "dev_cat": 1,
            "dev_family": 4,
            "dev_vendor": 47,
            "dev_id": 0
        }
    ]
}
```

**Go Struct**:
```go
type KnownClient struct {
    ID              string `json:"_id,omitempty"`
    SiteID          string `json:"site_id,omitempty"`
    MAC             string `json:"mac"`
    Name            string `json:"name,omitempty"`
    Hostname        string `json:"hostname,omitempty"`
    Note            string `json:"note,omitempty"`
    UseFixedIP      bool   `json:"use_fixedip"`
    FixedIP         string `json:"fixed_ip,omitempty"`
    NetworkID       string `json:"network_id,omitempty"`
    UsergroupID     string `json:"usergroup_id,omitempty"`
    Blocked         bool   `json:"blocked"`
    LastSeen        int64  `json:"last_seen,omitempty"`
    FirstSeen       int64  `json:"first_seen,omitempty"`
    OUI             string `json:"oui,omitempty"`
    DevCat          int    `json:"dev_cat,omitempty"`
    DevFamily       int    `json:"dev_family,omitempty"`
    DevVendor       int    `json:"dev_vendor,omitempty"`
    DevID           int    `json:"dev_id,omitempty"`
}
```

**Create/Update known client**:
```http
POST /proxy/network/api/s/default/rest/user HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "mac": "aa:bb:cc:dd:ee:ff",
    "name": "Server",
    "use_fixedip": true,
    "fixed_ip": "192.168.1.10",
    "network_id": "network_id_1"
}
```

---

#### rest/setting - Site Settings

Manages site-wide settings. This is a complex endpoint with nested setting categories.

**GET all settings**:
```http
GET /proxy/network/api/s/default/rest/setting HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response structure** (each setting type is a separate object):
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "setting_id_1",
            "key": "connectivity",
            "site_id": "site_id",
            "uplink_type": "gateway",
            "enabled": true
        },
        {
            "_id": "setting_id_2",
            "key": "mgmt",
            "site_id": "site_id",
            "advanced_feature_enabled": true,
            "led_enabled": true,
            "auto_upgrade": false,
            "unifi_idp_enabled": true
        },
        {
            "_id": "setting_id_3",
            "key": "country",
            "site_id": "site_id",
            "code": 840
        }
    ]
}
```

**Setting keys**:
| Key | Description |
|-----|-------------|
| `connectivity` | Network connectivity settings |
| `mgmt` | Management settings (LED, auto-upgrade, etc.) |
| `country` | Regulatory country code |
| `locale` | Locale/language settings |
| `ntp` | NTP server settings |
| `guest_access` | Guest portal settings |
| `super_mgmt` | Super admin management |
| `dpi` | Deep Packet Inspection settings |
| `ips` | Intrusion Prevention settings |
| `network_optimization` | Network optimization |
| `auto_speedtest` | Automatic speed test settings |
| `rsyslogd` | Remote syslog settings |
| `snmp` | SNMP settings |
| `radius` | RADIUS settings |

**Update a specific setting**:
```http
PUT /proxy/network/api/s/default/rest/setting/{setting_key} HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "led_enabled": false
}
```

---

### Command Endpoints (cmd/)

Command endpoints use POST requests to execute actions.

Base path: `/proxy/network/api/s/{site}/cmd/{manager}`

Request format:
```json
{
    "cmd": "command_name",
    "param1": "value1",
    "param2": "value2"
}
```

---

#### cmd/devmgr - Device Manager

**Adopt a device**:
```http
POST /proxy/network/api/s/default/cmd/devmgr HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "cmd": "adopt",
    "mac": "00:11:22:33:44:55"
}
```

**All devmgr commands**:

| Command | Required Parameters | Description |
|---------|---------------------|-------------|
| `adopt` | `mac` | Adopt a device |
| `restart` | `mac` | Restart a device |
| `force-provision` | `mac` | Force provision a device |
| `upgrade` | `mac` | Upgrade device firmware |
| `upgrade-external` | `mac`, `url` | Upgrade from external URL |
| `power-cycle` | `mac`, `port_idx` | Cycle PoE port on switch |
| `set-locate` | `mac` | Enable locating LED |
| `unset-locate` | `mac` | Disable locating LED |
| `speedtest` | - | Start speed test |
| `speedtest-status` | - | Get speed test status |
| `spectrum-scan` | `mac` | Start spectrum scan |

**Go Implementation**:
```go
type DeviceCommand struct {
    Cmd     string `json:"cmd"`
    MAC     string `json:"mac,omitempty"`
    PortIdx int    `json:"port_idx,omitempty"`
    URL     string `json:"url,omitempty"`
}

func (c *Client) RestartDevice(mac string) error {
    cmd := DeviceCommand{
        Cmd: "restart",
        MAC: mac,
    }

    resp, err := c.doAPIRequest("POST", "cmd/devmgr", cmd)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to restart device: %d", resp.StatusCode)
    }

    return nil
}

func (c *Client) AdoptDevice(mac string) error {
    cmd := DeviceCommand{
        Cmd: "adopt",
        MAC: mac,
    }

    resp, err := c.doAPIRequest("POST", "cmd/devmgr", cmd)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (c *Client) PowerCyclePOEPort(switchMAC string, portIndex int) error {
    cmd := DeviceCommand{
        Cmd:     "power-cycle",
        MAC:     switchMAC,
        PortIdx: portIndex,
    }

    resp, err := c.doAPIRequest("POST", "cmd/devmgr", cmd)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

---

#### cmd/stamgr - Station (Client) Manager

**Block a client**:
```http
POST /proxy/network/api/s/default/cmd/stamgr HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "cmd": "block-sta",
    "mac": "aa:bb:cc:dd:ee:ff"
}
```

**All stamgr commands**:

| Command | Required Parameters | Description |
|---------|---------------------|-------------|
| `block-sta` | `mac` | Block a client |
| `unblock-sta` | `mac` | Unblock a client |
| `kick-sta` | `mac` | Disconnect (kick) a client |
| `forget-sta` | `mac` | Forget a client (remove from known) |
| `authorize-guest` | `mac`, additional | Authorize guest |
| `unauthorize-guest` | `mac` | Unauthorize guest |

**Authorize guest with parameters**:
```json
{
    "cmd": "authorize-guest",
    "mac": "aa:bb:cc:dd:ee:ff",
    "minutes": 60,
    "up": 1024,
    "down": 2048,
    "bytes": 104857600,
    "ap_mac": "00:11:22:33:44:55"
}
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `minutes` | int | Authorization duration |
| `up` | int | Upload bandwidth limit (Kbps) |
| `down` | int | Download bandwidth limit (Kbps) |
| `bytes` | int | Data limit (bytes) |
| `ap_mac` | string | Specific AP to authorize on |

**Go Implementation**:
```go
type ClientCommand struct {
    Cmd     string `json:"cmd"`
    MAC     string `json:"mac"`
    Minutes int    `json:"minutes,omitempty"`
    Up      int    `json:"up,omitempty"`
    Down    int    `json:"down,omitempty"`
    Bytes   int64  `json:"bytes,omitempty"`
    APMAC   string `json:"ap_mac,omitempty"`
}

func (c *Client) BlockClient(mac string) error {
    cmd := ClientCommand{
        Cmd: "block-sta",
        MAC: strings.ToLower(strings.ReplaceAll(mac, ":", "")),
    }

    resp, err := c.doAPIRequest("POST", "cmd/stamgr", cmd)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (c *Client) UnblockClient(mac string) error {
    cmd := ClientCommand{
        Cmd: "unblock-sta",
        MAC: strings.ToLower(strings.ReplaceAll(mac, ":", "")),
    }

    resp, err := c.doAPIRequest("POST", "cmd/stamgr", cmd)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (c *Client) KickClient(mac string) error {
    cmd := ClientCommand{
        Cmd: "kick-sta",
        MAC: strings.ToLower(strings.ReplaceAll(mac, ":", "")),
    }

    resp, err := c.doAPIRequest("POST", "cmd/stamgr", cmd)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (c *Client) AuthorizeGuest(mac string, minutes int, upKbps, downKbps int, bytesLimit int64) error {
    cmd := ClientCommand{
        Cmd:     "authorize-guest",
        MAC:     strings.ToLower(strings.ReplaceAll(mac, ":", "")),
        Minutes: minutes,
        Up:      upKbps,
        Down:    downKbps,
        Bytes:   bytesLimit,
    }

    resp, err := c.doAPIRequest("POST", "cmd/stamgr", cmd)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

---

#### cmd/sitemgr - Site Manager

**Create a new site**:
```http
POST /proxy/network/api/s/default/cmd/sitemgr HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "cmd": "add-site",
    "name": "Branch Office",
    "desc": "New York Branch"
}
```

| Command | Required Parameters | Description |
|---------|---------------------|-------------|
| `add-site` | `name`, `desc` | Create new site |
| `delete-site` | `name` | Delete a site |
| `move-device` | `mac`, `site` | Move device to another site |

---

#### cmd/backup - Backup Manager

**List backups**:
```http
POST /proxy/network/api/s/default/cmd/backup HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "cmd": "list-backups"
}
```

| Command | Required Parameters | Description |
|---------|---------------------|-------------|
| `list-backups` | - | List available backups |
| `delete-backup` | `filename` | Delete a backup file |

---

#### cmd/system - System Commands

**Available commands**:

| Command | Description |
|---------|-------------|
| `reboot` | Reboot controller (requires CSRF token + Super Admin) |
| `poweroff` | Power off controller (requires CSRF token + Super Admin) |

---

### API v2 Endpoints

Base path: `/proxy/network/v2/api/site/{site}/`

The v2 API provides newer endpoints with enhanced functionality.

---

#### v2/api/site/{site}/trafficrules - Traffic Rules

Traffic rules provide higher-level rule management with device targeting.

**GET all traffic rules**:
```http
GET /proxy/network/v2/api/site/default/trafficrules HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

**Response**:
```json
{
    "meta": {"rc": "ok"},
    "data": [
        {
            "_id": "traffic_rule_id_1",
            "site_id": "site_id",
            "name": "Block Social Media",
            "enabled": true,
            "action": "BLOCK",
            "matching_target": "INTERNET",
            "target_devices": [
                {"client_mac": "aa:bb:cc:dd:ee:ff", "type": "CLIENT"}
            ],
            "app_category_ids": ["social-network"],
            "schedule": {
                "mode": "ALWAYS"
            }
        }
    ]
}
```

**Create traffic rule**:
```http
POST /proxy/network/v2/api/site/default/trafficrules HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "name": "Kids Internet Schedule",
    "enabled": true,
    "action": "BLOCK",
    "matching_target": "INTERNET",
    "target_devices": [
        {"client_mac": "aa:bb:cc:dd:ee:ff", "type": "CLIENT"},
        {"client_mac": "11:22:33:44:55:66", "type": "CLIENT"}
    ],
    "schedule": {
        "mode": "CUSTOM",
        "repeat_on_days": ["mon", "tue", "wed", "thu", "fri"],
        "time_all_day": false,
        "time_range_start": "22:00",
        "time_range_stop": "07:00"
    }
}
```

**Update traffic rule**:
```http
PUT /proxy/network/v2/api/site/default/trafficrules/{rule_id}/ HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
Content-Type: application/json

{
    "enabled": false
}
```

Note: PUT returns status 201 (not 200) on success.

**Delete traffic rule**:
```http
DELETE /proxy/network/v2/api/site/default/trafficrules/{rule_id}/ HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
X-CSRF-Token: {csrf_token}
```

**Go Struct**:
```go
type TrafficRule struct {
    ID              string         `json:"_id,omitempty"`
    SiteID          string         `json:"site_id,omitempty"`
    Name            string         `json:"name"`
    Enabled         bool           `json:"enabled"`
    Action          string         `json:"action"`
    MatchingTarget  string         `json:"matching_target"`
    TargetDevices   []TargetDevice `json:"target_devices,omitempty"`
    NetworkID       string         `json:"network_id,omitempty"`
    IPAddresses     []string       `json:"ip_addresses,omitempty"`
    Domains         []string       `json:"domains,omitempty"`
    AppCategoryIDs  []string       `json:"app_category_ids,omitempty"`
    AppIDs          []int          `json:"app_ids,omitempty"`
    Regions         []string       `json:"regions,omitempty"`
    IPRange         *IPRange       `json:"ip_range,omitempty"`
    Schedule        *Schedule      `json:"schedule,omitempty"`
    Bandwidth       *Bandwidth     `json:"bandwidth,omitempty"`
}

type TargetDevice struct {
    ClientMAC  string `json:"client_mac,omitempty"`
    NetworkID  string `json:"network_id,omitempty"`
    Type       string `json:"type"` // "CLIENT", "NETWORK", "ALL"
}

type Schedule struct {
    Mode           string   `json:"mode"` // "ALWAYS", "CUSTOM"
    RepeatOnDays   []string `json:"repeat_on_days,omitempty"`
    TimeAllDay     bool     `json:"time_all_day,omitempty"`
    TimeRangeStart string   `json:"time_range_start,omitempty"`
    TimeRangeStop  string   `json:"time_range_stop,omitempty"`
}

type Bandwidth struct {
    DownloadLimit int `json:"download_limit,omitempty"` // Kbps
    UploadLimit   int `json:"upload_limit,omitempty"`   // Kbps
}

type IPRange struct {
    Start string `json:"start"`
    Stop  string `json:"stop"`
}

// Actions
const (
    TrafficActionBlock       = "BLOCK"
    TrafficActionAllow       = "ALLOW"
    TrafficActionLimitBW     = "LIMIT_BANDWIDTH"
)

// Matching targets
const (
    MatchingTargetInternet   = "INTERNET"
    MatchingTargetLocalNet   = "LOCAL_NETWORK"
    MatchingTargetIP         = "IP_ADDRESS"
    MatchingTargetDomain     = "DOMAIN"
    MatchingTargetRegion     = "REGION"
    MatchingTargetApp        = "APP"
    MatchingTargetAppCat     = "APP_CATEGORY"
)
```

---

#### v2/api/site/{site}/notifications - Notifications

**GET notifications**:
```http
GET /proxy/network/v2/api/site/default/notifications HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

---

### Integrations API (v1)

Base path: `/proxy/network/integrations/v1/`

**GET sites (integrations)**:
```http
GET /proxy/network/integrations/v1/sites HTTP/1.1
Host: {udm-ip}
Cookie: TOKEN={session_token}
```

---

## Site Manager (Cloud) API Endpoints

Base URL: `https://api.ui.com/v1/`

Authentication: `X-API-KEY` header

---

### GET /v1/hosts

Returns all hosts (consoles) associated with your account.

```http
GET /v1/hosts HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key
Accept: application/json
```

**Response**:
```json
{
    "data": [
        {
            "id": "host_id_1",
            "hardwareId": "UDMPRO-SERIAL123",
            "type": "console",
            "reportedState": {
                "mac": "00:11:22:33:44:55",
                "hostname": "UDM-Pro",
                "firmwareVersion": "4.0.6",
                "isSetup": true,
                "isAdopted": true,
                "ipAddress": "192.168.1.1",
                "wan": {
                    "ip": "203.0.113.1"
                },
                "controllers": {
                    "network": {
                        "running": true,
                        "version": "10.0.1",
                        "releaseChannel": "release"
                    }
                },
                "hardware": {
                    "model": "UDMPRO",
                    "shortname": "Dream Machine Pro"
                }
            },
            "userData": {
                "alias": "Home Network"
            }
        }
    ],
    "meta": {
        "totalCount": 1,
        "offset": 0,
        "limit": 200
    }
}
```

**Go Struct**:
```go
type Host struct {
    ID            string        `json:"id"`
    HardwareID    string        `json:"hardwareId"`
    Type          string        `json:"type"`
    ReportedState ReportedState `json:"reportedState"`
    UserData      UserData      `json:"userData"`
}

type ReportedState struct {
    MAC             string            `json:"mac"`
    Hostname        string            `json:"hostname"`
    FirmwareVersion string            `json:"firmwareVersion"`
    IsSetup         bool              `json:"isSetup"`
    IsAdopted       bool              `json:"isAdopted"`
    IPAddress       string            `json:"ipAddress"`
    WAN             *WANInfo          `json:"wan,omitempty"`
    Controllers     map[string]Controller `json:"controllers,omitempty"`
    Hardware        HardwareInfo      `json:"hardware"`
}

type Controller struct {
    Running        bool   `json:"running"`
    Version        string `json:"version"`
    ReleaseChannel string `json:"releaseChannel"`
}

type HardwareInfo struct {
    Model     string `json:"model"`
    Shortname string `json:"shortname"`
}

type WANInfo struct {
    IP string `json:"ip"`
}

type UserData struct {
    Alias string `json:"alias"`
}
```

---

### GET /v1/hosts/{hostId}

Returns a specific host by ID.

```http
GET /v1/hosts/{hostId} HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key
Accept: application/json
```

---

### GET /v1/sites

Returns all sites across all hosts.

```http
GET /v1/sites HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key
Accept: application/json
```

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `hostIds` | string | Comma-separated host IDs to filter |

**Response**:
```json
{
    "data": [
        {
            "siteId": "site_id_1",
            "hostId": "host_id_1",
            "meta": {
                "name": "default",
                "description": "Default",
                "timezone": "America/New_York"
            },
            "statistics": {
                "counts": {
                    "offlineDevice": 0,
                    "pendingDevice": 0,
                    "activeDevice": 5,
                    "totalDevice": 5
                },
                "percentages": {
                    "txRetry": 1.2,
                    "satisfaction": 98
                },
                "wan": {
                    "status": "connected",
                    "latency": 15,
                    "download": 450000000,
                    "upload": 50000000
                }
            }
        }
    ]
}
```

**Go Struct**:
```go
type Site struct {
    SiteID     string          `json:"siteId"`
    HostID     string          `json:"hostId"`
    Meta       SiteMeta        `json:"meta"`
    Statistics SiteStatistics  `json:"statistics"`
}

type SiteMeta struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Timezone    string `json:"timezone"`
}

type SiteStatistics struct {
    Counts      DeviceCounts   `json:"counts"`
    Percentages Percentages    `json:"percentages"`
    WAN         WANStatistics  `json:"wan"`
}

type DeviceCounts struct {
    OfflineDevice int `json:"offlineDevice"`
    PendingDevice int `json:"pendingDevice"`
    ActiveDevice  int `json:"activeDevice"`
    TotalDevice   int `json:"totalDevice"`
}

type Percentages struct {
    TxRetry      float64 `json:"txRetry"`
    Satisfaction float64 `json:"satisfaction"`
}

type WANStatistics struct {
    Status   string `json:"status"`
    Latency  int    `json:"latency"`
    Download int64  `json:"download"`
    Upload   int64  `json:"upload"`
}
```

---

### GET /v1/devices

Returns all devices across all sites.

```http
GET /v1/devices HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key
Accept: application/json
```

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `hostIds` | string | Filter by host IDs |
| `siteIds` | string | Filter by site IDs |
| `macAddresses` | string | Filter by MAC addresses |
| `time` | string | ISO 8601 timestamp for historical data |

**Response**:
```json
{
    "data": [
        {
            "mac": "00:11:22:33:44:55",
            "ip": "192.168.1.10",
            "hostId": "host_id_1",
            "siteId": "site_id_1",
            "name": "Office AP",
            "model": "U6-Pro",
            "type": "uap",
            "status": "online",
            "version": "7.0.66",
            "uptime": 1234567,
            "lastSeen": "2025-01-15T10:00:00Z",
            "uidb": {
                "guid": "device_guid",
                "id": "internal_id",
                "images": {
                    "default": "https://images.ui.com/...",
                    "topology": "https://images.ui.com/..."
                }
            }
        }
    ]
}
```

**Go Struct**:
```go
type CloudDevice struct {
    MAC      string        `json:"mac"`
    IP       string        `json:"ip"`
    HostID   string        `json:"hostId"`
    SiteID   string        `json:"siteId"`
    Name     string        `json:"name"`
    Model    string        `json:"model"`
    Type     string        `json:"type"`
    Status   string        `json:"status"`
    Version  string        `json:"version"`
    Uptime   int64         `json:"uptime"`
    LastSeen string        `json:"lastSeen"`
    UIDB     *UIDB         `json:"uidb,omitempty"`
}

type UIDB struct {
    GUID   string     `json:"guid"`
    ID     string     `json:"id"`
    Images UIDImages  `json:"images"`
}

type UIDImages struct {
    Default  string `json:"default"`
    Topology string `json:"topology"`
}
```

---

### GET /v1/isp-metrics (Early Access)

Returns ISP metrics for all sites.

```http
GET /ea/v1/isp-metrics HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key
Accept: application/json
```

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `interval` | string | `5m` or `1h` |
| `begin` | string | ISO 8601 start time |
| `end` | string | ISO 8601 end time |

Note: 5-minute metrics available for 24 hours, 1-hour metrics for 30 days.

**Response**:
```json
{
    "data": [
        {
            "hostId": "host_id_1",
            "siteId": "site_id_1",
            "metrics": [
                {
                    "timestamp": "2025-01-15T10:00:00Z",
                    "latency": {
                        "avg": 15.5,
                        "max": 45.2
                    },
                    "download": 450000000,
                    "upload": 50000000,
                    "uptime": 100.0,
                    "packetLoss": 0.1,
                    "ispName": "Comcast"
                }
            ]
        }
    ]
}
```

---

### POST /v1/isp-metrics (Early Access)

Query ISP metrics with more options.

```http
POST /ea/v1/isp-metrics HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key
Content-Type: application/json

{
    "hostIds": ["host_id_1", "host_id_2"],
    "siteIds": ["site_id_1"],
    "interval": "1h",
    "begin": "2025-01-14T00:00:00Z",
    "end": "2025-01-15T00:00:00Z"
}
```

---

### GET /v1/sd-wan/configs

Returns SD-WAN configurations.

```http
GET /v1/sd-wan/configs HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key
Accept: application/json
```

---

### GET /v1/sd-wan/configs/{configId}

Returns specific SD-WAN configuration.

```http
GET /v1/sd-wan/configs/{configId} HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key
Accept: application/json
```

---

### GET /v1/sd-wan/configs/{configId}/status

Returns SD-WAN configuration status.

```http
GET /v1/sd-wan/configs/{configId}/status HTTP/1.1
Host: api.ui.com
X-API-KEY: your_api_key
Accept: application/json
```

---

## WebSocket API

The WebSocket API provides real-time event streaming.

### Connection

**URL**: `wss://{udm-ip}/proxy/network/wss/s/{site}/events`

**Authentication**: Include session cookie from HTTP login.

### Go Implementation

```go
package unifi

import (
    "encoding/json"
    "net/http"
    "net/url"

    "github.com/gorilla/websocket"
)

type WebSocketEvent struct {
    Meta struct {
        Message string `json:"message"`
        RC      string `json:"rc"`
    } `json:"meta"`
    Data []json.RawMessage `json:"data"`
}

type EventCallback func(eventType string, data json.RawMessage)

func (c *Client) ConnectWebSocket(callback EventCallback) error {
    u := url.URL{
        Scheme: "wss",
        Host:   c.getHost(),
        Path:   fmt.Sprintf("/proxy/network/wss/s/%s/events", c.Site),
    }

    // Get cookies from HTTP client
    cookies := c.HTTPClient.Jar.Cookies(c.parseURL())
    header := http.Header{}
    for _, cookie := range cookies {
        header.Add("Cookie", cookie.String())
    }

    dialer := websocket.Dialer{
        TLSClientConfig: c.HTTPClient.Transport.(*http.Transport).TLSClientConfig,
    }

    conn, _, err := dialer.Dial(u.String(), header)
    if err != nil {
        return err
    }
    defer conn.Close()

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            return err
        }

        var event WebSocketEvent
        if err := json.Unmarshal(message, &event); err != nil {
            continue
        }

        for _, data := range event.Data {
            callback(event.Meta.Message, data)
        }
    }
}
```

### Event Types

| Event Type | Description |
|------------|-------------|
| `sta:sync` | Client connect/disconnect |
| `device:sync` | Device status change |
| `device:update` | Device configuration update |
| `alarm` | Alarm triggered |
| `event` | General event |
| `speedtest:done` | Speed test completed |
| `backup:done` | Backup completed |
| `upgrade:progress` | Firmware upgrade progress |

### Event Data Structures

**Client Sync Event** (`sta:sync`):
```json
{
    "mac": "aa:bb:cc:dd:ee:ff",
    "site_id": "site_id",
    "ip": "192.168.1.100",
    "hostname": "device-name",
    "ap_mac": "00:11:22:33:44:55",
    "essid": "MyWiFi",
    "rssi": -45,
    "signal": -45,
    "_is_online": true
}
```

**Device Sync Event** (`device:sync`):
```json
{
    "mac": "00:11:22:33:44:55",
    "state": 1,
    "uptime": 1234567,
    "version": "10.0.1",
    "system-stats": {
        "cpu": "5.2",
        "mem": "45.3"
    }
}
```

---

## Go Implementation Guide

### Complete Client Structure

```go
package unifi

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "strings"
    "sync"
    "time"
)

// Client represents a UniFi API client
type Client struct {
    BaseURL    string
    Site       string
    HTTPClient *http.Client
    CSRFToken  string
    mu         sync.RWMutex
}

// Config holds client configuration
type Config struct {
    Host          string
    Port          int
    Site          string
    Username      string
    Password      string
    SkipTLSVerify bool
    Timeout       time.Duration
}

// NewClient creates a new UniFi API client
func NewClient(cfg Config) (*Client, error) {
    if cfg.Port == 0 {
        cfg.Port = 443
    }
    if cfg.Site == "" {
        cfg.Site = "default"
    }
    if cfg.Timeout == 0 {
        cfg.Timeout = 30 * time.Second
    }

    jar, err := cookiejar.New(nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create cookie jar: %w", err)
    }

    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: cfg.SkipTLSVerify,
        },
    }

    client := &Client{
        BaseURL: fmt.Sprintf("https://%s:%d", cfg.Host, cfg.Port),
        Site:    cfg.Site,
        HTTPClient: &http.Client{
            Jar:       jar,
            Transport: transport,
            Timeout:   cfg.Timeout,
        },
    }

    // Perform login
    if err := client.login(cfg.Username, cfg.Password); err != nil {
        return nil, err
    }

    return client, nil
}

// login authenticates with the controller
func (c *Client) login(username, password string) error {
    loginReq := struct {
        Username   string `json:"username"`
        Password   string `json:"password"`
        RememberMe bool   `json:"rememberMe"`
        Token      string `json:"token"`
    }{
        Username:   username,
        Password:   password,
        RememberMe: true,
        Token:      "",
    }

    body, err := json.Marshal(loginReq)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", c.BaseURL+"/api/auth/login", bytes.NewBuffer(body))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("login failed: %d - %s", resp.StatusCode, string(body))
    }

    // Store CSRF token
    if csrfToken := resp.Header.Get("X-CSRF-Token"); csrfToken != "" {
        c.mu.Lock()
        c.CSRFToken = csrfToken
        c.mu.Unlock()
    }

    return nil
}

// Logout terminates the session
func (c *Client) Logout() error {
    req, err := http.NewRequest("POST", c.BaseURL+"/proxy/network/api/logout", nil)
    if err != nil {
        return err
    }

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

// doRequest performs an HTTP request to the API
func (c *Client) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
    var reqBody io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        reqBody = bytes.NewBuffer(jsonBody)
    }

    url := fmt.Sprintf("%s/proxy/network/api/s/%s/%s", c.BaseURL, c.Site, endpoint)
    req, err := http.NewRequest(method, url, reqBody)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    // Include CSRF token for write operations
    if method != "GET" {
        c.mu.RLock()
        if c.CSRFToken != "" {
            req.Header.Set("X-CSRF-Token", c.CSRFToken)
        }
        c.mu.RUnlock()
    }

    return c.HTTPClient.Do(req)
}

// doV2Request performs a request to v2 API endpoints
func (c *Client) doV2Request(method, endpoint string, body interface{}) (*http.Response, error) {
    var reqBody io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        reqBody = bytes.NewBuffer(jsonBody)
    }

    url := fmt.Sprintf("%s/proxy/network/v2/api/site/%s/%s", c.BaseURL, c.Site, endpoint)
    req, err := http.NewRequest(method, url, reqBody)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    if method != "GET" {
        c.mu.RLock()
        if c.CSRFToken != "" {
            req.Header.Set("X-CSRF-Token", c.CSRFToken)
        }
        c.mu.RUnlock()
    }

    return c.HTTPClient.Do(req)
}

// parseResponse parses an API response
func parseResponse[T any](resp *http.Response) (*APIResponse[T], error) {
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Check for non-success status codes
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, parseError(resp.StatusCode, body)
    }

    var apiResp APIResponse[T]
    if err := json.Unmarshal(body, &apiResp); err != nil {
        return nil, err
    }

    if apiResp.Meta.RC != "ok" {
        return nil, fmt.Errorf("API error: %s", apiResp.Meta.Message)
    }

    return &apiResp, nil
}

func parseError(statusCode int, body []byte) error {
    switch statusCode {
    case 401:
        return ErrLoginRequired
    case 403:
        if bytes.Contains(body, []byte("CSRF")) {
            return ErrInvalidCSRFToken
        }
        return ErrNoPermission
    case 404:
        return ErrNotFound
    case 429:
        return ErrRateLimited
    default:
        return fmt.Errorf("API error %d: %s", statusCode, string(body))
    }
}

// normalizeMac converts MAC address to lowercase without colons
func normalizeMac(mac string) string {
    mac = strings.ToLower(mac)
    mac = strings.ReplaceAll(mac, ":", "")
    mac = strings.ReplaceAll(mac, "-", "")
    return mac
}
```

### Example Usage

```go
package main

import (
    "fmt"
    "log"

    "yourmodule/unifi"
)

func main() {
    // Create client
    client, err := unifi.NewClient(unifi.Config{
        Host:          "192.168.1.1",
        Site:          "default",
        Username:      "local_admin",
        Password:      "your_password",
        SkipTLSVerify: true, // For self-signed certs
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Logout()

    // Get site health
    health, err := client.GetSiteHealth()
    if err != nil {
        log.Fatal(err)
    }

    for _, subsystem := range health {
        fmt.Printf("%s: %s\n", subsystem.Subsystem, subsystem.Status)
    }

    // List active clients
    clients, err := client.ListActiveClients()
    if err != nil {
        log.Fatal(err)
    }

    for _, c := range clients {
        fmt.Printf("Client: %s (%s) - %s\n", c.Name, c.MAC, c.IP)
    }

    // List devices
    devices, err := client.ListDevices()
    if err != nil {
        log.Fatal(err)
    }

    for _, d := range devices {
        fmt.Printf("Device: %s (%s) - State: %d\n", d.Name, d.MAC, d.State)
    }

    // Restart a device
    err = client.RestartDevice("00:11:22:33:44:55")
    if err != nil {
        log.Printf("Failed to restart device: %v", err)
    }

    // Block a client
    err = client.BlockClient("aa:bb:cc:dd:ee:ff")
    if err != nil {
        log.Printf("Failed to block client: %v", err)
    }
}
```

### Cloud API Client

```go
package unifi

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// CloudClient represents a UniFi Site Manager API client
type CloudClient struct {
    APIKey     string
    HTTPClient *http.Client
    BaseURL    string
}

// NewCloudClient creates a new Site Manager API client
func NewCloudClient(apiKey string) *CloudClient {
    return &CloudClient{
        APIKey:  apiKey,
        BaseURL: "https://api.ui.com",
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *CloudClient) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
    var reqBody io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        reqBody = bytes.NewBuffer(jsonBody)
    }

    req, err := http.NewRequest(method, c.BaseURL+endpoint, reqBody)
    if err != nil {
        return nil, err
    }

    req.Header.Set("X-API-KEY", c.APIKey)
    req.Header.Set("Accept", "application/json")
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    return c.HTTPClient.Do(req)
}

// ListHosts retrieves all hosts
func (c *CloudClient) ListHosts() ([]Host, error) {
    resp, err := c.doRequest("GET", "/v1/hosts", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Data []Host `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Data, nil
}

// ListSites retrieves all sites
func (c *CloudClient) ListSites(hostIDs ...string) ([]Site, error) {
    endpoint := "/v1/sites"
    if len(hostIDs) > 0 {
        endpoint += "?hostIds=" + strings.Join(hostIDs, ",")
    }

    resp, err := c.doRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Data []Site `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Data, nil
}

// ListDevices retrieves all devices
func (c *CloudClient) ListDevices(hostIDs, siteIDs []string) ([]CloudDevice, error) {
    endpoint := "/v1/devices"
    params := url.Values{}

    if len(hostIDs) > 0 {
        params.Set("hostIds", strings.Join(hostIDs, ","))
    }
    if len(siteIDs) > 0 {
        params.Set("siteIds", strings.Join(siteIDs, ","))
    }

    if len(params) > 0 {
        endpoint += "?" + params.Encode()
    }

    resp, err := c.doRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Data []CloudDevice `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Data, nil
}
```

---

## Additional Notes

### MAC Address Format

- **Internal format**: Lowercase, no separators (`aabbccddeeff`)
- **Display format**: Colon-separated, lowercase (`aa:bb:cc:dd:ee:ff`)
- Always normalize MAC addresses before sending to the API

### Common HTTP Headers

| Header | Value | When Used |
|--------|-------|-----------|
| `Content-Type` | `application/json` | All requests with body |
| `Accept` | `application/json` | All requests |
| `X-CSRF-Token` | Token from login | Write operations (POST/PUT/DELETE) |
| `X-API-KEY` | API key | Cloud API requests |
| `Cookie` | Session cookie | Local API requests |

### SSL/TLS Considerations

- UDM Pro uses self-signed certificates by default
- Either skip verification or add the certificate to trusted roots
- Production environments should use proper certificates

### Important Caveats

1. **Undocumented API**: Most local endpoints are undocumented and may change between versions
2. **Rate limiting**: The local API can be overwhelmed; implement reasonable polling intervals
3. **CSRF tokens**: Some operations require CSRF tokens; always include when performing write operations
4. **Local admin accounts**: Required for API access after July 2024 MFA changes
5. **v2 API differences**: PUT operations return 201 (not 200) on success
6. **Site Manager API**: Currently read-only; write operations will require API key updates

---

## Reference Implementations

These community-maintained client libraries provide excellent references for understanding the API structure and can be used to cross-check endpoint behavior.

### PHP

| Project | URL | Notes |
|---------|-----|-------|
| **Art-of-WiFi/UniFi-API-client** | https://github.com/Art-of-WiFi/UniFi-API-client | Most comprehensive PHP client (2.7k+ stars). Actively maintained, supports UniFi OS 3.x-5.x and Network Application 5.x-10.x. Excellent source code documentation. |
| **Art-of-WiFi/UniFi-API-browser** | https://github.com/Art-of-WiFi/UniFi-API-browser | Web-based tool to explore API endpoints interactively. Useful for discovering available data. |

### Python

| Project | URL | Notes |
|---------|-----|-------|
| **unificontrol** | https://github.com/nickovs/unificontrol | High-level Python interface with good documentation. Supports controller versions 5.x-6.x. |
| **aiounifi** | https://github.com/Kane610/aiounifi | Async Python client used by Home Assistant. Production-tested, well-maintained. |
| **pyunifi** | https://github.com/finish06/pyunifi | Simple synchronous client, good for basic scripting. |
| **unifi-controller-api** | https://github.com/tnware/unifi-controller-api | Modern Python client with type hints and dataclasses. |

### TypeScript / JavaScript / Node.js

| Project | URL | Notes |
|---------|-----|-------|
| **unifi-client** | https://github.com/thib3113/unifi-client | Full-featured TypeScript client with comprehensive type definitions. Excellent reference for API structures. Documentation: https://thib3113.github.io/unifi-client/ |
| **node-unifi** | https://github.com/jens-maus/node-unifi | Popular Node.js client (400+ stars). Well-documented, supports UDM/UDM Pro. |
| **ubnt-unifi** | https://github.com/hobbyquaker/ubnt-unifi | Lightweight Node.js module with WebSocket event support. |

### Go

| Project | URL | Notes |
|---------|-----|-------|
| **unpoller/unifi** | https://github.com/unpoller/unifi | Go client from the UniFi Poller project. Production-tested for metrics collection. |
| **paultyng/go-unifi** | https://github.com/paultyng/go-unifi | Go client used by the Terraform UniFi Provider. Auto-generated from API exploration. |
| **dim13/unifi** | https://github.com/dim13/unifi | Minimal Go client for basic operations. |

### Other Languages

| Project | Language | URL | Notes |
|---------|----------|-----|-------|
| **Unifi.Net** | C# | https://github.com/schwoi/Unifi.Net | .NET client library |
| **unifi_ruby** | Ruby | https://github.com/dewski/unifi | Ruby client |

### Infrastructure & Automation Tools

| Project | URL | Notes |
|---------|-----|-------|
| **Terraform UniFi Provider** | https://github.com/paultyng/terraform-provider-unifi | Infrastructure as code for UniFi. Good reference for REST endpoint usage patterns. |
| **Home Assistant UniFi Integration** | https://github.com/home-assistant/core/tree/dev/homeassistant/components/unifi | Production integration used by thousands. Uses aiounifi under the hood. |
| **UniFi Poller** | https://github.com/unpoller/unpoller | Prometheus/InfluxDB metrics exporter. Go implementation with comprehensive device/client polling. |

---

## References

### Official Documentation

- [Ubiquiti Help Center - Getting Started with the Official UniFi API](https://help.ui.com/hc/en-us/articles/30076656117655-Getting-Started-with-the-Official-UniFi-API)
- [UniFi Developer Portal - Site Manager API](https://developer.ui.com/site-manager-api/gettingstarted)
- [UniFi Developer Portal - List Hosts](https://developer.ui.com/site-manager-api/list-hosts)
- [UniFi Developer Portal - List Sites](https://developer.ui.com/site-manager-api/list-sites)
- [UniFi Developer Portal - List Devices](https://developer.ui.com/site-manager-api/list-devices)
- [UniFi Developer Portal - ISP Metrics](https://developer.ui.com/site-manager-api/get-isp-metrics)

### Community Documentation

- [Ubiquiti Community Wiki - UniFi Controller API](https://ubntwiki.com/products/software/unifi-controller/api)
- [UniFi Controller API Professional Guide](https://github.com/uchkunrakhimow/unifi-controller-api-professional-guide)
- [UniFi Best Practices - API Guide](https://github.com/uchkunrakhimow/unifi-best-practices)
- [Nikhil's UniFi Notes](https://wiki.nikhil.io/Unifi_Notes/)
- [Art of WiFi Blog - Local Admin for API](https://artofwifi.net/blog/use-local-admin-account-unifi-api-captive-portal)

### Community Forums

- [Ubiquiti Community Forums](https://community.ui.com/)
- [Reddit r/Ubiquiti](https://www.reddit.com/r/Ubiquiti/)
