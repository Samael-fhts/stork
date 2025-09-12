package agent

import (
	"encoding/json"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gopkg.in/h2non/gock.v1"
	keaconfig "isc.org/stork/daemoncfg/kea"
	"isc.org/stork/daemonctrl/constant"
	keactrl "isc.org/stork/daemonctrl/kea"
	"isc.org/stork/testutil"
	storkutil "isc.org/stork/util"
)

// Test the case that the command is successfully sent to Kea.
func TestSendCommand(t *testing.T) {
	// Expect appropriate content type and the body. If they are not matched
	// an error will be raised.
	defer gock.Off()
	gock.New("http://localhost:45634").
		MatchHeader("Content-Type", "application/json").
		JSON(map[string]string{"command": "list-commands"}).
		Post("/").
		Reply(200).
		JSON([]map[string]int{{"result": 0}})

	command := keactrl.NewCommandBase(keactrl.ListCommands, constant.KeaDaemonNameCA)

	accessPoint := AccessPoint{Type: AccessPointControl, Address: "localhost", Port: 45634, Protocol: "http"}
	daemon := &KeaDaemon{
		daemon: daemon{
			Name:         constant.DaemonNameCA,
			AccessPoints: []AccessPoint{accessPoint},
		},
		connector: newKeaConnector(accessPoint, HTTPClientConfig{Interceptor: gock.InterceptClient}),
	}
	var response keactrl.Response
	err := daemon.sendCommand(command, &response)
	require.NoError(t, err)
}

// Test the case that the command is not successfully sent to Kea because
// there is no control access point.
func TestSendCommandNoAccessPoint(t *testing.T) {
	command := keactrl.NewCommandBase(keactrl.ListCommands, constant.KeaDaemonNameCA)

	daemon := &KeaDaemon{
		daemon: daemon{
			Name:         constant.DaemonNameDHCPv4,
			AccessPoints: []AccessPoint{},
		},
		connector: nil,
	}

	var response keactrl.Response
	err := daemon.sendCommand(command, &response)
	require.ErrorContains(t, err, "no control access point")
}

// Test the case when Kea returns invalid response to the command.
func TestSendCommandInvalidResponse(t *testing.T) {
	// Return invalid response. Arguments must be a map not an integer.
	defer gock.Off()
	gock.New("http://localhost:45634").
		MatchHeader("Content-Type", "application/json").
		JSON(map[string]string{"command": "version-get"}).
		Post("/").
		Reply(200).
		JSON([]map[string]interface{}{
			{"result": 0, "text": "1.0.0", "arguments": 1},
		})

	command := keactrl.NewCommandBase(keactrl.VersionGet, constant.KeaDaemonNameDHCPv4)

	accessPoint := AccessPoint{Type: AccessPointControl, Address: "localhost", Port: 45634, Protocol: "http"}
	daemon := &KeaDaemon{
		daemon: daemon{
			Name:         constant.DaemonNameDHCPv4,
			AccessPoints: []AccessPoint{accessPoint},
		},
		connector: newKeaConnector(accessPoint, HTTPClientConfig{Interceptor: gock.InterceptClient}),
	}

	type versionGet struct {
		keactrl.ResponseHeader
		Arguments struct {
			ExtendedVersion string
		}
	}
	var response versionGet
	err := daemon.sendCommand(command, &response)
	require.Error(t, err)
}

// Test the case when Kea server is unreachable.
func TestSendCommandNoKea(t *testing.T) {
	command := keactrl.NewCommandBase(keactrl.ListCommands, constant.KeaDaemonNameCA)
	accessPoint := AccessPoint{Type: AccessPointControl, Address: "localhost", Port: 45634, Protocol: "http"}
	daemon := &KeaDaemon{
		daemon: daemon{
			Name:         constant.DaemonNameCA,
			AccessPoints: []AccessPoint{accessPoint},
		},
		connector: newKeaConnector(accessPoint, HTTPClientConfig{}),
	}
	var response keactrl.Response
	err := daemon.sendCommand(command, &response)
	require.Error(t, err)
}

// Test the function which extracts the list of log files from the Kea
// daemon by sending the request to the Kea Control Agent and the
// daemons behind it.
func TestKeaAllowedLogs(t *testing.T) {
	// The first config-get command should go to the Kea Control Agent.
	// The logs should be extracted from there and the subsequent config-get
	// commands should be sent to the daemons with which the CA is configured
	// to communicate.
	defer gock.Off()
	caResponseJSON := `[{
        "result": 0,
        "arguments": {
            "Control-agent": {
                "control-sockets": {
                    "dhcp4": {
                        "socket-name": "/tmp/dhcp4.sock"
                    },
                    "dhcp6": {
                        "socket-name": "/tmp/dhcp6.sock"
                    }
                },
                "loggers": [
                    {
                        "output_options": [
                            {
                                "output": "/tmp/kea-ctrl-agent.log"
                            }
                        ]
                    }
                ]
            }
        }
    }]`
	caResponse := make([]map[string]interface{}, 1)
	err := json.Unmarshal([]byte(caResponseJSON), &caResponse)
	require.NoError(t, err)
	gock.New("https://localhost:45634").
		MatchHeader("Content-Type", "application/json").
		JSON(map[string]string{"command": "config-get"}).
		Post("/").
		Reply(200).
		JSON(caResponse)

	dhcpResponsesJSON := `[
        {
            "result": 0,
            "arguments": {
                "Dhcp4": {
                    "loggers": [
                        {
                            "output_options": [
                                {
                                    "output": "/tmp/kea-dhcp4.log"
                                }
                            ]
                        }
                    ]
                }
            }
        },
        {
            "result": 0,
            "arguments": {
                "Dhcp6": {
                    "loggers": [
                        {
                            "output_options": [
                                {
                                    "output": "/tmp/kea-dhcp6.log"
                                }
                            ]
                        }
                    ]
                }
            }
        }
    ]`
	dhcpResponses := make([]map[string]interface{}, 2)
	err = json.Unmarshal([]byte(dhcpResponsesJSON), &dhcpResponses)
	require.NoError(t, err)

	// The config-get command sent to the daemons behind CA should return
	// configurations of the DHCPv4 and DHCPv6 daemons.
	gock.New("https://localhost:45634").
		MatchHeader("Content-Type", "application/json").
		JSON(map[string]interface{}{"command": "config-get", "service": []string{"dhcp4", "dhcp6"}}).
		Post("/").
		Reply(200).
		JSON(dhcpResponses)

	accessPoint := AccessPoint{Type: AccessPointControl, Address: "localhost", Port: 45634, Protocol: "https"}
	daemon := &KeaDaemon{
		daemon: daemon{
			Name:         constant.DaemonNameCA,
			AccessPoints: []AccessPoint{accessPoint},
		},
		connector: newKeaConnector(accessPoint, HTTPClientConfig{Interceptor: gock.InterceptClient}),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	agentManager := NewMockAgentManager(ctrl)
	// We should have three log files recorded from the returned configurations.
	// One from CA, one from DHCPv4 and one from DHCPv6.
	agentManager.EXPECT().AllowLog(gomock.Any()).Times(3)

	err = daemon.Evaluate(agentManager)
	require.NoError(t, err)
}

// Test the function which extracts the list of log files from the Kea
// daemon by sending the request to the Kea Control Agent and the
// daemons behind it. This test variant uses output-options alias for
// logger configuration.
func TestKeaAllowedLogsOutputOptionsWithDash(t *testing.T) {
	// The first config-get command should go to the Kea Control Agent.
	// The logs should be extracted from there and the subsequent config-get
	// commands should be sent to the daemons with which the CA is configured
	// to communicate.
	defer gock.Off()
	caResponseJSON := `[{
        "result": 0,
        "arguments": {
            "Control-agent": {
                "control-sockets": {
                    "dhcp4": {
                        "socket-name": "/tmp/dhcp4.sock"
                    },
                    "dhcp6": {
                        "socket-name": "/tmp/dhcp6.sock"
                    }
                },
                "loggers": [
                    {
                        "output-options": [
                            {
                                "output": "/tmp/kea-ctrl-agent.log"
                            }
                        ]
                    }
                ]
            }
        }
    }]`
	caResponse := make([]map[string]interface{}, 1)
	err := json.Unmarshal([]byte(caResponseJSON), &caResponse)
	require.NoError(t, err)
	gock.New("https://localhost:45634").
		MatchHeader("Content-Type", "application/json").
		JSON(map[string]string{"command": "config-get"}).
		Post("/").
		Reply(200).
		JSON(caResponse)

	dhcpResponsesJSON := `[
        {
            "result": 0,
            "arguments": {
                "Dhcp4": {
                    "loggers": [
                        {
                            "output-options": [
                                {
                                    "output": "/tmp/kea-dhcp4.log"
                                }
                            ]
                        }
                    ]
                }
            }
        },
        {
            "result": 0,
            "arguments": {
                "Dhcp6": {
                    "loggers": [
                        {
                            "output-options": [
                                {
                                    "output": "/tmp/kea-dhcp6.log"
                                }
                            ]
                        }
                    ]
                }
            }
        }
    ]`
	dhcpResponses := make([]map[string]interface{}, 2)
	err = json.Unmarshal([]byte(dhcpResponsesJSON), &dhcpResponses)
	require.NoError(t, err)

	// The config-get command sent to the daemons behind CA should return
	// configurations of the DHCPv4 and DHCPv6 daemons.
	gock.New("https://localhost:45634").
		MatchHeader("Content-Type", "application/json").
		JSON(map[string]interface{}{"command": "config-get", "service": []string{"dhcp4", "dhcp6"}}).
		Post("/").
		Reply(200).
		JSON(dhcpResponses)

	accessPoint := AccessPoint{Type: AccessPointControl, Address: "localhost", Port: 45634, Protocol: "https"}
	daemon := &KeaDaemon{
		daemon: daemon{
			Name:         constant.DaemonNameCA,
			AccessPoints: []AccessPoint{accessPoint},
		},
		connector: newKeaConnector(accessPoint, HTTPClientConfig{Interceptor: gock.InterceptClient}),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	agentManager := NewMockAgentManager(ctrl)
	// We should have three log files recorded from the returned configurations.
	// One from CA, one from DHCPv4 and one from DHCPv6.
	agentManager.EXPECT().AllowLog(gomock.Any()).Times(3)

	err = daemon.Evaluate(agentManager)
	require.NoError(t, err)
}

// This test verifies that an error is returned when the number of responses
// from the Kea daemons is lower than the number of services specified in the
// command.
func TestKeaAllowedLogsFewerResponses(t *testing.T) {
	defer gock.Off()

	// Return only one response while the number of daemons is two.
	dhcpResponsesJSON := `[
        {
            "result": 0,
            "arguments": {
                "Dhcp4": {
                }
            }
        }
    ]`
	dhcpResponses := make([]map[string]interface{}, 1)
	err := json.Unmarshal([]byte(dhcpResponsesJSON), &dhcpResponses)
	require.NoError(t, err)

	gock.New("https://localhost:45634").
		MatchHeader("Content-Type", "application/json").
		JSON(map[string]interface{}{"command": "config-get", "service": []string{"dhcp4", "dhcp6"}}).
		Post("/").
		Reply(200).
		JSON(dhcpResponses)

	accessPoint := AccessPoint{Type: AccessPointControl, Address: "localhost", Port: 45634, Protocol: "https"}
	daemon := &KeaDaemon{
		daemon: daemon{
			Name:         constant.DaemonNameCA,
			AccessPoints: []AccessPoint{accessPoint},
		},
		connector: newKeaConnector(accessPoint, HTTPClientConfig{Interceptor: gock.InterceptClient}),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	agentManager := NewMockAgentManager(ctrl)

	err = daemon.Evaluate(agentManager)
	require.Error(t, err)
}

// Test that cleaning up the daemon doesn't panic.
func TestKeaDaemonCleanup(t *testing.T) {
	daemon := &KeaDaemon{}
	require.NotPanics(t, func() {
		daemon.Cleanup()
	})
}

// Test that the client credentials are retrieved properly.
func TestReadClientCredentials(t *testing.T) {
	t.Run("Nil authentication", func(t *testing.T) {
		// Arrange
		var authentication *keaconfig.Authentication

		// Act & Assert
		require.Panics(t, func() {
			_, _ = readClientCredentials(authentication)
		})
	})

	t.Run("No clients", func(t *testing.T) {
		// Arrange
		authentication := &keaconfig.Authentication{
			Clients: nil,
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.NoError(t, err)
		require.Empty(t, clients)
	})

	t.Run("User and password", func(t *testing.T) {
		// Arrange
		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					User:     storkutil.Ptr("user"),
					Password: storkutil.Ptr("password"),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.NoError(t, err)
		require.Len(t, clients, 1)
		require.Equal(t, "user", clients[0].User)
		require.Equal(t, "password", clients[0].Password)
	})

	t.Run("No client properties are set", func(t *testing.T) {
		// Arrange
		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.Error(t, err)
		require.Empty(t, clients)
	})

	t.Run("User file is set only", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()
		dir, _ := sb.JoinDir("empty")
		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     storkutil.Ptr(path.Join(dir, "user")),
					PasswordFile: nil,
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.ErrorContains(t, err, "could not read the user file")
		require.Empty(t, clients)
	})

	t.Run("Non-existing user file and password file", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()
		passwordFile, _ := sb.Join("password")

		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     storkutil.Ptr(path.Join(sb.BasePath, "user")),
					PasswordFile: storkutil.Ptr(passwordFile),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.ErrorContains(t, err, "could not read the user file")
		require.Empty(t, clients)
	})

	t.Run("User file and password file - non-existing user file", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()
		dir, _ := sb.JoinDir("test")
		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     storkutil.Ptr(path.Join(dir, "user")),
					PasswordFile: storkutil.Ptr(path.Join(dir, "password")),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.ErrorContains(t, err, "could not read the user file")
		require.Empty(t, clients)
	})

	t.Run("User file and password file - non-existing password file", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		userFile, _ := sb.Join("user")

		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     storkutil.Ptr(userFile),
					PasswordFile: storkutil.Ptr(path.Join(sb.BasePath, "password")),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.ErrorContains(t, err, "could not read the password file")
		require.Empty(t, clients)
	})

	t.Run("User file and password file with default directory", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		userFile, _ := sb.Write("user", "foo")
		passwordFile, _ := sb.Write("password", "bar")

		authentication := &keaconfig.Authentication{
			Directory: nil,
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     storkutil.Ptr(userFile),
					PasswordFile: storkutil.Ptr(passwordFile),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.NoError(t, err)
		require.Len(t, clients, 1)
		require.Equal(t, "foo", clients[0].User)
		require.Equal(t, "bar", clients[0].Password)
	})

	t.Run("User file and password file with custom directory", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		_, _ = sb.Write("user", "foo")
		_, _ = sb.Write("password", "bar")

		authentication := &keaconfig.Authentication{
			Directory: storkutil.Ptr(sb.BasePath),
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     storkutil.Ptr("user"),
					PasswordFile: storkutil.Ptr("password"),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.NoError(t, err)
		require.Len(t, clients, 1)
		require.Equal(t, "foo", clients[0].User)
		require.Equal(t, "bar", clients[0].Password)
	})

	t.Run("Password file only - non-existing file", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     nil,
					PasswordFile: storkutil.Ptr(path.Join(sb.BasePath, "password")),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.ErrorContains(t, err, "could not read the password file")
		require.Empty(t, clients)
	})

	t.Run("Password file only - default directory", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		passwordFile, _ := sb.Write("password", "foo:bar")

		authentication := &keaconfig.Authentication{
			Directory: nil,
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     nil,
					PasswordFile: storkutil.Ptr(passwordFile),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.NoError(t, err)
		require.Len(t, clients, 1)
		require.Equal(t, "foo", clients[0].User)
		require.Equal(t, "bar", clients[0].Password)
	})

	t.Run("Password file only - invalid content", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		passwordFile, _ := sb.Write("password", "foo-bar")

		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     nil,
					PasswordFile: storkutil.Ptr(passwordFile),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.ErrorContains(t, err, "invalid format of the password file")
		require.Empty(t, clients)
	})

	t.Run("Password file only - custom directory", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		_, _ = sb.Write("password", "foo:bar")

		authentication := &keaconfig.Authentication{
			Directory: storkutil.Ptr(sb.BasePath),
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile:     nil,
					PasswordFile: storkutil.Ptr("password"),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.NoError(t, err)
		require.Len(t, clients, 1)
		require.Equal(t, "foo", clients[0].User)
		require.Equal(t, "bar", clients[0].Password)
	})

	t.Run("User string and password file", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		passwordFile, _ := sb.Write("password", "bar")

		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					User:         storkutil.Ptr("foo"),
					PasswordFile: storkutil.Ptr(passwordFile),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.NoError(t, err)
		require.Len(t, clients, 1)
		require.Equal(t, "foo", clients[0].User)
		require.Equal(t, "bar", clients[0].Password)
	})

	t.Run("User file and password string", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		userFile, _ := sb.Write("user", "foo")

		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					UserFile: storkutil.Ptr(userFile),
					Password: storkutil.Ptr("bar"),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.NoError(t, err)
		require.Len(t, clients, 1)
		require.Equal(t, "foo", clients[0].User)
		require.Equal(t, "bar", clients[0].Password)
	})

	t.Run("All methods at once", func(t *testing.T) {
		// Arrange
		sb := testutil.NewSandbox()
		defer sb.Close()

		userFile, _ := sb.Write("user", "foo")
		passwordFile, _ := sb.Write("password", "bar")
		singlePasswordFile, _ := sb.Write("password-single", "baz:boz")

		authentication := &keaconfig.Authentication{
			Clients: []keaconfig.ClientCredentials{
				{
					// User and password as strings.
					User:     storkutil.Ptr("bim"),
					Password: storkutil.Ptr("bom"),
				},
				{
					// User and password as files.
					UserFile:     storkutil.Ptr(userFile),
					PasswordFile: storkutil.Ptr(passwordFile),
				},
				{
					// User and password in a single file.
					UserFile:     nil,
					PasswordFile: storkutil.Ptr(singlePasswordFile),
				},
				{
					// User as a string and password as a file.
					User:         storkutil.Ptr("ding"),
					PasswordFile: storkutil.Ptr(passwordFile),
				},
				{
					// User as a file and password as a string.
					UserFile: storkutil.Ptr(userFile),
					Password: storkutil.Ptr("dong"),
				},
			},
		}

		// Act
		clients, err := readClientCredentials(authentication)

		// Assert
		require.NoError(t, err)
		require.Len(t, clients, 5)

		require.Equal(t, "bim", clients[0].User)
		require.Equal(t, "bom", clients[0].Password)

		require.Equal(t, "foo", clients[1].User)
		require.Equal(t, "bar", clients[1].Password)

		require.Equal(t, "baz", clients[2].User)
		require.Equal(t, "boz", clients[2].Password)

		require.Equal(t, "ding", clients[3].User)
		require.Equal(t, "bar", clients[3].Password)

		require.Equal(t, "foo", clients[4].User)
		require.Equal(t, "dong", clients[4].Password)
	})
}
