package storage

import (
	"errors"
	"sync"
)

type InMemoryURLStorage struct {
	urlMap map[string]string
	mutex  sync.RWMutex
}

func NewInMemoryURLStorage() *InMemoryURLStorage {
	return &InMemoryURLStorage{
		urlMap: make(map[string]string),
	}
}

func (s *InMemoryURLStorage) Get(shortKey string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	originalURL, exists := s.urlMap[shortKey]
	if !exists {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (s *InMemoryURLStorage) Set(shortKey string, originalURL string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.urlMap[shortKey] = originalURL
}
