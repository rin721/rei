package sample

import (
	"context"
	"fmt"

	"github.com/rin721/rei/internal/repository"
	"github.com/rin721/rei/types"
)

// Dependencies 描述示例服务依赖。
type Dependencies struct {
	Samples repository.SampleRepository
	Demos   []ToolkitDemo
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
	if len(deps.Demos) == 0 {
		deps.Demos = DefaultToolkitDemos()
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

// Tooling returns business-facing demo previews for isolated pkg helpers.
func (s *Service) Tooling(ctx context.Context) ([]types.SampleToolkitDemoResponse, error) {
	demos := make([]types.SampleToolkitDemoResponse, 0, len(s.deps.Demos))
	for _, demo := range s.deps.Demos {
		item, err := demo.Build(ctx)
		if err != nil {
			return nil, fmt.Errorf("build %s demo: %w", demo.Module(), err)
		}
		demos = append(demos, item)
	}

	return demos, nil
}
