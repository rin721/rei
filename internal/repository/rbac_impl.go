package repository

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/rin721/rei/internal/models"
	pkgdbtx "github.com/rin721/rei/pkg/dbtx"
	"gorm.io/gorm"
)

type roleRepository struct {
	*gormRepository[models.Role]
}

type userRoleRepository struct {
	*gormRepository[models.UserRole]
}

type policyRepository struct {
	*gormRepository[models.Policy]
}

// NewRoleRepository 创建角色仓储。
func NewRoleRepository(db *gorm.DB, tx *pkgdbtx.Manager) RoleRepository {
	return &roleRepository{
		gormRepository: newGormRepository[models.Role](db, tx),
	}
}

// NewUserRoleRepository 创建用户角色关系仓储。
func NewUserRoleRepository(db *gorm.DB, tx *pkgdbtx.Manager) UserRoleRepository {
	return &userRoleRepository{
		gormRepository: newGormRepository[models.UserRole](db, tx),
	}
}

// NewPolicyRepository 创建策略仓储。
func NewPolicyRepository(db *gorm.DB, tx *pkgdbtx.Manager) PolicyRepository {
	return &policyRepository{
		gormRepository: newGormRepository[models.Policy](db, tx),
	}
}

func (r *roleRepository) Ensure(ctx context.Context, role *models.Role) error {
	role.Name = strings.TrimSpace(role.Name)
	return r.Upsert(ctx, role, "name")
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.query(ctx).Where("name = ?", strings.TrimSpace(name)).First(&role).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	if err := r.query(ctx).Order("name asc").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *userRoleRepository) Assign(ctx context.Context, userRole *models.UserRole) error {
	userRole.RoleName = strings.TrimSpace(userRole.RoleName)
	return r.Upsert(ctx, userRole, "user_id", "role_name")
}

func (r *userRoleRepository) Revoke(ctx context.Context, userID, roleName string) error {
	return r.query(ctx).Where("user_id = ? AND role_name = ?", userID, strings.TrimSpace(roleName)).Delete(&models.UserRole{}).Error
}

func (r *userRoleRepository) ListRolesByUser(ctx context.Context, userID string) ([]string, error) {
	var roles []string
	if err := r.query(ctx).Model(&models.UserRole{}).Where("user_id = ?", userID).Order("role_name asc").Pluck("role_name", &roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *userRoleRepository) ListUsersByRole(ctx context.Context, roleName string) ([]string, error) {
	var userIDs []string
	if err := r.query(ctx).Model(&models.UserRole{}).Where("role_name = ?", strings.TrimSpace(roleName)).Order("user_id asc").Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (r *userRoleRepository) List(ctx context.Context) ([]models.UserRole, error) {
	var bindings []models.UserRole
	if err := r.query(ctx).Order("user_id asc, role_name asc").Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

func (r *policyRepository) Add(ctx context.Context, policy *models.Policy) error {
	policy.Subject = strings.TrimSpace(policy.Subject)
	policy.Object = strings.TrimSpace(policy.Object)
	policy.Action = strings.TrimSpace(strings.ToLower(policy.Action))
	return r.Upsert(ctx, policy, "subject", "object", "action")
}

func (r *policyRepository) Remove(ctx context.Context, subject, object, action string) error {
	return r.query(ctx).
		Where("subject = ? AND object = ? AND action = ?", strings.TrimSpace(subject), strings.TrimSpace(object), strings.TrimSpace(strings.ToLower(action))).
		Delete(&models.Policy{}).Error
}

func (r *policyRepository) List(ctx context.Context) ([]models.Policy, error) {
	var policies []models.Policy
	if err := r.query(ctx).Order("subject asc, object asc, action asc").Find(&policies).Error; err != nil {
		return nil, err
	}
	return policies, nil
}
