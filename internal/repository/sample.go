package repository

import (
	"context"

	"github.com/rei0721/go-scaffold2/internal/models"
)

// SampleRepository 定义示例模块仓储契约。
type SampleRepository interface {
	Repository[models.Sample]
	Ensure(context.Context, *models.Sample) error
	ListEnabled(context.Context) ([]models.Sample, error)
}
