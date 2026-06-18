package database
package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shandialamp/dedenne/config"
	"go.uber.org/zap"
)

// NewDB 初始化数据库连接
func NewDB(cfg *config.DatabaseConfig, logger *zap.Logger) (*sql.DB, error) {
	var driverName string
	var dsn string

	switch cfg.Type {
	case "sqlite":
		driverName = "sqlite3"
		dsn = cfg.DSN
		logger.Info("Connecting to SQLite database", zap.String("dsn", dsn))

	case "mysql":
		driverName = "mysql"
		dsn = cfg.DSN
		logger.Info("Connecting to MySQL database", zap.String("host", "***"))

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.MaxOpenConn)
	db.SetMaxIdleConns(cfg.MaxIdleConn)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// 验证连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully",
		zap.String("type", cfg.Type),
		zap.Int("maxOpenConn", cfg.MaxOpenConn),
		zap.Int("maxIdleConn", cfg.MaxIdleConn),
		zap.Int("maxLifetime", cfg.MaxLifetime),
	)

	return db, nil
}

// Close 关闭数据库连接
func Close(db *sql.DB, logger *zap.Logger) error {
	if db == nil {
		return nil
	}
	if err := db.Close(); err != nil {
		logger.Error("Failed to close database", zap.Error(err))
		return err
	}
	logger.Info("Database connection closed")
	return nil
}
