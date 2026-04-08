package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Config 描述数据库连接配置。
type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Database 提供线程安全的 GORM 数据库封装。
type Database struct {
	mu    sync.RWMutex
	cfg   Config
	db    *gorm.DB
	sqlDB *sql.DB
}

// New 创建一个新的数据库封装实例。
func New(cfg Config) (*Database, error) {
	db, sqlDB, normalized, err := open(cfg)
	if err != nil {
		return nil, err
	}

	return &Database{
		cfg:   normalized,
		db:    db,
		sqlDB: sqlDB,
	}, nil
}

// DB 返回当前 GORM 数据库实例。
func (d *Database) DB() *gorm.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.db
}

// Ping 对当前数据库连接执行存活检查。
func (d *Database) Ping(ctx context.Context) error {
	d.mu.RLock()
	sqlDB := d.sqlDB
	d.mu.RUnlock()

	if sqlDB == nil {
		return errors.New("database is closed")
	}

	return sqlDB.PingContext(ctx)
}

// Close 关闭数据库连接。
func (d *Database) Close() error {
	d.mu.Lock()
	sqlDB := d.sqlDB
	d.sqlDB = nil
	d.db = nil
	d.mu.Unlock()

	if sqlDB == nil {
		return nil
	}

	return sqlDB.Close()
}

// Reload 使用新配置建立连接并原子替换旧实例。
func (d *Database) Reload(cfg Config) error {
	nextDB, nextSQLDB, normalized, err := open(cfg)
	if err != nil {
		return err
	}

	d.mu.Lock()
	oldSQLDB := d.sqlDB
	d.cfg = normalized
	d.db = nextDB
	d.sqlDB = nextSQLDB
	d.mu.Unlock()

	if oldSQLDB != nil {
		return oldSQLDB.Close()
	}

	return nil
}

func open(cfg Config) (*gorm.DB, *sql.DB, Config, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, nil, Config{}, err
	}

	dialector, err := dialectorFor(normalized)
	if err != nil {
		return nil, nil, Config{}, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, nil, Config{}, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, Config{}, err
	}

	sqlDB.SetMaxOpenConns(normalized.MaxOpenConns)
	sqlDB.SetMaxIdleConns(normalized.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(normalized.ConnMaxLifetime)

	return db, sqlDB, normalized, nil
}

func normalizeConfig(cfg Config) (Config, error) {
	if cfg.Driver == "" {
		return Config{}, errors.New("database driver is required")
	}
	if cfg.DSN == "" {
		return Config{}, errors.New("database dsn is required")
	}
	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns < 0 {
		cfg.MaxIdleConns = 0
	}
	return cfg, nil
}

func dialectorFor(cfg Config) (gorm.Dialector, error) {
	switch cfg.Driver {
	case "sqlite":
		return sqlite.Open(cfg.DSN), nil
	case "mysql":
		return mysql.Open(cfg.DSN), nil
	case "postgres":
		return postgres.Open(cfg.DSN), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}
}
