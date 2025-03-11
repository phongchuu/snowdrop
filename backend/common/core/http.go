package core

import (
	"bytes"
	"encoding/json"
	"net/http"

	"go.uber.org/fx"
)

type RouteTag = int

type HTTPHandler interface {
	http.Handler
	Tags() []RouteTag
	Pattern() string
}

type ResponseController struct {
	w http.ResponseWriter
	r *http.Request
}

const PublicRoute RouteTag = 0
const PrivateRoute RouteTag = 1

func NewResponseController(w http.ResponseWriter, r *http.Request) *ResponseController {
	return &ResponseController{w, r}
}

func (r ResponseController) Status(code int) {
	r.w.WriteHeader(code)
}

func (r ResponseController) JSON(value any) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(value); err != nil {
		http.Error(r.w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.w.Header().Set("Content-Type", "application/json")
	_, _ = r.w.Write(buf.Bytes())
}

func Route(function any) any {
	return fx.Annotate(
		function,
		fx.As(new(HTTPHandler)),
		fx.ResultTags(`group:"http_routes"`),
	)
}
