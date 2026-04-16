package repository

import (
	"context"

	"github.com/rin721/rei/internal/models"
	pkgdbtx "github.com/rin721/rei/pkg/dbtx"
	"gorm.io/gorm"
)

type sampleRepository struct {
	*gormRepository[models.Sample]
}

// NewSampleRepository 创建示例模块仓储。
func NewSampleRepository(db *gorm.DB, tx *pkgdbtx.Manager) SampleRepository {
	return &sampleRepository{
		gormRepository: newGormRepository[models.Sample](db, tx),
	}
}

func (r *sampleRepository) Ensure(ctx context.Context, sample *models.Sample) error {
	return r.Upsert(ctx, sample, "name")
}

func (r *sampleRepository) ListEnabled(ctx context.Context) ([]models.Sample, error) {
	var samples []models.Sample
	if err := r.query(ctx).Where("enabled = ?", true).Order("name asc").Find(&samples).Error; err != nil {
		return nil, err
	}
	return samples, nil
}
