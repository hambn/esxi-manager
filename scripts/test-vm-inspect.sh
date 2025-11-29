#!/bin/bash

# Test script to verify vm-inspect works on all VMs on an ESXi host
# Usage: ./test-vm-inspect.sh <esxi_host> <username> <password>

if [ $# -lt 3 ]; then
    echo "Usage: $0 <esxi_host> <username> <password>"
    exit 1
fi

ESXI_HOST=$1
USERNAME=$2
PASSWORD=$3
BINARY="${4:-.../cmd/esxi-manager/esxi-manager}"

echo "=========================================="
echo "VM Inspect Test Suite"
echo "=========================================="
echo "ESXi Host: $ESXI_HOST"
echo ""

# Get list of all VMs
echo "Fetching VM list from ESXi host..."
VMS=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no "$USERNAME@$ESXI_HOST" "vim-cmd vmsvc/getallvms 2>/dev/null" | tail -n +2 | awk '{print $1 ":" $2}')

if [ -z "$VMS" ]; then
    echo "ERROR: Could not fetch VMs from ESXi host"
    exit 1
fi

echo "Found VMs:"
echo "$VMS" | while IFS=: read -r id name; do
    echo "  [$id] $name"
done
echo ""

# Test each VM
TEST_NUM=1
echo "$VMS" | while IFS=: read -r id name; do
    echo "=========================================="
    echo "Test $TEST_NUM: Inspecting [$id] $name"
    echo "=========================================="

    $BINARY \
        --esxi-host-uri="$ESXI_HOST" \
        --esxi-host-username="$USERNAME" \
        --esxi-host-password="$PASSWORD" \
        --command="vm-inspect" \
        --vm-inspect-id="$id"

    echo ""
    TEST_NUM=$((TEST_NUM + 1))
done

echo "=========================================="
echo "All tests completed"
echo "=========================================="
