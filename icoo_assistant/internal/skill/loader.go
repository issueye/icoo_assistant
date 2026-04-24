package skill

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Entry struct {
	Name        string
	Description string
	Body        string
	Path        string
}

type Loader struct {
	entries map[string]Entry
}

func Load(skillsDir string) (*Loader, error) {
	loader := &Loader{entries: map[string]Entry{}}
	if skillsDir == "" {
		return loader, nil
	}
	if _, err := os.Stat(skillsDir); err != nil {
		if os.IsNotExist(err) {
			return loader, nil
		}
		return nil, err
	}
	paths, err := filepath.Glob(filepath.Join(skillsDir, "*", "SKILL.md"))
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		entry := parseSkill(path, string(data))
		loader.entries[entry.Name] = entry
	}
	return loader, nil
}

func parseSkill(path, content string) Entry {
	name := filepath.Base(filepath.Dir(path))
	description := "No description"
	body := strings.TrimSpace(content)
	re := regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)
	matches := re.FindStringSubmatch(content)
	if len(matches) == 3 {
		for _, line := range strings.Split(matches[1], "\n") {
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			switch key {
			case "name":
				if value != "" {
					name = value
				}
			case "description":
				if value != "" {
					description = value
				}
			}
		}
		body = strings.TrimSpace(matches[2])
	}
	return Entry{Name: name, Description: description, Body: body, Path: path}
}

func (l *Loader) Descriptions() string {
	if l == nil || len(l.entries) == 0 {
		return "(no skills available)"
	}
	names := make([]string, 0, len(l.entries))
	for name := range l.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	lines := make([]string, 0, len(names))
	for _, name := range names {
		entry := l.entries[name]
		lines = append(lines, "  - "+entry.Name+": "+entry.Description)
	}
	return strings.Join(lines, "\n")
}

func (l *Loader) Load(name string) string {
	if l == nil {
		return "Error: Unknown skill '" + name + "'."
	}
	entry, ok := l.entries[name]
	if !ok {
		return "Error: Unknown skill '" + name + "'."
	}
	return "<skill name=\"" + entry.Name + "\">\n" + entry.Body + "\n</skill>"
}
