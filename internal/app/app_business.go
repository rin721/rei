package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rei0721/go-scaffold2/internal/handler"
	"github.com/rei0721/go-scaffold2/internal/models"
	"github.com/rei0721/go-scaffold2/internal/repository"
	"github.com/rei0721/go-scaffold2/internal/router"
	"github.com/rei0721/go-scaffold2/internal/service"
	authservice "github.com/rei0721/go-scaffold2/internal/service/auth"
	rbacservice "github.com/rei0721/go-scaffold2/internal/service/rbac"
	sampleservice "github.com/rei0721/go-scaffold2/internal/service/sample"
	userservice "github.com/rei0721/go-scaffold2/internal/service/user"
)

func (a *App) initBusiness() error {
	if a.handlers != nil {
		return nil
	}
	if a.database == nil {
		return fmt.Errorf("init business: database is required")
	}
	if a.dbtx == nil {
		return fmt.Errorf("init business: dbtx is required")
	}
	if a.cache == nil {
		return fmt.Errorf("init business: cache is required")
	}
	if a.crypto == nil {
		return fmt.Errorf("init business: crypto service is required")
	}
	if a.jwt == nil {
		return fmt.Errorf("init business: jwt manager is required")
	}
	if a.rbac == nil {
		return fmt.Errorf("init business: rbac manager is required")
	}
	if a.idGen == nil {
		return fmt.Errorf("init business: id generator is required")
	}

	if err := a.database.DB().AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.Policy{},
		&models.Sample{},
	); err != nil {
		return fmt.Errorf("auto migrate business models: %w", err)
	}

	repos := repository.NewSet(a.database.DB(), a.dbtx)
	if err := a.seedBusiness(context.Background(), repos); err != nil {
		return fmt.Errorf("seed business data: %w", err)
	}

	authSvc, err := authservice.New(authservice.Dependencies{
		Users:           repos.Users,
		Roles:           repos.Roles,
		UserRoles:       repos.UserRoles,
		IDProvider:      a.idGen,
		Password:        a.crypto,
		Tokens:          a.jwt,
		Cache:           a.cache,
		Tx:              a.dbtx,
		RoleManager:     a.rbac,
		RefreshTokenTTL: time.Duration(a.cfg.JWT.RefreshTokenTTLHours) * time.Hour,
	})
	if err != nil {
		return fmt.Errorf("init auth service: %w", err)
	}

	userSvc, err := userservice.New(userservice.Dependencies{
		Users:     repos.Users,
		UserRoles: repos.UserRoles,
	})
	if err != nil {
		return fmt.Errorf("init user service: %w", err)
	}

	rbacSvc, err := rbacservice.New(rbacservice.Dependencies{
		Users:       repos.Users,
		Roles:       repos.Roles,
		UserRoles:   repos.UserRoles,
		Policies:    repos.Policies,
		IDProvider:  a.idGen,
		Tx:          a.dbtx,
		RoleManager: a.rbac,
	})
	if err != nil {
		return fmt.Errorf("init rbac service: %w", err)
	}
	if err := rbacSvc.LoadFromStore(context.Background()); err != nil {
		return fmt.Errorf("load rbac state from store: %w", err)
	}

	sampleSvc, err := sampleservice.New(sampleservice.Dependencies{
		Samples: repos.Samples,
	})
	if err != nil {
		return fmt.Errorf("init sample service: %w", err)
	}

	a.handlers = handler.NewBundle(authSvc, userSvc, rbacSvc, sampleSvc)
	return nil
}

func (a *App) seedBusiness(ctx context.Context, repos *repository.Set) error {
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

	sampleID, err := nextBusinessID(a.idGen)
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
