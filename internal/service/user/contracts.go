package user

import (
	"context"
	"time"

	domainuser "github.com/rin721/rei/internal/domain/user"
)

// UseCase defines the user application contract.
type UseCase interface {
	GetProfile(context.Context, GetProfileQuery) (Profile, error)
	UpdateProfile(context.Context, UpdateProfileCommand) (Profile, error)
}

// GetProfileQuery describes the input needed to fetch a profile.
type GetProfileQuery struct {
	UserID string
}

// UpdateProfileCommand describes the input needed to update a profile.
type UpdateProfileCommand struct {
	UserID      string
	DisplayName string
	Email       string
}

// Profile is the application-layer result returned by the user usecase.
type Profile struct {
	ID          string
	Username    string
	DisplayName string
	Email       string
	Roles       []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserStore defines the persistence port for user entities.
type UserStore interface {
	FindByID(context.Context, string) (*domainuser.User, error)
	Save(context.Context, *domainuser.User) error
}

// RoleBindingReader defines the read port for user role memberships.
type RoleBindingReader interface {
	ListRolesByUser(context.Context, string) ([]string, error)
}
