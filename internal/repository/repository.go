package repository

import (
	"context"
	stderrors "errors"

	pkgdbtx "github.com/rin721/go-scaffold2/pkg/dbtx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository 定义通用仓储最小契约。
type Repository[T any] interface {
	Create(context.Context, *T) error
	Save(context.Context, *T) error
	FindByID(context.Context, string) (*T, error)
}

type gormRepository[T any] struct {
	db *gorm.DB
	tx *pkgdbtx.Manager
}

func newGormRepository[T any](db *gorm.DB, tx *pkgdbtx.Manager) *gormRepository[T] {
	return &gormRepository[T]{
		db: db,
		tx: tx,
	}
}

func (r *gormRepository[T]) query(ctx context.Context) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	if r.tx != nil {
		return r.tx.GetDB(ctx).WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

func (r *gormRepository[T]) Create(ctx context.Context, value *T) error {
	return r.query(ctx).Create(value).Error
}

func (r *gormRepository[T]) Save(ctx context.Context, value *T) error {
	return r.query(ctx).Save(value).Error
}

func (r *gormRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	var model T
	if err := r.query(ctx).First(&model, "id = ?", id).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model, nil
}

func (r *gormRepository[T]) FirstWhere(ctx context.Context, model *T, query any, args ...any) error {
	return r.query(ctx).Where(query, args...).First(model).Error
}

func (r *gormRepository[T]) FindWhere(ctx context.Context, query any, args ...any) ([]T, error) {
	var result []T
	if err := r.query(ctx).Where(query, args...).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *gormRepository[T]) Upsert(ctx context.Context, value *T, columns ...string) error {
	onConflict := clause.OnConflict{DoNothing: true}
	if len(columns) > 0 {
		onConflict.Columns = make([]clause.Column, 0, len(columns))
		for _, column := range columns {
			onConflict.Columns = append(onConflict.Columns, clause.Column{Name: column})
		}
	}
	return r.query(ctx).Clauses(onConflict).Create(value).Error
}

// Set 统一持有业务仓储。
type Set struct {
	Users     UserRepository
	Roles     RoleRepository
	UserRoles UserRoleRepository
	Policies  PolicyRepository
	Samples   SampleRepository
}

// NewSet 创建一组基于 GORM 的业务仓储。
func NewSet(db *gorm.DB, tx *pkgdbtx.Manager) *Set {
	return &Set{
		Users:     NewUserRepository(db, tx),
		Roles:     NewRoleRepository(db, tx),
		UserRoles: NewUserRoleRepository(db, tx),
		Policies:  NewPolicyRepository(db, tx),
		Samples:   NewSampleRepository(db, tx),
	}
}
