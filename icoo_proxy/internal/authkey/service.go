package authkey

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"

	"icoo_proxy/internal/storage"
)

type Record struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SecretMasked string `json:"secret_masked"`
	Enabled      bool   `json:"enabled"`
	Description  string `json:"description"`
	UpdatedAt    string `json:"updated_at"`
	CreatedAt    string `json:"created_at"`
}

type keyModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"index"`
	Secret      string `gorm:"uniqueIndex"`
	Enabled     bool
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (keyModel) TableName() string {
	return "auth_keys"
}

type UpsertInput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Secret      string `json:"secret"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

type Service struct {
	db *gorm.DB
}

func NewService(root string) (*Service, error) {
	db, err := storage.Open(root)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&keyModel{}); err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}

func (s *Service) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *Service) List() []Record {
	var rows []keyModel
	if err := s.db.Order("lower(name) asc").Find(&rows).Error; err != nil {
		return nil
	}
	items := make([]Record, 0, len(rows))
	for _, item := range rows {
		items = append(items, toRecord(item))
	}
	return items
}

func (s *Service) EnabledSecrets() []string {
	var rows []keyModel
	if err := s.db.Where("enabled = ?", true).Order("lower(name) asc").Find(&rows).Error; err != nil {
		return nil
	}
	items := make([]string, 0, len(rows))
	for _, item := range rows {
		if secret := strings.TrimSpace(item.Secret); secret != "" {
			items = append(items, secret)
		}
	}
	return items
}

func (s *Service) Upsert(input UpsertInput) (Record, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Record{}, fmt.Errorf("auth key name is required")
	}
	secret := strings.TrimSpace(input.Secret)

	id := strings.TrimSpace(input.ID)
	var existing keyModel
	found := false
	if id != "" {
		found = s.db.Limit(1).Find(&existing, "id = ?", id).RowsAffected > 0
	}
	if !found && secret != "" {
		found = s.db.Limit(1).Find(&existing, "secret = ?", secret).RowsAffected > 0
	}
	if !found && secret == "" {
		secret = generateSecret()
	}
	if found && secret == "" {
		secret = existing.Secret
	}
	if secret == "" {
		return Record{}, fmt.Errorf("auth key secret is required")
	}

	now := time.Now()
	current := keyModel{
		ID:          generateID(name),
		Name:        name,
		Secret:      secret,
		Enabled:     input.Enabled,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if found {
		current.ID = existing.ID
		current.CreatedAt = existing.CreatedAt
	} else if id != "" {
		current.ID = id
	}
	if err := s.db.Save(&current).Error; err != nil {
		return Record{}, err
	}
	return toRecord(current), nil
}

func (s *Service) Delete(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("auth key id is required")
	}
	result := s.db.Delete(&keyModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("auth key not found")
	}
	return nil
}

func (s *Service) GetSecret(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("auth key id is required")
	}
	var item keyModel
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("auth key not found")
		}
		return "", err
	}
	return item.Secret, nil
}

func toRecord(item keyModel) Record {
	return Record{
		ID:           item.ID,
		Name:         item.Name,
		SecretMasked: maskSecret(item.Secret),
		Enabled:      item.Enabled,
		Description:  item.Description,
		UpdatedAt:    item.UpdatedAt.Format(time.RFC3339),
		CreatedAt:    item.CreatedAt.Format(time.RFC3339),
	}
}

func maskSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 10 {
		return strings.Repeat("*", len(value))
	}
	return value[:6] + strings.Repeat("*", 6) + value[len(value)-4:]
}

func generateID(name string) string {
	base := strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == ' ':
			b.WriteRune('-')
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" {
		id = "auth-key"
	}
	return fmt.Sprintf("%s-%d", id, time.Now().UnixNano())
}

func generateSecret() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("icoo_%d", time.Now().UnixNano())
	}
	return "icoo_" + hex.EncodeToString(buf)
}

func MergeSecrets(primary string, extras []string) []string {
	values := make([]string, 0, len(extras)+1)
	for _, item := range append([]string{primary}, extras...) {
		for _, part := range strings.Split(item, ",") {
			value := strings.TrimSpace(part)
			if value != "" && !slices.Contains(values, value) {
				values = append(values, value)
			}
		}
	}
	return values
}
