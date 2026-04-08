package repository

import (
	"context"

	"github.com/rei0721/go-scaffold2/internal/models"
)

// UserRepository 定义用户仓储契约。
type UserRepository interface {
	Repository[models.User]
	FindByUsername(context.Context, string) (*models.User, error)
	Count(context.Context) (int64, error)
}
