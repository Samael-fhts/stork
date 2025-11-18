package daemonname_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"isc.org/stork/datamodel/daemonname"
)

// Test that the Kea daemon names are indicated properly.
func TestDaemonNameIsKea(t *testing.T) {
	require.True(t, daemonname.CA.IsKea())
	require.True(t, daemonname.D2.IsKea())
	require.True(t, daemonname.DHCPv4.IsKea())
	require.True(t, daemonname.DHCPv4.IsKea())

	require.False(t, daemonname.NetConf.IsKea())
	require.False(t, daemonname.Bind9.IsKea())
	require.False(t, daemonname.PDNS.IsKea())
}

// Test that the DHCP daemon names are indicated properly.
func TestDaemonNameIsDNS(t *testing.T) {
	require.True(t, daemonname.Bind9.IsDNS())
	require.True(t, daemonname.PDNS.IsDNS())
	require.False(t, daemonname.CA.IsDNS())
	require.False(t, daemonname.D2.IsDNS())
	require.False(t, daemonname.DHCPv4.IsDNS())
	require.False(t, daemonname.NetConf.IsDNS())
}

// Test that the DHCP daemon names are indicated properly.
func TestDaemonNameIsDHCP(t *testing.T) {
	require.True(t, daemonname.DHCPv4.IsDHCP())
	require.True(t, daemonname.DHCPv6.IsDHCP())
	require.False(t, daemonname.CA.IsDHCP())
	require.False(t, daemonname.D2.IsDHCP())
	require.False(t, daemonname.Bind9.IsDHCP())
	require.False(t, daemonname.NetConf.IsDHCP())
	require.False(t, daemonname.PDNS.IsDHCP())
}

// Test that parsing daemon names from strings works properly.
func TestParseDaemonName(t *testing.T) {
	t.Run("BIND 9", func(t *testing.T) {
		dn, err := daemonname.Parse("named")
		require.NoError(t, err)
		require.Equal(t, daemonname.Bind9, dn)
	})

	t.Run("Kea DHCPv4", func(t *testing.T) {
		dn, err := daemonname.Parse("dhcp4")
		require.NoError(t, err)
		require.Equal(t, daemonname.DHCPv4, dn)
	})

	t.Run("Kea DHCPv6", func(t *testing.T) {
		dn, err := daemonname.Parse("dhcp6")
		require.NoError(t, err)
		require.Equal(t, daemonname.DHCPv6, dn)
	})

	t.Run("Kea D2", func(t *testing.T) {
		dn, err := daemonname.Parse("d2")
		require.NoError(t, err)
		require.Equal(t, daemonname.D2, dn)
	})

	t.Run("Kea CA", func(t *testing.T) {
		dn, err := daemonname.Parse("ca")
		require.NoError(t, err)
		require.Equal(t, daemonname.CA, dn)
	})

	t.Run("PowerDNS", func(t *testing.T) {
		dn, err := daemonname.Parse("pdns")
		require.NoError(t, err)
		require.Equal(t, daemonname.PDNS, dn)
	})

	t.Run("NetConf", func(t *testing.T) {
		dn, err := daemonname.Parse("netconf")
		require.NoError(t, err)
		require.Equal(t, daemonname.NetConf, dn)
	})

	t.Run("Unknown Daemon", func(t *testing.T) {
		dn, err := daemonname.Parse("unknown-daemon")
		require.Error(t, err)
		require.Empty(t, dn)
	})
}
