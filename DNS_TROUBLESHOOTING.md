# DNS Troubleshooting Guide

## Problem
The `addfixedip` example successfully creates DNS records in UniFi (output shows `DNS: wone -> 192.168.4.35`), but local DNS resolution doesn't work (`ping wone` fails).

## Why This Happens

UniFi creating a DNS record doesn't automatically make it resolvable on your system. Your system needs to:
1. Use the UniFi gateway as its DNS server
2. The UniFi DNS service needs to be properly enabled
3. DNS cache may need to be flushed

## Step 1: Verify DNS Record in UniFi

Check that the record exists:
```bash
# List all DNS records
bin/examples/fixedips -k

# Or create a simple test
curl -k "https://192.168.4.1/proxy/network/v2/api/site/default/static-dns" \
  -H "Cookie: $(cat ~/.unifi-cookie)" | jq
```

## Step 2: Check Your System DNS Configuration

### macOS
```bash
# Check current DNS servers
scutil --dns | grep nameserver

# Check /etc/resolv.conf
cat /etc/resolv.conf
```

### Linux
```bash
# Check DNS servers
cat /etc/resolv.conf

# Check systemd-resolved (if using systemd)
resolvectl status
```

**Expected**: You should see `192.168.4.1` (your UniFi gateway) listed as a nameserver.

## Step 3: Test DNS Resolution Directly

Test DNS resolution against the UniFi gateway directly:

```bash
# Using nslookup
nslookup wone 192.168.4.1

# Using dig
dig @192.168.4.1 wone

# Using host
host wone 192.168.4.1
```

**If this works**: Your system isn't configured to use UniFi DNS.
**If this fails**: UniFi DNS service may not be enabled properly.

## Step 4: Enable UniFi DNS Service

1. Go to UniFi Network UI → Settings → Networks
2. Select your network (e.g., "Default")
3. In the "Advanced" section, check:
   - **DHCP Name Server**: Should be "Auto" or explicitly set to the gateway IP
   - **Domain Name**: Set to a domain (e.g., "local", "home", "lan")
   - Ensure "DHCP DNS Service" is enabled

## Step 5: Configure Your System to Use UniFi DNS

### macOS (per-interface)
```bash
# For Wi-Fi
networksetup -setdnsservers Wi-Fi 192.168.4.1

# For Ethernet
networksetup -setdnsservers Ethernet 192.168.4.1

# Verify
scutil --dns | grep nameserver
```

### macOS (system-wide)
Go to System Settings → Network → (your connection) → DNS → Add `192.168.4.1`

### Linux (using systemd-resolved)
Edit `/etc/systemd/resolved.conf`:
```ini
[Resolve]
DNS=192.168.4.1
Domains=~.
```

Then restart:
```bash
sudo systemctl restart systemd-resolved
```

### Linux (using /etc/resolv.conf)
Edit `/etc/resolv.conf`:
```
nameserver 192.168.4.1
search local
```

## Step 6: Flush DNS Cache

### macOS
```bash
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

### Linux (systemd-resolved)
```bash
sudo resolvectl flush-caches
```

### Linux (nscd)
```bash
sudo systemctl restart nscd
```

## Step 7: Test Again

After configuration:
```bash
# Test direct resolution
nslookup wone 192.168.4.1

# Test system resolution
ping wone

# Test with FQDN (if you set a domain)
ping wone.local
```

## Common Issues

### 1. mDNS/Bonjour Interference
If you set domain to "local", macOS mDNS might interfere:
- Use a different domain like "home" or "lan"
- Or use fully qualified name: `wone.local.`

### 2. DHCP Not Pushing DNS
If you got your IP via DHCP before enabling UniFi DNS:
```bash
# macOS - renew DHCP lease
sudo ipconfig set en0 DHCP

# Linux - renew DHCP lease
sudo dhclient -r && sudo dhclient
```

### 3. DNS Record Not in Correct Zone
UniFi may require the domain suffix. Try:
```bash
# If your domain is "local"
ping wone.local

# Check what UniFi thinks the domain is
# In UniFi UI → Settings → Networks → Advanced → Domain Name
```

### 4. DNS Service Not Enabled
Check UniFi controller logs:
```bash
# SSH to UniFi gateway
ssh admin@192.168.4.1

# Check dnsmasq (UniFi's DNS service)
ps aux | grep dnsmasq
cat /run/dnsmasq.conf.d/*.conf | grep wone
```

## Quick Test Script

Create a test script:
```bash
#!/bin/bash
echo "=== DNS Configuration Test ==="
echo ""
echo "1. System DNS servers:"
scutil --dns 2>/dev/null | grep nameserver || cat /etc/resolv.conf
echo ""
echo "2. Direct query to UniFi:"
nslookup wone 192.168.4.1 2>&1 | head -10
echo ""
echo "3. System resolution:"
nslookup wone 2>&1 | head -10
echo ""
echo "4. Ping test:"
ping -c 1 wone 2>&1 | head -3
```

## Verification Commands

After following the steps above, verify everything works:

```bash
# 1. DNS record exists in UniFi
bin/examples/addfixedip -H 192.168.4.1 -m 20:05:b7:01:00:20 -i 192.168.4.35 -n wone -k
# Should show: DNS: wone -> 192.168.4.35

# 2. UniFi resolves it
dig @192.168.4.1 wone
# Should show: wone.        0    IN    A    192.168.4.35

# 3. System uses UniFi DNS
scutil --dns | grep "192.168.4.1"
# Should show 192.168.4.1 in nameserver list

# 4. Resolution works
ping -c 1 wone
# Should succeed
```

## Alternative: Use FQDN

If you can't change system DNS, use fully qualified domain names:

1. Set domain in UniFi to "home" or "lan"
2. Add to fixed IP:
   ```bash
   bin/examples/addfixedip -H 192.168.4.1 -m 20:05:b7:01:00:20 -i 192.168.4.35 -n wone -k
   ```
3. Use FQDN:
   ```bash
   ping wone.home
   # or
   ping wone.lan
   ```

## Still Not Working?

If DNS still doesn't work after these steps:

1. Check UniFi controller logs for errors
2. Verify the network has "DNS Service" enabled
3. Try creating a DNS record manually in UniFi UI to test
4. Check if a firewall is blocking DNS (port 53)
5. Try different domain name (not "local")

## Success Check

You'll know it's working when:
```bash
$ dig @192.168.4.1 wone
wone.                   0       IN      A       192.168.4.35

$ ping wone
PING wone (192.168.4.35): 56 data bytes
64 bytes from 192.168.4.35: icmp_seq=0 ttl=64 time=1.234 ms
```
