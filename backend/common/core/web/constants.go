package web

import "time"

// readHeaderTimeout specifies the maximum duration for reading
// the headers of a request. This helps to prevent slowloris attacks
// by limiting the time allowed to read the headers.
const readHeaderTimeout = 2 * time.Second

// defaultPage is the default page number for pagination.
const defaultPage = 1

// defaultPageSize is the default number of items per page for pagination.
const defaultPageSize = 10

// maxPageSize is the maximum number of items per page for pagination.
const maxPageSize = 100

// PublicRoute is a route tag for public routes.
const PublicRoute RouteTag = 0

// PrivateRoute is a route tag for private routes.
const PrivateRoute RouteTag = 1
