package models

// Policy 定义 RBAC 策略模型。
type Policy struct {
	BaseModel
	Subject string `gorm:"size:64;not null;uniqueIndex:uidx_policy_binding,priority:1" json:"subject"`
	Object  string `gorm:"size:255;not null;uniqueIndex:uidx_policy_binding,priority:2" json:"object"`
	Action  string `gorm:"size:32;not null;uniqueIndex:uidx_policy_binding,priority:3" json:"action"`
}

// TableName 返回 policies 表名。
func (Policy) TableName() string {
	return "policies"
}
