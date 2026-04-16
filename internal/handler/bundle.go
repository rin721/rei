package handler

import "github.com/rin721/rei/internal/service"

// Bundle 统一持有业务处理器集合。
type Bundle struct {
	Auth   *AuthHandler
	User   *UserHandler
	RBAC   *RBACHandler
	Sample *SampleHandler
}

// NewBundle 创建业务处理器集合。
func NewBundle(authService service.AuthService, userService service.UserService, rbacService service.RBACService, sampleService service.SampleService) *Bundle {
	return &Bundle{
		Auth:   NewAuthHandler(authService),
		User:   NewUserHandler(userService),
		RBAC:   NewRBACHandler(rbacService),
		Sample: NewSampleHandler(sampleService),
	}
}
