package app

import (
	"context"
	"fmt"

	"github.com/rin721/rei/internal/models"
	"github.com/rin721/rei/internal/repository"
)

type sampleBusinessSeeder struct{}

func (sampleBusinessSeeder) Name() string {
	return "sample"
}

func (sampleBusinessSeeder) Seed(ctx context.Context, deps businessProvisioning, repos *repository.Set) error {
	sampleID, err := nextBusinessID(deps.idGen)
	if err != nil {
		return fmt.Errorf("generate sample id: %w", err)
	}
	if err := repos.Samples.Ensure(ctx, &models.Sample{
		BaseModel: models.BaseModel{
			ID: sampleID,
		},
		Name:        "welcome",
		Description: "Phase 7 sample module is ready",
		Enabled:     true,
	}); err != nil {
		return fmt.Errorf("seed sample data: %w", err)
	}
	return nil
}
