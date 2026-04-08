package models

// User 定义系统用户模型。
type User struct {
	BaseModel
	Username     string `gorm:"size:64;not null;uniqueIndex" json:"username"`
	Email        string `gorm:"size:128;index" json:"email,omitempty"`
	DisplayName  string `gorm:"size:128;not null" json:"displayName"`
	PasswordHash string `gorm:"column:password_hash;size:255;not null" json:"-"`
	Status       string `gorm:"size:32;not null;default:active" json:"status"`
}

// TableName 返回 users 表名。
func (User) TableName() string {
	return "users"
}
