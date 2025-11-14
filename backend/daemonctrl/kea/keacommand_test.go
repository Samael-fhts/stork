package keactrl

import (
	"testing"

	require "github.com/stretchr/testify/require"
	"isc.org/stork/datamodel/constants/daemonname"
)

const valuesSetCommand CommandName = "values-set"

// Test successful creation of the Kea command with daemons and arguments.
func TestNewCommand(t *testing.T) {
	cmd := newCommand(valuesSetCommand, daemonname.DHCPv4, map[string]any{
		"value-a": 1,
		"value-b": 2,
		"value-c": []any{1, 2, 3},
	})

	require.NotNil(t, cmd)
	require.NotNil(t, cmd.Daemons)
	require.NotNil(t, cmd.Arguments)

	require.Equal(t, valuesSetCommand, cmd.Command)
	require.Len(t, cmd.Daemons, 1)
	require.Contains(t, cmd.Daemons, daemonname.DHCPv4)
	arguments, ok := cmd.Arguments.(map[string]any)
	require.True(t, ok)
	require.Contains(t, arguments, "value-a")
	require.Contains(t, arguments, "value-b")
	require.Contains(t, arguments, "value-c")
	require.NotContains(t, arguments, "value-d")
}

// Test successful creation of the Kea command with arguments specified as a structure.
func TestNewCommandWithStructArgs(t *testing.T) {
	type argsType struct {
		ValueA int
		ValueB int
		ValueC []int
	}
	inputArguments := argsType{
		ValueA: 2,
		ValueB: 3,
		ValueC: []int{5, 6, 7},
	}
	cmd := newCommand(valuesSetCommand, daemonname.DHCPv4, inputArguments)
	require.NotNil(t, cmd)
	require.NotNil(t, cmd.Daemons)
	require.NotNil(t, cmd.Arguments)
	require.Equal(t, valuesSetCommand, cmd.Command)
	require.Len(t, cmd.Daemons, 1)
	require.Contains(t, cmd.Daemons, daemonname.DHCPv4)
	outputArguments, ok := cmd.Arguments.(argsType)
	require.True(t, ok)
	require.Equal(t, inputArguments, outputArguments)
}

// Test successful creation of the Kea command with arguments specified as a pointer
// to a structure.
func TestNewCommandWithStructPtrArgs(t *testing.T) {
	type argsType struct {
		ValueA int
	}
	args := argsType{
		ValueA: 2,
	}
	cmd := newCommand(valuesSetCommand, daemonname.DHCPv4, args)
	require.NotNil(t, cmd)
	require.NotNil(t, cmd.Daemons)
	require.NotNil(t, cmd.Arguments)
	require.Equal(t, valuesSetCommand, cmd.Command)
	require.Len(t, cmd.Daemons, 1)
	require.Contains(t, cmd.Daemons, daemonname.DHCPv4)
	outputArgs, ok := cmd.Arguments.(argsType)
	require.True(t, ok)
	require.Equal(t, args, outputArgs)
}

// Test that the command is not created when the arguments have an invalid type.
func TestNewCommandWithInvalidArgTypes(t *testing.T) {
	require.Nil(t, newCommand(valuesSetCommand, daemonname.DHCPv4, 123))
	require.Nil(t, newCommand(valuesSetCommand, daemonname.DHCPv4, []int{123, 345}))
	m := make(map[string]interface{})
	require.Nil(t, newCommand(valuesSetCommand, daemonname.DHCPv4, &m))
}

// Test that command name must be non-empty.
func TestNewCommandEmptyName(t *testing.T) {
	cmd := NewCommandBase("", daemonname.DHCPv4)
	require.Nil(t, cmd)
}

// Test instantiating a command with no arguments.
func TestNewCommandWithNoArgs(t *testing.T) {
	command := NewCommandBase(ListCommands, daemonname.DHCPv4)
	require.NotNil(t, command)
	require.Equal(t, ListCommands, command.Command)
	require.Len(t, command.Daemons, 1)
	require.Equal(t, daemonname.DHCPv4, command.Daemons[0])
	require.Nil(t, command.Arguments)
}

// Test instantiating a command with no arguments and no daemons.
func TestNewCommandWithNoArgsNoDaemons(t *testing.T) {
	command := NewCommandBase(ListCommands, daemonname.DHCPv4)
	require.NotNil(t, command)
	require.Equal(t, ListCommands, command.Command)
	require.Len(t, command.Daemons, 1)
	marshaledBytes, err := command.Marshal()
	require.NoError(t, err)
	require.JSONEq(t, `{
		"command": "list-commands",
		"service": ["dhcp4"]
	}`, string(marshaledBytes))
}

// Test creating a new command with non-array arguments.
func TestNewCommandWithArgs(t *testing.T) {
	command := newCommand(CommandName("test"), daemonname.DHCPv4, map[string]any{
		"element":  5,
		"element2": "foo",
	})
	require.NotNil(t, command)
	marshaledBytes, err := command.Marshal()
	require.NoError(t, err)
	require.JSONEq(t, `{
		"command": "test",
		"service": ["dhcp4"],
		"arguments": {
			"element": 5,
			"element2": "foo"
		}
	}`, string(marshaledBytes))
}

// Tests creating a new command with array argument.
func TestNewCommandWithArrayArgs(t *testing.T) {
	command := newCommand(CommandName("test"), daemonname.DHCPv4, map[string]any{
		"element":  []any{5, 9},
		"element2": []any{"foo"},
	})
	require.NotNil(t, command)
	marshaledBytes, err := command.Marshal()
	require.NoError(t, err)
	require.JSONEq(t, `{
		"command": "test",
		"service": ["dhcp4"],
		"arguments": {
			"element": [ 5, 9 ],
			"element2": [ "foo" ]
		}
	}`, string(marshaledBytes))
}

// Test that JSON representation of the command is created correctly when
// both daemon name (service in Kea terms) and arguments are present.
func TestKeaCommandMarshal(t *testing.T) {
	cmd := newCommand(valuesSetCommand, daemonname.DHCPv4, map[string]any{
		"value-a": 1,
		"value-b": 2,
		"value-c": []any{1, 2, 3},
	})
	require.NotNil(t, cmd)

	marshaled, err := cmd.Marshal()
	require.NoError(t, err)
	require.JSONEq(t,
		`{
             "command":"values-set",
             "service":["dhcp4"],
             "arguments": {
                 "value-a":1,
                 "value-b":2,
                 "value-c": [1,2,3]
             }
         }`,
		string(marshaled))
}

// Test that the error is returned when the arguments cannot be marshaled to
// JSON.
func TestKeaCommandMarshalError(t *testing.T) {
	// Arrange
	payload := map[string]any{}
	payload["payload"] = payload // Circular reference to cause marshaling error.

	cmd := newCommand(valuesSetCommand, daemonname.DHCPv4, payload)

	// Act
	marshaled, err := cmd.Marshal()

	// Assert
	require.Error(t, err)
	require.Nil(t, marshaled)

	// The daemon list should be unchanged after marshaling.
	require.Len(t, cmd.Daemons, 1)
	require.Equal(t, daemonname.DHCPv4, cmd.Daemons[0])

}

// Test that JSON representation of the command is created correctly when
// arguments are specified in a structure.
func TestKeaCommandMarshalWithStructArgs(t *testing.T) {
	type argsType struct {
		ValueA int   `json:"value-a"`
		ValueB int   `json:"value-b"`
		ValueC []int `json:"value-c"`
	}
	args := argsType{
		ValueA: 222,
		ValueB: 333,
		ValueC: []int{123, 234},
	}
	cmd := newCommand(valuesSetCommand, daemonname.DHCPv4, &args)
	require.NotNil(t, cmd)

	marshaled, err := cmd.Marshal()
	require.NoError(t, err)
	require.JSONEq(t,
		`{
             "command":"values-set",
             "service":["dhcp4"],
             "arguments": {
                 "value-a":222,
                 "value-b":333,
                 "value-c": [123,234]
             }
         }`,
		string(marshaled))

	// The daemon list should be unchanged after marshaling.
	require.Len(t, cmd.Daemons, 1)
	require.Equal(t, daemonname.DHCPv4, cmd.Daemons[0])

}

// Test that no service list is included when daemons list is empty.
func TestKeaCommandMarshalEmptyDaemonsArguments(t *testing.T) {
	cmd := newCommand(valuesSetCommand, daemonname.DHCPv4, map[string]any{})
	require.NotNil(t, cmd)

	marshaled, err := cmd.Marshal()
	require.NoError(t, err)
	require.JSONEq(t,
		`{
             "command":"values-set",
             "service":["dhcp4"],
             "arguments": { }
         }`,
		string(marshaled))
}

// Test that it is possible to send a command without arguments and without
// daemons list.
func TestKeaCommandMarshalCommandOnly(t *testing.T) {
	cmd := NewCommandBase(ListCommands, daemonname.DHCPv4)
	require.NotNil(t, cmd)

	marshaled, err := cmd.Marshal()
	require.NoError(t, err)
	require.JSONEq(t,
		`{
             "command":"list-commands",
             "service":["dhcp4"]
         }`,
		string(marshaled))
}

// Test that the service (daemon list) is not included in the commands directed
// to the Kea CA daemon.
func TestKeaCommandMarshalServicesIsMissingForCA(t *testing.T) {
	// Arrange
	cmd := NewCommandBase(ListCommands, daemonname.CA)
	// The command has non-empty daemon list before marshaling.
	require.Len(t, cmd.Daemons, 1)
	require.Equal(t, daemonname.CA, cmd.Daemons[0])

	marshaled, err := cmd.Marshal()
	require.NoError(t, err)
	// There is no service field in the marshaled command.
	require.JSONEq(t,
		`{
             "command":"list-commands"
         }`,
		string(marshaled))

	// The daemon list should be unchanged after marshaling.
	require.Len(t, cmd.Daemons, 1)
	require.Equal(t, daemonname.CA, cmd.Daemons[0])
}

// Test that GetCommand() function returns the command name.
func TestGetCommand(t *testing.T) {
	command := NewCommandBase(ListCommands, daemonname.DHCPv4)
	require.NotNil(t, command)
	require.Equal(t, ListCommands, command.GetCommand())
}

// Test that Response properly implements the ExaminableResponse interface.
func TestExaminableResponse(t *testing.T) {
	arguments := []byte("{}")
	response := Response{
		ResponseHeader: ResponseHeader{
			Result: ResponseError,
			Text:   "a response text",
		},
		Arguments: arguments,
	}
	require.EqualValues(t, 1, response.GetResult())
	require.Equal(t, "a response text", response.GetText())
	require.Equal(t, arguments, []byte(response.GetArguments()))
}

// Test returning an error for a response with error status.
func TestGetResponseError(t *testing.T) {
	response := Response{
		ResponseHeader: ResponseHeader{
			Result: ResponseError,
			Text:   "another response text",
		},
	}
	err := GetResponseError(response)
	require.ErrorContains(t, err, "error status (1) returned by Kea")
}

// Test returning an error for a response with unsupported command status.
func TestGetResponseUnsupportedCommand(t *testing.T) {
	response := Response{
		ResponseHeader: ResponseHeader{
			Result: ResponseCommandUnsupported,
			Text:   "it is unsupported",
		},
	}
	err := GetResponseError(response)
	require.ErrorContains(t, err, "unsupported command status (2) returned by Kea with text: 'it is unsupported'")
}

// Test that no error is returned for a response with empty status.
func TestGetResponseEmpty(t *testing.T) {
	response := Response{
		ResponseHeader: ResponseHeader{
			Result: ResponseEmpty,
		},
	}
	require.Nil(t, GetResponseError(response))
}

// Test that no error is returned for a response with success status.
func TestGetResponseSuccess(t *testing.T) {
	response := Response{
		ResponseHeader: ResponseHeader{
			Result: ResponseSuccess,
		},
	}
	require.Nil(t, GetResponseError(response))
}

// Test that the error is constructed properly.
func TestResponseHeaderError(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		require.Nil(t, (ResponseHeader{Result: 0}).GetError())
	})

	t.Run("error without text", func(t *testing.T) {
		require.ErrorContains(t,
			(ResponseHeader{Result: 42}).GetError(),
			"non-success response result from Kea: 42",
		)
	})

	t.Run("error with text", func(t *testing.T) {
		require.ErrorContains(t,
			(ResponseHeader{
				Result: 42,
				Text:   "foobar",
			}).GetError(),
			"non-success response result from Kea: 42, text: foobar",
		)
	})

	t.Run("empty response is not an error", func(t *testing.T) {
		require.Nil(t, (ResponseHeader{Result: ResponseEmpty}).GetError())
	})

	t.Run("unsupported operation", func(t *testing.T) {
		header := ResponseHeader{
			Result: ResponseCommandUnsupported,
			Text:   "unsupported operation",
		}
		err := header.GetError()
		require.ErrorAs(t, err, &UnsupportedOperationKeaError{})
		require.ErrorContains(t,
			err,
			"non-success response result from Kea: 2, text: unsupported operation",
		)
	})

	t.Run("number overflow", func(t *testing.T) {
		header := ResponseHeader{
			Result: ResponseError,
			Text:   "Number overflow",
		}
		err := header.GetError()
		require.ErrorAs(t, err, &NumberOverflowKeaError{})
		require.ErrorContains(t,
			err,
			"non-success response result from Kea: 1, text: Number overflow",
		)
	})

	t.Run("connectivity error", func(t *testing.T) {
		header := ResponseHeader{
			Result: ResponseError,
			Text:   "server is likely to be offline",
		}
		err := header.GetError()
		require.ErrorAs(t, err, &ConnectivityIssueKeaError{})
		require.ErrorContains(t,
			err,
			"non-success response result from Kea: 1, text: server is likely to be offline",
		)
	})
}
