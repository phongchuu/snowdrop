package web

import (
	"bytes"
	"encoding/json"
	"net/http"

	"go.uber.org/fx"
)

type RouteTag = int

type HTTPHandler interface {
	http.Handler

	Method() string
	Path() string
	Tags() []RouteTag
}

type ResponseController struct {
	w http.ResponseWriter
	r *http.Request
}

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

func HTTPRoute(function any) any {
	return fx.Annotate(
		function,
		fx.As(new(HTTPHandler)),
		fx.ResultTags(`group:"http_routes"`),
	)
}
