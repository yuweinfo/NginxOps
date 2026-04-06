package database

import (
	"database/sql"
	"fmt"
	"log"
	"nginxops/internal/config"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// getMigrationsPath 获取迁移文件路径
func getMigrationsPath() string {
	// 优先检查 /app/migrations (Docker 环境)
	if _, err := os.Stat("/app/migrations"); err == nil {
		return "file:///app/migrations"
	}
	// 开发环境
	return "file://migrations"
}

// RunMigrations 执行数据库迁移
func RunMigrations() error {
	cfg := config.AppConfig.Database

	// 构建数据库连接字符串
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	// 打开数据库连接
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// 创建postgres驱动实例
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// 创建迁移实例
	migrationsPath := getMigrationsPath()
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// 执行迁移
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	log.Printf("Database migration completed. Version: %d, Dirty: %v", version, dirty)

	return nil
}

// RollbackMigration 回滚最后一次迁移
func RollbackMigration() error {
	cfg := config.AppConfig.Database

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationsPath := getMigrationsPath()
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	version, dirty, _ := m.Version()
	log.Printf("Database migration rollback completed. Version: %d, Dirty: %v", version, dirty)

	return nil
}

// GetMigrationVersion 获取当前迁移版本
func GetMigrationVersion() (uint, bool, error) {
	cfg := config.AppConfig.Database

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return 0, false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationsPath := getMigrationsPath()
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m.Version()
}
