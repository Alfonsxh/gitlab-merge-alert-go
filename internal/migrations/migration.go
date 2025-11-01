package migrations

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// Migration 迁移接口
type Migration interface {
	ID() string
	Description() string
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

// MigrationRecord 迁移记录表
type MigrationRecord struct {
	ID        string    `gorm:"primaryKey;size:255"`
	AppliedAt time.Time `gorm:"not null"`
}

// TableName 指定表名
func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

// Migrator 迁移管理器
type Migrator struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrator 创建迁移管理器
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: []Migration{},
	}
}

// AddMigration 添加迁移
func (m *Migrator) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// initMigrationsTable 初始化迁移记录表
func (m *Migrator) initMigrationsTable() error {
	return m.db.AutoMigrate(&MigrationRecord{})
}

// getAppliedMigrations 获取已应用的迁移
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	var records []MigrationRecord
	err := m.db.Find(&records).Error
	if err != nil {
		return nil, err
	}

	applied := make(map[string]bool)
	for _, record := range records {
		applied[record.ID] = true
	}
	return applied, nil
}

// sortMigrations 按ID排序迁移
func (m *Migrator) sortMigrations() {
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].ID() < m.migrations[j].ID()
	})
}

// Up 执行迁移
func (m *Migrator) Up() error {
	// 初始化迁移记录表
	if err := m.initMigrationsTable(); err != nil {
		return fmt.Errorf("failed to init migrations table: %w", err)
	}

	// 获取已应用的迁移
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 排序迁移
	m.sortMigrations()

	// 执行未应用的迁移
	for _, migration := range m.migrations {
		if applied[migration.ID()] {
			continue
		}

		fmt.Printf("Running migration: %s - %s\n", migration.ID(), migration.Description())

		// 开始事务
		tx := m.db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", migration.ID(), tx.Error)
		}

		// 执行迁移
		if err := migration.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to run migration %s: %w", migration.ID(), err)
		}

		// 记录迁移
		record := MigrationRecord{
			ID:        migration.ID(),
			AppliedAt: time.Now(),
		}
		if err := tx.Create(&record).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.ID(), err)
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.ID(), err)
		}

		fmt.Printf("Migration completed: %s\n", migration.ID())
	}

	return nil
}

// Status 查看迁移状态
func (m *Migrator) Status() error {
	if err := m.initMigrationsTable(); err != nil {
		return fmt.Errorf("failed to init migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	m.sortMigrations()

	fmt.Println("Migration Status:")
	fmt.Println("================")
	for _, migration := range m.migrations {
		status := "pending"
		if applied[migration.ID()] {
			status = "applied"
		}
		fmt.Printf("%-20s %-10s %s\n", migration.ID(), status, migration.Description())
	}

	return nil
}

// Rollback 回滚最后一个迁移
func (m *Migrator) Rollback() error {
	if err := m.initMigrationsTable(); err != nil {
		return fmt.Errorf("failed to init migrations table: %w", err)
	}

	// 获取最后一个已应用的迁移
	var lastRecord MigrationRecord
	err := m.db.Order("applied_at DESC").First(&lastRecord).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	// 找到对应的迁移
	var targetMigration Migration
	for _, migration := range m.migrations {
		if migration.ID() == lastRecord.ID {
			targetMigration = migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %s not found", lastRecord.ID)
	}

	fmt.Printf("Rolling back migration: %s - %s\n", targetMigration.ID(), targetMigration.Description())

	// 开始事务
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 执行回滚
	if err := targetMigration.Down(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback migration %s: %w", targetMigration.ID(), err)
	}

	// 删除迁移记录
	if err := tx.Delete(&lastRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete migration record %s: %w", targetMigration.ID(), err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback %s: %w", targetMigration.ID(), err)
	}

	fmt.Printf("Rollback completed: %s\n", targetMigration.ID())
	return nil
}
