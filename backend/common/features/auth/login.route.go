package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"internal.snowdrop/common/features/usermgt"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/framework/session"
	"internal.snowdrop/framework/web"
)

type LoginRoute struct {
	validator      *validator.Validate
	userService    usermgt.UserService
	sessionManager *session.Manager
	logger         *slog.Logger
	tracer         trace.Tracer
}

type LoginRouteParams struct {
	fx.In
	Validator      *validator.Validate
	UserService    usermgt.UserService
	SessionManager *session.Manager
	Logger         *slog.Logger
	Tracer         trace.Tracer
}

type LoginFormData struct {
	Username string `json:"username" validate:"required,min=4,max=255"`
	Password string `json:"password" validate:"required,min=6,max=73"`
}

var _ snowdrop.HTTPHandler = (*LoginRoute)(nil)

func NewLoginRoute(p LoginRouteParams) LoginRoute {
	return LoginRoute{
		validator:      p.Validator,
		userService:    p.UserService,
		sessionManager: p.SessionManager,
		logger:         p.Logger,
		tracer:         p.Tracer,
	}
}

// Method implements web.HTTPHandler.
func (loginRoute LoginRoute) Method() string {
	return http.MethodPost
}

// Path implements web.HTTPHandler.
func (loginRoute LoginRoute) Path() string {
	return "/auth/login"
}

// Tags implements web.HTTPHandler.
func (loginRoute LoginRoute) Tags() []snowdrop.RouteTag {
	return []snowdrop.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (loginRoute LoginRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := response.NewBuilder(w, r)
	binder := web.NewBinder(web.WithValidator(loginRoute.validator))

	_, bindingDataSpan := loginRoute.tracer.Start(
		r.Context(),
		"Binding request body to struct",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer bindingDataSpan.End()

	var formData LoginFormData
	if err := binder.JSON(r, &formData); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			bindingDataSpan.SetStatus(codes.Error, "Request body is invalid")
			response.Status(http.StatusBadRequest).Errors(validationErrs).JSON()

			return
		}

		bindingDataSpan.RecordError(err)
		bindingDataSpan.SetStatus(codes.Error, "Something went wrong when binding request body")
		response.Status(http.StatusInternalServerError).JSON()

		return
	}

	userDTO, err := loginRoute.userService.Login(r.Context(), formData.Username, formData.Password)
	if err != nil {
		response.Status(http.StatusBadRequest)

		return
	}

	session, err := loginRoute.sessionManager.UpgradeSession(w, r, userDTO.ID)
	if err != nil {
		response.Status(http.StatusBadRequest).Message(err.Error())

		return
	}

	response.Status(http.StatusOK).Data(session)
}
