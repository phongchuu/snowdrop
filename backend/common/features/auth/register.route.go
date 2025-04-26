package auth

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"internal.snowdrop/common/features/usermgt"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/framework/web"
)

type RegisterRoute struct {
	validator   *validator.Validate
	userService usermgt.UserService
}

type RegisterRouteParams struct {
	fx.In
	Validator   *validator.Validate
	UserService usermgt.UserService
}

type RegisterFormData struct {
	Username string `json:"username" validate:"required,min=4,max=255"`
	Email    string `json:"email"    validate:"required,email,min=6,max=255"`
	Password string `json:"password" validate:"required,min=6,max=73"`
}

var _ snowdrop.HTTPHandler = (*RegisterRoute)(nil)

func NewRegisterRoute(p RegisterRouteParams) RegisterRoute {
	return RegisterRoute{
		validator:   p.Validator,
		userService: p.UserService,
	}
}

// Method implements web.HTTPHandler.
func (RegisterRoute) Method() string {
	return http.MethodPost
}

// Path implements web.HTTPHandler.
func (RegisterRoute) Path() string {
	return "/auth/register"
}

// Tags implements web.HTTPHandler.
func (RegisterRoute) Tags() []snowdrop.RouteTag {
	return []snowdrop.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (registerRoute RegisterRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	responseBuilder := response.NewBuilder(w, r)
	binder := web.NewBinder(web.WithValidator(registerRoute.validator))

	var formData RegisterFormData

	if err := binder.JSON(r, &formData); err != nil {
		responseBuilder.Status(http.StatusBadRequest).Errors(err).JSON()

		return
	}

	userDTO, err := registerRoute.userService.CreateUser(r.Context(), usermgt.CreateUserDTO{
		Username:    formData.Username,
		RawPassword: formData.Password,
		Email:       formData.Email,
	})
	if err != nil {
		responseBuilder.Status(http.StatusBadRequest).Message(err.Error()).JSON()

		return
	}

	responseBuilder.Status(http.StatusCreated).Data(userDTO).JSON()
}
