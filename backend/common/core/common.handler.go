package core

import (
	"net/http"

	"internal.snowdrop/framework/response"
	"internal.snowdrop/framework/web"
)

// NewNotFoundHandler returns a handler function that responds with 404 Not Found status
// using JSON format for all unmatched routes.
func NewNotFoundHandler() web.NotFoundHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		response.NewBuilder(w, r).Status(http.StatusNotFound).JSON()
	}
}

// NewMethodNotAllowedHandler returns a handler function that responds with 405 Method Not Allowed status
// when a request uses an unsupported HTTP method for a route.
func NewMethodNotAllowedHandler() web.MethodNotAllowedHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		response.NewBuilder(w, r).Status(http.StatusMethodNotAllowed).JSON()
	}
}
