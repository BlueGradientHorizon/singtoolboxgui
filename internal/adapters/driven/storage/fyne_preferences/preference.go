package fyne_preferences

import (
	"encoding/json"

	"fyne.io/fyne/v2"
)

type FynePreference[T any] struct {
	key          string
	DefaultValue T
	prefs        fyne.Preferences
	getter       func(p fyne.Preferences, key string, fallback T) T
	setter       func(p fyne.Preferences, key string, value T)
}

func (p *FynePreference[T]) Get() T {
	return p.getter(p.prefs, p.key, p.DefaultValue)
}

func (p *FynePreference[T]) Set(val T) {
	p.setter(p.prefs, p.key, val)
}

func (p *FynePreference[T]) EnsureDefault() {
	isSetKey := p.key + "_set"
	if !p.prefs.BoolWithFallback(isSetKey, false) {
		p.Set(p.DefaultValue)
		p.prefs.SetBool(isSetKey, true)
		println("EnsureDefault", isSetKey)
	}
}

func NewFynePreference[T any](prefs fyne.Preferences, key string, defaultValue T) FynePreference[T] {
	p := FynePreference[T]{
		prefs:        prefs,
		key:          key,
		DefaultValue: defaultValue,
	}

	switch any(defaultValue).(type) {
	case bool:
		p.getter = func(p fyne.Preferences, k string, fb T) T {
			return any(p.BoolWithFallback(k, any(fb).(bool))).(T)
		}
		p.setter = func(p fyne.Preferences, k string, v T) {
			p.SetBool(k, any(v).(bool))
		}
	case []bool:
		p.getter = func(p fyne.Preferences, k string, fb T) T {
			return any(p.BoolListWithFallback(k, any(fb).([]bool))).(T)
		}
		p.setter = func(p fyne.Preferences, k string, v T) {
			p.SetBoolList(k, any(v).([]bool))
		}
	case int:
		p.getter = func(p fyne.Preferences, k string, fb T) T {
			return any(p.IntWithFallback(k, any(fb).(int))).(T)
		}
		p.setter = func(p fyne.Preferences, k string, v T) {
			p.SetInt(k, any(v).(int))
		}
	case []int:
		p.getter = func(p fyne.Preferences, k string, fb T) T {
			return any(p.IntListWithFallback(k, any(fb).([]int))).(T)
		}
		p.setter = func(p fyne.Preferences, k string, v T) {
			p.SetIntList(k, any(v).([]int))
		}
	case float64:
		p.getter = func(p fyne.Preferences, k string, fb T) T {
			return any(p.FloatWithFallback(k, any(fb).(float64))).(T)
		}
		p.setter = func(p fyne.Preferences, k string, v T) {
			p.SetFloat(k, any(v).(float64))
		}
	case []float64:
		p.getter = func(p fyne.Preferences, k string, fb T) T {
			return any(p.FloatListWithFallback(k, any(fb).([]float64))).(T)
		}
		p.setter = func(p fyne.Preferences, k string, v T) {
			p.SetFloatList(k, any(v).([]float64))
		}
	case string:
		p.getter = func(p fyne.Preferences, k string, fb T) T {
			return any(p.StringWithFallback(k, any(fb).(string))).(T)
		}
		p.setter = func(p fyne.Preferences, k string, v T) {
			p.SetString(k, any(v).(string))
		}
	case []string:
		p.getter = func(p fyne.Preferences, k string, fb T) T {
			return any(p.StringListWithFallback(k, any(fb).([]string))).(T)
		}
		p.setter = func(p fyne.Preferences, k string, v T) {
			p.SetStringList(k, any(v).([]string))
		}
	default:
		p.getter = func(p fyne.Preferences, k string, fb T) T {
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
		p.setter = func(p fyne.Preferences, k string, v T) {
			data, _ := json.Marshal(v)
			p.SetString(k, string(data))
		}
	}

	return p
}
