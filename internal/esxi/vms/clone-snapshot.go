package vms

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/esxi-manager/esxi-manager/internal/esxi/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// CloneSnapshotResponse represents the result of a snapshot clone operation
type CloneSnapshotResponse struct {
	Status          string                 `json:"status"`
	Message         string                 `json:"message"`
	SourceVM        string                 `json:"source_vm"`
	Snapshot        string                 `json:"snapshot"`
	DestinationVM   string                 `json:"destination_vm"`
	Datastore       string                 `json:"datastore"`
	Configuration   *CloneConfiguration    `json:"configuration,omitempty"`
	OperationSteps  []OperationStep        `json:"operation_steps,omitempty"`
	Error           string                 `json:"error,omitempty"`
}

// CloneConfiguration holds the configuration applied to the cloned VM
type CloneConfiguration struct {
	CPUCores       int    `json:"cpu_cores"`
	MemoryMB       int    `json:"memory_mb"`
	DiskType       string `json:"disk_type"`
	DiskSizeGB     int    `json:"disk_size_gb,omitempty"`
	Network        string `json:"network,omitempty"`
	ConfigSource   string `json:"config_source"` // "source" or "custom"
}

// OperationStep represents a single step in the clone operation
type OperationStep struct {
	Step   int    `json:"step"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// cloneVMFromSnapshot clones a VM from a snapshot with customizable configuration
func cloneVMFromSnapshot(params *config.Params) (string, error) {
	response := &CloneSnapshotResponse{
		Status:        "failed",
		OperationSteps: []OperationStep{},
	}

	// Validate required parameters
	if params.SourceVMName == "" && params.SourceVMID == "" {
		response.Error = "either source-vm-name or source-vm-id must be provided"
		return formatCloneResponse(response)
	}

	if params.DestVMName == "" {
		response.Error = "dest-vm-name is required"
		return formatCloneResponse(response)
	}

	if params.DestDiskStore == "" {
		response.Error = "dest-vm-disk-store is required"
		return formatCloneResponse(response)
	}

	if params.SnapshotName == "" && params.SnapshotID == "" {
		response.Error = "either snapshot-name or snapshot-id must be provided"
		return formatCloneResponse(response)
	}

	// Create SSH manager
	manager, err := utils.NewSSHManager(params)
	if err != nil {
		response.Error = fmt.Sprintf("SSH connection failed: %v", err)
		return formatCloneResponse(response)
	}
	defer manager.Close()

	// Step 1: Get source VM details
	addStep(response, 1, "Validating source VM", "in_progress", "")
	sourceVMID := params.SourceVMID
	sourceVMName := params.SourceVMName

	if sourceVMID == "" {
		// Find VM ID by name
		id, err := findVMIDByName(manager, sourceVMName)
		if err != nil {
			response.Error = fmt.Sprintf("failed to find VM: %v", err)
			addStep(response, 1, "Validating source VM", "failed", response.Error)
			return formatCloneResponse(response)
		}
		sourceVMID = id
	} else if sourceVMName == "" {
		// Find VM name by ID
		name, err := findVMNameByID(manager, sourceVMID)
		if err != nil {
			response.Error = fmt.Sprintf("failed to find VM: %v", err)
			addStep(response, 1, "Validating source VM", "failed", response.Error)
			return formatCloneResponse(response)
		}
		sourceVMName = name
	}

	response.SourceVM = sourceVMName
	addStep(response, 1, "Validating source VM", "success", fmt.Sprintf("VM ID: %s", sourceVMID))

	// Step 2: Find and validate snapshot
	addStep(response, 2, "Finding snapshot", "in_progress", "")
	snapshotID := params.SnapshotID
	snapshotName := params.SnapshotName

	if snapshotID == "" {
		// Find snapshot ID by name
		id, err := findSnapshotIDByName(manager, sourceVMID, snapshotName)
		if err != nil {
			response.Error = fmt.Sprintf("failed to find snapshot: %v", err)
			addStep(response, 2, "Finding snapshot", "failed", response.Error)
			return formatCloneResponse(response)
		}
		snapshotID = id
	} else if snapshotName == "" {
		// Find snapshot name by ID
		name, err := findSnapshotNameByID(manager, sourceVMID, snapshotID)
		if err != nil {
			response.Error = fmt.Sprintf("failed to find snapshot: %v", err)
			addStep(response, 2, "Finding snapshot", "failed", response.Error)
			return formatCloneResponse(response)
		}
		snapshotName = name
	}

	response.Snapshot = snapshotName
	response.Datastore = params.DestDiskStore
	addStep(response, 2, "Finding snapshot", "success", fmt.Sprintf("Snapshot ID: %s", snapshotID))

	// Step 3: Get source VM configuration
	addStep(response, 3, "Reading source VM configuration", "in_progress", "")
	srcConfig, err := getVMConfiguration(manager, sourceVMID)
	if err != nil {
		response.Error = fmt.Sprintf("failed to read source VM config: %v", err)
		addStep(response, 3, "Reading source VM configuration", "failed", response.Error)
		return formatCloneResponse(response)
	}
	addStep(response, 3, "Reading source VM configuration", "success", "")

	// Determine target configuration
	targetCPU := params.DestCPU
	targetMemoryMB := params.DestRAM
	targetNetwork := params.DestNetwork
	configSource := "custom"

	if targetCPU == 0 && targetMemoryMB == 0 {
		// Use source configuration
		targetCPU = srcConfig.CPU
		targetMemoryMB = srcConfig.MemoryMB
		if targetNetwork == "" {
			targetNetwork = srcConfig.Network
		}
		configSource = "source"
	}

	// Step 4: Verify destination VM doesn't exist
	addStep(response, 4, "Checking destination VM name", "in_progress", "")
	exists, err := vmExists(manager, params.DestVMName)
	if err != nil {
		response.Error = fmt.Sprintf("failed to check VM existence: %v", err)
		addStep(response, 4, "Checking destination VM name", "failed", response.Error)
		return formatCloneResponse(response)
	}
	if exists {
		response.Error = fmt.Sprintf("destination VM '%s' already exists", params.DestVMName)
		addStep(response, 4, "Checking destination VM name", "failed", response.Error)
		return formatCloneResponse(response)
	}
	addStep(response, 4, "Checking destination VM name", "success", "")

	// Step 5: Get source VM path information
	addStep(response, 5, "Getting source VM details", "in_progress", "")
	srcVMPath, srcDatastore, err := getVMPath(manager, sourceVMID)
	if err != nil {
		response.Error = fmt.Sprintf("failed to get VM path: %v", err)
		addStep(response, 5, "Getting source VM details", "failed", response.Error)
		return formatCloneResponse(response)
	}
	addStep(response, 5, "Getting source VM details", "success", fmt.Sprintf("Datastore: %s", srcDatastore))

	// Step 6: Create destination directory
	addStep(response, 6, "Creating destination directory", "in_progress", "")
	destPath := fmt.Sprintf("/vmfs/volumes/%s/%s", params.DestDiskStore, params.DestVMName)
	cmd := fmt.Sprintf("mkdir -p '%s'", destPath)
	_, err = manager.RunCommand(cmd)
	if err != nil {
		response.Error = fmt.Sprintf("failed to create directory: %v", err)
		addStep(response, 6, "Creating destination directory", "failed", response.Error)
		return formatCloneResponse(response)
	}
	addStep(response, 6, "Creating destination directory", "success", destPath)

	// Step 7: Clone snapshot VMDK
	addStep(response, 7, "Cloning snapshot disk", "in_progress", "")

	// Find snapshot VMDK file - use the snapshot descriptor file
	snapshotVmdkFile := fmt.Sprintf("%s-%s.vmdk", sourceVMName, padSnapshotID(snapshotID))
	srcVmdkFullPath := fmt.Sprintf("/vmfs/volumes/%s/%s/%s", srcDatastore, srcVMPath, snapshotVmdkFile)
	destVmdkFullPath := fmt.Sprintf("%s/%s.vmdk", destPath, params.DestVMName)

	// Handle disk type
	diskTypeArg := "thin"
	if params.DestDiskType != "" {
		diskTypeArg = params.DestDiskType
	}

	// Clone the snapshot VMDK (flattens the chain)
	cloneCmd := fmt.Sprintf("vmkfstools -i '%s' '%s' -d %s", srcVmdkFullPath, destVmdkFullPath, diskTypeArg)
	_, err = manager.RunCommand(cloneCmd)
	if err != nil {
		response.Error = fmt.Sprintf("failed to clone VMDK: %v", err)
		addStep(response, 7, "Cloning snapshot disk", "failed", response.Error)
		// Cleanup: remove the created directory
		cleanupCmd := fmt.Sprintf("rm -rf '%s'", destPath)
		manager.RunCommand(cleanupCmd)
		return formatCloneResponse(response)
	}

	// Step 7b: Resize disk if custom size is specified (only expand, not shrink)
	resizeDetail := fmt.Sprintf("Disk Type: %s", diskTypeArg)
	if params.DestDiskSize > 0 {
		// Query current VMDK size (RW line is typically around line 12 after extent description)
		queryCmd := fmt.Sprintf("grep '^RW' '%s' | awk '{print $2}'", destVmdkFullPath)
		sectorOutput, _ := manager.RunCommand(queryCmd)
		sectorOutput = strings.TrimSpace(sectorOutput)

		if sectorOutput != "" {
			if sectors, err := strconv.ParseInt(sectorOutput, 10, 64); err == nil {
				currentSizeGB := (sectors * 512) / 1073741824
				if params.DestDiskSize > int(currentSizeGB) {
					// Can expand
					resizeCmd := fmt.Sprintf("vmkfstools -X %dG '%s' 2>&1", params.DestDiskSize, destVmdkFullPath)
					resizeOutput, resizeErr := manager.RunCommand(resizeCmd)

					if resizeErr == nil || strings.Contains(resizeOutput, "100% done") || strings.Contains(resizeOutput, "Grow: 100%") {
						resizeDetail = fmt.Sprintf("Disk Type: %s, Resized to %dGB", diskTypeArg, params.DestDiskSize)
					} else {
						resizeDetail = fmt.Sprintf("Disk Type: %s (resize to %dGB failed, using %dGB)", diskTypeArg, params.DestDiskSize, currentSizeGB)
					}
				} else if params.DestDiskSize < int(currentSizeGB) {
					resizeDetail = fmt.Sprintf("Disk Type: %s (cannot shrink from %dGB to %dGB, using %dGB)", diskTypeArg, currentSizeGB, params.DestDiskSize, currentSizeGB)
				} else {
					resizeDetail = fmt.Sprintf("Disk Type: %s, Size: %dGB", diskTypeArg, currentSizeGB)
				}
			}
		}
	}
	addStep(response, 7, "Cloning snapshot disk", "success", resizeDetail)

	// Step 8: Copy and modify VMX file
	addStep(response, 8, "Creating VM configuration", "in_progress", "")
	srcVmxPath := fmt.Sprintf("/vmfs/volumes/%s/%s/%s.vmx", srcDatastore, srcVMPath, sourceVMName)
	destVmxPath := fmt.Sprintf("%s/%s.vmx", destPath, params.DestVMName)

	// Copy VMX file
	cpCmd := fmt.Sprintf("cp '%s' '%s'", srcVmxPath, destVmxPath)
	_, err = manager.RunCommand(cpCmd)
	if err != nil {
		response.Error = fmt.Sprintf("failed to copy VMX file: %v", err)
		addStep(response, 8, "Creating VM configuration", "failed", response.Error)
		cleanupCmd := fmt.Sprintf("rm -rf '%s'", destPath)
		manager.RunCommand(cleanupCmd)
		return formatCloneResponse(response)
	}

	// Modify VMX file: replace VM name references
	sedCmd := fmt.Sprintf("sed -i 's/%s/%s/g' '%s'", sourceVMName, params.DestVMName, destVmxPath)
	_, err = manager.RunCommand(sedCmd)
	if err != nil {
		response.Error = fmt.Sprintf("failed to update VMX file: %v", err)
		addStep(response, 8, "Creating VM configuration", "failed", response.Error)
		cleanupCmd := fmt.Sprintf("rm -rf '%s'", destPath)
		manager.RunCommand(cleanupCmd)
		return formatCloneResponse(response)
	}

	// Remove snapshot references from VMX
	cleanupVmxCmd := fmt.Sprintf("sed -i '/snapshot\\./d; /\\.redo/d; /checkpoint\\./d' '%s'", destVmxPath)
	manager.RunCommand(cleanupVmxCmd)

	// Fix disk file references to point to base VMDK (remove snapshot references)
	fixDiskCmd := fmt.Sprintf("sed -i 's/scsi0:0\\.fileName = \".*\"/scsi0:0.fileName = \"%s.vmdk\"/' '%s'", params.DestVMName, destVmxPath)
	manager.RunCommand(fixDiskCmd)

	// Determine network configuration
	networkList := params.DestNetworkList
	if networkList == "" {
		// Extract from source VM VMX
		networkList = extractNetworksFromVMX(manager, srcVmxPath)
	}

	// Remove existing network config from VMX
	removeNetCmd := fmt.Sprintf("sed -i '/ethernet[0-9]*\\./d' '%s'", destVmxPath)
	manager.RunCommand(removeNetCmd)

	// Add network configuration
	if networkList != "" {
		networks := strings.Split(networkList, ",")
		for idx, network := range networks {
			network = strings.TrimSpace(network)
			if network != "" {
				netConfig := fmt.Sprintf("ethernet%d.present = \"TRUE\"\nethernet%d.virtualDev = \"e1000\"\nethernet%d.networkName = \"%s\"\nethernet%d.addressType = \"generated\"\n",
					idx, idx, idx, network, idx)
				// Append network config to VMX
				appendCmd := fmt.Sprintf("echo '%s' >> '%s'", escapeForShell(netConfig), destVmxPath)
				manager.RunCommand(appendCmd)
			}
		}
	}

	// Update hardware settings
	updateCPUCmd := fmt.Sprintf("sed -i 's/numvcpus = .*/numvcpus = \"%d\"/' '%s'", targetCPU, destVmxPath)
	manager.RunCommand(updateCPUCmd)

	updateMemCmd := fmt.Sprintf("sed -i 's/memSize = .*/memSize = \"%d\"/' '%s'", targetMemoryMB, destVmxPath)
	manager.RunCommand(updateMemCmd)

	addStep(response, 8, "Creating VM configuration", "success", "")

	// Step 9: Register VM
	addStep(response, 9, "Registering VM", "in_progress", "")
	registerCmd := fmt.Sprintf("vim-cmd solo/registervm '%s'", destVmxPath)
	_, err = manager.RunCommand(registerCmd)
	if err != nil {
		response.Error = fmt.Sprintf("failed to register VM: %v", err)
		addStep(response, 9, "Registering VM", "failed", response.Error)
		cleanupCmd := fmt.Sprintf("rm -rf '%s'", destPath)
		manager.RunCommand(cleanupCmd)
		return formatCloneResponse(response)
	}
	addStep(response, 9, "Registering VM", "success", "")

	// Success!
	response.Status = "success"
	response.Message = fmt.Sprintf("Successfully cloned VM '%s' from snapshot '%s' to '%s'", sourceVMName, snapshotName, params.DestVMName)
	response.Configuration = &CloneConfiguration{
		CPUCores:     targetCPU,
		MemoryMB:     targetMemoryMB,
		DiskType:     diskTypeArg,
		DiskSizeGB:   params.DestDiskSize,
		Network:      targetNetwork,
		ConfigSource: configSource,
	}

	return formatCloneResponse(response)
}

// Helper functions

func findVMIDByName(mgr *utils.SSHManager, vmName string) (string, error) {
	output, err := mgr.RunCommand("vim-cmd vmsvc/getallvms 2>/dev/null")
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == vmName {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("VM not found: %s", vmName)
}

func findVMNameByID(mgr *utils.SSHManager, vmID string) (string, error) {
	output, err := mgr.RunCommand("vim-cmd vmsvc/getallvms 2>/dev/null")
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == vmID {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("VM not found: %s", vmID)
}

func findSnapshotIDByName(mgr *utils.SSHManager, vmID string, snapshotName string) (string, error) {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/snapshot.get %s 2>/dev/null", vmID))
	if err != nil {
		return "", err
	}

	var currentID string
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Snapshot Name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) == snapshotName {
				// Next line should have the ID
				continue
			}
		}
		if strings.Contains(line, "Snapshot Id") && currentID == "" {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentID = strings.TrimSpace(parts[1])
			}
		}
	}

	if currentID != "" {
		return currentID, nil
	}
	return "", fmt.Errorf("snapshot not found: %s", snapshotName)
}

func findSnapshotNameByID(mgr *utils.SSHManager, vmID string, snapshotID string) (string, error) {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/snapshot.get %s 2>/dev/null", vmID))
	if err != nil {
		return "", err
	}

	var currentName string
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Snapshot Id") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) == snapshotID {
				// Look for name in previous/next lines
				continue
			}
		}
		if strings.Contains(line, "Snapshot Name") && currentName == "" {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentName = strings.TrimSpace(parts[1])
			}
		}
	}

	if currentName != "" {
		return currentName, nil
	}
	return "", fmt.Errorf("snapshot not found: %s", snapshotID)
}

func getVMConfiguration(mgr *utils.SSHManager, vmID string) (*VMConfig, error) {
	output, err := mgr.RunCommand(fmt.Sprintf("vim-cmd vmsvc/get.config %s 2>/dev/null", vmID))
	if err != nil {
		return nil, err
	}

	config := &VMConfig{CPU: 1, MemoryMB: 1024}

	// Parse CPU
	if match := regexp.MustCompile(`numCpus\s*=\s*(\d+)`).FindStringSubmatch(output); len(match) > 1 {
		if cpu, err := strconv.Atoi(match[1]); err == nil {
			config.CPU = cpu
		}
	}

	// Parse Memory
	if match := regexp.MustCompile(`memSize\s*=\s*(\d+)`).FindStringSubmatch(output); len(match) > 1 {
		if mem, err := strconv.Atoi(match[1]); err == nil {
			config.MemoryMB = mem
		}
	}

	// Parse Network
	if match := regexp.MustCompile(`ethernet\d+\.networkName\s*=\s*"([^"]+)"`).FindStringSubmatch(output); len(match) > 1 {
		config.Network = match[1]
	}

	return config, nil
}

func vmExists(mgr *utils.SSHManager, vmName string) (bool, error) {
	output, err := mgr.RunCommand("vim-cmd vmsvc/getallvms 2>/dev/null")
	if err != nil {
		return false, err
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == vmName {
			return true, nil
		}
	}
	return false, nil
}

func getVMPath(mgr *utils.SSHManager, vmID string) (string, string, error) {
	output, err := mgr.RunCommand("vim-cmd vmsvc/getallvms 2>/dev/null")
	if err != nil {
		return "", "", err
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] != vmID {
			continue
		}

		// Parse [datastore] path
		if match := regexp.MustCompile(`\[([^\]]+)\]\s+(.+?)(?:\s+\w+Guest|\s*$)`).FindStringSubmatch(line); len(match) >= 3 {
			datastore := match[1]
			vmPath := match[2]
			// Remove filename part if present
			if idx := strings.LastIndex(vmPath, "/"); idx >= 0 {
				vmPath = vmPath[:idx]
			}
			return vmPath, datastore, nil
		}
	}
	return "", "", fmt.Errorf("VM path not found for ID: %s", vmID)
}

func padSnapshotID(id string) string {
	// Pad snapshot ID to 6 digits
	numID, _ := strconv.Atoi(id)
	return fmt.Sprintf("%06d", numID)
}

func extractNetworksFromVMX(mgr *utils.SSHManager, vmxPath string) string {
	// Extract network portgroups from VMX file
	output, err := mgr.RunCommand(fmt.Sprintf("grep 'ethernet.*\\.networkName' '%s' 2>/dev/null | sed 's/.*networkName = \"\\([^\"]*\\)\".*/\\1/' | sort | uniq", vmxPath))
	if err != nil || strings.TrimSpace(output) == "" {
		return ""
	}
	// Convert newline-separated values to comma-separated
	networks := strings.Split(strings.TrimSpace(output), "\n")
	return strings.Join(networks, ",")
}

func escapeForShell(s string) string {
	// Escape single quotes for shell
	return strings.ReplaceAll(s, "'", "'\\''")
}

func addStep(response *CloneSnapshotResponse, stepNum int, name string, status string, detail string) {
	response.OperationSteps = append(response.OperationSteps, OperationStep{
		Step:   stepNum,
		Name:   name,
		Status: status,
		Detail: detail,
	})
}

func formatCloneResponse(response *CloneSnapshotResponse) (string, error) {
	return utils.FormatAsJSON(response)
}

// VMConfig holds VM configuration details
type VMConfig struct {
	CPU      int
	MemoryMB int
	Network  string
}

func init() {
	config.RegisterFunc("clone-vm-from-snapshot", cloneVMFromSnapshot)
}
