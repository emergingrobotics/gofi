# OpenWrt DNS Configuration Guide

## Scenario 1: OpenWrt Using UniFi for DNS

If your OpenWrt One is behind a UniFi gateway and you want it to resolve local hostnames managed by UniFi.

### Via Web Interface (LuCI)

1. Go to **Network → Interfaces**
2. Click **Edit** on your WAN interface (or the interface connecting to UniFi)
3. Under **Advanced Settings** tab:
   - Uncheck "Use DNS servers advertised by peer" (if you want to force UniFi DNS)
   - Add custom DNS servers: `192.168.4.1` (your UniFi gateway)
4. Click **Save & Apply**

### Via Command Line (SSH)

```bash
# SSH to OpenWrt
ssh root@openwrt.lan

# Edit network config
vi /etc/config/network

# Find your WAN interface section and add/modify:
config interface 'wan'
    option device 'eth1'
    option proto 'dhcp'
    option peerdns '0'          # Don't use DHCP-provided DNS
    option dns '192.168.4.1'    # Use UniFi gateway as DNS

# Restart network
/etc/init.d/network restart

# Verify DNS configuration
cat /tmp/resolv.conf.d/resolv.conf.auto
```

### Configure dnsmasq to Forward to UniFi

```bash
# Edit dnsmasq config
vi /etc/config/dhcp

# Add or modify the dnsmasq section:
config dnsmasq
    option domainneeded '1'
    option localise_queries '1'
    option rebind_protection '1'
    option rebind_localhost '1'
    option local '/lan/'
    option domain 'lan'
    option expandhosts '1'
    option authoritative '1'
    option readethers '1'
    option leasefile '/tmp/dhcp.leases'
    option resolvfile '/tmp/resolv.conf.d/resolv.conf.auto'
    option localservice '1'
    list server '192.168.4.1'   # Forward to UniFi

# Restart dnsmasq
/etc/init.d/dnsmasq restart
```

---

## Scenario 2: OpenWrt as Primary DNS Server

If you want OpenWrt to handle DNS for your local network and use the gofi examples to manage UniFi records.

### Configure OpenWrt dnsmasq

```bash
# SSH to OpenWrt
ssh root@openwrt.lan

# Edit dnsmasq config
vi /etc/config/dhcp
```

Add this configuration:

```
config dnsmasq
    option domainneeded '1'
    option localise_queries '1'
    option rebind_protection '1'
    option rebind_localhost '1'
    option local '/lan/'
    option domain 'lan'           # Your local domain
    option expandhosts '1'
    option authoritative '1'
    option readethers '1'
    option leasefile '/tmp/dhcp.leases'
    option resolvfile '/tmp/resolv.conf.d/resolv.conf.auto'
    option localservice '1'
    option noresolv '0'

# Upstream DNS servers (Google, Cloudflare, etc.)
config dnsmasq
    list server '8.8.8.8'
    list server '1.1.1.1'

# Static DNS entries (hostnames)
config domain
    option name 'wone.lan'
    option ip '192.168.4.35'

config domain
    option name 'router.lan'
    option ip '192.168.1.1'
```

Restart dnsmasq:
```bash
/etc/init.d/dnsmasq restart
```

### Alternative: Use /etc/hosts

OpenWrt's dnsmasq reads `/etc/hosts` automatically:

```bash
# Edit hosts file
vi /etc/hosts

# Add entries
192.168.4.35    wone wone.lan
192.168.1.100   myserver myserver.lan
192.168.1.200   printer printer.lan

# Reload dnsmasq
/etc/init.d/dnsmasq reload
```

### Alternative: Use dnsmasq.conf Directly

For more advanced configurations:

```bash
# Create custom dnsmasq config
vi /etc/dnsmasq.conf

# Add entries
address=/wone.lan/192.168.4.35
address=/myserver.lan/192.168.1.100

# Restart dnsmasq
/etc/init.d/dnsmasq restart
```

---

## Scenario 3: Hybrid Setup (OpenWrt + UniFi DNS)

If you want OpenWrt to handle some local DNS but also query UniFi for its records.

### Configure Conditional Forwarding

```bash
vi /etc/config/dhcp
```

```
config dnsmasq
    option domainneeded '1'
    option localise_queries '1'
    option rebind_protection '1'
    option rebind_localhost '1'
    option local '/lan/'
    option domain 'lan'
    option expandhosts '1'
    option authoritative '1'
    option readethers '1'
    option leasefile '/tmp/dhcp.leases'
    option resolvfile '/tmp/resolv.conf.d/resolv.conf.auto'
    option localservice '1'

# Static DNS entries managed by OpenWrt
config domain
    option name 'openwrt.lan'
    option ip '192.168.1.1'

# Forward specific domains to UniFi
# Note: This forwards ALL queries, not just specific domains
# dnsmasq doesn't support domain-specific forwarding in UCI easily
```

For domain-specific forwarding, use dnsmasq.conf:

```bash
vi /etc/dnsmasq.conf

# Forward queries for .unifi domain to UniFi gateway
server=/unifi/192.168.4.1

# Forward everything else upstream
server=8.8.8.8
server=1.1.1.1
```

---

## Integrate with gofi Examples

### Option 1: Sync UniFi DNS to OpenWrt

Create a script to sync DNS records from UniFi to OpenWrt:

```bash
#!/bin/sh
# /root/sync-unifi-dns.sh

# Export credentials
export UNIFI_UDM_IP="192.168.4.1"
export UNIFI_USERNAME="admin"
export UNIFI_PASSWORD="your-password"

# Get DNS records from UniFi (requires gofi fixedips example with DNS output)
# For now, manually maintain /etc/hosts or use the API

# Example: Query UniFi API for fixed IPs and update /etc/hosts
# (This would require curl + jq or the gofi CLI tools)

# Reload dnsmasq
/etc/init.d/dnsmasq reload
```

### Option 2: Let UniFi Manage DNS, OpenWrt Forwards

Simplest approach:

```bash
vi /etc/config/dhcp

config dnsmasq
    list server '192.168.4.1'   # Forward all DNS to UniFi
    option noresolv '1'          # Don't use /etc/resolv.conf
```

This way:
- Run `bin/examples/addfixedip` to create DNS in UniFi
- OpenWrt forwards all DNS queries to UniFi
- All clients behind OpenWrt get UniFi DNS resolution

---

## Testing DNS Configuration

### Test DNS Resolution on OpenWrt

```bash
# Query local dnsmasq
nslookup wone 127.0.0.1

# Query UniFi directly
nslookup wone 192.168.4.1

# Check what DNS servers OpenWrt is using
cat /tmp/resolv.conf.d/resolv.conf.auto

# Check dnsmasq status
logread | grep dnsmasq

# Test from OpenWrt command line
ping wone
```

### Check dnsmasq Configuration

```bash
# Show running config
ps | grep dnsmasq

# Check lease file
cat /tmp/dhcp.leases

# Check resolv.conf
cat /etc/resolv.conf
```

### Debug dnsmasq

Enable query logging:

```bash
vi /etc/config/dhcp

config dnsmasq
    option logqueries '1'

/etc/init.d/dnsmasq restart

# Watch logs
logread -f | grep dnsmasq
```

---

## DHCP Configuration for Clients

Make sure clients behind OpenWrt get the right DNS server.

### Via Web Interface

1. Go to **Network → Interfaces**
2. Click **Edit** on **LAN** interface
3. Under **DHCP Server** tab:
   - **DNS Server**: Leave blank to advertise OpenWrt as DNS, or set to `192.168.4.1` to advertise UniFi

### Via Command Line

```bash
vi /etc/config/dhcp

config dhcp 'lan'
    option interface 'lan'
    option start '100'
    option limit '150'
    option leasetime '12h'
    option dhcpv4 'server'
    list dhcp_option '6,192.168.4.1'  # Option 6 = DNS Server
    # Or leave blank to advertise OpenWrt itself (192.168.1.1)

/etc/init.d/dnsmasq restart
```

---

## Recommended Setup for Your Use Case

Based on your gofi examples, I recommend:

### Setup: UniFi Manages DNS, OpenWrt Forwards

**On OpenWrt:**
```bash
# 1. Configure OpenWrt to forward DNS to UniFi
vi /etc/config/dhcp

config dnsmasq
    option domainneeded '1'
    option localise_queries '1'
    option rebind_protection '1'
    option rebind_localhost '1'
    option local '/lan/'
    option domain 'lan'
    option expandhosts '1'
    option authoritative '1'
    option readethers '1'
    option leasefile '/tmp/dhcp.leases'
    option localservice '1'
    list server '192.168.4.1'    # Forward to UniFi
    option noresolv '1'           # Don't use /etc/resolv.conf

# 2. Restart services
/etc/init.d/dnsmasq restart
/etc/init.d/network restart

# 3. Test
nslookup wone 127.0.0.1
```

**On Your Workstation:**
```bash
# Use gofi to manage UniFi DNS
export UNIFI_UDM_IP=192.168.4.1
bin/examples/addfixedip -m 20:05:b7:01:00:20 -i 192.168.4.35 -n wone -k

# Test resolution through OpenWrt
ping wone
```

This way:
- UniFi is the source of truth for DNS (via gofi)
- OpenWrt forwards DNS queries to UniFi
- All devices behind OpenWrt can resolve UniFi-managed hostnames
- You manage everything through the gofi CLI tools

---

## Troubleshooting

### DNS Not Resolving

```bash
# Check dnsmasq is running
ps | grep dnsmasq

# Check dnsmasq logs
logread | grep dnsmasq

# Verify DNS server in config
uci show dhcp.@dnsmasq[0]

# Test direct query to UniFi
nslookup wone 192.168.4.1
```

### Clients Not Getting DNS

```bash
# Check DHCP leases
cat /tmp/dhcp.leases

# Verify DHCP options
uci show dhcp.lan

# Check what DNS is being advertised
tcpdump -i br-lan port 67 -vv
```

### OpenWrt Can't Reach UniFi

```bash
# Test connectivity
ping 192.168.4.1

# Check routing
ip route
traceroute 192.168.4.1

# Verify firewall isn't blocking
iptables -L -n | grep 192.168.4.1
```

---

## Summary

**For your setup with gofi + OpenWrt + UniFi:**

1. Configure OpenWrt to forward DNS to UniFi (192.168.4.1)
2. Use `bin/examples/addfixedip` to create DNS records in UniFi
3. All clients (behind OpenWrt or directly on UniFi network) can resolve hostnames
4. Manage everything via gofi CLI tools

This gives you centralized DNS management through UniFi's API while OpenWrt acts as a transparent DNS forwarder.
