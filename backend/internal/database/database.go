package database

import (
	"fmt"
	"log"
	"nginxops/internal/config"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	cfg := config.AppConfig.Database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// InitDBWithRetry 带重试的数据库初始化
func InitDBWithRetry(maxRetries int, retryInterval time.Duration) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		err := InitDB()
		if err == nil {
			return nil
		}
		lastErr = err
		log.Printf("Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}
	return fmt.Errorf("failed to connect database after %d retries: %w", maxRetries, lastErr)
}

func GetDB() *gorm.DB {
	return DB
}

// IsConnected 检查数据库是否已连接
func IsConnected() bool {
	return DB != nil
}
