package app

import (
	"context"
	"fmt"

	"github.com/rin721/rei/internal/models"
	"github.com/rin721/rei/internal/repository"
	"github.com/rin721/rei/internal/router"
	"github.com/rin721/rei/internal/service"
)

type rbacBusinessSeeder struct{}

func (rbacBusinessSeeder) Name() string {
	return "rbac"
}

func (rbacBusinessSeeder) Seed(ctx context.Context, a *App, repos *repository.Set) error {
	defaultRoles := []string{
		service.DefaultRoleAdmin,
		service.DefaultRoleUser,
	}
	for _, roleName := range defaultRoles {
		id, err := nextBusinessID(a.idGen)
		if err != nil {
			return fmt.Errorf("generate role id: %w", err)
		}
		if err := repos.Roles.Ensure(ctx, &models.Role{
			BaseModel: models.BaseModel{
				ID: id,
			},
			Name:        roleName,
			Description: roleDescription(roleName),
		}); err != nil {
			return fmt.Errorf("seed role %q: %w", roleName, err)
		}
	}

	defaultPolicies := []struct {
		subject string
		object  string
		action  string
	}{
		{service.DefaultRoleAdmin, router.RouteRBACCheck, "get"},
		{service.DefaultRoleAdmin, router.RouteRBACAssignRole, "post"},
		{service.DefaultRoleAdmin, router.RouteRBACRevokeRole, "post"},
		{service.DefaultRoleAdmin, router.RouteRBACUserRoles, "get"},
		{service.DefaultRoleAdmin, router.RouteRBACRoleUsers, "get"},
		{service.DefaultRoleAdmin, router.RouteRBACPolicies, "post"},
		{service.DefaultRoleAdmin, router.RouteRBACPolicies, "delete"},
		{service.DefaultRoleAdmin, router.RouteRBACPolicies, "get"},
		{service.DefaultRoleAdmin, router.RouteUserMe, "get"},
		{service.DefaultRoleAdmin, router.RouteUserMe, "put"},
		{service.DefaultRoleUser, router.RouteUserMe, "get"},
		{service.DefaultRoleUser, router.RouteUserMe, "put"},
	}
	for _, item := range defaultPolicies {
		policy, err := newPolicyModel(a.idGen, item.subject, item.object, item.action)
		if err != nil {
			return fmt.Errorf("generate policy model: %w", err)
		}
		if err := repos.Policies.Add(ctx, &policy); err != nil {
			return fmt.Errorf("seed policy %q %q %q: %w", policy.Subject, policy.Object, policy.Action, err)
		}
	}

	return nil
}
