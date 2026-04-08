package user

// RegisterRequest 定义用户注册请求。
type RegisterRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// LoginRequest 定义用户登录请求。
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordRequest 定义修改密码请求。
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// RefreshTokenRequest 定义刷新令牌请求。
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// UpdateProfileRequest 定义更新个人资料请求。
type UpdateProfileRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// Profile 定义统一用户资料响应。
type Profile struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	Email       string   `json:"email,omitempty"`
	Roles       []string `json:"roles"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
}

// TokenPair 定义认证成功时返回的令牌信息。
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
}

// AuthResponse 定义认证相关接口的统一响应。
type AuthResponse struct {
	User   Profile   `json:"user"`
	Tokens TokenPair `json:"tokens"`
}
