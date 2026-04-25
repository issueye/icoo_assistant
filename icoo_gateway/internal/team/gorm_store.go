package team

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormTeamModel struct {
	ID           string    `gorm:"primaryKey;size:64"`
	Name         string    `gorm:"size:128;index"`
	Description  string    `gorm:"type:text"`
	EntryAgentID string    `gorm:"size:128;index"`
	Status       string    `gorm:"size:64;index"`
	CreatedAt    time.Time `gorm:"index"`
	UpdatedAt    time.Time `gorm:"index"`
}

func (GormTeamModel) TableName() string {
	return "teams"
}

type GormMemberModel struct {
	ID             string    `gorm:"primaryKey;size:64"`
	TeamID         string    `gorm:"size:64;index"`
	AgentID        string    `gorm:"size:128;index"`
	Role           string    `gorm:"size:64"`
	SortOrder      int       `gorm:"index"`
	Status         string    `gorm:"size:64;index"`
	Responsibility string    `gorm:"type:text"`
	CreatedAt      time.Time `gorm:"index"`
	UpdatedAt      time.Time `gorm:"index"`
}

func (GormMemberModel) TableName() string {
	return "team_members"
}

type GormStore struct {
	db *gorm.DB
}

var _ Store = (*GormStore)(nil)

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) Create(input CreateInput) (Team, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Team{}, fmt.Errorf("name required")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}
	now := time.Now().UTC()
	record := GormTeamModel{
		ID:           fmt.Sprintf("team-%s", uuid.NewString()),
		Name:         name,
		Description:  strings.TrimSpace(input.Description),
		EntryAgentID: strings.TrimSpace(input.EntryAgentID),
		Status:       status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return Team{}, err
	}
	return modelToTeam(record), nil
}

func (s *GormStore) Get(id string) (Team, bool) {
	var record GormTeamModel
	if err := s.db.First(&record, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return Team{}, false
	}
	return modelToTeam(record), true
}

func (s *GormStore) List() []Team {
	var records []GormTeamModel
	if err := s.db.Order("created_at asc").Find(&records).Error; err != nil {
		return nil
	}
	items := make([]Team, 0, len(records))
	for _, record := range records {
		items = append(items, modelToTeam(record))
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items
}

func (s *GormStore) Update(id string, input UpdateInput) (Team, error) {
	var record GormTeamModel
	if err := s.db.First(&record, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return Team{}, fmt.Errorf("team not found")
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return Team{}, fmt.Errorf("name required")
		}
		record.Name = name
	}
	if input.Description != nil {
		record.Description = strings.TrimSpace(*input.Description)
	}
	if input.EntryAgentID != nil {
		record.EntryAgentID = strings.TrimSpace(*input.EntryAgentID)
	}
	if input.Status != nil {
		status := strings.TrimSpace(*input.Status)
		if status == "" {
			return Team{}, fmt.Errorf("status required")
		}
		record.Status = status
	}
	record.UpdatedAt = time.Now().UTC()
	if err := s.db.Save(&record).Error; err != nil {
		return Team{}, err
	}
	return modelToTeam(record), nil
}

func (s *GormStore) AddMember(teamID string, input AddMemberInput) (Member, error) {
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return Member{}, fmt.Errorf("team id required")
	}
	agentID := strings.TrimSpace(input.AgentID)
	if agentID == "" {
		return Member{}, fmt.Errorf("agent_id required")
	}
	role := strings.TrimSpace(input.Role)
	if role == "" {
		role = "member"
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}
	var teamRecord GormTeamModel
	if err := s.db.First(&teamRecord, "id = ?", teamID).Error; err != nil {
		return Member{}, fmt.Errorf("team not found")
	}
	var existing int64
	if err := s.db.Model(&GormMemberModel{}).Where("team_id = ? AND agent_id = ?", teamID, agentID).Count(&existing).Error; err != nil {
		return Member{}, err
	}
	if existing > 0 {
		return Member{}, fmt.Errorf("agent already added to team")
	}
	now := time.Now().UTC()
	record := GormMemberModel{
		ID:             fmt.Sprintf("team-member-%s", uuid.NewString()),
		TeamID:         teamID,
		AgentID:        agentID,
		Role:           role,
		SortOrder:      input.SortOrder,
		Status:         status,
		Responsibility: strings.TrimSpace(input.Responsibility),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return Member{}, err
	}
	teamRecord.UpdatedAt = now
	_ = s.db.Save(&teamRecord).Error
	return modelToMember(record), nil
}

func (s *GormStore) ListMembers(teamID string) ([]Member, bool) {
	if _, ok := s.Get(teamID); !ok {
		return nil, false
	}
	var records []GormMemberModel
	if err := s.db.Where("team_id = ?", strings.TrimSpace(teamID)).Order("sort_order asc, created_at asc").Find(&records).Error; err != nil {
		return nil, true
	}
	items := make([]Member, 0, len(records))
	for _, record := range records {
		items = append(items, modelToMember(record))
	}
	return items, true
}

func (s *GormStore) UpdateMember(teamID, memberID string, input UpdateMemberInput) (Member, error) {
	teamID = strings.TrimSpace(teamID)
	memberID = strings.TrimSpace(memberID)
	if _, ok := s.Get(teamID); !ok {
		return Member{}, fmt.Errorf("team not found")
	}
	var record GormMemberModel
	if err := s.db.First(&record, "id = ? AND team_id = ?", memberID, teamID).Error; err != nil {
		return Member{}, fmt.Errorf("team member not found")
	}
	if input.Role != nil {
		role := strings.TrimSpace(*input.Role)
		if role == "" {
			return Member{}, fmt.Errorf("role required")
		}
		record.Role = role
	}
	if input.SortOrder != nil {
		record.SortOrder = *input.SortOrder
	}
	if input.Status != nil {
		status := strings.TrimSpace(*input.Status)
		if status == "" {
			return Member{}, fmt.Errorf("status required")
		}
		record.Status = status
	}
	if input.Responsibility != nil {
		record.Responsibility = strings.TrimSpace(*input.Responsibility)
	}
	record.UpdatedAt = time.Now().UTC()
	if err := s.db.Save(&record).Error; err != nil {
		return Member{}, err
	}
	return modelToMember(record), nil
}

func (s *GormStore) DeleteMember(teamID, memberID string) (Member, error) {
	teamID = strings.TrimSpace(teamID)
	memberID = strings.TrimSpace(memberID)
	if _, ok := s.Get(teamID); !ok {
		return Member{}, fmt.Errorf("team not found")
	}
	var record GormMemberModel
	if err := s.db.First(&record, "id = ? AND team_id = ?", memberID, teamID).Error; err != nil {
		return Member{}, fmt.Errorf("team member not found")
	}
	if err := s.db.Delete(&record).Error; err != nil {
		return Member{}, err
	}
	return modelToMember(record), nil
}

func (s *GormStore) HasMember(teamID, agentID string) bool {
	var count int64
	err := s.db.Model(&GormMemberModel{}).
		Where("team_id = ? AND agent_id = ? AND status = ?", strings.TrimSpace(teamID), strings.TrimSpace(agentID), "active").
		Count(&count).Error
	return err == nil && count > 0
}

func modelToTeam(record GormTeamModel) Team {
	return Team{
		ID:           record.ID,
		Name:         record.Name,
		Description:  record.Description,
		EntryAgentID: record.EntryAgentID,
		Status:       record.Status,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
}

func modelToMember(record GormMemberModel) Member {
	return Member{
		ID:             record.ID,
		TeamID:         record.TeamID,
		AgentID:        record.AgentID,
		Role:           record.Role,
		SortOrder:      record.SortOrder,
		Status:         record.Status,
		Responsibility: record.Responsibility,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
}
