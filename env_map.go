package envy

import (
	"sync"
)

// envMap wraps sync.Map and uses the following types:
// key:   string
// value: string
type envMap struct {
	data *sync.Map
	init sync.Once
}

func (m *envMap) Data() *sync.Map {
	(&m.init).Do(func() {
		if m.data == nil {
			m.data = &sync.Map{}
		}
	})
	return m.data
}

// Delete the key from the map
func (m *envMap) Delete(key string) {
	m.Data().Delete(key)
}

// Load the key from the map.
// Returns string or bool.
// A false return indicates either the key was not found
// or the value is not of type string
func (m *envMap) Load(key string) (string, bool) {
	i, ok := m.Data().Load(key)
	if !ok {
		return "", false
	}
	s, ok := i.(string)
	return s, ok
}

// LoadOrStore will return an existing key or
// store the value if not already in the map
func (m *envMap) LoadOrStore(key string, value string) (string, bool) {
	i, _ := m.Data().LoadOrStore(key, value)
	s, ok := i.(string)
	return s, ok
}

// Range over the string values in the map
func (m *envMap) Range(f func(key string, value string) bool) {
	m.Data().Range(func(k, v interface{}) bool {
		key, ok := k.(string)
		if !ok {
			return false
		}
		value, ok := v.(string)
		if !ok {
			return false
		}
		return f(key, value)
	})
}

// Store a string in the map
func (m *envMap) Store(key string, value string) {
	m.Data().Store(key, value)
}
