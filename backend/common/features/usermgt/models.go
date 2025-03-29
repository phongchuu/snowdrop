package usermgt

import (
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"internal.snowdrop/common/store"
)

type UserDTO struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	UpdatedAt *time.Time `json:"updatedAt"`
	UpdatedBy *string    `json:"updatedBy"`
}

type UserSetter struct {
	Username string
	Email    string
	Password string
}

// ToUserDTO converts a store.UserModel instance into a UserDTO by mapping its fields.
// It transfers the core user details and conditionally assigns the UpdatedAt and UpdatedBy fields as pointers
// if their corresponding values are valid.
func ToUserDTO(model store.UserModel) UserDTO {
	return UserDTO{
		ID:        model.ID,
		Username:  model.Username,
		Email:     model.Email,
		CreatedAt: model.CreatedAt,
		CreatedBy: model.CreatedBy,
		UpdatedAt: lo.Ternary(model.UpdatedAt.Valid, &model.UpdatedAt.Time, nil),
		UpdatedBy: lo.Ternary(model.UpdatedBy.Valid, &model.UpdatedBy.String, nil),
	}
}
