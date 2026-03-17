package config_registry

import (
	"fmt"
	"sort"
	"strings"
)

type Registry struct {
	entries map[string]Entry
}

func NewRegistry() *Registry {
	return &Registry{entries: map[string]Entry{}}
}

func (r *Registry) Register(entry Entry) error {
	key := strings.TrimSpace(entry.Key)
	if key == "" {
		return fmt.Errorf("governance registry key is required")
	}
	if entry.Kind == "" {
		return fmt.Errorf("governance registry kind is required")
	}
	if entry.ValueType == "" {
		return fmt.Errorf("governance registry value type is required")
	}
	if len(entry.Scopes) == 0 {
		return fmt.Errorf("governance registry scopes are required")
	}
	if _, exists := r.entries[key]; exists {
		return fmt.Errorf("governance registry key already registered: %s", key)
	}
	entry.Key = key
	r.entries[key] = entry
	return nil
}

func (r *Registry) MustRegister(entry Entry) {
	if err := r.Register(entry); err != nil {
		panic(err)
	}
}

func (r *Registry) Get(key string) (Entry, bool) {
	entry, ok := r.entries[strings.TrimSpace(key)]
	return entry, ok
}

func (r *Registry) List() []Entry {
	out := make([]Entry, 0, len(r.entries))
	for _, entry := range r.entries {
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
