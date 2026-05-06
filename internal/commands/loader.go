package commands

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Loader struct {
	entries map[string]string
}

func Load(dir string) (*Loader, error) {
	loader := &Loader{entries: map[string]string{}}
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return loader, nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return loader, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return loader, nil
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	for _, item := range paths {
		data, err := os.ReadFile(item)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(filepath.Base(item), filepath.Ext(item))
		loader.entries[strings.ToLower(strings.TrimSpace(name))] = strings.TrimSpace(string(data))
	}
	return loader, nil
}

func (l *Loader) Has(name string) bool {
	if l == nil {
		return false
	}
	_, ok := l.entries[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

func (l *Loader) Render(name, args string) string {
	if l == nil {
		return ""
	}
	body, ok := l.entries[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return ""
	}
	args = strings.TrimSpace(args)
	if args == "" {
		return body
	}
	return body + "\n\nArguments:\n" + args
}

func (l *Loader) Names() []string {
	if l == nil || len(l.entries) == 0 {
		return nil
	}
	names := make([]string, 0, len(l.entries))
	for name := range l.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
