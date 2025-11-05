package storkutil

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// This type allows assign a JSON raw message (byte array) or any other type.
// It is dedicated to construct objects from the Golang objects or already
// marshalled data without redundant marshalling or unmarshalling.
//
// If this type is marshalled and the inner type is json.RawMessage, it is used
// as-is (as byte array). If the inner type is any other type, it is marshalled
// to JSON.
//
// If this type is unmarshalled and the inner type is json.RawMessage or nil,
// it  is unmarshalled as json.RawMessage (so the bytes are just stored into
// it). If the inner type is any other type, it is unmarshalled from JSON.
type RawMessageOrAny struct {
	value any
}

var (
	_ json.Marshaler   = (*RawMessageOrAny)(nil)
	_ json.Unmarshaler = (*RawMessageOrAny)(nil)
)

// Constructs a new object instance.
func NewRawMessageOrAny(value any) *RawMessageOrAny {
	if IsNilPtr(value) {
		return nil
	}
	return &RawMessageOrAny{value: value}
}

// MarshalJSON implements json.Marshaler.
func (r RawMessageOrAny) MarshalJSON() (bytes []byte, err error) {
	if raw, ok := r.value.(json.RawMessage); ok {
		bytes, err = json.Marshal(raw)
	} else {
		bytes, err = json.Marshal(r.value)
	}
	return bytes, errors.WithStack(err)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *RawMessageOrAny) UnmarshalJSON(bytes []byte) error {
	isRawMessage := IsNilPtr(r.value)
	if !isRawMessage {
		_, isRawMessage = r.value.(json.RawMessage)
	}

	if isRawMessage {
		// Unmarshal as new RawMessage.
		var raw json.RawMessage
		err := json.Unmarshal(bytes, &raw)
		if err != nil {
			return errors.WithStack(err)
		}
		r.value = raw
		return nil
	}
	// Try to unmarshal into the inner object.
	err := json.Unmarshal(bytes, &r.value)
	return errors.WithStack(err)
}

// Unmarshal the inner value into the provided target.
func (r RawMessageOrAny) UnmarshalInto(target any) error {
	var bytes []byte
	var err error
	if raw, ok := r.value.(json.RawMessage); ok {
		// For raw message, unmarshal it.
		bytes = raw
	} else {
		// For other types, marshal it and unmarshal into target.
		bytes, err = json.Marshal(r.value)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	err = json.Unmarshal(bytes, target)
	return errors.WithStack(err)
}
