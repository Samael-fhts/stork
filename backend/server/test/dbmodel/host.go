package storktestdbmodel

import (
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/stretchr/testify/require"
	keaconfig "isc.org/stork/daemoncfg/kea"
	dhcpmodel "isc.org/stork/datamodel/dhcp"
	dbmodel "isc.org/stork/server/database/model"
	storkutil "isc.org/stork/util"
)

// This function creates multiple hosts used in tests which fetch and
// filter hosts.
func AddTestHosts(t *testing.T, db *pg.DB) (hosts []dbmodel.Host, allDaemons []*dbmodel.Daemon) {
	// Add two apps.
	for i := 0; i < 2; i++ {
		m := &dbmodel.Machine{
			ID:        0,
			Address:   "cool.example.org",
			AgentPort: int64(8080 + i),
		}
		err := dbmodel.AddMachine(db, m)
		require.NoError(t, err)

		accessPoints := []*dbmodel.AccessPoint{}
		accessPoints = dbmodel.AppendAccessPoint(accessPoints, dbmodel.AccessPointControl, "localhost", "", int64(1234+i), "https")

		daemons := []*dbmodel.Daemon{
			dbmodel.NewKeaDaemon(m.ID, dbmodel.DaemonNameDHCPv4, true),
			dbmodel.NewKeaDaemon(m.ID, dbmodel.DaemonNameDHCPv6, true),
		}

		err = daemons[0].SetConfigFromJSON(`{
            "Dhcp4": {
				"client-classes": [
					{
						"name": "class2"
					},
					{
						"name": "class1"
					}
				],
                "subnet4": [
                    {
                        "id": 111,
                        "subnet": "192.0.2.0/24"
                    }
                ],
                "hooks-libraries": [
                    {
                        "library": "libdhcp_host_cmds.so"
                    }
                ]
            }
        }`)
		require.NoError(t, err)

		err = daemons[1].SetConfigFromJSON(`{
            "Dhcp6": {
				"client-classes": [
					{
						"name": "class2"
					},
					{
						"name": "class3"
					}
				],
                "subnet6": [
                    {
                        "id": 222,
                        "subnet": "2001:db8:1::/64"
                    }
                ],
                "hooks-libraries": [
                    {
                        "library": "libdhcp_host_cmds.so"
                    }
                ]
            }
        }`)
		require.NoError(t, err)
		allDaemons = append(allDaemons, daemons...)
	}

	subnets := []dbmodel.Subnet{
		{
			ID:     1,
			Prefix: "192.0.2.0/24",
		},
		{
			ID:     2,
			Prefix: "2001:db8:1::/64",
		},
	}
	for i, s := range subnets {
		subnet := s
		err := dbmodel.AddSubnet(db, &subnet)
		require.NoError(t, err)
		require.NotZero(t, subnet.ID)
		subnets[i] = subnet
	}

	// Add daemons to the database.
	for i, d := range allDaemons {
		err := dbmodel.AddDaemon(db, d)
		require.NoError(t, err)
		require.NotZero(t, d.ID)
		// Associate the daemons with the subnets.
		err = dbmodel.AddDaemonToSubnet(db, &subnets[i%2], d)
		require.NoError(t, err)
	}

	hasher := keaconfig.NewHasher()
	hosts = []dbmodel.Host{
		// Host 0.
		{
			SubnetID: 1,
			HostIdentifiers: []dbmodel.HostIdentifier{
				{
					Type:  "hw-address",
					Value: []byte{1, 2, 3, 4, 5, 6},
				},
				{
					Type:  "circuit-id",
					Value: []byte{1, 2, 3, 4},
				},
			},
			LocalHosts: []dbmodel.LocalHost{
				{
					DaemonID:       allDaemons[0].ID,
					Hostname:       "first.example.org",
					DataSource:     dbmodel.HostDataSourceAPI,
					NextServer:     "192.2.2.2",
					ServerHostname: "stork.example.org",
					BootFileName:   "/tmp/boot.xyz",
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "192.0.2.4",
						},
						{
							Address: "192.0.2.5",
						},
					},
				},
				{
					DaemonID:       allDaemons[2].ID,
					Hostname:       "first.example.org",
					DataSource:     dbmodel.HostDataSourceAPI,
					NextServer:     "192.2.2.2",
					ServerHostname: "stork.example.org",
					BootFileName:   "/tmp/boot.xyz",
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "192.0.2.4",
						},
						{
							Address: "192.0.2.5",
						},
					},
				},
			},
		},
		// Host 1.
		{
			HostIdentifiers: []dbmodel.HostIdentifier{
				{
					Type:  "hw-address",
					Value: []byte{2, 3, 4, 5, 6, 7},
				},
				{
					Type:  "circuit-id",
					Value: []byte{2, 3, 4, 5},
				},
			},
			LocalHosts: []dbmodel.LocalHost{
				{
					DaemonID:   allDaemons[0].ID,
					DataSource: dbmodel.HostDataSourceConfig,
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "192.0.2.6",
						},
						{
							Address: "192.0.2.7",
						},
					},
				},
				{
					DaemonID:   allDaemons[2].ID,
					DataSource: dbmodel.HostDataSourceAPI,
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "192.0.2.6",
						},
						{
							Address: "192.0.2.7",
						},
					},
				},
			},
		},
		// Host 2.
		{
			SubnetID: 2,
			HostIdentifiers: []dbmodel.HostIdentifier{
				{
					Type:  "hw-address",
					Value: []byte{1, 2, 3, 4, 5, 6},
				},
			},
			LocalHosts: []dbmodel.LocalHost{
				{
					DaemonID:   allDaemons[1].ID,
					DataSource: dbmodel.HostDataSourceConfig,
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "2001:db8:1::1",
						},
					},
				},
				{
					DaemonID:   allDaemons[3].ID,
					DataSource: dbmodel.HostDataSourceAPI,
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "2001:db8:1::1",
						},
					},
				},
			},
		},
		// Host 3.
		{
			HostIdentifiers: []dbmodel.HostIdentifier{
				{
					Type:  "duid",
					Value: []byte{1, 2, 3, 4},
				},
			},
			LocalHosts: []dbmodel.LocalHost{
				{
					DaemonID:   allDaemons[1].ID,
					DataSource: dbmodel.HostDataSourceConfig,
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "2001:db8:1::2",
						},
						{
							Address: "3001::/48",
						},
					},
				},
				{
					DaemonID:   allDaemons[3].ID,
					DataSource: dbmodel.HostDataSourceAPI,
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "2001:db8:1::2",
						},
						{
							Address: "3001::/48",
						},
					},
				},
			},
		},
		// Host 4.
		{
			HostIdentifiers: []dbmodel.HostIdentifier{
				{
					Type:  "duid",
					Value: []byte{2, 2, 2, 2},
				},
			},
			LocalHosts: []dbmodel.LocalHost{
				{
					DaemonID:   allDaemons[1].ID,
					DataSource: dbmodel.HostDataSourceConfig,
					ClientClasses: []string{
						"foo",
						"bar",
					},
					DHCPOptionSet: dbmodel.NewDHCPOptionSet([]dbmodel.DHCPOption{
						{
							Code: 23,
							Fields: []dbmodel.DHCPOptionField{
								{
									FieldType: dhcpmodel.IPv6AddressField,
									Values:    []any{"3001:dbef:1e5::"},
								},
								{
									FieldType: dhcpmodel.IPv6AddressField,
									Values:    []any{"3002:abc::"},
								},
							},
							Name:     "dns-servers",
							Space:    dhcpmodel.DHCPv6OptionSpace,
							Universe: storkutil.IPv6,
						},
					}, hasher),
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "3000::/48",
						},
					},
				},
				{
					DaemonID:   allDaemons[3].ID,
					DataSource: dbmodel.HostDataSourceAPI,
					ClientClasses: []string{
						"foo",
						"bar",
					},
					DHCPOptionSet: dbmodel.NewDHCPOptionSet([]dbmodel.DHCPOption{
						{
							Code: 23,
							Fields: []dbmodel.DHCPOptionField{
								{
									FieldType: dhcpmodel.IPv6AddressField,
									Values:    []any{"3001:dbef:1e5::"},
								},
								{
									FieldType: dhcpmodel.IPv6AddressField,
									Values:    []any{"3002:abc::"},
								},
							},
							Name:     "dns-servers",
							Space:    dhcpmodel.DHCPv6OptionSpace,
							Universe: storkutil.IPv6,
						},
					}, hasher),
					IPReservations: []dbmodel.IPReservation{
						{
							Address: "3000::/48",
						},
					},
				},
			},
		},
	}

	// Add hosts to the database.
	for i, h := range hosts {
		host := h
		err := dbmodel.AddHost(db, &host)
		require.NoError(t, err)
		require.NotZero(t, host.ID)
		hosts[i] = host
	}
	return hosts, allDaemons
}
