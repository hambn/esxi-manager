# ESXi Manager Diagnostic Scripts

This directory contains helper scripts for testing and debugging the `vm-inspect` command across different ESXi hosts and configurations.

## Scripts

### 1. `debug-vm-inspect.sh`

Extract and display raw `vim-cmd` output for a specific VM to understand the exact format and field names used by the ESXi host.

**Usage:**
```bash
./debug-vm-inspect.sh <esxi_host> <username> <password> <vm_name_or_id> [output_file]
```

**Example:**
```bash
./debug-vm-inspect.sh 192.168.0.186 root "mypassword" "VM-OS-Prometheus"
```

**Output:**
- Captures raw output from `vim-cmd vmsvc/get.summary`
- Captures raw output from `vim-cmd vmsvc/get.config`
- Extracts all field names to help identify what fields are available
- Saves to `.debug-vm-inspect-output.txt` (or custom file)

**When to Use:**
- vm-inspect shows empty fields that should have values
- Need to understand the exact vim-cmd output format
- Debugging parser issues

**What to Look For:**
- Field naming conventions (e.g., `guestFullName` vs `guest.fullname`)
- Spacing around separators (e.g., `field =` vs `field=`)
- Separator types (e.g., `=` vs `:`)
- Placeholder values (e.g., `<unset>`, `(unset)`)
- Any fields you expect but don't see

### 2. `test-vm-inspect.sh`

Run vm-inspect on ALL VMs on an ESXi host to verify it works across your entire infrastructure.

**Usage:**
```bash
./test-vm-inspect.sh <esxi_host> <username> <password> [esxi_manager_binary_path]
```

**Example:**
```bash
./test-vm-inspect.sh 192.168.0.186 root "mypassword" "./cmd/esxi-manager/esxi-manager"
```

**Output:**
- Lists all VMs on the host
- Inspects each VM with vm-inspect
- Shows which VMs have complete data and which may have issues

**When to Use:**
- Before deployment to verify vm-inspect works broadly
- Testing new ESXi versions
- Regression testing after code changes

## Troubleshooting Guide

### Symptom: Empty Fields in vm-inspect Output

1. **Run the debug script:**
   ```bash
   ./scripts/debug-vm-inspect.sh 192.168.0.186 root "password" "VM-Name"
   ```

2. **Check the output file for:**
   - Does the field appear in the raw output at all?
   - What is the exact field name and format?
   - Does it use quotes? Trailing commas? Different separators?

3. **Update the parser if needed:**
   - Edit `internal/esxi/virtual-machines/inspect.go`
   - Update the `parseConfigValueFlexible()` calls with new field name patterns
   - Rebuild and test

### Symptom: Config Path Not Found

The vm-inspect command uses multiple strategies to find the VM's configuration path:
1. Looks in `vim-cmd vmsvc/get.summary` output
2. Falls back to `vim-cmd vmsvc/get.config`
3. Last resort: searches filesystem with `find /vmfs/volumes`

If all fail, the dependent features (VMX file parsing, VMDK parsing) won't work.

**To debug:**
```bash
./scripts/debug-vm-inspect.sh 192.168.0.186 root "password" "VM-Name" | grep -i path
```

### Symptom: Guest OS Shows `<unset>`

This is normal for VMs without VMware Tools installed. The parser intentionally skips `<unset>` values and displays them as "N/A".

### Symptom: Some VMs Work, Others Don't

This often indicates version differences in vim-cmd output. ESXi 6.x and 7.x have different output formats. The parsing is designed to be flexible, but:

1. Run debug script on both working and non-working VMs
2. Compare the field names and formats
3. Note any differences
4. Update parser with additional field name patterns

## How the Flexible Parser Works

The current parser (`parseConfigValueFlexible`) uses these strategies:

1. **Case-insensitive matching** - handles `FieldName`, `fieldname`, `FIELDNAME`
2. **Multiple key patterns** - tries 3-5 different field names per field
3. **Separator flexibility** - recognizes both `field = value` and `field: value`
4. **Quote/comma cleanup** - removes quotes and trailing commas
5. **Placeholder detection** - skips `<unset>`, `(unset)`, `<unknown>`, etc.

If a field still isn't found, you likely need to add another key pattern.

## Adding Support for New Field Names

If you find a field isn't being extracted:

1. Note the exact field name from debug output
2. Edit `internal/esxi/virtual-machines/inspect.go`
3. Find the `parseConfigValueFlexible()` call for that field
4. Add the new field name pattern to the keys array

Example:
```go
// Old - only tried 3 patterns
parseConfigValueFlexible(output, []string{"guestFullName =", "guestFullName=", "guest.fullname ="}, &info.GuestOS)

// New - added 2 more patterns
parseConfigValueFlexible(output, []string{
    "guestFullName =", "guestFullName=",
    "guest.fullname =", "guestOS =",
    "config.guestFullName ="
}, &info.GuestOS)
```

5. Rebuild: `go build -o ./cmd/esxi-manager/esxi-manager ./cmd/esxi-manager/`
6. Test: `./cmd/esxi-manager/esxi-manager ... --command="vm-inspect" ...`

## Requirements

The debug scripts require:
- `sshpass` - for non-interactive SSH password authentication
- `grep`, `awk`, `sed` - standard Unix tools
- SSH access to the ESXi host
