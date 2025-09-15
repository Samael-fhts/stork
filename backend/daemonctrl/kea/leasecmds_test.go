package keactrl

import (
	"testing"

	require "github.com/stretchr/testify/require"
	"isc.org/stork/daemonctrl/daemonname"
)

// Tests lease4-get command.
func TestNewCommandLease4Get(t *testing.T) {
	command := NewCommandLease4Get("192.0.2.1", daemonname.DHCPv4)
	require.NotNil(t, command)
	require.Len(t, command.Daemons, 1)
	bytes, err := command.Marshal()
	require.NoError(t, err)
	require.JSONEq(t, `{
		"command": "lease4-get",
		"service": ["dhcp4"],
		"arguments": {
			"ip-address": "192.0.2.1"
		}

	}`, string(bytes))
}

// Tests lease6-get command.
func TestNewCommandLease6Get(t *testing.T) {
	command := NewCommandLease6Get(LeaseTypeNA, "2001:db8:1::1", daemonname.DHCPv6)
	require.NotNil(t, command)
	require.Len(t, command.Daemons, 1)
	bytes, err := command.Marshal()
	require.NoError(t, err)
	require.JSONEq(t, `{
		"command": "lease6-get",
		"service": ["dhcp6"],
		"arguments": {
			"type": "IA_NA",
			"ip-address": "2001:db8:1::1"
		}

	}`, string(bytes))
}
