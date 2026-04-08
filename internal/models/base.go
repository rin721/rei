package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 定义业务模型共用字段。
type BaseModel struct {
	ID        string         `gorm:"primaryKey;size:32" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
