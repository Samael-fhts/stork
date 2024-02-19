package keactrl

func NewLease4GetCommandByAddress(ip string) *Command {
	arguments := map[string]interface{}{
		"ip-address": ip,
	}
	return NewCommand("lease4-get", wrapDHCPService(DHCPDaemonNameDHCP4), arguments)
}

func NewLease4GetCommandByIdentifier(subnetID int64, identifierType, identifier string) *Command {
	arguments := map[string]interface{}{
		"subnet-id":       subnetID,
		"identifier-type": identifierType,
		"identifier":      identifier,
	}
	return NewCommand("lease4-get", wrapDHCPService(DHCPDaemonNameDHCP4), arguments)
}
