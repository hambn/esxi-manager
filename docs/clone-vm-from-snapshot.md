# Clone VM from Snapshot

This command clones an ESXi virtual machine from a snapshot, creating a new independent VM with customizable hardware configuration.

## Command

```bash
clone-vm-from-snapshot
```

## Description

The `clone-vm-from-snapshot` command creates a new VM by cloning from a specific snapshot of an existing source VM. The cloned VM can either inherit the source VM's configuration (CPU, memory, network) or use custom hardware specifications.

### Key Features

- **Snapshot Selection**: Find snapshots by name or ID
- **Flexible Configuration**: Use source VM config or specify custom hardware
- **Multiple Disk Types**: Support for thin provisioned, thick provisioned (lazy-zeroed), and thick provisioned (eager-zeroed) disks
- **Disk Sizing**: Optionally expand disk size during cloning (shrinking not supported)
- **Network Configuration**: Inherit from source VM or specify one or multiple portgroups
- **Detailed Operation Log**: Returns step-by-step operation details in JSON format
- **Smart Defaults**: Defaults to thin provisioning and source VM networks if not specified

## Required Parameters

```
--source-vm-name <string>      Source VM name to clone from
  OR
--source-vm-id <string>        Source VM ID to clone from

--snapshot-name <string>       Snapshot name to clone from
  OR
--snapshot-id <string>         Snapshot ID to clone from

--dest-vm-name <string>        Destination VM name (must not exist)

--dest-vm-disk-store <string>  Destination datastore for VM

--esxi-host-uri <string>       ESXi host URI/IP address
--esxi-host-username <string>  ESXi host username
--esxi-host-password <string>  ESXi host password
```

## Optional Parameters

```
--dest-vm-cpu <int>            CPU cores for destination VM (uses source if not specified)
--dest-vm-ram <int>            RAM in MB for destination VM (uses source if not specified)
--dest-disk-type <string>      Disk provisioning type (default: thin):
                               - thin - Thin provisioned (minimal initial space)
                               - zeroedthick - Thick provisioned, lazy-zeroed
                               - eagerzeroedthick - Thick provisioned, eager-zeroed
--dest-disk-size <int>         Disk size in GB (expands if larger than source,
                               shrinking not supported, inherits source size if not specified)
--dest-network-list <string>   Comma-separated portgroup names for destination VM
                               (uses source VM portgroups if not specified)
--esxi-host-port <int>         ESXi host SSH port (default: 22)
--raw-json                     Output raw JSON without CLI formatting
```

## Examples

### Clone with Source Configuration

Clone snapshot to a new VM using the same hardware as the source VM:

```bash
./esxi-manager \
  --esxi-host-uri=192.168.1.100 \
  --esxi-host-username=root \
  --esxi-host-password='password' \
  --command=clone-vm-from-snapshot \
  --source-vm-name=template-vm \
  --snapshot-name=production-ready \
  --dest-vm-name=clone-vm-01 \
  --dest-vm-disk-store=datastore1
```

### Clone with Custom Hardware Configuration

Clone snapshot with custom CPU and memory:

```bash
./esxi-manager \
  --esxi-host-uri=192.168.1.100 \
  --esxi-host-username=root \
  --esxi-host-password='password' \
  --command=clone-vm-from-snapshot \
  --source-vm-id=42 \
  --snapshot-id=5 \
  --dest-vm-name=custom-clone \
  --dest-vm-disk-store=datastore2 \
  --dest-vm-cpu=8 \
  --dest-vm-ram=16384 \
  --dest-disk-type=zeroedthick \
  --raw-json
```

### Clone with Disk Expansion

Clone snapshot and expand the disk size:

```bash
./esxi-manager \
  --esxi-host-uri=192.168.1.100 \
  --esxi-host-username=root \
  --esxi-host-password='password' \
  --command=clone-vm-from-snapshot \
  --source-vm-name=template-vm \
  --snapshot-name=base-snapshot \
  --dest-vm-name=large-clone \
  --dest-vm-disk-store=datastore1 \
  --dest-vm-cpu=8 \
  --dest-vm-ram=16384 \
  --dest-disk-size=50 \
  --dest-disk-type=zeroedthick \
  --raw-json
```

**Note**: If requested size is smaller than the source snapshot, the disk will retain the source size and report the limitation in the operation steps.

### Clone with Single Network Portgroup

Specify a single custom network for the cloned VM:

```bash
./esxi-manager \
  --esxi-host-uri=192.168.1.100 \
  --esxi-host-username=root \
  --esxi-host-password='password' \
  --command=clone-vm-from-snapshot \
  --source-vm-name=web-template \
  --snapshot-name=stable-v2.1 \
  --dest-vm-name=web-server-03 \
  --dest-vm-disk-store=fast-storage \
  --dest-network-list=prod-network \
  --dest-vm-cpu=4 \
  --dest-vm-ram=8192
```

### Clone with Multiple Network Portgroups

Assign multiple network adapters to the cloned VM:

```bash
./esxi-manager \
  --esxi-host-uri=192.168.1.100 \
  --esxi-host-username=root \
  --esxi-host-password='password' \
  --command=clone-vm-from-snapshot \
  --source-vm-name=db-template \
  --snapshot-name=production \
  --dest-vm-name=db-server-01 \
  --dest-vm-disk-store=fast-storage \
  --dest-network-list="prod-network,mgmt-network,backup-network" \
  --dest-vm-cpu=16 \
  --dest-vm-ram=32768 \
  --dest-disk-size=100 \
  --dest-disk-type=eagerzeroedthick \
  --raw-json
```

This creates a VM with 3 network adapters:
- ethernet0 → prod-network
- ethernet1 → mgmt-network
- ethernet2 → backup-network

## Output Format

The command returns JSON with the following structure:

```json
{
  "status": "success",
  "message": "Successfully cloned VM 'template-vm' from snapshot 'production-ready' to 'clone-vm-01'",
  "source_vm": "template-vm",
  "snapshot": "production-ready",
  "destination_vm": "clone-vm-01",
  "datastore": "datastore1",
  "configuration": {
    "cpu_cores": 4,
    "memory_mb": 8192,
    "disk_type": "thin",
    "disk_size_gb": 16,
    "network": "portGroup-test-01,portGroup-test-02",
    "config_source": "source"
  },
  "operation_steps": [
    {
      "step": 1,
      "name": "Validating source VM",
      "status": "success",
      "detail": "VM ID: 42"
    },
    {
      "step": 2,
      "name": "Finding snapshot",
      "status": "success",
      "detail": "Snapshot ID: 5"
    },
    ...
  ]
}
```

### Response Fields

- **status**: `success` or `failed`
- **message**: Human-readable status message
- **source_vm**: Name of the source VM
- **snapshot**: Name of the snapshot that was cloned
- **destination_vm**: Name of the newly created VM
- **datastore**: Target datastore where the VM was created
- **configuration**: The configuration applied to the cloned VM
  - **cpu_cores**: Number of CPU cores
  - **memory_mb**: Memory in megabytes
  - **disk_type**: Disk provisioning type used (thin, zeroedthick, or eagerzeroedthick)
  - **disk_size_gb**: Final disk size in GB (may differ from requested if shrinking was attempted)
  - **network**: Network/portgroup(s) assigned (comma-separated if multiple)
  - **config_source**: Whether config came from "source" VM or "custom" specification
- **operation_steps**: Array of steps executed during cloning, each with:
  - **step**: Step number
  - **name**: Step name
  - **status**: `success`, `failed`, or `in_progress`
  - **detail**: Additional information about the step
- **error**: Error message (if status is `failed`)

## Operation Steps

The command performs the following steps:

1. **Validate source VM** - Ensure source VM exists and retrieve its VM ID
2. **Find snapshot** - Locate and validate the specified snapshot by name or ID
3. **Read source VM configuration** - Extract CPU, memory, disk, and network settings from source
4. **Check destination VM name** - Ensure the destination VM name doesn't already exist
5. **Get source VM details** - Extract source VM datastore location and path
6. **Create destination directory** - Create the target VM folder on the destination datastore
7. **Clone snapshot disk** - Copy snapshot VMDK file to new location and optionally expand disk:
   - Clones the snapshot disk file using vmkfstools
   - If `--dest-disk-size` is specified and larger than source, expands using `vmkfstools -X`
   - If `--dest-disk-size` is smaller than source, uses source size and reports limitation
   - Reports actual disk configuration in operation detail
8. **Create VM configuration** - Generate VMX file with customized settings:
   - Sets CPU cores and memory from parameters or source
   - Configures network adapters from `--dest-network-list` or inherits from source
   - Removes snapshot references and points to base VMDK
9. **Register VM** - Register the new VM with ESXi host to make it available

## Failure Scenarios

If any step fails, the command:
- Stops execution
- Returns `status: "failed"` with error details
- Attempts to clean up created resources (directories)
- Provides detailed error message in the `error` field

## Notes

- **VM Names**: Must be unique on the ESXi host
- **Snapshots**: The snapshot must exist on the source VM; cloning from a snapshot does not create a live snapshot
- **Disk Type**:
  - Thin provisioning is the default and uses least initial disk space
  - Eager-zeroed thick is recommended for high-performance storage
  - Lazy-zeroed thick provides a balance between performance and compatibility
- **Disk Sizing**:
  - Can only expand disks during cloning, not shrink
  - If a smaller size is requested, the source snapshot size is used
  - Expansion uses vmkfstools and may fail if insufficient space is available
- **Network Configuration**:
  - If `--dest-network-list` is not specified, the cloned VM automatically inherits all portgroups from the source VM
  - Use comma-separated values (no spaces) to specify multiple networks: `"network1,network2,network3"`
  - Each specified network becomes an ethernet adapter in sequence (ethernet0, ethernet1, etc.)
- **Post-Clone**: When first powering on the cloned VM in vSphere, select "I Copied It" to update machine IDs
- **SSH Access**: Requires SSH access to the ESXi host and sufficient permissions to modify datastores

## Troubleshooting

### "Snapshot not found"
- Verify the snapshot name or ID is correct
- Check the snapshot exists on the source VM using `vm-inspect`

### "VM already exists"
- Choose a different destination VM name
- Use list-vms to see existing VMs

### "Failed to create directory"
- Verify the target datastore exists and has sufficient permissions
- Check datastore has available space

### "Failed to clone VMDK"
- Ensure adequate disk space on target datastore
- Verify SSH permissions for ESXi user

### Disk size shows "cannot shrink from 16GB to 10GB"
- This is expected behavior - vmkfstools can only expand disks, not shrink them
- To use a smaller disk, specify a size larger than or equal to the source snapshot
- The operation succeeds but retains the source disk size

### "Failed to extend disk: There is not enough space"
- The target datastore doesn't have enough free space for the expanded disk
- Check available space using `list-datastores`
- Either request a smaller disk size or use a different datastore

### VM has no network connectivity after cloning
- If `--dest-network-list` was used, verify the portgroup names are correct
- The portgroup must exist on the ESXi host
- If not specified, the VM should inherit networks from the source VM
- Check network configuration in VMX file: `cat /vmfs/volumes/datastore/vm-name/vm-name.vmx | grep ethernet`

### Multiple network adapters not appearing in VM
- Use comma-separated format without spaces: `"network1,network2,network3"`
- With spaces: `"network1, network2"` will fail to parse correctly
- Verify portgroup names exactly match those on the ESXi host

## See Also

- `list-vms` - List all virtual machines
- `vm-inspect` - Get detailed VM information including snapshots
- `list-datastores` - List available datastores
