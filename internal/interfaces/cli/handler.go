package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/esxi-manager/esxi-manager/internal/esxi"
)

// PrintUsage prints the CLI usage help text
func PrintUsage() {
	fmt.Fprintf(os.Stderr, `ESXi Manager CLI

USAGE:
  esxi-manager [options]

REQUIRED FLAGS:
  --esxi-host-uri         ESXi host URI/IP address
  --esxi-host-username    ESXi host username
  --esxi-host-password    ESXi host password
  --command              Command to execute

OPTIONAL FLAGS:
  --esxi-host-port       ESXi host SSH port (default: 22)
  --raw-json             Output raw JSON without CLI formatting

AVAILABLE COMMANDS:
  test-connection           Test SSH connectivity to ESXi host
  get-version              Get ESXi host version information

  Virtual Machines:
    list-vms               List all virtual machines on ESXi host
    vm-inspect             Inspect specific VM (requires --vm-name or --vm-id)

  Storage:
    list-datastores        List all datastores on ESXi host
    datastore-inspect      Inspect specific datastore (requires --datastore-name)

  Networking:
    list-portgroups        List all port groups on ESXi host
    portgroup-inspect      Inspect specific port group (requires --portgroup-name)
    list-vswitches         List all virtual switches
    vswitch-inspect        Inspect specific vswitch (requires --vswitch-name)
    list-network-adapters  List all physical network adapters
    list-vmknics           List all VM kernel NICs

EXAMPLE:
  esxi-manager \
    --esxi-host-uri="192.168.0.186" \
    --esxi-host-port="22" \
    --esxi-host-username="root" \
    --esxi-host-password="H@med1382" \
    --command="test-connection"

NOTE: All flags must come BEFORE any positional arguments.
`)
}

// ParseFlags parses command-line flags and returns parsed parameters
// Returns: (params, commandName, error)
func ParseFlags() (*esxi.Params, string, error) {
	params := &esxi.Params{}
	params.RegisterFlags()

	commandName := flag.String("command", "", "Command to execute (test-connection, get-version, list-vms, etc)")
	flag.Parse()

	// Reject positional arguments
	if flag.NArg() > 0 {
		args := flag.Args()
		return nil, "", fmt.Errorf("unexpected positional argument(s): %v\n\nHint: All flags (including --command) must come at the start, before any positional arguments.\nDid you put '%s' before your flags?", args, args[0])
	}

	// Validate connection parameters
	if err := esxi.ValidateConnectionParams(params); err != nil {
		return nil, "", err
	}

	// Validate command is specified
	if *commandName == "" {
		return nil, "", esxi.NewValidationError("command", "is required")
	}

	return params, *commandName, nil
}

// Execute dispatches and executes the specified command
// Handles output formatting based on the RawJSON parameter:
//   - If RawJSON is true: outputs the raw JSON from the esxi command as-is
//   - If RawJSON is false: parses JSON and formats as human-readable table/text
func Execute(commandName string, params *esxi.Params) error {
	// Dispatch command from registry
	cmd, err := esxi.Dispatch(commandName, params)
	if err != nil {
		return esxi.WrapError(err, "failed to dispatch command")
	}

	// Execute the command and get the JSON output
	jsonOutput, err := cmd.Execute()
	if err != nil {
		return esxi.WrapError(err, "command execution failed")
	}

	// If RawJSON is true, output the raw JSON from esxi layer
	if params.RawJSON {
		fmt.Println(jsonOutput)
		return nil
	}

	// Otherwise, format the JSON output as human-readable table/text
	// Parse the JSON to determine the structure (array vs object)
	var data interface{}
	if err := json.Unmarshal([]byte(jsonOutput), &data); err != nil {
		// If JSON parsing fails, just output the raw string
		fmt.Println(jsonOutput)
		return nil
	}

	// Format the output based on data type
	switch v := data.(type) {
	case []interface{}:
		// Array of objects - format as table
		if len(v) > 0 {
			if mapData, ok := v[0].(map[string]interface{}); ok {
				// Get column names from first object and sort them for consistent output
				var columns []string
				for col := range mapData {
					columns = append(columns, col)
				}
				// Sort columns for consistent output
				sort.Strings(columns)

				// Convert to rows format for table formatter
				var rows []map[string]interface{}
				for _, item := range v {
					rows = append(rows, item.(map[string]interface{}))
				}

				formatted := FormatAsTable(rows, columns)
				fmt.Println(formatted)
			} else {
				fmt.Println(jsonOutput)
			}
		} else {
			fmt.Println("No data")
		}
	case map[string]interface{}:
		// Single object - output as readable key-value pairs
		for key, value := range v {
			fmt.Printf("%s: %v\n", key, value)
		}
	default:
		// Fallback to raw JSON
		fmt.Println(jsonOutput)
	}

	return nil
}
