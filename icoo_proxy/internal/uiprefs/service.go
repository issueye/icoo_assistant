package uiprefs

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"icoo_proxy/internal/storage"
)

const defaultKey = "main"

type Preferences struct {
	Theme      string `json:"theme"`
	ButtonSize string `json:"buttonSize"`
}

type uiPrefModel struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

func (uiPrefModel) TableName() string {
	return "ui_prefs"
}

type Service struct {
	db *gorm.DB
}

func NewService(root string) (*Service, error) {
	db, err := storage.Open(root)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&uiPrefModel{}); err != nil {
		return nil, err
	}
	s := &Service{db: db}
	if _, err := s.Get(); err != nil {
		defaults := Preferences{Theme: "blue", ButtonSize: "md"}
		if saveErr := s.Save(defaults); saveErr != nil {
			return nil, saveErr
		}
	}
	return s, nil
}

func (s *Service) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *Service) Get() (Preferences, error) {
	var row uiPrefModel
	err := s.db.Limit(1).Find(&row, "key = ?", defaultKey).Error
	if err != nil {
		return Preferences{}, fmt.Errorf("load ui preferences: %w", err)
	}
	if row.Key == "" {
		return Preferences{}, fmt.Errorf("no ui preferences found")
	}
	var prefs Preferences
	if err := json.Unmarshal([]byte(row.Value), &prefs); err != nil {
		return Preferences{}, fmt.Errorf("parse ui preferences: %w", err)
	}
	return prefs, nil
}

func (s *Service) Save(input Preferences) error {
	theme := strings.TrimSpace(input.Theme)
	if theme == "" {
		theme = "blue"
	}
	buttonSize := strings.TrimSpace(input.ButtonSize)
	if buttonSize == "" {
		buttonSize = "md"
	}
	normalized := Preferences{Theme: theme, ButtonSize: buttonSize}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("serialize ui preferences: %w", err)
	}
	row := uiPrefModel{
		Key:   defaultKey,
		Value: string(raw),
	}
	return s.db.Save(&row).Error
}
