package main

import (
	"flag"
	"fmt"
	"log"

	"gitlab-merge-alert-go/internal/config"
	"gitlab-merge-alert-go/internal/database"
)

func main() {
	// 命令行参数
	var (
		status   = flag.Bool("status", false, "显示迁移状态")
		rollback = flag.Bool("rollback", false, "回滚最后一个迁移")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 根据参数执行不同操作
	switch {
	case *status:
		if err := database.MigrationsStatus(db); err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
		}
	case *rollback:
		if err := database.RollbackMigration(db); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
	default:
		// 默认执行迁移
		fmt.Println("Running database migrations...")
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully!")
	}
}
