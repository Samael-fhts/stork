package keaconfig

// A structure representing the configuration of multiple control sockets
// in the Kea Control Agent. They are used to manage Kea daemons remotely.
type ManagementControlSockets struct {
	D2      *ControlSocket `json:"d2,omitempty"`
	Dhcp4   *ControlSocket `json:"dhcp4,omitempty"`
	Dhcp6   *ControlSocket `json:"dhcp6,omitempty"`
	NetConf *ControlSocket `json:"netconf,omitempty"`
}

// Returns a list of daemons for which management sockets have been configured.
func (cs *ManagementControlSockets) GetManagedDaemonNames() (names []string) {
	if cs == nil {
		return
	}

	if cs.D2 != nil {
		names = append(names, "d2")
	}
	if cs.Dhcp4 != nil {
		names = append(names, "dhcp4")
	}
	if cs.Dhcp6 != nil {
		names = append(names, "dhcp6")
	}
	if cs.NetConf != nil {
		names = append(names, "netconf")
	}

	return
}

// Returns true if any management control socket is configured.
func (cs *ManagementControlSockets) HasAnyManagedDaemon() bool {
	return cs != nil && (cs.D2 != nil || cs.Dhcp4 != nil || cs.Dhcp6 != nil || cs.NetConf != nil)
}

// Represents the HTTP headers in the Kea configuration.
type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// A structure representing the client credentials in the Kea Control Agent.
type ClientCredentials struct {
	User         *string `json:"user"`
	Password     *string `json:"password"`
	UserFile     *string `json:"user-file"`
	PasswordFile *string `json:"password-file"`
}

// A structure representing a configuration of the authentication credentials
// in the Kea Control Agent.
type Authentication struct {
	Type      string              `json:"type"`
	Realm     string              `json:"realm"`
	Directory *string             `json:"directory"`
	Clients   []ClientCredentials `json:"clients"`
}

// Indicates whether the basic auth method is used.
func (a Authentication) IsBasicAuth() bool {
	return a.Type == "basic"
}

// A structure representing a configuration of a single control socket in
// the Kea.
type ControlSocket struct {
	// Only for unix sockets.
	SocketName *string `json:"socket-name"`
	// Available values are "unix", "http" and "https".
	// The "http" and "https" types are supported since Kea 2.7.2.
	SocketType     string          `json:"socket-type"`
	SocketAddress  *string         `json:"socket-address,omitempty"`
	SocketPort     *int64          `json:"socket-port,omitempty"`
	HTTPHeaders    []HTTPHeader    `json:"http-headers,omitempty"`
	TrustAnchor    *string         `json:"trust-anchor,omitempty"`
	CertFile       *string         `json:"cert-file,omitempty"`
	KeyFile        *string         `json:"key-file,omitempty"`
	CertRequired   *bool           `json:"cert-required,omitempty"`
	Authentication *Authentication `json:"authentication,omitempty"`
}

// Indicates the name of the protocol used by the control socket:
// "unix", "http" or "https".
func (cs ControlSocket) GetProtocol() string {
	return cs.SocketType
}

// Returns a port number or the default port if the port is not set.
func (cs ControlSocket) GetPort() int64 {
	if cs.SocketType == "unix" {
		return 0
	}

	if cs.SocketPort != nil {
		return *cs.SocketPort
	}
	return 8000
}

// Return a socket address or socket path. It normalizes some special values.
func (cs ControlSocket) GetAddress() string {
	if cs.SocketType == "unix" {
		if cs.SocketName == nil {
			return ""
		}
		return *cs.SocketName
	}

	if cs.SocketAddress == nil {
		return "127.0.0.1"
	}
	address := *cs.SocketAddress
	switch address {
	case "0.0.0.0", "":
		address = "127.0.0.1"
	case "::":
		address = "::1"
	}
	return address
}
