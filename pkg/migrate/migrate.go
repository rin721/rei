package migrate

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

// Migrator executes versioned SQL migrations from a scripts directory.
type Migrator struct {
	db         *gorm.DB
	dialect    string
	scriptsDir string
	history    *History
}

func New(db *gorm.DB, dialect, scriptsDir string) *Migrator {
	return &Migrator{
		db:         db,
		dialect:    dialect,
		scriptsDir: scriptsDir,
		history:    newHistory(db),
	}
}

func (m *Migrator) Migrate(ctx context.Context, dryRun bool) ([]string, error) {
	if !dryRun {
		if err := m.history.EnsureTable(ctx); err != nil {
			return nil, err
		}
	}

	migrations, err := Scan(m.scriptsDir)
	if err != nil {
		return nil, fmt.Errorf("scan migrations: %w", err)
	}
	if len(migrations) == 0 {
		return nil, nil
	}

	appliedSet, err := m.appliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	executed := make([]string, 0, len(migrations))
	for _, mig := range migrations {
		if _, ok := appliedSet[mig.Version]; ok {
			if err := m.verifyChecksum(ctx, mig); err != nil {
				return executed, err
			}
			continue
		}

		sql, err := os.ReadFile(mig.UpFile)
		if err != nil {
			return executed, fmt.Errorf("read up script %q: %w", mig.UpFile, err)
		}

		if dryRun {
			fmt.Printf("[dry-run] would execute migration %s:\n%s\n", mig.Version, string(sql))
			executed = append(executed, mig.Version)
			continue
		}

		if err := m.execInTx(ctx, mig, string(sql)); err != nil {
			return executed, err
		}
		executed = append(executed, mig.Version)
	}

	return executed, nil
}

func (m *Migrator) appliedVersions(ctx context.Context) (map[string]struct{}, error) {
	exists, err := m.history.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return map[string]struct{}{}, nil
	}

	applied, err := m.history.ListApplied(ctx)
	if err != nil {
		return nil, err
	}

	appliedSet := make(map[string]struct{}, len(applied))
	for _, rec := range applied {
		appliedSet[rec.Version] = struct{}{}
	}
	return appliedSet, nil
}

func (m *Migrator) verifyChecksum(ctx context.Context, mig *Migration) error {
	var records []MigrationRecord
	if err := m.db.WithContext(ctx).
		Where("version = ?", mig.Version).
		Find(&records).Error; err != nil || len(records) == 0 {
		return nil
	}
	rec := records[0]
	if rec.Checksum != "" && rec.Checksum != mig.Checksum {
		return fmt.Errorf(
			"migration %q checksum mismatch: file changed after being applied (recorded=%s, current=%s)",
			mig.Version, rec.Checksum, mig.Checksum,
		)
	}
	return nil
}

func (m *Migrator) execInTx(ctx context.Context, mig *Migration, sql string) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, stmt := range splitStatements(sql) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("exec migration %q statement %q: %w", mig.Version, truncate(stmt, 80), err)
			}
		}

		rec := newRecord(mig)
		if err := tx.Create(&rec).Error; err != nil {
			return fmt.Errorf("mark migration %q applied: %w", mig.Version, err)
		}
		return nil
	})
}

func (m *Migrator) Rollback(ctx context.Context, steps int, dryRun bool) ([]string, error) {
	if steps < 1 {
		steps = 1
	}

	exists, err := m.history.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	var records []MigrationRecord
	if err := m.db.WithContext(ctx).
		Order("applied_at DESC").
		Limit(steps).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("query last applied: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}

	migrations, err := Scan(m.scriptsDir)
	if err != nil {
		return nil, fmt.Errorf("scan migrations: %w", err)
	}
	migIndex := make(map[string]*Migration, len(migrations))
	for _, mig := range migrations {
		migIndex[mig.Version] = mig
	}

	rolledBack := make([]string, 0, len(records))
	for _, rec := range records {
		mig, ok := migIndex[rec.Version]
		if !ok {
			return rolledBack, fmt.Errorf("migration %q not found in scripts dir", rec.Version)
		}
		if mig.DownFile == "" {
			return rolledBack, fmt.Errorf("migration %q has no down script, cannot rollback", rec.Version)
		}

		sql, err := os.ReadFile(mig.DownFile)
		if err != nil {
			return rolledBack, fmt.Errorf("read down script %q: %w", mig.DownFile, err)
		}

		if dryRun {
			fmt.Printf("[dry-run] would rollback migration %s:\n%s\n", rec.Version, string(sql))
			rolledBack = append(rolledBack, rec.Version)
			continue
		}

		if err := m.rollbackInTx(ctx, rec.Version, string(sql)); err != nil {
			return rolledBack, err
		}
		rolledBack = append(rolledBack, rec.Version)
	}

	return rolledBack, nil
}

func (m *Migrator) rollbackInTx(ctx context.Context, version string, sql string) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, stmt := range splitStatements(sql) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("exec rollback %q stmt %q: %w", version, truncate(stmt, 80), err)
			}
		}
		return tx.Where("version = ?", version).Delete(&MigrationRecord{}).Error
	})
}

func (m *Migrator) Status(ctx context.Context) (*MigrationStatus, error) {
	applied := []MigrationRecord{}
	appliedSet := map[string]struct{}{}

	exists, err := m.history.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		applied, err = m.history.ListApplied(ctx)
		if err != nil {
			return nil, err
		}
		for _, rec := range applied {
			appliedSet[rec.Version] = struct{}{}
		}
	}

	migrations, err := Scan(m.scriptsDir)
	if err != nil {
		return nil, fmt.Errorf("scan migrations: %w", err)
	}

	pending := make([]*Migration, 0, len(migrations))
	for _, mig := range migrations {
		if _, ok := appliedSet[mig.Version]; !ok {
			pending = append(pending, mig)
		}
	}

	return &MigrationStatus{
		Applied: applied,
		Pending: pending,
	}, nil
}

func splitStatements(sql string) []string {
	var result []string
	for _, part := range strings.Split(sql, ";") {
		trimmed := trimComments(strings.TrimSpace(part))
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func trimComments(sql string) string {
	var lines []string
	for _, line := range strings.Split(sql, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "--") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
