# Examples Changes Summary

## Overview

Updated three fixed IP management examples to:
1. Automatically manage DNS records when adding/removing fixed IPs
2. Support `UNIFI_UDM_IP` environment variable for controller host address

## Modified Files

### 1. examples/addfixedip/main.go

**Changes:**
- Added `UNIFI_UDM_IP` environment variable support
  - If `-H` flag not provided, checks for `UNIFI_UDM_IP` environment variable
  - Updated usage text to document this feature
  - Updated error message when host is missing

- Added automatic DNS record creation
  - New function `createDNSRecord()` that:
    - Checks if DNS record already exists for the IP
    - Updates existing record if hostname differs
    - Creates new DNS record if none exists
    - Creates A records with hostname -> IP mapping
  - Called automatically after fixed IP assignment succeeds
  - Shows warning if DNS creation fails but continues (fixed IP still created)
  - Displays DNS record info: `DNS: hostname -> IP`

**Usage:**
```bash
# With environment variable
export UNIFI_UDM_IP=192.168.4.1
bin/examples/addfixedip -m aa:bb:cc:dd:ee:ff -i 192.168.4.35 -n wone -k

# With flag (overrides environment)
bin/examples/addfixedip -H 192.168.4.1 -m aa:bb:cc:dd:ee:ff -i 192.168.4.35 -n wone -k
```

**Output:**
```
Created fixed IP assignment:
  Name:    wone
  MAC:     aa:bb:cc:dd:ee:ff
  IP:      192.168.4.35
  Network: Default
  DNS:     wone -> 192.168.4.35
```

### 2. examples/delfixedip/main.go

**Changes:**
- Added `UNIFI_UDM_IP` environment variable support
  - If `-H` flag not provided, checks for `UNIFI_UDM_IP` environment variable
  - Updated usage text to document this feature
  - Updated error message when host is missing
  - Added example showing usage without `-H` flag

**Note:** DNS deletion was already implemented - no changes needed to DNS logic.

**Usage:**
```bash
# With environment variable
export UNIFI_UDM_IP=192.168.4.1
bin/examples/delfixedip -m aa:bb:cc:dd:ee:ff -k

# With flag (overrides environment)
bin/examples/delfixedip -H 192.168.4.1 -m aa:bb:cc:dd:ee:ff -k
```

### 3. examples/fixedips/main.go

**Changes:**
- Added `UNIFI_UDM_IP` environment variable support
  - If `-H` flag not provided, checks for `UNIFI_UDM_IP` environment variable
  - Updated usage text to document this feature
  - Updated error message when host is missing
  - Added example showing usage without `-H` flag

**Usage:**
```bash
# With environment variable
export UNIFI_UDM_IP=192.168.4.1
bin/examples/fixedips -k

# With flag (overrides environment)
bin/examples/fixedips -H 192.168.4.1 -k
```

## Environment Variable Priority

All three examples follow this priority:
1. Command line `-H` flag (highest priority)
2. `UNIFI_UDM_IP` environment variable
3. Error if neither provided

This allows users to set a default host while still being able to override it per command.

## DNS Management Details

### addfixedip DNS Logic

When a fixed IP is assigned:
1. Check if any DNS record points to the target IP
2. If record exists with same hostname: do nothing
3. If record exists with different hostname: update hostname
4. If no record exists: create new A record

DNS records are created with:
- `Key`: hostname (from `-n` flag)
- `Value`: IP address (from `-i` flag)
- `RecordType`: "A"
- `Enabled`: true

### delfixedip DNS Logic (unchanged)

When a fixed IP is removed:
1. Find all DNS records pointing to the fixed IP
2. Display them to user
3. Delete them (unless `--keep-dns` flag used)
4. Remove fixed IP assignment

## Testing

All modified files pass syntax validation:
```bash
gofmt -e examples/addfixedip/main.go   # Syntax OK
gofmt -e examples/delfixedip/main.go   # Syntax OK
gofmt -e examples/fixedips/main.go     # Syntax OK
```

## Environment Setup Example

Create a `.env` file:
```bash
export UNIFI_USERNAME=admin
export UNIFI_PASSWORD=your-password
export UNIFI_UDM_IP=192.168.4.1
```

Source it before running examples:
```bash
source .env
bin/examples/addfixedip -m aa:bb:cc:dd:ee:ff -i 192.168.4.35 -n mydevice -k
bin/examples/fixedips -k
bin/examples/delfixedip -m aa:bb:cc:dd:ee:ff -k
```

## Error Handling

### DNS Creation Failures

If DNS record creation fails, `addfixedip` will:
- Print a warning to stderr
- Inform user that fixed IP was created but DNS must be configured manually
- Exit with success code (0)
- Fixed IP assignment remains in place

This ensures that DNS issues don't prevent fixed IP assignment.

## Backwards Compatibility

All changes are backwards compatible:
- Existing scripts using `-H` flag continue to work unchanged
- New environment variable is optional
- DNS creation is automatic but failures are non-fatal
