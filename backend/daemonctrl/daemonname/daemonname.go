package daemonname

import "github.com/pkg/errors"

// Defines the consistent names of the daemons we support. They are intended
// to be used throughout the codebase. This file should be put in a package
// without any imports.
type Name string

const (
	Bind9   Name = "named"
	DHCPv4  Name = "dhcp4"
	DHCPv6  Name = "dhcp6"
	NetConf Name = "netconf"
	D2      Name = "d2"
	CA      Name = "ca"
	PDNS    Name = "pdns"
)

// Indicates if the daemon name is a Kea daemon name.
func (dn Name) IsKea() bool {
	switch dn {
	case DHCPv4, DHCPv6, D2, CA:
		return true
	default:
		return false
	}
}

// Indicates if the daemon name is a DHCP daemon name.
func (dn Name) IsDHCP() bool {
	switch dn {
	case DHCPv4, DHCPv6:
		return true
	default:
		return false
	}
}

// Indicates if the daemon name is a DNS daemon name.
func (dn Name) IsDNS() bool {
	switch dn {
	case Bind9, PDNS:
		return true
	default:
		return false
	}
}

// Parses the daemon name from string. It returns an error if the
// daemon name is not recognized.
func Parse(name string) (Name, error) {
	switch name {
	case string(Bind9):
		return Bind9, nil
	case string(DHCPv4):
		return DHCPv4, nil
	case string(DHCPv6):
		return DHCPv6, nil
	case string(NetConf):
		return NetConf, nil
	case string(D2):
		return D2, nil
	case string(CA):
		return CA, nil
	case string(PDNS):
		return PDNS, nil
	default:
		return Name(""), errors.Errorf("unrecognized daemon name: %s", name)
	}
}
