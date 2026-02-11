# OpenWrt One - Initial Setup Guide

## Device Overview

The OpenWrt One is a reference platform router with:
- MediaTek MT7981B SoC (dual-core ARM Cortex-A53)
- 1GB DDR4 RAM
- 128MB SPI-NAND flash
- 2× 2.5 Gigabit Ethernet ports
- MediaTek MT7976C wireless (WiFi 6, 2.4GHz + 5GHz)
- USB 2.0 port
- M.2 2042 slot

---

## Step 1: Initial Connection

### Physical Setup

1. **Power**: Connect USB-C power supply (5V/3A minimum)
2. **Ethernet**: Connect your computer to **LAN port** (typically eth0, the port closest to USB)
3. **WAN**: Connect internet source to **WAN port** (typically eth1)

### First Access

**Default IP**: `192.168.1.1`
**Default credentials**:
- Username: `root`
- Password: (none - empty password initially)

#### Option A: Web Interface (LuCI)

1. Open browser: http://192.168.1.1
2. Login as `root` with no password
3. **IMPORTANT**: Set a password immediately!

#### Option B: SSH

```bash
# Connect via SSH
ssh root@192.168.1.1

# You'll see a warning about no password - this is expected
# Set a password immediately
passwd
# Enter new password twice
```

---

## Step 2: Set Root Password (CRITICAL!)

**Do this first before anything else:**

### Via Web Interface
1. Go to **System → Administration**
2. Under "Router Password", enter new password twice
3. Click **Save & Apply**

### Via SSH
```bash
passwd
# Enter new password twice
```

---

## Step 3: Update Package Lists

```bash
# SSH to device
ssh root@192.168.1.1

# Update package lists
opkg update

# Optional: Upgrade installed packages
opkg list-upgradable
opkg upgrade [package-name]
```

---

## Step 4: Network Configuration

### Check Current Configuration

```bash
# View network interfaces
cat /etc/config/network

# View current IP setup
ifconfig
ip addr show
```

### Basic Network Setup

#### Scenario A: Router Mode (Most Common)

OpenWrt One acts as your main router with DHCP server.

**Web Interface:**
1. Go to **Network → Interfaces**
2. **LAN Interface**:
   - IPv4 address: `192.168.1.1` (or your preferred subnet)
   - Netmask: `255.255.255.0`
   - DHCP Server: Enabled
   - DHCP range: `192.168.1.100 - 192.168.1.250`

3. **WAN Interface**:
   - Protocol: `DHCP client` (if your ISP uses DHCP)
   - Or `Static address` (if you have static IP from ISP)
   - Or `PPPoE` (if required by ISP)

**SSH Configuration:**

```bash
# Edit network config
vi /etc/config/network

# LAN configuration
config interface 'lan'
    option device 'br-lan'
    option proto 'static'
    option ipaddr '192.168.1.1'
    option netmask '255.255.255.0'
    option ip6assign '60'

# WAN configuration (DHCP from ISP)
config interface 'wan'
    option device 'eth1'
    option proto 'dhcp'

# Restart network
/etc/init.d/network restart
```

#### Scenario B: Bridge/Access Point Mode

OpenWrt One extends existing network without routing.

```bash
vi /etc/config/network

# Disable WAN, configure LAN as client
config interface 'lan'
    option device 'br-lan'
    option proto 'static'
    option ipaddr '192.168.1.2'      # Different from your main router
    option netmask '255.255.255.0'
    option gateway '192.168.1.1'     # Your main router IP
    option dns '192.168.1.1'

# Disable DHCP server
/etc/init.d/dnsmasq stop
/etc/init.d/dnsmasq disable
```

---

## Step 5: Wireless Configuration

### Enable Wireless

OpenWrt typically ships with wireless disabled by default.

#### Via Web Interface

1. Go to **Network → Wireless**
2. Click **Enable** on both radios (2.4GHz and 5GHz)
3. Click **Edit** on each radio
4. Configure:
   - **ESSID**: Your network name
   - **Mode**: Access Point
   - **Network**: lan
   - **Encryption**: WPA2-PSK or WPA3-SAE
   - **Key**: Your WiFi password (minimum 8 characters)
5. Click **Save & Apply**

#### Via SSH

```bash
# Edit wireless config
vi /etc/config/wireless

# 2.4GHz Radio
config wifi-device 'radio0'
    option type 'mac80211'
    option path 'platform/soc/11280000.pcie/pci0000:00/0000:00:00.0/0000:01:00.0'
    option channel '6'
    option band '2g'
    option htmode 'HE20'
    option disabled '0'              # Enable radio

config wifi-iface 'default_radio0'
    option device 'radio0'
    option network 'lan'
    option mode 'ap'
    option ssid 'OpenWrt-2G'         # Your network name
    option encryption 'psk2'
    option key 'YourPassword123'     # Your WiFi password

# 5GHz Radio
config wifi-device 'radio1'
    option type 'mac80211'
    option path 'platform/soc/11280000.pcie/pci0000:00/0000:00:00.0/0000:01:00.0+1'
    option channel '36'
    option band '5g'
    option htmode 'HE80'
    option disabled '0'              # Enable radio

config wifi-iface 'default_radio1'
    option device 'radio1'
    option network 'lan'
    option mode 'ap'
    option ssid 'OpenWrt-5G'         # Your network name
    option encryption 'psk2'
    option key 'YourPassword123'     # Your WiFi password

# Restart wireless
wifi reload
```

### Check Wireless Status

```bash
# View wireless status
wifi status

# View associated clients
iw dev wlan0 station dump
iw dev wlan1 station dump
```

---

## Step 6: Firewall Configuration

### Basic Firewall Rules

OpenWrt comes with sensible firewall defaults:
- LAN → WAN: Allowed (MASQUERADE)
- WAN → LAN: Blocked
- LAN → Router: Allowed

#### Via Web Interface

1. Go to **Network → Firewall**
2. Review zones:
   - **LAN zone**: Input: Accept, Output: Accept, Forward: Accept
   - **WAN zone**: Input: Reject, Output: Accept, Forward: Reject
3. **WAN → LAN**: Should be REJECT by default

#### Via SSH

```bash
# View firewall config
cat /etc/config/firewall

# Basic zones are pre-configured
# Usually no changes needed for initial setup
```

### Allow Services from WAN (Optional)

**ONLY if you need external access - not recommended for security!**

```bash
# Example: Allow SSH from WAN (NOT RECOMMENDED)
vi /etc/config/firewall

config rule
    option name 'Allow-SSH-WAN'
    option src 'wan'
    option proto 'tcp'
    option dest_port '22'
    option target 'ACCEPT'

/etc/init.d/firewall restart
```

**Better approach**: Use VPN or WireGuard instead of exposing SSH.

---

## Step 7: DHCP & DNS Configuration

### Configure DHCP Server

```bash
vi /etc/config/dhcp

config dhcp 'lan'
    option interface 'lan'
    option start '100'               # First IP in pool
    option limit '150'               # Number of IPs
    option leasetime '12h'
    option dhcpv4 'server'
    option dhcpv6 'server'
    option ra 'server'
    option ra_management '1'

# Restart DHCP
/etc/init.d/dnsmasq restart
```

### Configure DNS

```bash
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
    # Add upstream DNS servers
    list server '1.1.1.1'
    list server '8.8.8.8'

/etc/init.d/dnsmasq restart
```

---

## Step 8: System Hardening

### Disable Unnecessary Services

```bash
# List running services
/etc/init.d/*

# Disable uhttpd if you don't need web interface remotely
/etc/init.d/uhttpd stop
/etc/init.d/uhttpd disable

# Or restrict to LAN only
vi /etc/config/uhttpd

config uhttpd 'main'
    list listen_http '192.168.1.1:80'
    list listen_https '192.168.1.1:443'
    # Remove 0.0.0.0:80 entries
```

### Configure SSH Security

```bash
vi /etc/config/dropbear

config dropbear
    option Port '22'
    option Interface 'lan'           # Only listen on LAN
    option RootPasswordAuth 'on'
    option PasswordAuth 'on'
    option GatewayPorts 'off'

/etc/init.d/dropbear restart
```

### Set Timezone

```bash
# Via Web Interface
# System → System → General Settings → Timezone

# Via SSH
vi /etc/config/system

config system
    option hostname 'OpenWrt-One'
    option timezone 'PST8PDT,M3.2.0,M11.1.0'
    option zonename 'America/Los Angeles'

/etc/init.d/system restart
```

---

## Step 9: Install Essential Packages

```bash
# Update package list
opkg update

# Useful packages
opkg install luci-app-sqm          # SQM QoS (bufferbloat mitigation)
opkg install luci-app-statistics   # System monitoring
opkg install htop                  # Better process viewer
opkg install tcpdump               # Network debugging
opkg install nano                  # Easier text editor
opkg install curl                  # HTTP client
opkg install iperf3                # Network performance testing

# Optional: VPN
opkg install luci-app-wireguard    # WireGuard VPN
opkg install luci-app-openvpn      # OpenVPN
```

---

## Step 10: Backup Configuration

**Always backup after configuration!**

### Via Web Interface

1. Go to **System → Backup / Flash Firmware**
2. Click **Generate archive**
3. Download and save the `.tar.gz` file

### Via SSH

```bash
# Create backup
sysupgrade -b /tmp/backup-$(date +%Y%m%d).tar.gz

# Copy to your computer
scp root@192.168.1.1:/tmp/backup-*.tar.gz ~/Downloads/
```

---

## Common Configurations

### Configure Static DHCP Leases

```bash
vi /etc/config/dhcp

config host
    option name 'myserver'
    option mac 'AA:BB:CC:DD:EE:FF'
    option ip '192.168.1.50'

/etc/init.d/dnsmasq restart
```

### Port Forwarding

```bash
vi /etc/config/firewall

config redirect
    option name 'Web Server'
    option src 'wan'
    option src_dport '80'
    option dest 'lan'
    option dest_ip '192.168.1.100'
    option dest_port '80'
    option proto 'tcp'

/etc/init.d/firewall restart
```

### QoS/SQM (Bufferbloat Control)

```bash
# Install SQM
opkg update
opkg install luci-app-sqm

# Configure via web interface
# Network → SQM QoS
# Enable on WAN interface
# Set download/upload speeds to 85-95% of your actual speeds
```

---

## Troubleshooting

### Can't Access Web Interface

```bash
# Check if uhttpd is running
ps | grep uhttpd

# Restart web server
/etc/init.d/uhttpd restart

# Check IP configuration
ip addr show br-lan
```

### No Internet Connection

```bash
# Check WAN status
ifstatus wan

# Check if DNS is working
nslookup google.com

# Check default route
ip route

# Restart network
/etc/init.d/network restart
```

### Wireless Not Working

```bash
# Check wireless status
wifi status

# Enable wireless
uci set wireless.radio0.disabled='0'
uci set wireless.radio1.disabled='0'
uci commit wireless
wifi reload

# Check logs
logread | grep -i wireless
```

### Reset to Defaults

**Hardware reset:**
1. Power on device
2. Wait for boot (LEDs settle)
3. Press and hold reset button for 10+ seconds
4. Release - device will reboot with factory defaults

**Software reset:**
```bash
firstboot -y
reboot
```

---

## Performance Tuning

### Enable Hardware Flow Offloading

```bash
# Enable hardware NAT acceleration
vi /etc/config/firewall

config defaults
    option syn_flood '1'
    option input 'ACCEPT'
    option output 'ACCEPT'
    option forward 'REJECT'
    option flow_offloading '1'         # Enable
    option flow_offloading_hw '1'      # Enable hardware offload

/etc/init.d/firewall restart
```

### Optimize Wireless

```bash
vi /etc/config/wireless

config wifi-device 'radio1'
    option txpower '20'                # Adjust transmit power
    option legacy_rates '0'            # Disable legacy rates for better performance
    option beacon_int '100'            # Default beacon interval

config wifi-iface 'default_radio1'
    option isolate '0'                 # Allow clients to talk to each other
    option wmm '1'                     # Enable WMM (QoS)
```

---

## Next Steps

After basic setup, consider:

1. **VPN Setup**: Configure WireGuard for secure remote access
2. **AdBlocking**: Install adblock-fast or AdGuard Home
3. **Monitoring**: Set up statistics and monitoring
4. **Guest Network**: Create isolated guest WiFi
5. **VLAN Setup**: Segment network with VLANs
6. **USB Storage**: Mount USB drive for additional storage/logs

---

## Resources

- **Official Docs**: https://openwrt.org/docs/start
- **OpenWrt One Page**: https://openwrt.org/toh/openwrt/one
- **Forum**: https://forum.openwrt.org/
- **Package List**: https://openwrt.org/packages/start

---

## Quick Reference Commands

```bash
# Network
/etc/init.d/network restart       # Restart networking
ifconfig                          # Show interfaces
ip addr show                      # Show IP addresses
ip route                          # Show routing table

# Wireless
wifi status                       # Wireless status
wifi reload                       # Reload wireless config
iw dev                            # Show wireless devices

# Services
/etc/init.d/dnsmasq restart      # Restart DHCP/DNS
/etc/init.d/firewall restart     # Restart firewall
/etc/init.d/uhttpd restart       # Restart web interface

# System
logread                          # View system log
logread -f                       # Follow log in real-time
ps                               # List processes
top / htop                       # System monitor
reboot                           # Reboot device

# Package Management
opkg update                      # Update package list
opkg list                        # List available packages
opkg install <package>           # Install package
opkg remove <package>            # Remove package
opkg list-installed              # List installed packages
```
