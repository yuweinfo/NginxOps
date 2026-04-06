package main

import (
	"flag"
	"fmt"
	"log"
	"nginxops/internal/config"
	"nginxops/internal/database"
	"os"
)

func main() {
	// 定义命令行参数
	action := flag.String("action", "up", "Migration action: up, down, version, force")
	forceVersion := flag.Int("force-version", 0, "Force version for force action")
	flag.Parse()

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 执行相应操作
	switch *action {
	case "up":
		fmt.Println("Running migrations...")
		if err := database.RunMigrations(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✅ Migrations completed successfully")

	case "down":
		fmt.Println("Rolling back last migration...")
		if err := database.RollbackMigration(); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("✅ Rollback completed successfully")

	case "version":
		version, dirty, err := database.GetMigrationVersion()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("Current migration version: %d, Dirty: %v\n", version, dirty)

	case "force":
		if *forceVersion == 0 {
			fmt.Println("❌ Error: force-version is required for force action")
			os.Exit(1)
		}
		fmt.Printf("Forcing migration version to %d...\n", *forceVersion)
		// 需要实现force函数
		fmt.Println("⚠️  Force action not implemented yet")

	default:
		fmt.Printf("❌ Unknown action: %s\n", *action)
		fmt.Println("Available actions: up, down, version, force")
		os.Exit(1)
	}
}
