package auth

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"internal.snowdrop/common/core/session"
	"internal.snowdrop/common/core/web"
	"internal.snowdrop/common/features/usermgt"
)

type LoginRoute struct {
	validator      *validator.Validate
	userService    usermgt.UserService
	sessionManager *session.Manager
}

type LoginRouteParams struct {
	fx.In
	Validator      *validator.Validate
	UserService    usermgt.UserService
	SessionManager *session.Manager
}

type LoginFormData struct {
	Username string `json:"username" validate:"required,min=4,max=255"`
	Password string `json:"password" validate:"required,min=6,max=73"`
}

var _ web.HTTPHandler = (*LoginRoute)(nil)

func NewLoginRoute(p LoginRouteParams) LoginRoute {
	return LoginRoute{
		validator:      p.Validator,
		userService:    p.UserService,
		sessionManager: p.SessionManager,
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
func (loginRoute LoginRoute) Tags() []web.RouteTag {
	return []web.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (loginRoute LoginRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := web.NewResponseBuilder(w, r)
	binder := web.NewBinder(web.WithValidator(loginRoute.validator))

	var formData LoginFormData

	if err := binder.JSON(r, &formData); err != nil {
		response.Status(http.StatusBadRequest).Message(err.Error()).JSON()

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
