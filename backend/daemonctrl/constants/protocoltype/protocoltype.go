package protocoltype

import "github.com/pkg/errors"

// Supported protocol types in communication between the Stork agent and
// daemons.
type ProtocolType string

const (
	Unspecified ProtocolType = ""
	HTTP        ProtocolType = "http"
	HTTPS       ProtocolType = "https"
	Socket      ProtocolType = "unix"
	RNDC        ProtocolType = "rndc"
)

// Indicates whether the protocol type is secure.
func (pt ProtocolType) IsSecure() bool {
	return pt == HTTPS || pt == RNDC
}

// Parses the protocol type from string. It returns an error if the
// protocol type is not recognized.
func Parse(protocolType string) (ProtocolType, error) {
	switch protocolType {
	case string(HTTP):
		return HTTP, nil
	case string(HTTPS):
		return HTTPS, nil
	case string(Socket):
		return Socket, nil
	case string(RNDC):
		return RNDC, nil
	default:
		return Unspecified, errors.Errorf("unknown protocol type: %s", protocolType)
	}
}
