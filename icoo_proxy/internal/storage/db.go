package storage

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Open(root string) (*gorm.DB, error) {
	storeDir := filepath.Join(root, ".data")
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return nil, err
	}
	return gorm.Open(sqlite.Open(filepath.Join(storeDir, "icoo_proxy.db")), &gorm.Config{})
}
