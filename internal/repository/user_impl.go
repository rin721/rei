package repository

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/rin721/rei/internal/models"
	pkgdbtx "github.com/rin721/rei/pkg/dbtx"
	"gorm.io/gorm"
)

type userRepository struct {
	*gormRepository[models.User]
}

// NewUserRepository 创建用户仓储。
func NewUserRepository(db *gorm.DB, tx *pkgdbtx.Manager) UserRepository {
	return &userRepository{
		gormRepository: newGormRepository[models.User](db, tx),
	}
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.query(ctx).Where("username = ?", strings.TrimSpace(strings.ToLower(username))).First(&user).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.query(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
