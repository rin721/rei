package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rin721/rei/internal/models"
	"github.com/rin721/rei/internal/repository"
	"github.com/rin721/rei/internal/service"
)

type businessSeeder interface {
	Name() string
	Seed(context.Context, businessProvisioning, *repository.Set) error
}

func (p businessProvisioning) seed(ctx context.Context, repos *repository.Set) error {
	for _, seeder := range p.seeders() {
		if err := seeder.Seed(ctx, p, repos); err != nil {
			return fmt.Errorf("seed %s data: %w", seeder.Name(), err)
		}
	}
	return nil
}

func (p businessProvisioning) seeders() []businessSeeder {
	return []businessSeeder{
		rbacBusinessSeeder{},
		sampleBusinessSeeder{},
	}
}

func newPolicyModel(idProvider service.IDProvider, subject, object, action string) (models.Policy, error) {
	id, err := nextBusinessID(idProvider)
	if err != nil {
		return models.Policy{}, err
	}
	return models.Policy{
		BaseModel: models.BaseModel{
			ID: id,
		},
		Subject: subject,
		Object:  object,
		Action:  action,
	}, nil
}

func nextBusinessID(provider service.IDProvider) (string, error) {
	id, err := provider.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func roleDescription(roleName string) string {
	switch roleName {
	case service.DefaultRoleAdmin:
		return "system administrator"
	case service.DefaultRoleUser:
		return "registered user"
	default:
		return "custom role"
	}
}
