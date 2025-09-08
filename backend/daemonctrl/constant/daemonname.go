package constant

import "github.com/pkg/errors"

// Defines the consistent names of the daemons we support. They are intended
// to be used throughout the codebase. This file should be put in a package
// without any imports.
type DaemonName string

const (
	DaemonNameBind9   DaemonName = "named"
	DaemonNameDHCPv4  DaemonName = "dhcp4"
	DaemonNameDHCPv6  DaemonName = "dhcp6"
	DaemonNameNetConf DaemonName = "netconf"
	DaemonNameD2      DaemonName = "d2"
	DaemonNameCA      DaemonName = "ca"
	DaemonNamePDNS    DaemonName = "pdns"
)

// Converts a DaemonName to the corresponding KeaDaemonName.
func (dn DaemonName) ToKeaDaemonName() (KeaDaemonName, error) {
	switch dn {
	case DaemonNameDHCPv4:
		return KeaDaemonNameDHCPv4, nil
	case DaemonNameDHCPv6:
		return KeaDaemonNameDHCPv6, nil
	case DaemonNameD2:
		return KeaDaemonNameD2, nil
	case DaemonNameCA:
		return KeaDaemonNameCA, nil
	default:
		return KeaDaemonName(dn), errors.Errorf("cannot convert daemon name %s to Kea daemon name", dn)
	}
}

// Converts a DaemonName to the corresponding KeaDHCPDaemonName.
func (dn DaemonName) ToKeaDHCPDaemonName() (KeaDHCPDaemonName, error) {
	switch dn {
	case DaemonNameDHCPv4:
		return KeaDHCPDaemonNameDHCPv4, nil
	case DaemonNameDHCPv6:
		return KeaDHCPDaemonNameDHCPv6, nil
	default:
		return KeaDHCPDaemonName(dn), errors.Errorf("cannot convert daemon name %s to Kea DHCP daemon name", dn)
	}
}

// Parses the daemon name from string. It returns an error if the
// daemon name is not recognized.
func ParseDaemonName(name string) (DaemonName, error) {
	switch name {
	case string(DaemonNameBind9):
		return DaemonNameBind9, nil
	case string(DaemonNameDHCPv4):
		return DaemonNameDHCPv4, nil
	case string(DaemonNameDHCPv6):
		return DaemonNameDHCPv6, nil
	case string(DaemonNameNetConf):
		return DaemonNameNetConf, nil
	case string(DaemonNameD2):
		return DaemonNameD2, nil
	case string(DaemonNameCA):
		return DaemonNameCA, nil
	case string(DaemonNamePDNS):
		return DaemonNamePDNS, nil
	default:
		return DaemonName(""), errors.Errorf("unrecognized daemon name: %s", name)
	}
}

// The names of Kea daemons. They are intended to be used throughout
// the codebase.
type KeaDaemonName string

const (
	KeaDaemonNameDHCPv4  KeaDaemonName = KeaDaemonName(DaemonNameDHCPv4)
	KeaDaemonNameDHCPv6  KeaDaemonName = KeaDaemonName(DaemonNameDHCPv6)
	KeaDaemonNameD2      KeaDaemonName = KeaDaemonName(DaemonNameD2)
	KeaDaemonNameCA      KeaDaemonName = KeaDaemonName(DaemonNameCA)
	KeaDaemonNameNetConf KeaDaemonName = KeaDaemonName(DaemonNameNetConf)
)

// Converts a KeaDaemonName to the corresponding DaemonName.
func (kdn KeaDaemonName) ToDaemonName() DaemonName {
	return DaemonName(kdn)
}

// The names of the Kea DHCP daemons. They are intended to be used throughout
// the codebase.
type KeaDHCPDaemonName string

const (
	KeaDHCPDaemonNameDHCPv4 KeaDHCPDaemonName = KeaDHCPDaemonName(KeaDaemonNameDHCPv4)
	KeaDHCPDaemonNameDHCPv6 KeaDHCPDaemonName = KeaDHCPDaemonName(KeaDaemonNameDHCPv6)
)
