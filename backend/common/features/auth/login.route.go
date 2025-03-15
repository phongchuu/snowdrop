package auth

import (
	"net/http"

	"github.com/gorilla/schema"
	"internal.snowdrop/common/core/web"
)

type LoginRoute struct {
	schemaDecoder *schema.Decoder
}

type LoginFormData struct {
	Username string `schema:"username"`
	Password string `schema:"password"`
}

var _ web.HTTPHandler = (*LoginRoute)(nil)

func NewLoginRoute(schemaDecoder *schema.Decoder) *LoginRoute {
	return &LoginRoute{
		schemaDecoder: schemaDecoder,
	}
}

// Method implements web.HTTPHandler.
func (l LoginRoute) Method() string {
	return http.MethodPost
}

// Path implements web.HTTPHandler.
func (l LoginRoute) Path() string {
	return "/auth/login"
}

// Tags implements web.HTTPHandler.
func (l LoginRoute) Tags() []web.RouteTag {
	return []web.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (l LoginRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	respCtrl := web.NewResponseBuilder(w, r)

	var formData LoginFormData

	if err := l.schemaDecoder.Decode(&formData, r.PostForm); err != nil {
		respCtrl.Status(http.StatusBadRequest).Message(err.Error()).JSON()
	}
}
