package web

import (
	"time"

	snowdrop "internal.snowdrop/framework"
)

// readHeaderTimeout specifies the maximum duration for reading
// the headers of a request. This helps to prevent slowloris attacks
// by limiting the time allowed to read the headers.
const readHeaderTimeout = 2 * time.Second

// PublicRoute is a route tag for public routes.
const PublicRoute snowdrop.RouteTag = 0

// PrivateRoute is a route tag for private routes.
const PrivateRoute snowdrop.RouteTag = 1
