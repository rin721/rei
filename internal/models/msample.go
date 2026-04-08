package models

// Sample 定义示例业务模块模型。
type Sample struct {
	BaseModel
	Name        string `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Description string `gorm:"size:255" json:"description"`
	Enabled     bool   `gorm:"not null;default:true" json:"enabled"`
}

// TableName 返回 samples 表名。
func (Sample) TableName() string {
	return "samples"
}
