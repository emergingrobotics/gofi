# gofimac Security Review

**Date**: 2026-02-23
**Reviewer**: AI Security Audit (Revised)
**Scope**: `utilities/gofimac/` (main.go, oui.go, operations.go, format.go) and supporting gofi library transport layer

---

## Findings Summary

| Severity | Count |
|----------|-------|
| Critical | 2     |
| High     | 3     |
| Medium   | 6     |
| Low      | 5     |

---

## CRITICAL SEVERITY

### C-1: No Response Body Size Limit on OUI Download

**File**: `utilities/gofimac/oui.go:170`
**OWASP**: A05:2021 Security Misconfiguration

The OUI download uses `io.Copy(file, response.Body)` with no size limit. A compromised or MITM'd server could serve an arbitrarily large response, exhausting disk space. The real IEEE OUI file is approximately 4 MB.

```go
bytesWritten, err := io.Copy(file, response.Body)
```

**Recommendation**: Use `io.LimitReader(response.Body, maxOUIFileSize)` with a limit around 50 MB.

```go
const maxOUIFileSize = 50 * 1024 * 1024
limitedBody := io.LimitReader(response.Body, maxOUIFileSize)
bytesWritten, err := io.Copy(file, limitedBody)
```

### C-2: No Response Body Size Limit on UDM API Responses

**File**: `transport/transport.go:149`
**OWASP**: A05:2021 Security Misconfiguration

The gofi transport layer reads the entire response body into memory with no limit:

```go
body, err := io.ReadAll(httpResp.Body)
```

A malicious or compromised UDM could return an arbitrarily large response, causing OOM. This affects every API call gofimac makes.

**Recommendation**: Use `io.LimitReader` on `httpResp.Body` before `ReadAll`, with a configurable maximum (e.g., 64 MB).

---

## HIGH SEVERITY

### H-1: Temporary File Created with Default umask Permissions

**File**: `utilities/gofimac/oui.go:165`
**OWASP**: A01:2021 Broken Access Control

The temporary OUI file is created with `os.Create()`, which uses mode `0666` masked by umask. On systems with permissive umask, this file is world-writable. A world-writable temp file in a predictable location allows an attacker to replace its contents before the atomic rename.

```go
temporaryPath := databasePath + ".tmp"
file, err := os.Create(temporaryPath)
```

**Recommendation**: Use `os.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)`.

### H-2: Predictable Temporary File Path Enables Symlink Attack

**File**: `utilities/gofimac/oui.go:164`
**OWASP**: A01:2021 Broken Access Control

The temporary file path is always `<databasePath>.tmp`. An attacker with write access to the same directory could pre-create a symlink at this path, causing the tool to write downloaded content to an arbitrary file location.

```go
temporaryPath := databasePath + ".tmp"
```

**Recommendation**: Use `os.CreateTemp(directory, "oui-*.tmp")` to generate a random filename, then `os.Rename` to the final path.

### H-3: XDG_DATA_HOME Not Validated for Path Safety

**File**: `utilities/gofimac/oui.go:101-109`
**OWASP**: A03:2021 Injection

The `XDG_DATA_HOME` environment variable is used directly in path construction without any validation. A malicious value could direct file writes to unexpected locations (e.g., `XDG_DATA_HOME=/etc` writes to `/etc/gofimac/oui.txt`).

```go
dataHome := os.Getenv("XDG_DATA_HOME")
// ...
dataHome = filepath.Join(dataHome, "gofimac")
```

While environment variables are generally user-controlled, defense-in-depth requires validation. `filepath.Join` normalizes paths but does not prevent writing to system directories.

**Recommendation**: Validate that `XDG_DATA_HOME` is an absolute path and does not contain `..` components.

---

## MEDIUM SEVERITY

### M-1: OUI HTTP Client Follows Redirects Without Restriction

**File**: `utilities/gofimac/oui.go:147`
**OWASP**: A07:2021 Identification and Authentication Failures

The OUI download uses a default `http.Client` which follows up to 10 redirects. An open redirect on the IEEE server (or DNS poisoning) could redirect to a malicious host that serves a crafted response.

Note: The gofi transport correctly disables redirect following (`transport/transport.go:80-83`), but the OUI download HTTP client does not.

```go
httpClient := &http.Client{Timeout: ouiDownloadTimeout}
```

**Recommendation**: Add a `CheckRedirect` function that restricts redirects to the same host or limits the count, and enforce TLS 1.2 minimum.

```go
httpClient := &http.Client{
    Timeout: ouiDownloadTimeout,
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        if len(via) >= 3 {
            return fmt.Errorf("too many redirects")
        }
        return nil
    },
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
    },
}
```

### M-2: No Host or Port Input Validation

**File**: `utilities/gofimac/main.go:78-83`
**OWASP**: A03:2021 Injection

The `--host` value is passed to the gofi client without validation. A malicious host value containing path components or special characters could cause unexpected URL construction. The `--port` flag accepts negative values.

```go
if *host == "" {
    *host = os.Getenv(envUDMIP)
}
```

**Recommendation**: Validate host is a valid IP address or hostname (`[a-zA-Z0-9._:-]+`). Validate port is in range 1-65535.

### M-3: Site Name Not Validated

**File**: `utilities/gofimac/main.go:22`
**OWASP**: A03:2021 Injection

The `--site` parameter is used in URL path construction (`/api/s/{site}/...`). If the site name contains path traversal characters (`../`), it could alter the API path.

**Recommendation**: Validate site name matches `^[a-zA-Z0-9_-]+$`.

### M-4: Terminal Escape Injection via Client Hostnames

**File**: `utilities/gofimac/operations.go:115-123`, `utilities/gofimac/format.go:23`
**OWASP**: A03:2021 Injection

Client `Name` and `Hostname` from the UDM API are passed through to text output without sanitization. A malicious hostname containing ANSI escape sequences could execute terminal injection attacks (cursor manipulation, arbitrary text overwrite, or clipboard injection in vulnerable terminals).

```go
func resolveClientHostname(client types.Client) string {
    if client.Name != "" {
        return client.Name
    }
```

JSON output is safe because Go's `json.Encoder` escapes control characters.

**Recommendation**: Strip or escape control characters (bytes < 0x20 and 0x7F) from all string fields before text output.

### M-5: OUI Data Directory Created with World-Readable Permissions

**File**: `utilities/gofimac/oui.go:143`

```go
if err := os.MkdirAll(directory, 0o755); err != nil {
```

The data directory is created with world-readable and world-executable permissions. While OUI data is public, tighter permissions (`0o700`) are more appropriate for user data directories following the principle of least privilege.

### M-6: No Downloaded Content Format Validation Before Persist

**File**: `utilities/gofimac/oui.go:164-187`
**OWASP**: A08:2021 Software and Data Integrity Failures

The downloaded file is persisted without any format validation. Only empty downloads are rejected. An attacker who compromises the download could inject arbitrary content that would be cached for up to 30 days.

**Recommendation**: After download but before the atomic rename, open the temp file and verify it contains a minimum number of `(hex)` markers (e.g., at least 1000 entries expected). Alternatively, validate the file can be parsed by `ParseOUIDatabase` before committing.

---

## LOW SEVERITY

### L-1: No Context Timeout for Overall API Operations

**File**: `utilities/gofimac/main.go:116`

```go
ctx := context.Background()
```

No overall timeout wraps the API operations. While the transport has a 30s per-request timeout, the overall program could hang if the UDM responds slowly but within per-request limits across many requests.

**Recommendation**: Use `context.WithTimeout(context.Background(), 2*time.Minute)`.

### L-2: OUI Parser Has No Entry Count Limit

**File**: `utilities/gofimac/oui.go:66-97`

The parser reads all entries into memory with no cap. A crafted OUI file with millions of `(hex)` lines could exhaust memory. The real database has approximately 40,000 entries.

**Recommendation**: Add a `maxOUIEntries = 200000` guard.

### L-3: No File Locking on OUI Database Operations

**File**: `utilities/gofimac/oui.go:114-189`

Multiple concurrent gofimac invocations could race on the OUI database. The atomic rename mitigates corruption, but simultaneous downloads waste bandwidth and could interact poorly.

**Recommendation**: Use advisory file locking (`flock`) on a lockfile during download.

### L-4: Fallback to Current Directory When Home Directory Unknown

**File**: `utilities/gofimac/oui.go:103-105`

```go
homeDirectory, err := os.UserHomeDir()
if err != nil {
    homeDirectory = "."
}
```

If `os.UserHomeDir()` fails, the fallback to `.` creates the OUI database in whatever the current working directory is. This is unpredictable and could write to unexpected locations.

**Recommendation**: Return an error instead of silently falling back.

### L-5: FormatJSON Emits `null` for Nil Slice

**File**: `utilities/gofimac/format.go:31-34`, `utilities/gofimac/operations.go:61`

If no clients match the filter, `ListClients` returns a nil slice. `json.Encode(nil)` outputs `null`, not `[]`. The spec requires empty JSON array `[]` for no clients.

```go
var entries []ClientEntry  // nil, not initialized
```

**Recommendation**: Initialize as `entries := []ClientEntry{}` or guard in `FormatJSON`.

---

## POSITIVE FINDINGS

These security practices are correctly implemented:

1. **Credentials via environment variables only** (`main.go:86-93`). Passwords are never accepted as CLI flags, preventing exposure in `/proc/*/cmdline`.
2. **Credentials never logged or printed**. Error messages reference variable names, not values.
3. **HTTPS for OUI download** (`oui.go:15`). Hardcoded `https://` URL.
4. **HTTPS for UDM connection**. Default port 443 with TLS.
5. **CSRF token handling** (`transport/transport.go:131-134`). Uses `atomic.Value` for thread-safe storage.
6. **Atomic file replacement** (`oui.go:184`). `os.Rename` prevents partial reads of OUI database.
7. **HTTP timeout on OUI download** (`oui.go:22`). 60-second timeout prevents indefinite hangs.
8. **HTTP timeout on transport** (`transport/config.go:67`). 30-second default.
9. **Redirect suppression on UDM transport** (`transport/transport.go:80-83`). `http.ErrUseLastResponse` prevents redirect-following attacks on authenticated sessions.
10. **Empty download detection** (`oui.go:179-182`). Zero-byte downloads rejected.
11. **Graceful stale cache fallback** (`oui.go:129-133`). Download failure with existing cache prints warning and continues.
12. **JSON encoding uses stdlib** (`format.go:32-34`). Proper escaping, no injection vectors.
13. **Cleanup on write errors** (`oui.go:175`). Temporary file removed if write fails.
14. **No user-controlled URLs**. OUI URL is a constant; no SSRF vector.

---

## OWASP TOP 10 (2021) MAPPING

| Category | Status | Findings |
|----------|--------|----------|
| A01: Broken Access Control | HIGH | H-1, H-2 (file permissions, symlink) |
| A02: Cryptographic Failures | PASS | HTTPS used, no crypto misuse |
| A03: Injection | MEDIUM | M-2, M-3, M-4 (host/site/terminal injection) |
| A04: Insecure Design | PASS | Sound architecture |
| A05: Security Misconfiguration | CRITICAL | C-1, C-2 (unbounded reads) |
| A06: Vulnerable Components | N/A | Not audited (gofi library itself) |
| A07: Auth Failures | MEDIUM | M-1 (redirect following on OUI client) |
| A08: Data Integrity Failures | MEDIUM | M-6 (no content validation) |
| A09: Logging/Monitoring | PASS | Appropriate for CLI tool |
| A10: SSRF | PASS | No user-controlled URLs |

---

## RECOMMENDATIONS PRIORITY

### Immediate (before release)
1. C-1: Add `io.LimitReader` to OUI download
2. C-2: Add `io.LimitReader` to transport `ReadAll` (gofi library fix)
3. H-1: Use explicit file permissions on temp file
4. H-2: Use `os.CreateTemp` for random temp filename
5. H-3: Validate `XDG_DATA_HOME`

### Next iteration
6. M-1: Restrict redirects and enforce TLS 1.2 on OUI HTTP client
7. M-2: Validate host and port inputs
8. M-3: Validate site name
9. M-4: Strip control characters from text output
10. M-5: Use `0o700` for data directory
11. M-6: Validate OUI format before persisting

### When convenient
12. L-1 through L-5: Context timeout, entry count limit, file locking, home dir fallback, nil slice

---

## COMPLIANCE

### CLAUDE.md Security Rules

| Rule | Status |
|------|--------|
| Never commit API keys, passwords, or credentials | PASS |
| Validate all external inputs at system boundaries | PARTIAL (M-2, M-3, M-4, M-6) |
| Use parameterized queries for database access | N/A |
| HTTPS for all external API calls | PASS |
| No silenced warnings or linter ignores | PASS |

**Overall Risk Rating**: MEDIUM-HIGH (due to unbounded read vulnerabilities)
