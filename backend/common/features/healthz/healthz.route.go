package healthz

import (
	"database/sql"
	"log/slog"
	"net/http"
	"slices"

	"github.com/samber/lo"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/log"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/framework/web"
)

type HealthCheckRoute struct {
	logger *slog.Logger
	db     *sql.DB
}

var _ snowdrop.HTTPHandler = (*HealthCheckRoute)(nil)

func NewHealthCheckRoute(
	logger *slog.Logger,
	db *sql.DB,
) HealthCheckRoute {
	return HealthCheckRoute{
		logger,
		db,
	}
}

// Method implements web.HTTPHandler.
func (h HealthCheckRoute) Method() string {
	return http.MethodGet
}

// Path implements web.HTTPHandler.
func (h HealthCheckRoute) Path() string {
	return "/healthz"
}

// Tags implements web.HTTPHandler.
func (h HealthCheckRoute) Tags() []snowdrop.RouteTag {
	return []snowdrop.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (h HealthCheckRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	responseBuilder := response.NewBuilder(w, r)
	data := map[string]string{
		"database": "available",
	}

	if err := h.db.PingContext(r.Context()); err != nil {
		data["database"] = "unavailable"

		h.logger.ErrorContext(r.Context(), "database unavailable", log.ErrorLogAttr(err))
	}

	if slices.Contains(lo.Values(data), "unavailable") {
		responseBuilder.Status(http.StatusServiceUnavailable)
	}

	responseBuilder.Data(data).JSON()
}
