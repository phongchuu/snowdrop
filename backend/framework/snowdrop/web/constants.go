package web

import (
	"net/http"
	"time"

	snowdrop "internal.snowdrop/framework"
)

type (
	NotFoundHandler         http.HandlerFunc
	MethodNotAllowedHandler http.HandlerFunc
)

// readHeaderTimeout specifies the maximum duration for reading
// the headers of a request. This helps to prevent slowloris attacks
// by limiting the time allowed to read the headers.
const readHeaderTimeout = 2 * time.Second

const (
	// PublicRoute is a route tag for public routes.
	PublicRoute snowdrop.RouteTag = iota

	// PrivateRoute is a route tag for private routes.
	PrivateRoute
)
