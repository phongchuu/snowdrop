package openapi

import (
	"net/http"

	"internal.snowdrop/common/core/web"
)

type DocumentationRoute struct{}

var _ web.HTTPHandler = (*DocumentationRoute)(nil)

func NewDocumentationRoute() *DocumentationRoute {
	return &DocumentationRoute{}
}

// Method implements web.HTTPHandler.
func (o *DocumentationRoute) Method() string {
	return http.MethodGet
}

// Path implements web.HTTPHandler.
func (o *DocumentationRoute) Path() string {
	return "/api/openapi"
}

// Tags implements web.HTTPHandler.
func (o *DocumentationRoute) Tags() []web.RouteTag {
	return []web.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (o *DocumentationRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	responseCtrl := web.NewResponseBuilder(w, r)
	responseCtrl.HTML(`
	<!doctype html>
	<html lang="en">
	<head>
		<title>API Documentation</title>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<script id="api-reference" data-url="/public/static/openapi.yaml"></script>
		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	</body>
	</html>`)
}
