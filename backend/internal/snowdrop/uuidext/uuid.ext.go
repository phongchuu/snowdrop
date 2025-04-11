package uuidext

import "github.com/google/uuid"

func MustUUIDV7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return id
}

func IsNil(value *uuid.UUID) bool {
	return value == nil || *value == uuid.Nil
}
