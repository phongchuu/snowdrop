package staticfile

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"internal.snowdrop/common/core/config"
	"internal.snowdrop/common/core/web"
)

type FileServerRoute struct {
	embedResourcesFolder fs.FS
}

var _ web.HTTPHandler = (*FileServerRoute)(nil)

func NewStaticRoute(config config.Manager) *FileServerRoute {
	return &FileServerRoute{
		embedResourcesFolder: config.GetEmbedResourceFolder(),
	}
}

// Method implements web.HTTPHandler.
func (s *FileServerRoute) Method() string {
	return http.MethodGet
}

// Path implements web.HTTPHandler.
func (s *FileServerRoute) Path() string {
	return "/public/*"
}

// Tags implements web.HTTPHandler.
func (s *FileServerRoute) Tags() []web.RouteTag {
	return []web.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (s *FileServerRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fSys, _ := fs.Sub(s.embedResourcesFolder, "resources/public")
	rctx := chi.RouteContext(r.Context())
	pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
	fs := http.StripPrefix(pathPrefix, http.FileServer(http.FS(fSys)))
	fs.ServeHTTP(w, r)
}
