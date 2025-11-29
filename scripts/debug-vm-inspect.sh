#!/bin/bash

# Debug script to extract raw vim-cmd output for a VM
# Usage: ./debug-vm-inspect.sh <esxi_host> <username> <password> <vm_id_or_name>

if [ $# -lt 4 ]; then
    echo "Usage: $0 <esxi_host> <username> <password> <vm_id_or_name>"
    exit 1
fi

ESXi_HOST=$1
USERNAME=$2
PASSWORD=$3
VM_SPEC=$4

echo "=========================================="
echo "DEBUG: Extracting raw vim-cmd output"
echo "=========================================="
echo ""

# Try to resolve VM ID if name is provided
echo "[1] Resolving VM ID..."
VM_ID=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no "$USERNAME@$ESXi_HOST" "vim-cmd vmsvc/getallvms | grep '$VM_SPEC' | awk '{print \$1}'" 2>/dev/null)

if [ -z "$VM_ID" ]; then
    echo "ERROR: Could not resolve VM ID for '$VM_SPEC'"
    exit 1
fi

echo "VM ID: $VM_ID"
echo ""

# Get summary
echo "[2] vim-cmd vmsvc/get.summary $VM_ID"
echo "---"
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no "$USERNAME@$ESXi_HOST" "vim-cmd vmsvc/get.summary $VM_ID" 2>/dev/null
echo ""
echo ""

# Get config
echo "[3] vim-cmd vmsvc/get.config $VM_ID"
echo "---"
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no "$USERNAME@$ESXi_HOST" "vim-cmd vmsvc/get.config $VM_ID" 2>/dev/null | head -50
echo "(truncated...)"
echo ""
