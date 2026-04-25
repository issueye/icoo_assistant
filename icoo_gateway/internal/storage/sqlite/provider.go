package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sqlitegorm "github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"icoo_gateway/internal/audit"
	"icoo_gateway/internal/config"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	path := strings.TrimSpace(cfg.SQLitePath)
	if path == "" {
		return nil, fmt.Errorf("sqlite_path required when storage driver is sqlite")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlitegorm.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&audit.GormModel{}); err != nil {
		return nil, err
	}
	return db, nil
}
