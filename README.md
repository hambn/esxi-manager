# esxi-manager

A powerful, multi-platform ESXi infrastructure management tool written in Go. Manage your virtualized infrastructure with simple YAML configurations, CLI parameters, Terraform integration, or Docker containers.

## Features

- **Multi-language Interface**: Use YAML configuration files, CLI parameters, or Terraform modules—all with the same capabilities
- **Sequential Task Execution**: Define multiple tasks in your YAML configuration and execute them one-by-one
- **Flexible Deployment Options**:
  - Standalone binary/package
  - Docker container
  - Terraform module/provider
- **Rich Operations Support**:
  - VM cloning (`--clone-vm`)
  - Virtual switch creation (`--create-vswitch`)
  - And much more...
- **Cross-Platform Support**: Run on multiple operating systems
- **Extensible Architecture**: Easy to add new features and integrations (future Ansible support planned)

## Installation

### Binary/Package

```bash
# Build from source
go build -o esxi-manager ./cmd/esxi-manager
```

### Docker

```bash
docker build -t esxi-manager .
docker run esxi-manager --help
```

### Terraform

Include the esxi-manager module in your Terraform configuration:

```hcl
module "esxi_manager" {
  source = "./terraform"
  # configuration...
}
```

## Usage

### YAML Configuration File

Define your infrastructure tasks in a YAML file:

```yaml
tasks:
  - type: clone-vm
    source: "template-vm"
    destination: "new-vm-clone"
    host: "esxi-host.example.com"

  - type: create-vswitch
    name: "vswitch-prod"
    host: "esxi-host.example.com"
    mtu: 1500

  - type: add-network-adapter
    vm: "new-vm-clone"
    vswitch: "vswitch-prod"
```

Run with:

```bash
esxi-manager --config=infrastructure.yml
```

### CLI Parameters

Execute operations directly via command-line parameters:

```bash
esxi-manager --esxi-host="192.168.1.100" --clone-vm --source="template" --destination="prod-vm"
esxi-manager --esxi-host="192.168.1.100" --create-vswitch --vswitch-name="prod-switch"
```

### Docker Container

```bash
docker run \
  -v $(pwd)/config.yml:/etc/esxi-manager/config.yml \
  esxi-manager --config=/etc/esxi-manager/config.yml
```

### Terraform Integration

All operations are available through Terraform resources:

```hcl
resource "esxi-manager_vm_clone" "example" {
  esxi_host   = "192.168.1.100"
  source_vm   = "template"
  dest_vm     = "prod-vm"
}

resource "esxi-manager_vswitch" "prod" {
  esxi_host = "192.168.1.100"
  name      = "prod-switch"
  mtu       = 1500
}
```

## Project Structure

This project follows the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) directory structure

## Configuration

### YAML Schema

Tasks support the following operations:

- `clone-vm`: Clone a virtual machine
- `create-vswitch`: Create a new virtual switch
- `add-network-adapter`: Add network adapter to VM
- `remove-vm`: Delete a virtual machine
- And more...

## Future Roadmap

- Ansible-like playbook support for even more flexible automation
- Additional provider integrations
- Advanced scheduling and monitoring
- Custom plugin system for extensibility

## Contributing

Contributions are welcome! To add new features:

1. Add the operation type to the supported task types
2. Implement handlers in the appropriate package
3. Update documentation and examples
4. Ensure the feature works across YAML, CLI, and Terraform interfaces

## License

[Add your license here]
