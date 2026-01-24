package basic_preferences

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type JSONStore struct {
	path string
	data map[string]any
	mu   sync.RWMutex
}

func NewJSONStore(appID string) *JSONStore {
	configDir, _ := os.UserConfigDir()
	storagePath := filepath.Join(configDir, "fyne", appID, "preferences.json")

	s := &JSONStore{
		path: storagePath,
		data: make(map[string]any),
	}
	s.load()
	return s
}

func (s *JSONStore) load() {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.ReadFile(s.path)
	if err == nil {
		_ = json.Unmarshal(file, &s.data)
	}
}

func (s *JSONStore) save() {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = os.MkdirAll(filepath.Dir(s.path), 0755)
	data, _ := json.MarshalIndent(s.data, "", "  ")
	_ = os.WriteFile(s.path, data, 0644)
}

func (s *JSONStore) getValue(key string, fallback any) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.data[key]; ok {
		return v
	}
	return fallback
}

func (s *JSONStore) setValue(key string, value any) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
	s.save()
}

func (s *JSONStore) BoolWithFallback(k string, fb bool) bool {
	v, ok := s.getValue(k, fb).(bool)
	if !ok {
		return fb
	}
	return v
}

func (s *JSONStore) SetBool(k string, v bool) { s.setValue(k, v) }

func (s *JSONStore) BoolListWithFallback(k string, fb []bool) []bool {
	raw := s.getValue(k, fb)
	if v, ok := raw.([]bool); ok {
		return v
	}
	return fb
}

func (s *JSONStore) SetBoolList(k string, v []bool) { s.setValue(k, v) }

func (s *JSONStore) IntWithFallback(k string, fb int) int {
	raw := s.getValue(k, fb)
	if v, ok := raw.(float64); ok {
		return int(v)
	}
	if v, ok := raw.(int); ok {
		return v
	}
	return fb
}

func (s *JSONStore) SetInt(k string, v int) { s.setValue(k, v) }

func (s *JSONStore) IntListWithFallback(k string, fb []int) []int {
	raw := s.getValue(k, fb)
	if items, ok := raw.([]any); ok {
		res := make([]int, len(items))
		for i, v := range items {
			if f, ok := v.(float64); ok {
				res[i] = int(f)
			}
		}
		return res
	}
	if v, ok := raw.([]int); ok {
		return v
	}
	return fb
}

func (s *JSONStore) SetIntList(k string, v []int) { s.setValue(k, v) }

func (s *JSONStore) FloatWithFallback(k string, fb float64) float64 {
	raw := s.getValue(k, fb)
	if v, ok := raw.(float64); ok {
		return v
	}
	return fb
}

func (s *JSONStore) SetFloat(k string, v float64) { s.setValue(k, v) }

func (s *JSONStore) FloatListWithFallback(k string, fb []float64) []float64 {
	raw := s.getValue(k, fb)
	if items, ok := raw.([]any); ok {
		res := make([]float64, len(items))
		for i, v := range items {
			if f, ok := v.(float64); ok {
				res[i] = f
			}
		}
		return res
	}
	if v, ok := raw.([]float64); ok {
		return v
	}
	return fb
}

func (s *JSONStore) SetFloatList(k string, v []float64) { s.setValue(k, v) }

func (s *JSONStore) StringWithFallback(k string, fb string) string {
	v, ok := s.getValue(k, fb).(string)
	if !ok {
		return fb
	}
	return v
}

func (s *JSONStore) SetString(k string, v string) { s.setValue(k, v) }

func (s *JSONStore) StringListWithFallback(k string, fb []string) []string {
	raw := s.getValue(k, fb)
	if items, ok := raw.([]any); ok {
		res := make([]string, len(items))
		for i, v := range items {
			if str, ok := v.(string); ok {
				res[i] = str
			}
		}
		return res
	}
	if v, ok := raw.([]string); ok {
		return v
	}
	return fb
}

func (s *JSONStore) SetStringList(k string, v []string) { s.setValue(k, v) }

type BasicPreference[T any] struct {
	key          string
	DefaultValue T
	prefs        *JSONStore
	getter       func(p *JSONStore, key string, fallback T) T
	setter       func(p *JSONStore, key string, value T)
}

func (p *BasicPreference[T]) Get() T {
	return p.getter(p.prefs, p.key, p.DefaultValue)
}

func (p *BasicPreference[T]) Set(val T) {
	p.setter(p.prefs, p.key, val)
}

func (p *BasicPreference[T]) EnsureDefault() {
	isSetKey := p.key + "_set"
	if !p.prefs.BoolWithFallback(isSetKey, false) {
		p.Set(p.DefaultValue)
		p.prefs.SetBool(isSetKey, true)
	}
}

func NewBasicPreference[T any](prefs *JSONStore, key string, defaultValue T) BasicPreference[T] {
	p := BasicPreference[T]{
		prefs:        prefs,
		key:          key,
		DefaultValue: defaultValue,
	}

	switch any(defaultValue).(type) {
	case bool:
		p.getter = func(p *JSONStore, k string, fb T) T {
			return any(p.BoolWithFallback(k, any(fb).(bool))).(T)
		}
		p.setter = func(p *JSONStore, k string, v T) {
			p.SetBool(k, any(v).(bool))
		}
	case []bool:
		p.getter = func(p *JSONStore, k string, fb T) T {
			return any(p.BoolListWithFallback(k, any(fb).([]bool))).(T)
		}
		p.setter = func(p *JSONStore, k string, v T) {
			p.SetBoolList(k, any(v).([]bool))
		}
	case int:
		p.getter = func(p *JSONStore, k string, fb T) T {
			return any(p.IntWithFallback(k, any(fb).(int))).(T)
		}
		p.setter = func(p *JSONStore, k string, v T) {
			p.SetInt(k, any(v).(int))
		}
	case []int:
		p.getter = func(p *JSONStore, k string, fb T) T {
			return any(p.IntListWithFallback(k, any(fb).([]int))).(T)
		}
		p.setter = func(p *JSONStore, k string, v T) {
			p.SetIntList(k, any(v).([]int))
		}
	case float64:
		p.getter = func(p *JSONStore, k string, fb T) T {
			return any(p.FloatWithFallback(k, any(fb).(float64))).(T)
		}
		p.setter = func(p *JSONStore, k string, v T) {
			p.SetFloat(k, any(v).(float64))
		}
	case []float64:
		p.getter = func(p *JSONStore, k string, fb T) T {
			return any(p.FloatListWithFallback(k, any(fb).([]float64))).(T)
		}
		p.setter = func(p *JSONStore, k string, v T) {
			p.SetFloatList(k, any(v).([]float64))
		}
	case string:
		p.getter = func(p *JSONStore, k string, fb T) T {
			return any(p.StringWithFallback(k, any(fb).(string))).(T)
		}
		p.setter = func(p *JSONStore, k string, v T) {
			p.SetString(k, any(v).(string))
		}
	case []string:
		p.getter = func(p *JSONStore, k string, fb T) T {
			return any(p.StringListWithFallback(k, any(fb).([]string))).(T)
		}
		p.setter = func(p *JSONStore, k string, v T) {
			p.SetStringList(k, any(v).([]string))
		}
	default:
		p.getter = func(p *JSONStore, k string, fb T) T {
			fbStr, err := json.Marshal(fb)
			if err != nil {
				panic("cannot JSON marshal fallback value")
			}
			str := p.StringWithFallback(key, string(fbStr))
			var res T
			err = json.Unmarshal([]byte(str), &res)
			if err != nil {
				return fb
			}
			return res
		}
		p.setter = func(p *JSONStore, k string, v T) {
			data, _ := json.Marshal(v)
			p.SetString(k, string(data))
		}
	}

	return p
}
