package storage

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/command"
	"github.com/esxi-manager/esxi-manager/internal/esxi/utils"
)

// Datastore represents a datastore entry
type Datastore struct {
	MountPoint string
	VolumeName string
	UUID       string
	Mounted    string
	Type       string
	Size       string
	Free       string
}

// ListDatastoresCommand lists all datastores on the ESXi host
type ListDatastoresCommand struct {
	host *config.ESXiHost
}

func (c *ListDatastoresCommand) Validate() error {
	return nil
}

func (c *ListDatastoresCommand) Execute() error {
	// Manage connection internally
	manager, err := utils.NewSSHManager(c.host)
	if err != nil {
		return fmt.Errorf("failed to create SSH manager: %w", err)
	}
	defer manager.Close()

	// Connect to host
	if err := manager.Connect(); err != nil {
		return fmt.Errorf("failed to connect to ESXi host: %w", err)
	}

	// Get client
	client, err := manager.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get SSH client: %w", err)
	}

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// List datastores using esxcli command
	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run("esxcli storage filesystem list"); err != nil {
		return fmt.Errorf("failed to execute list datastores command: %w", err)
	}

	// Parse and format output
	if err := c.formatOutput(stdout.String()); err != nil {
		return err
	}
	return nil
}

func (c *ListDatastoresCommand) formatOutput(output string) error {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		fmt.Println(output)
		return nil
	}

	// Parse datastores - skip header (first line) and separator (second line)
	var datastores []Datastore
	for i := 2; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split by multiple spaces to parse columns
		fields := strings.Fields(line)
		if len(fields) >= 7 {
			ds := Datastore{
				MountPoint: fields[0],
				VolumeName: fields[1],
				UUID:       fields[2],
				Mounted:    fields[3],
				Type:       fields[4],
				Size:       fields[5],
				Free:       fields[6],
			}
			datastores = append(datastores, ds)
		}
	}

	// Format and display using tabwriter
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Name\tType\tSize\tFree\tMounted\n")

	for _, ds := range datastores {
		sizeBytes := parseBytes(ds.Size)
		freeBytes := parseBytes(ds.Free)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ds.VolumeName,
			ds.Type,
			utils.FormatBytes(sizeBytes),
			utils.FormatBytes(freeBytes),
			ds.Mounted,
		)
	}
	w.Flush()
	fmt.Print(buf.String())
	return nil
}

// parseBytes converts string representation of bytes to int64
func parseBytes(s string) int64 {
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}

// init registers the list-datastores command
func init() {
	command.Register("list-datastores", func(params command.Params, host *config.ESXiHost) command.Interface {
		return &ListDatastoresCommand{
			host: host,
		}
	})
}
