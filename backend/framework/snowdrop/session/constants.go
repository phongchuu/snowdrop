package session

import "errors"

// ErrNoSessionInContext is returned when attempting to retrieve a session from a context
// where no session information has been stored or attached.
var ErrNoSessionInContext = errors.New("cannot get the session from context")
