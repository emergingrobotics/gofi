#!/bin/bash
# Quick DNS test script

HOST="${1:-wone}"

echo "=== DNS Resolution Test for '$HOST' ==="
echo ""

echo "1. Testing with nslookup:"
nslookup "$HOST" 2>&1 | head -6
echo ""

echo "2. Testing with ping (1 packet):"
ping -c 1 "$HOST" 2>&1 | head -3
echo ""

echo "3. Checking what resolved:"
if command -v dig &> /dev/null; then
    dig "$HOST" +short
else
    host "$HOST" 2>&1 | head -1
fi
echo ""

echo "4. Testing port 22 (SSH):"
if command -v nc &> /dev/null; then
    timeout 2 nc -zv "$HOST" 22 2>&1
elif command -v telnet &> /dev/null; then
    timeout 2 telnet "$HOST" 22 2>&1 | head -3
else
    echo "(nc/telnet not available - install netcat to test ports)"
fi
echo ""

echo "=== Summary ==="
if ping -c 1 -W 1 "$HOST" &> /dev/null; then
    echo "✓ DNS resolution: SUCCESS"
    echo "✓ Host is reachable"
    IP=$(ping -c 1 "$HOST" 2>&1 | grep -oE '\([0-9.]+\)' | head -1 | tr -d '()')
    echo "  Resolved to: $IP"

    if timeout 2 nc -zv "$HOST" 22 &> /dev/null 2>&1; then
        echo "✓ SSH port (22): OPEN"
    else
        echo "✗ SSH port (22): CLOSED or FILTERED"
        echo "  (Device is up but not running SSH server)"
    fi
else
    echo "✗ DNS resolution or connectivity: FAILED"
fi
