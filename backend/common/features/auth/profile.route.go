package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"reflect"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"internal.snowdrop/common/features/usermgt"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/opentelemetry"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/framework/web"
)

type ProfileRoute struct {
	userService usermgt.UserService
	tracer      trace.Tracer
}

type ProfileRouteParams struct {
	fx.In
	UserService usermgt.UserService
	Tracer      trace.Tracer
}

var _ snowdrop.HTTPHandler = (*ProfileRoute)(nil)

func NewProfileRoute(p ProfileRouteParams) ProfileRoute {
	return ProfileRoute{
		userService: p.UserService,
		tracer:      p.Tracer,
	}
}

func (ProfileRoute) Method() string {
	return http.MethodGet
}

func (ProfileRoute) Path() string {
	return "/auth/profile"
}

func (ProfileRoute) Tags() []snowdrop.RouteTag {
	return []snowdrop.RouteTag{web.PrivateRoute}
}

func (h ProfileRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	spanCtx, span := h.tracer.Start(
		r.Context(),
		opentelemetry.BuildSpanName(reflect.TypeOf(h).Name(), "ServeHTTP"),
	)
	defer span.End()

	responseBuilder := response.NewBuilder(w, r.WithContext(spanCtx))

	sesionModel, err := snowdrop.GetCurrentSession(r.WithContext(spanCtx))
	if err != nil {
		if errors.Is(err, snowdrop.ErrNoSession) {
			span.SetStatus(codes.Error, snowdrop.ErrNoSession.Error())
			responseBuilder.Status(http.StatusUnauthorized).JSON()
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		responseBuilder.Status(http.StatusInternalServerError).JSON()

		return
	}

	userDTO, err := h.userService.GetUserByID(spanCtx, sesionModel.UserID.UUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			responseBuilder.Status(http.StatusUnauthorized).JSON()

			return
		}

		responseBuilder.Status(http.StatusInternalServerError).JSON()

		return
	}

	responseBuilder.Status(http.StatusOK).Data(userDTO).JSON()
}
