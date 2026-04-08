package sample

import (
	"context"
	"fmt"

	"github.com/rei0721/go-scaffold2/internal/repository"
	"github.com/rei0721/go-scaffold2/types"
)

// Dependencies 描述示例服务依赖。
type Dependencies struct {
	Samples repository.SampleRepository
}

// Service 实现示例业务模块。
type Service struct {
	deps Dependencies
}

// New 创建示例服务。
func New(deps Dependencies) (*Service, error) {
	if deps.Samples == nil {
		return nil, fmt.Errorf("samples repository is required")
	}
	return &Service{deps: deps}, nil
}

// List 返回启用中的示例数据。
func (s *Service) List(ctx context.Context) ([]types.SampleItemResponse, error) {
	samples, err := s.deps.Samples.ListEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled samples: %w", err)
	}

	items := make([]types.SampleItemResponse, 0, len(samples))
	for _, sample := range samples {
		items = append(items, types.SampleItemResponse{
			ID:          sample.ID,
			Name:        sample.Name,
			Description: sample.Description,
		})
	}

	return items, nil
}
