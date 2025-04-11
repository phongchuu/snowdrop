package snowdrop

import "net/http"

// RouteTag represents a tag identifier for grouping and categorizing HTTP routes.
type RouteTag = int

// HTTPHandler extends the standard http.Handler interface with additional routing information.
// It provides methods to retrieve the HTTP method, path, and tags associated with the handler.
type HTTPHandler interface {
	http.Handler

	// Method returns the HTTP method (GET, POST, etc.) that this handler responds to.
	Method() string

	// Path returns the URL path pattern that this handler is registered for.
	Path() string

	// Tags returns a slice of RouteTag that categorize or label this handler for documentation
	// and organizational purposes.
	Tags() []RouteTag
}
