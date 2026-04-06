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
	// 重试连接数据库，最多尝试 30 次，每次间隔 1 秒
	for i := 0; i < 30; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			log.Println("Database connected successfully")
			return nil
		}
		log.Printf("Database connection attempt %d/30 failed: %v", i+1, err)
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("failed to connect database after 30 retries: %w", err)
}

func GetDB() *gorm.DB {
	return DB
}
