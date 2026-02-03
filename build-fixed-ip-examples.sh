#!/bin/bash
# Build script for fixed IP examples

set -e

cd "$(dirname "$0")"

echo "Building fixed IP management examples..."

echo "  - addfixedip"
go build -o bin/examples/addfixedip ./examples/addfixedip

echo "  - delfixedip"
go build -o bin/examples/delfixedip ./examples/delfixedip

echo "  - fixedips"
go build -o bin/examples/fixedips ./examples/fixedips

echo ""
echo "Build complete! Examples are in bin/examples/"
echo ""
echo "Usage:"
echo "  export UNIFI_UDM_IP=192.168.4.1"
echo "  export UNIFI_USERNAME=admin"
echo "  export UNIFI_PASSWORD=your-password"
echo ""
echo "  bin/examples/addfixedip -m 20:05:b7:01:00:20 -i 192.168.4.35 -n wone -k"
echo "  bin/examples/fixedips -k"
echo "  bin/examples/delfixedip -m 20:05:b7:01:00:20 -k"
