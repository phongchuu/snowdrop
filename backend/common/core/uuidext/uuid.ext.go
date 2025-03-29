package uuidext

import "github.com/google/uuid"

// MustUUIDV7 generates a new UUID version 7 and panics if generation fails.
func MustUUIDV7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return id
}

// IsNil returns true if the provided pointer is nil or if it points to a UUID equal to uuid.Nil.
func IsNil(value *uuid.UUID) bool {
	return value == nil || *value == uuid.Nil
}
