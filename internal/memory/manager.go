package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultLongTermFile = "long_term.jsonl"
	personalityFile     = "personality.json"
	userProfileFile     = "user_profile.json"

	maxShortTermMemories = 100
	maxIndexKeywordLen   = 80
)

type Manager struct {
	Dir        string
	shortTerm  []Memory
	sessionID  string
	mu         sync.Mutex
	now        func() time.Time
}

func DefaultDir(root string) string {
	return filepath.Join(root, ".memory")
}

func NewManager(dir string) (*Manager, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("memory dir required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	sessionDir := filepath.Join(dir, "sessions")
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return nil, err
	}
	return &Manager{
		Dir:       dir,
		shortTerm: make([]Memory, 0, maxShortTermMemories),
		now:       time.Now,
	}, nil
}

func (m *Manager) SetSessionID(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessionID = id
}

func (m *Manager) Store(input StoreInput) (Memory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	memory, err := m.buildMemory(input)
	if err != nil {
		return Memory{}, err
	}
	switch memory.Type {
	case TypeShortTerm:
		return m.storeShortTermLocked(memory), nil
	case TypeLongTerm:
		return m.storeLongTermLocked(memory)
	case TypeSessionSummary:
		return m.storeSessionSummaryLocked(memory)
	case TypeAIPersonality:
		return m.storePersonalityLocked(memory)
	case TypeUserProfile:
		return m.storeProfileLocked(memory)
	default:
		return Memory{}, fmt.Errorf("unsupported memory type %q", memory.Type)
	}
}

func (m *Manager) Recall(input QueryInput) ([]Memory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var candidates []Memory
	var err error

	switch input.Type {
	case TypeShortTerm:
		candidates = m.shortTerm
	case TypeLongTerm:
		candidates, err = m.readLongTermLocked()
	case TypeSessionSummary:
		candidates, err = m.readSessionSummariesLocked(input.SessionID)
	case TypeAIPersonality:
		candidates, err = m.readPersonalityLocked()
	case TypeUserProfile:
		candidates, err = m.readProfileLocked()
	default:
		candidates, err = m.readAllLocked()
		candidates = append(candidates, m.shortTerm...)
	}
	if err != nil {
		return nil, err
	}

	candidates = m.filterByQuery(candidates, input.Query, input.Tags, input.MinImportance)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Importance == candidates[j].Importance {
			return candidates[i].UpdatedAt.After(candidates[j].UpdatedAt)
		}
		return candidates[i].Importance > candidates[j].Importance
	})
	if input.Limit > 0 && len(candidates) > input.Limit {
		candidates = candidates[:input.Limit]
	}
	for i := range candidates {
		candidates[i].AccessCount++
	}
	return candidates, nil
}

func (m *Manager) Delete(id, memType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch memType {
	case TypeShortTerm:
		return m.deleteShortTermLocked(id)
	case TypeLongTerm:
		return m.deleteLongTermLocked(id)
	case TypeSessionSummary:
		return m.deleteSessionSummaryLocked(id)
	case TypeAIPersonality:
		return m.deletePersonalityLocked()
	case TypeUserProfile:
		return m.deleteProfileLocked()
	default:
		return fmt.Errorf("unsupported memory type %q", memType)
	}
}

func (m *Manager) Update(id string, content string, tags []string, importance float64) (Memory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mem, memType, err := m.findByIDLocked(id)
	if err != nil {
		return Memory{}, err
	}
	if content != "" {
		mem.Content = content
	}
	if tags != nil {
		mem.Tags = tags
	}
	if importance >= 0 && importance <= 1 {
		mem.Importance = importance
	}
	mem.UpdatedAt = m.now().UTC()

	switch memType {
	case TypeShortTerm:
		for i, existing := range m.shortTerm {
			if existing.ID == id {
				m.shortTerm[i] = mem
				break
			}
		}
		return mem, nil
	case TypeLongTerm:
		return m.updateLongTermLocked(mem)
	case TypeAIPersonality:
		return m.storePersonalityLocked(mem)
	case TypeUserProfile:
		return m.storeProfileLocked(mem)
	case TypeSessionSummary:
		return m.updateSessionSummaryLocked(mem)
	default:
		return Memory{}, fmt.Errorf("unsupported memory type %q", memType)
	}
}

func (m *Manager) GenerateSessionContext() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var builder strings.Builder

	mems, err := m.readPersonalityLocked()
	if err == nil && len(mems) > 0 {
		builder.WriteString("<ai_personality>\n")
		builder.WriteString(mems[0].Content)
		builder.WriteString("\n</ai_personality>\n")
	}

	mems, err = m.readProfileLocked()
	if err == nil && len(mems) > 0 {
		builder.WriteString("<user_profile>\n")
		builder.WriteString(mems[0].Content)
		builder.WriteString("\n</user_profile>\n")
	}

	longTerm, err := m.readLongTermLocked()
	if err == nil && len(longTerm) > 0 {
		sort.Slice(longTerm, func(i, j int) bool {
			return longTerm[i].Importance > longTerm[j].Importance
		})
		limit := 10
		if len(longTerm) < limit {
			limit = len(longTerm)
		}
		builder.WriteString("<long_term_memories>\n")
		for _, mem := range longTerm[:limit] {
			builder.WriteString(fmt.Sprintf("- [%s] %s\n", strings.Join(mem.Tags, ", "), mem.Content))
		}
		builder.WriteString("</long_term_memories>\n")
	}

	mems, err = m.readSessionSummariesLocked("")
	if err == nil && len(mems) > 0 {
		sort.Slice(mems, func(i, j int) bool {
			return mems[i].CreatedAt.After(mems[j].CreatedAt)
		})
		limit := 3
		if len(mems) < limit {
			limit = len(mems)
		}
		builder.WriteString("<previous_session_summaries>\n")
		for _, mem := range mems[:limit] {
			builder.WriteString(fmt.Sprintf("- [%s]\n%s\n", mem.CreatedAt.Format("2006-01-02"), mem.Content))
		}
		builder.WriteString("</previous_session_summaries>\n")
	}

	return builder.String()
}

func (m *Manager) buildMemory(input StoreInput) (Memory, error) {
	input.Type = strings.ToLower(strings.TrimSpace(input.Type))
	if input.Type == "" {
		return Memory{}, fmt.Errorf("memory type required")
	}
	switch input.Type {
	case TypeShortTerm, TypeLongTerm, TypeSessionSummary, TypeAIPersonality, TypeUserProfile:
	default:
		return Memory{}, fmt.Errorf("unsupported memory type %q", input.Type)
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return Memory{}, fmt.Errorf("content required")
	}

	now := m.now().UTC()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = fmt.Sprintf("mem-%d", now.UnixNano())
	}

	importance := input.Importance
	if importance < 0 {
		importance = 0.5
	}
	if importance > 1 {
		importance = 1
	}

	tags := make([]string, 0, len(input.Tags))
	for _, tag := range input.Tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	sessionID := strings.TrimSpace(input.SessionID)
	if sessionID == "" {
		sessionID = m.sessionID
	}

	return Memory{
		ID:         id,
		Type:       input.Type,
		Content:    content,
		Tags:       tags,
		Importance: importance,
		SessionID:  sessionID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (m *Manager) storeShortTermLocked(memory Memory) Memory {
	if len(m.shortTerm) >= maxShortTermMemories {
		m.shortTerm = m.shortTerm[1:]
	}
	m.shortTerm = append(m.shortTerm, memory)
	return memory
}

func (m *Manager) storeLongTermLocked(memory Memory) (Memory, error) {
	path := filepath.Join(m.Dir, defaultLongTermFile)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return Memory{}, err
	}
	defer f.Close()

	mems, err := m.readLongTermLocked()
	if err != nil {
		return Memory{}, err
	}
	for i, existing := range mems {
		if existing.ID == memory.ID {
			mems[i] = memory
			if err := m.writeLongTermLocked(mems); err != nil {
				return Memory{}, err
			}
			return memory, nil
		}
	}

	data, err := json.Marshal(memory)
	if err != nil {
		return Memory{}, err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return Memory{}, err
	}
	return memory, nil
}

func (m *Manager) storeSessionSummaryLocked(memory Memory) (Memory, error) {
	if memory.SessionID == "" {
		memory.SessionID = memory.ID
	}
	path := filepath.Join(m.Dir, "sessions", fmt.Sprintf("session_%s.json", memory.SessionID))
	data, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return Memory{}, err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return Memory{}, err
	}
	return memory, nil
}

func (m *Manager) storePersonalityLocked(memory Memory) (Memory, error) {
	path := filepath.Join(m.Dir, personalityFile)
	d, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return Memory{}, err
	}
	if err := os.WriteFile(path, append(d, '\n'), 0o644); err != nil {
		return Memory{}, err
	}
	return memory, nil
}

func (m *Manager) storeProfileLocked(memory Memory) (Memory, error) {
	path := filepath.Join(m.Dir, userProfileFile)
	d, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return Memory{}, err
	}
	if err := os.WriteFile(path, append(d, '\n'), 0o644); err != nil {
		return Memory{}, err
	}
	return memory, nil
}

func (m *Manager) readLongTermLocked() ([]Memory, error) {
	path := filepath.Join(m.Dir, defaultLongTermFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	mems := make([]Memory, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var mem Memory
		if err := json.Unmarshal([]byte(line), &mem); err != nil {
			continue
		}
		mems = append(mems, mem)
	}
	return mems, nil
}

func (m *Manager) writeLongTermLocked(mems []Memory) error {
	path := filepath.Join(m.Dir, defaultLongTermFile)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, mem := range mems {
		d, err := json.Marshal(mem)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(d, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) updateLongTermLocked(memory Memory) (Memory, error) {
	mems, err := m.readLongTermLocked()
	if err != nil {
		return Memory{}, err
	}
	found := false
	for i, existing := range mems {
		if existing.ID == memory.ID {
			mems[i] = memory
			found = true
			break
		}
	}
	if !found {
		return Memory{}, fmt.Errorf("memory %s not found", memory.ID)
	}
	if err := m.writeLongTermLocked(mems); err != nil {
		return Memory{}, err
	}
	return memory, nil
}

func (m *Manager) readSessionSummariesLocked(sessionID string) ([]Memory, error) {
	sessionDir := filepath.Join(m.Dir, "sessions")
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	mems := make([]Memory, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "session_") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if sessionID != "" {
			expected := fmt.Sprintf("session_%s.json", sessionID)
			if entry.Name() != expected {
				continue
			}
		}
		data, err := os.ReadFile(filepath.Join(sessionDir, entry.Name()))
		if err != nil {
			continue
		}
		var mem Memory
		if err := json.Unmarshal(data, &mem); err != nil {
			continue
		}
		mems = append(mems, mem)
	}
	return mems, nil
}

func (m *Manager) updateSessionSummaryLocked(memory Memory) (Memory, error) {
	return m.storeSessionSummaryLocked(memory)
}

func (m *Manager) readPersonalityLocked() ([]Memory, error) {
	path := filepath.Join(m.Dir, personalityFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var mem Memory
	if err := json.Unmarshal(data, &mem); err != nil {
		return nil, err
	}
	return []Memory{mem}, nil
}

func (m *Manager) readProfileLocked() ([]Memory, error) {
	path := filepath.Join(m.Dir, userProfileFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var mem Memory
	if err := json.Unmarshal(data, &mem); err != nil {
		return nil, err
	}
	return []Memory{mem}, nil
}

func (m *Manager) readAllLocked() ([]Memory, error) {
	var all []Memory

	longTerm, err := m.readLongTermLocked()
	if err == nil {
		all = append(all, longTerm...)
	}

	summaries, err := m.readSessionSummariesLocked("")
	if err == nil {
		all = append(all, summaries...)
	}

	personality, err := m.readPersonalityLocked()
	if err == nil {
		all = append(all, personality...)
	}

	profile, err := m.readProfileLocked()
	if err == nil {
		all = append(all, profile...)
	}

	return all, nil
}

func (m *Manager) filterByQuery(mems []Memory, query string, tags []string, minImportance float64) []Memory {
	filtered := make([]Memory, 0, len(mems))
	query = strings.ToLower(strings.TrimSpace(query))

	for _, mem := range mems {
		if minImportance > 0 && mem.Importance < minImportance {
			continue
		}
		if len(tags) > 0 {
			hasTag := false
			for _, requiredTag := range tags {
				for _, memTag := range mem.Tags {
					if memTag == requiredTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		if query != "" {
			lowerContent := strings.ToLower(mem.Content)
			if !strings.Contains(lowerContent, query) {
				tagMatch := false
				for _, tag := range mem.Tags {
					if strings.Contains(tag, query) {
						tagMatch = true
						break
					}
				}
				if !tagMatch {
					continue
				}
			}
		}
		filtered = append(filtered, mem)
	}
	return filtered
}

func (m *Manager) findByIDLocked(id string) (Memory, string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Memory{}, "", fmt.Errorf("id required")
	}

	for _, mem := range m.shortTerm {
		if mem.ID == id {
			return mem, TypeShortTerm, nil
		}
	}

	longTerm, err := m.readLongTermLocked()
	if err == nil {
		for _, mem := range longTerm {
			if mem.ID == id {
				return mem, TypeLongTerm, nil
			}
		}
	}

	personality, err := m.readPersonalityLocked()
	if err == nil && len(personality) > 0 && personality[0].ID == id {
		return personality[0], TypeAIPersonality, nil
	}

	profile, err := m.readProfileLocked()
	if err == nil && len(profile) > 0 && profile[0].ID == id {
		return profile[0], TypeUserProfile, nil
	}

	summaries, err := m.readSessionSummariesLocked("")
	if err == nil {
		for _, mem := range summaries {
			if mem.ID == id {
				return mem, TypeSessionSummary, nil
			}
		}
	}

	return Memory{}, "", fmt.Errorf("memory %s not found", id)
}

func (m *Manager) deleteShortTermLocked(id string) error {
	for i, mem := range m.shortTerm {
		if mem.ID == id {
			m.shortTerm = append(m.shortTerm[:i], m.shortTerm[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("short term memory %s not found", id)
}

func (m *Manager) deleteLongTermLocked(id string) error {
	mems, err := m.readLongTermLocked()
	if err != nil {
		return err
	}
	found := false
	for i, mem := range mems {
		if mem.ID == id {
			mems = append(mems[:i], mems[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("long term memory %s not found", id)
	}
	return m.writeLongTermLocked(mems)
}

func (m *Manager) deleteSessionSummaryLocked(sessionID string) error {
	path := filepath.Join(m.Dir, "sessions", fmt.Sprintf("session_%s.json", sessionID))
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session summary %s not found", sessionID)
		}
		return err
	}
	return nil
}

func (m *Manager) deletePersonalityLocked() error {
	path := filepath.Join(m.Dir, personalityFile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (m *Manager) deleteProfileLocked() error {
	path := filepath.Join(m.Dir, userProfileFile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
