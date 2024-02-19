package keactrl

import keaconfig "isc.org/stork/appcfg/kea"

func NewSubnet4GetCommand(services ...DHCPServiceName) *Command {
	return NewCommand("subnet4-get", wrapDHCPServices(services), nil)
}

func NewSubnet4AddCommand(subnet *keaconfig.Subnet4) *Command {
	arguments := map[string]any{
		"subnet4": []*keaconfig.Subnet4{subnet},
	}
	return NewCommand("subnet4-add", wrapDHCPService(DHCPDaemonNameDHCP4), arguments)
}

func NewSubnet4UpdateCommand(subnet *keaconfig.Subnet4) *Command {
	arguments := map[string]any{
		"subnet4": []*keaconfig.Subnet4{subnet},
	}
	return NewCommand("subnet4-update", wrapDHCPService(DHCPDaemonNameDHCP4), arguments)
}

func NewNetwork4SubnetAddCommand(id int64, name string) *Command {
	arguments := map[string]any{
		"id":   id,
		"name": name,
	}
	return NewCommand("network4-subnet-add", wrapDHCPServices([]DHCPServiceName{DHCPDaemonNameDHCP4}), arguments)
}

func NewNetwork4SubnetDelCommand(id int64, name string) *Command {
	arguments := map[string]any{
		"id":   id,
		"name": name,
	}
	return NewCommand("network4-subnet-del", wrapDHCPServices([]DHCPServiceName{DHCPDaemonNameDHCP4}), arguments)
}

func NewSubnet6AddCommand(subnet *keaconfig.Subnet6) *Command {
	arguments := map[string]any{
		"subnet6": []*keaconfig.Subnet6{subnet},
	}
	return NewCommand("subnet6-add", wrapDHCPServices([]DHCPServiceName{DHCPDaemonNameDHCP6}), arguments)
}

func NewSubnet6UpdateCommand(subnet *keaconfig.Subnet6) *Command {
	arguments := map[string]any{
		"subnet6": []*keaconfig.Subnet6{subnet},
	}
	return NewCommand("subnet6-update", wrapDHCPServices([]DHCPServiceName{DHCPDaemonNameDHCP6}), arguments)
}

func NewNetwork6SubnetAddCommand(id int64, name string) *Command {
	arguments := map[string]any{
		"id":   id,
		"name": name,
	}
	return NewCommand("network6-subnet-add", wrapDHCPServices([]DHCPServiceName{DHCPDaemonNameDHCP6}), arguments)
}

func NewNetwork6SubnetDelCommand(id int64, name string) *Command {
	arguments := map[string]any{
		"id":   id,
		"name": name,
	}
	return NewCommand("network6-subnet-del", wrapDHCPServices([]DHCPServiceName{DHCPDaemonNameDHCP6}), arguments)
}
