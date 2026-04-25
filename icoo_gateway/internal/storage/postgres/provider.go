package postgres

import (
	"fmt"
	"strings"

	"icoo_gateway/internal/config"
)

func ValidateConfig(cfg config.Config) error {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return fmt.Errorf("database_url required when storage driver is postgres")
	}
	return fmt.Errorf("postgres storage provider not implemented yet")
}
