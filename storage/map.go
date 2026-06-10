package storage

import "sync"

type MapStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func (m *MapStorage) Set(key string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	return nil
}

func (m *MapStorage) Get(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.data[key]
	if !ok {
		return "", &NotFoundError{}
	}
	return value, nil
}

func (m *MapStorage) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		data: make(map[string]string),
	}
}
