package staticfile

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/web"
)

type FileServerRoute struct {
	embedResourcesFolder fs.FS
}

var _ snowdrop.HTTPHandler = (*FileServerRoute)(nil)

func NewStaticRoute(config snowdrop.ConfigManager) FileServerRoute {
	return FileServerRoute{
		embedResourcesFolder: config.GetEmbedResourceFolder(),
	}
}

// Method implements web.HTTPHandler.
func (FileServerRoute) Method() string {
	return http.MethodGet
}

// Path implements web.HTTPHandler.
func (FileServerRoute) Path() string {
	return "/public/*"
}

// Tags implements web.HTTPHandler.
func (FileServerRoute) Tags() []snowdrop.RouteTag {
	return []snowdrop.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (s FileServerRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fSys, _ := fs.Sub(s.embedResourcesFolder, "resources/public")
	rctx := chi.RouteContext(r.Context())
	pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
	fileServer := http.StripPrefix(pathPrefix, http.FileServer(http.FS(fSys)))
	fileServer.ServeHTTP(w, r)
}
