package dbtx

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type txRecord struct {
	ID   int64 `gorm:"primaryKey"`
	Name string
}

func TestManagerWithTxCommits(t *testing.T) {
	t.Parallel()

	db := openTxTestDB(t, t.Name())
	manager, err := New(db)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := manager.WithTx(context.Background(), func(ctx context.Context, tx *gorm.DB) error {
		if manager.GetDB(ctx) != tx {
			t.Fatal("GetDB() did not return transactional db")
		}
		return tx.Create(&txRecord{Name: "created"}).Error
	}); err != nil {
		t.Fatalf("WithTx() returned error: %v", err)
	}

	var count int64
	if err := db.Model(&txRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("Count() returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func TestManagerWithTxRollsBackOnError(t *testing.T) {
	t.Parallel()

	db := openTxTestDB(t, t.Name())
	manager, err := New(db)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	err = manager.WithTx(context.Background(), func(_ context.Context, tx *gorm.DB) error {
		if err := tx.Create(&txRecord{Name: "rolled-back"}).Error; err != nil {
			return err
		}
		return errors.New("stop")
	})
	if err == nil {
		t.Fatal("WithTx() returned nil error")
	}

	var count int64
	if err := db.Model(&txRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("Count() returned error: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func openTxTestDB(t *testing.T, name string) *gorm.DB {
	t.Helper()

	dsn := "file:" + strings.ReplaceAll(name, "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() returned error: %v", err)
	}

	if err := db.AutoMigrate(&txRecord{}); err != nil {
		t.Fatalf("AutoMigrate() returned error: %v", err)
	}

	return db
}
