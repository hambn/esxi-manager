#!/bin/bash

# Debug script to extract raw vim-cmd output for a VM
# Usage: ./debug-vm-inspect.sh <esxi_host> <username> <password> <vm_id_or_name>

if [ $# -lt 4 ]; then
    echo "Usage: $0 <esxi_host> <username> <password> <vm_id_or_name>"
    exit 1
fi

ESXI_HOST=$1
USERNAME=$2
PASSWORD=$3
VM_SPEC=$4
OUTPUT_FILE="${5:-.debug-vm-inspect-output.txt}"

echo "=========================================="
echo "DEBUG: Extracting raw vim-cmd output"
echo "=========================================="
echo ""

# Try to resolve VM ID if name is provided
echo "[1] Resolving VM ID..."
VM_ID=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no "$USERNAME@$ESXI_HOST" "vim-cmd vmsvc/getallvms | grep '$VM_SPEC' | awk '{print \$1}'" 2>/dev/null)

if [ -z "$VM_ID" ]; then
    echo "ERROR: Could not resolve VM ID for '$VM_SPEC'"
    exit 1
fi

echo "VM ID: $VM_ID"
echo "Output: $OUTPUT_FILE"
echo ""

# Redirect to file
{
    echo "=========================================="
    echo "RAW VIM-CMD OUTPUT FOR VM: $VM_SPEC (ID: $VM_ID)"
    echo "=========================================="
    echo ""

    echo "[1] vim-cmd vmsvc/get.summary $VM_ID"
    echo "=========================================="
    sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no "$USERNAME@$ESXI_HOST" "vim-cmd vmsvc/get.summary $VM_ID" 2>/dev/null
    echo ""
    echo ""

    echo "[2] vim-cmd vmsvc/get.config $VM_ID (first 100 lines)"
    echo "=========================================="
    sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no "$USERNAME@$ESXI_HOST" "vim-cmd vmsvc/get.config $VM_ID" 2>/dev/null | head -100
    echo ""
    echo ""

    echo "[3] List of all fields found in summary"
    echo "=========================================="
    sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no "$USERNAME@$ESXI_HOST" "vim-cmd vmsvc/get.summary $VM_ID" 2>/dev/null | grep "=" | sed 's/.*\([a-zA-Z][a-zA-Z.]*\) *=.*/\1/' | sort -u
    echo ""

} > "$OUTPUT_FILE"

echo "Output saved to: $OUTPUT_FILE"
cat "$OUTPUT_FILE"
