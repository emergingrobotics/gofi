# Security Review: gofips Implementation

**Review Date:** 2026-02-23
**Reviewed Files:**
- `/er/gofi/utilities/gofips/main.go`
- `/er/gofi/utilities/gofips/parse.go`
- `/er/gofi/utilities/gofips/format.go`
- `/er/gofi/utilities/gofips/operations.go`

**Scope:** Security audit against OWASP Top 10 and domain-specific risks including input validation, credential handling, file operations, error disclosure, DNS manipulation, command injection, resource exhaustion, and TLS handling.

---

## Executive Summary

The gofips implementation demonstrates generally sound security practices with rigorous input validation and proper credential handling. However, several findings require attention, particularly around path traversal risks, error message information disclosure, and DNS record manipulation safety.

**Critical Findings:** 0
**High Findings:** 1
**Medium Findings:** 4
**Low Findings:** 3
**Informational:** 2

---

## Findings by Severity

### HIGH SEVERITY

#### H-1: Path Traversal Vulnerability in File Input (main.go:211-215)

**Location:** `main.go`, `parseSetInput()` function, lines 211-215

**Issue:**
```go
if len(args) > 0 {
    f, err := os.Open(args[0])
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
```

The file path from command-line arguments is passed directly to `os.Open()` without validation. An attacker can specify arbitrary paths including:
- Absolute paths to sensitive files: `/etc/shadow`, `/root/.ssh/id_rsa`
- Relative paths with traversal: `../../../etc/passwd`
- Symlink attacks to read files the user has access to

**Impact:**
- Unauthorized file disclosure if the user running gofips has read access
- Information leakage of system configuration
- Potential credential exposure if sensitive files are readable

**Recommendation:**
1. Validate that the provided path does not contain `..` components
2. Resolve the path to its absolute form and verify it's within expected directories
3. Check that the resolved path is a regular file (not a symlink, device, etc.)
4. Consider restricting input to current directory only or implementing a whitelist

**Example Mitigation:**
```go
func validateInputPath(path string) error {
    if strings.Contains(path, "..") {
        return fmt.Errorf("path traversal detected")
    }

    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("invalid path: %w", err)
    }

    info, err := os.Lstat(absPath)  // Use Lstat to detect symlinks
    if err != nil {
        return err
    }

    if info.Mode()&os.ModeSymlink != 0 {
        return fmt.Errorf("symlinks not allowed")
    }

    if !info.Mode().IsRegular() {
        return fmt.Errorf("path must be a regular file")
    }

    return nil
}
```

---

### MEDIUM SEVERITY

#### M-1: Error Messages Leak Internal System Information (multiple locations)

**Locations:**
- `main.go:164` - `"failed to create client: " + err.Error()`
- `main.go:169` - `"failed to connect: " + err.Error()`
- `operations.go:34` - `"failed to list users: %w"`
- `operations.go:92` - `"failed to list networks: %w"`
- `operations.go:217` - `"failed to update user: %w"`

**Issue:**
Error messages directly expose internal API errors, which may contain:
- Internal IP addresses and network topology
- Database error messages with schema information
- API endpoint paths and structure
- Authentication state details

**Example Scenarios:**
```
Error: failed to connect: dial tcp 10.0.1.5:443: connection refused
Error: failed to list users: API error 401: invalid session token abc123xyz
Error: failed to create client: x509: certificate signed by unknown authority (CN=internal-ca.corp.local)
```

**Impact:**
- Information disclosure aids reconnaissance for attackers
- Reveals internal network structure
- Exposes authentication mechanisms
- May leak session tokens or credentials in error context

**Recommendation:**
1. Create generic error messages for end users
2. Log detailed errors to stderr or a log file with appropriate permissions
3. Sanitize error messages to remove sensitive information
4. Use error codes instead of exposing internal error strings

**Example:**
```go
func safeError(userMsg string, internalErr error) error {
    log.Printf("Internal error: %v", internalErr)  // Log detailed error
    return fmt.Errorf("%s (see logs for details)", userMsg)  // Return sanitized error
}

// Usage:
if err := client.Connect(ctx); err != nil {
    exitError(safeError("connection failed", err).Error())
}
```

---

#### M-2: DNS Record Manipulation Without Ownership Verification (operations.go:382-404)

**Location:** `operations.go`, `ensureDNSRecord()` function

**Issue:**
```go
func ensureDNSRecord(ctx context.Context, client gofi.Client, site, hostname, ip string) error {
    existing, _ := client.DNS().GetByName(ctx, site, hostname)
    if existing != nil {
        if existing.Value == ip {
            return nil
        }
        existing.Value = ip
        if _, err := client.DNS().Update(ctx, site, existing); err != nil {
            return fmt.Errorf("failed to update DNS record: %w", err)
        }
        return nil
    }
```

The function updates existing DNS records without verifying ownership or checking if the record was created by gofips. This allows:
- Hijacking existing DNS records
- Overwriting manually created records
- DNS spoofing within the network

**Attack Scenario:**
1. Attacker creates entry for `host adminpanel { ... fixed-address 10.0.1.200; }`
2. If `adminpanel` DNS record exists pointing to legitimate server `10.0.1.10`, it gets silently overridden to `10.0.1.200`
3. Internal services now resolve `adminpanel` to attacker-controlled IP

**Impact:**
- DNS hijacking within the network
- Man-in-the-middle attacks on internal services
- Disruption of legitimate services

**Recommendation:**
1. Add metadata to DNS records created by gofips (e.g., comment field)
2. Require `--force` flag to override existing DNS records
3. Warn users when DNS records already exist with different IPs
4. Implement dry-run mode that reports conflicts without making changes

**Example:**
```go
func ensureDNSRecord(ctx context.Context, client gofi.Client, site, hostname, ip string, force bool) error {
    existing, _ := client.DNS().GetByName(ctx, site, hostname)
    if existing != nil {
        if existing.Value == ip {
            return nil
        }

        // Check if record was created by gofips
        if !strings.Contains(existing.Comment, "gofips-managed") && !force {
            return fmt.Errorf("DNS record %s exists (points to %s), use --force to override", hostname, existing.Value)
        }

        existing.Value = ip
        if _, err := client.DNS().Update(ctx, site, existing); err != nil {
            return fmt.Errorf("failed to update DNS record: %w", err)
        }
        return nil
    }

    record := &types.DNSRecord{
        Key:        hostname,
        Value:      ip,
        RecordType: types.DNSRecordTypeA,
        Enabled:    true,
        Comment:    "gofips-managed",  // Mark for future identification
    }
    // ... rest of creation logic
}
```

---

#### M-3: Potential Integer Overflow in IP Sorting (format.go:92-103)

**Location:** `format.go`, `ipToUint32()` function

**Issue:**
```go
func ipToUint32(ipStr string) uint32 {
    ip := net.ParseIP(ipStr)
    if ip == nil {
        return 0
    }
    ip4 := ip.To4()
    if ip4 == nil {
        return 0
    }
    return binary.BigEndian.Uint32(ip4)
}
```

The function returns `0` for invalid IPs, causing all invalid IPs to sort to the beginning. While this doesn't cause a traditional integer overflow, it creates a potential denial-of-service through sort instability if large numbers of invalid IPs are processed.

**Impact:**
- Incorrect sorting of output (low impact)
- Potential resource exhaustion with malicious input containing many invalid IPs
- Sort algorithm instability could cause excessive CPU usage

**Recommendation:**
1. Return an error or validate IPs before sorting
2. Filter invalid IPs during parsing rather than handling them during sort
3. Document that invalid IPs sort to position 0

**Note:** This is rated Medium because the parser already validates IPs, so this code path should not execute with valid input. However, it represents defensive coding weakness.

---

#### M-4: Credentials Logged in Error Context (main.go:153-165)

**Location:** `main.go`, client configuration

**Issue:**
While credentials are properly read from environment variables and not directly logged, the `gofi.Config` struct is passed to `gofi.New()`, which may log or expose it in error messages:

```go
config := &gofi.Config{
    Host:          *host,
    Port:          *port,
    Username:      username,
    Password:      password,
    Site:          *site,
    SkipTLSVerify: *insecure,
}

client, err := gofi.New(config)
if err != nil {
    exitError("failed to create client: " + err.Error())  // May expose config details
}
```

**Impact:**
- Credentials may appear in error messages if gofi.New() fails
- Credentials may be logged by the gofi library
- Process memory dumps could expose the config struct

**Recommendation:**
1. Review gofi library to ensure it doesn't log credentials
2. Zero out password after authentication succeeds
3. Implement `String()` method on config struct that redacts sensitive fields
4. Use secure string types that clear memory

---

### LOW SEVERITY

#### L-1: Hostname Validation Accepts Underscore (parse.go:219)

**Location:** `parse.go`, `isDNSSafe()` function, line 219

**Issue:**
```go
if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '_') {
    return false
}
```

The validation accepts underscore (`_`) in hostnames. While RFC 952 forbids underscores in hostnames, RFC 2181 relaxed this for DNS labels. However, many systems still reject underscores, and this can cause interoperability issues.

**Impact:**
- Hostnames may be rejected by some DNS servers
- Interoperability issues with strict RFC 952 implementations
- Inconsistent behavior across network infrastructure

**Recommendation:**
1. Remove underscore from allowed characters for strict RFC 952 compliance
2. Or add a flag `--allow-underscore` if underscore support is desired
3. Document the hostname validation policy clearly

**Note:** This is marked Low because it's more of an interoperability concern than a security issue, though DNS inconsistencies can be exploited.

---

#### L-2: No Rate Limiting on Bulk Operations

**Location:** `operations.go`, `DoSet()` function

**Issue:**
The `DoSet()` function processes all entries in a tight loop without rate limiting:

```go
for _, e := range entries {
    // ... process entry ...
    if _, err := client.Users().Create(ctx, site, newUser); err != nil {
        // handle error
    }
    if err := ensureDNSRecord(ctx, client, site, e.Hostname, e.IP); err != nil {
        // handle error
    }
}
```

Processing thousands of entries rapidly could:
- Trigger API rate limits, causing bulk import to fail partway through
- Overwhelm the UDM controller
- Be flagged as anomalous behavior

**Impact:**
- Operational failure on large imports
- Potential temporary API lockout
- Reduced reliability

**Recommendation:**
1. Add configurable rate limiting between API calls
2. Implement exponential backoff on errors
3. Add progress reporting for large batches
4. Consider batch API calls if the UniFi API supports them

---

#### L-3: No Maximum File Size Check

**Location:** `main.go`, `parseSetInput()` and `parse.go`, `Parse()` function

**Issue:**
The parser reads files of arbitrary size into memory without limit:

```go
scanner := bufio.NewScanner(r)
// ... reads entire file line by line
```

While `bufio.Scanner` has built-in line length limits, there's no check on total file size or entry count.

**Impact:**
- Memory exhaustion with extremely large files
- Denial of service through resource exhaustion
- System instability

**Recommendation:**
1. Add maximum file size check (e.g., 10MB)
2. Add maximum entry count (e.g., 10,000 hosts)
3. Stream processing for very large imports
4. Clear error messages when limits are exceeded

**Example:**
```go
const (
    maxFileSize    = 10 * 1024 * 1024  // 10MB
    maxEntryCount  = 10000
)

func parseSetInput(args []string) (*ParseResult, error) {
    var r io.Reader
    if len(args) > 0 {
        info, err := os.Stat(args[0])
        if err != nil {
            return nil, err
        }
        if info.Size() > maxFileSize {
            return nil, fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxFileSize)
        }
        // ... rest of function
    }
    // ... parsing logic with entry count check
}
```

---

### INFORMATIONAL

#### I-1: TLS Certificate Verification Can Be Disabled

**Location:** `main.go`, line 26 and line 159

**Issue:**
The `--insecure` flag allows skipping TLS certificate verification:

```go
insecure = flag.Bool("insecure", false, "Skip TLS certificate verification")
// ...
SkipTLSVerify: *insecure,
```

**Context:**
This is a common and sometimes necessary feature for self-signed certificates in private networks. However, it should be used with caution.

**Recommendation:**
1. Document that `--insecure` should only be used in trusted networks
2. Warn users when `--insecure` is enabled
3. Consider supporting certificate pinning as an alternative
4. Log when insecure mode is used for audit trail

---

#### I-2: Dry Run Mode Not Implemented for All Operations

**Location:** `main.go`, line 36 and various operation functions

**Issue:**
The `--dry-run` flag is parsed but only implemented for `DoSet()`. It's not implemented for:
- `DoAdd()` - should show what would be created
- `DoDel()` - should show what would be deleted

**Impact:**
- Inconsistent user experience
- Users may accidentally modify data in add/del operations
- Reduced safety for testing commands

**Recommendation:**
Implement dry-run mode for all operations that modify data. This is a usability and safety feature more than a security issue.

---

## Security Strengths

The following security practices are implemented correctly:

### ✓ Input Validation (High Quality)

1. **Hostname Validation** (`parse.go:199-224`): Comprehensive DNS hostname validation including:
   - Length checks (253 total, 63 per label)
   - Character whitelist
   - Leading/trailing character restrictions
   - Empty label detection

2. **MAC Address Validation** (`parse.go:25, 102`): Strict regex validation
   - Format: `^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`
   - Normalized to lowercase

3. **IP Address Validation** (`parse.go:119-127`): Uses `net.ParseIP()` with IPv4-only enforcement
   - Rejects IPv6 addresses
   - Rejects malformed IPs

4. **Duplicate Detection** (`parse.go:169-196`): Checks for duplicate hostnames, MACs, and IPs within input

### ✓ Credential Handling (Correct)

1. **Environment Variables** (`main.go:108-115`): Credentials read from environment, never from command-line arguments
2. **No Credential Logging**: Credentials are not printed in normal output
3. **Secure Transmission**: Credentials passed to gofi client library for HTTPS transmission

### ✓ Parsing Safety

1. **Semicolon Enforcement** (`parse.go:97, 115`): Syntax validation prevents malformed input
2. **Block Closure Tracking** (`parse.go:138-139`): Detects unclosed blocks
3. **Error Accumulation** (`parse.go:142-144`): All parse errors reported before processing

### ✓ No Command Injection Vectors

The tool does not execute shell commands or pass unsanitized input to system calls. All operations are through the gofi API client library.

---

## OWASP Top 10 Coverage

### A01:2021 - Broken Access Control
**Status:** Low Risk
User authentication is handled by the gofi library and UDM controller. The tool requires valid credentials.

### A02:2021 - Cryptographic Failures
**Status:** Medium Risk
TLS can be disabled with `--insecure` flag (see I-1). Otherwise, HTTPS is enforced.

### A03:2021 - Injection
**Status:** Low Risk
No SQL, command, or code injection vectors identified. Input is validated and used through typed API calls.

### A04:2021 - Insecure Design
**Status:** Medium Risk
DNS record override without verification (M-2) is a design flaw.

### A05:2021 - Security Misconfiguration
**Status:** Low Risk
Default configuration is secure. TLS is enabled by default.

### A06:2021 - Vulnerable and Outdated Components
**Status:** Cannot Assess
Requires review of `github.com/unifi-go/gofi` dependency for known vulnerabilities.

### A07:2021 - Identification and Authentication Failures
**Status:** Low Risk
Authentication is delegated to UDM controller. Credentials are properly protected.

### A08:2021 - Software and Data Integrity Failures
**Status:** Low Risk
No code execution or deserialization vulnerabilities.

### A09:2021 - Security Logging and Monitoring Failures
**Status:** Medium Risk
Limited audit logging. No tracking of who makes changes or when.

### A10:2021 - Server-Side Request Forgery (SSRF)
**Status:** Low Risk
Host is user-specified but used only for intended UDM API connection. No SSRF vector.

---

## Recommendations Summary

### Immediate Actions (High Priority)

1. **Fix H-1**: Implement path traversal validation for file input
2. **Fix M-2**: Add DNS record ownership verification and require --force for overrides
3. **Fix M-1**: Sanitize error messages to prevent information disclosure

### Short-Term Actions (Medium Priority)

4. **Fix M-4**: Ensure credentials cannot leak through error messages
5. **Fix M-3**: Add explicit IP validation before sorting
6. **Fix L-2**: Implement rate limiting for bulk operations
7. **Fix L-3**: Add file size and entry count limits

### Long-Term Improvements

8. Implement comprehensive audit logging
9. Add dry-run support for all operations
10. Consider certificate pinning for TLS
11. Add security scanning to CI/CD pipeline
12. Conduct penetration testing of complete workflow

---

## Conclusion

The gofips implementation demonstrates mature security practices in input validation and credential handling. The most critical issue is the path traversal vulnerability in file input (H-1), which should be addressed immediately. DNS record manipulation safety (M-2) and error message sanitization (M-1) should also be prioritized.

Overall security posture: **Good with areas for improvement**.

The codebase follows the principle of "do not write forgiving code" and validates inputs strictly, which aligns with the project's security requirements. With the recommended fixes, this tool will be production-ready from a security perspective.

---

**Reviewer Notes:**
- This review covers static analysis of the Go code
- Dynamic testing and fuzzing recommended for complete coverage
- Review should be updated when dependencies are updated
- The gofi library itself should undergo separate security review
