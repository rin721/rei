package models

// Role 定义 RBAC 角色模型。
type Role struct {
	BaseModel
	Name        string `gorm:"size:64;not null;uniqueIndex" json:"name"`
	Description string `gorm:"size:255" json:"description"`
}

// TableName 返回 roles 表名。
func (Role) TableName() string {
	return "roles"
}

// UserRole 定义用户与角色的关系模型。
type UserRole struct {
	BaseModel
	UserID   string `gorm:"size:32;not null;index:idx_user_role_user,priority:1;uniqueIndex:uidx_user_role_binding,priority:1" json:"userId"`
	RoleName string `gorm:"size:64;not null;index:idx_user_role_role,priority:1;uniqueIndex:uidx_user_role_binding,priority:2" json:"role"`
}

// TableName 返回 user_roles 表名。
func (UserRole) TableName() string {
	return "user_roles"
}
