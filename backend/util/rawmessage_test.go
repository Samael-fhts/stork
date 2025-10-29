package storkutil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test that the instance is created correctly.
func TestNewRawMessageOrAny(t *testing.T) {
	// Arrange & Act
	type payload struct {
		Key string `json:"key"`
	}

	itemFromNil := NewRawMessageOrAny(nil)
	itemFromBytes := NewRawMessageOrAny(json.RawMessage(`{"key":"value"}`))
	itemFromInt := NewRawMessageOrAny(42)
	itemFromStruct := NewRawMessageOrAny(payload{
		Key: "value",
	})

	// Assert
	require.Nil(t, itemFromNil)
	require.NotNil(t, itemFromBytes)
	require.Equal(t, json.RawMessage(`{"key":"value"}`), itemFromBytes.value)
	require.NotNil(t, itemFromInt)
	require.Equal(t, 42, itemFromInt.value)
	require.NotNil(t, itemFromStruct)
	require.Equal(t, payload{
		Key: "value",
	}, itemFromStruct.value)
}

// Test that marshalling RawMessage keeps the bytes as-is.
func TestRawMessageOrAnyMarshalBytes(t *testing.T) {
	// Arrange
	item := NewRawMessageOrAny(json.RawMessage(`{"key":"value"}`))

	// Act
	bytes, err := json.Marshal(item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, `{"key":"value"}`, string(bytes))
}

// Test that marshalling primitive types works correctly.
func TestRawMessageOrAnyMarshalPrimitive(t *testing.T) {
	// Arrange
	item := NewRawMessageOrAny(42)

	// Act
	bytes, err := json.Marshal(item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "42", string(bytes))
}

// Test that marshalling structs works correctly.
func TestRawMessageOrAnyMarshalStruct(t *testing.T) {
	// Arrange
	item := NewRawMessageOrAny(struct {
		Key string `json:"key"`
	}{
		Key: "value",
	})

	// Act
	bytes, err := json.Marshal(item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, `{"key":"value"}`, string(bytes))

}

// Test that marshalling nil results in JSON null.
func TestRawMessageOrAnyMarshalNil(t *testing.T) {
	// Arrange
	item := NewRawMessageOrAny(nil)

	// Act
	bytes, err := json.Marshal(item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, "null", string(bytes))
}

// Test that unmarshalling with the inner type as json.RawMessage works keeps
// the new bytes as-is.
func TestRawMessageOrAnyUnmarshalOnBytes(t *testing.T) {
	// Arrange
	item := NewRawMessageOrAny(json.RawMessage(`{"key":"value"}`))

	// Act
	err := json.Unmarshal([]byte(`{"foo":"bar"}`), item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, json.RawMessage(`{"foo":"bar"}`), item.value)
}

// Test that unmarshalling with the inner struct type unmarshals into this
// struct.
func TestRawMessageOrAnyUnmarshalOnStruct(t *testing.T) {
	// Arrange
	type payload struct {
		Value string `json:"value"`
	}

	item := NewRawMessageOrAny(&payload{Value: "foo"})

	// Act
	err := json.Unmarshal([]byte(`{"value":"bar"}`), &item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, &payload{Value: "bar"}, item.value)
}

// Test that unmarshalling with nil inner type results in json.RawMessage.
func TestRawMessageOrAnyUnmarshalOnNil(t *testing.T) {
	// Arrange
	item := NewRawMessageOrAny(nil)

	// Act
	err := json.Unmarshal([]byte(`{"foo":"bar"}`), &item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, json.RawMessage(`{"foo":"bar"}`), item.value)
}

// Test that the value can be unmarshalled into a target object.
func TestRawMessageOrAnyUnmarshalInto(t *testing.T) {
	// Arrange
	type payload struct {
		Value string `json:"value"`
	}

	item := NewRawMessageOrAny(json.RawMessage(`{"value":"bar"}`))

	// Act
	var target payload
	err := item.UnmarshalInto(&target)

	// Assert
	require.NoError(t, err)
	require.Equal(t, payload{Value: "bar"}, target)
}

// Test that RawMessageOrAny can be used as a member of another struct.
func TestRawMessageOrAnyAsMember(t *testing.T) {
	// Arrange
	type container struct {
		Data *RawMessageOrAny `json:"data"`
	}

	var item any = container{
		Data: NewRawMessageOrAny(json.RawMessage(`{"key":"value"}`)),
	}

	// Act
	bytes, err := json.Marshal(item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, `{"data":{"key":"value"}}`, string(bytes))
}
