package keactrl

// An enum representing the name of the Kea daemon.
type ServiceName string
type DHCPServiceName string

const (
	DaemonNameDHCP4        ServiceName     = "dhcp4"
	DaemonNameDHCP6        ServiceName     = "dhcp6"
	DaemonNameControlAgent ServiceName     = ""
	DaemonNameD2           ServiceName     = "d2"
	DHCPDaemonNameDHCP4    DHCPServiceName = "dhcp4"
	DHCPDaemonNameDHCP6    DHCPServiceName = "dhcp6"
)

func wrapServices(services []ServiceName) []string {
	var names []string

	for _, service := range services {
		if service == DaemonNameControlAgent {
			continue
		}
		names = append(names, string(service))
	}

	return names
}

func wrapService(service ServiceName) []string {
	if service == DaemonNameControlAgent {
		return nil
	}
	return []string{string(service)}

}

func wrapDHCPServices(services []DHCPServiceName) []string {
	var names []string

	for _, service := range services {
		names = append(names, string(service))
	}

	return names
}

func wrapDHCPService(service DHCPServiceName) []string {
	return []string{string(service)}
}

func NewListCommandsCommand() *Command {
	return NewCommand("list-commands", nil, nil)
}

func NewVersionGetCommand(services ...ServiceName) *Command {
	return NewCommand("version-get", wrapServices(services), nil)
}

func NewStatusGetCommand(services ...DHCPServiceName) *Command {
	return NewCommand("status-get", wrapDHCPServices(services), nil)
}

func NewConfigGetCommand(services ...ServiceName) *Command {
	return NewCommand("config-get", wrapServices(services), nil)
}

func NewConfigWriteCommand(service ServiceName) *Command {
	return NewCommand("config-write", wrapService(service), nil)
}

func NewConfigReloadCommand(service ServiceName) *Command {
	return NewCommand("config-reload", wrapService(service), nil)
}
