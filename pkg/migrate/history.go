package migrate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// History manages the schema_migrations table.
type History struct {
	db *gorm.DB
}

func newHistory(db *gorm.DB) *History {
	return &History{db: db}
}

func (h *History) Exists(ctx context.Context) (bool, error) {
	if h.db == nil {
		return false, fmt.Errorf("history database is nil")
	}
	return h.db.WithContext(ctx).Migrator().HasTable(MigrationRecord{}), nil
}

func (h *History) EnsureTable(ctx context.Context) error {
	sql, err := historyTableDDL(h.db.Dialector.Name())
	if err != nil {
		return err
	}
	if err := h.db.WithContext(ctx).Exec(sql).Error; err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	return nil
}

func historyTableDDL(dialect string) (string, error) {
	switch dialect {
	case "sqlite":
		return `CREATE TABLE IF NOT EXISTS "schema_migrations" (
  "version" TEXT PRIMARY KEY,
  "description" TEXT,
  "applied_at" DATETIME NOT NULL,
  "checksum" TEXT NOT NULL
)`, nil
	case "mysql":
		return `CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(64) NOT NULL,
  description VARCHAR(255),
  applied_at DATETIME NOT NULL,
  checksum VARCHAR(64) NOT NULL,
  PRIMARY KEY (version)
)`, nil
	case "postgres":
		return `CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(64) PRIMARY KEY,
  description VARCHAR(255),
  applied_at TIMESTAMPTZ NOT NULL,
  checksum VARCHAR(64) NOT NULL
)`, nil
	default:
		return "", fmt.Errorf("ensure schema_migrations table: unsupported dialect %q", dialect)
	}
}

func (h *History) IsApplied(ctx context.Context, version string) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).
		Model(&MigrationRecord{}).
		Where("version = ?", version).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check migration %q: %w", version, err)
	}
	return count > 0, nil
}

func (h *History) MarkApplied(ctx context.Context, rec MigrationRecord) error {
	if err := h.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return fmt.Errorf("mark migration %q applied: %w", rec.Version, err)
	}
	return nil
}

func (h *History) Unmark(ctx context.Context, version string) error {
	result := h.db.WithContext(ctx).
		Where("version = ?", version).
		Delete(&MigrationRecord{})
	if result.Error != nil {
		return fmt.Errorf("unmark migration %q: %w", version, result.Error)
	}
	return nil
}

func (h *History) ListApplied(ctx context.Context) ([]MigrationRecord, error) {
	var records []MigrationRecord
	if err := h.db.WithContext(ctx).
		Order("applied_at ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	return records, nil
}

func (h *History) LastApplied(ctx context.Context) (*MigrationRecord, error) {
	var rec MigrationRecord
	err := h.db.WithContext(ctx).
		Order("applied_at DESC").
		First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get last applied migration: %w", err)
	}
	return &rec, nil
}

func newRecord(m *Migration) MigrationRecord {
	return MigrationRecord{
		Version:     m.Version,
		Description: m.Description,
		AppliedAt:   time.Now().UTC(),
		Checksum:    m.Checksum,
	}
}
