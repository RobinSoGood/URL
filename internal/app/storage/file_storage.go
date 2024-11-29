package storage

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type FileURLStorage struct {
	urlMap      map[string]string
	mutex       sync.RWMutex
	filePath    string
	initialized bool
}

func NewFileURLStorage(filePath string) *FileURLStorage {
	return &FileURLStorage{
		urlMap:   make(map[string]string),
		filePath: filePath,
	}
}

func (s *FileURLStorage) initialize() {
	if s.initialized {
		return
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to read file %s: %v", s.filePath, err)
	}

	if len(data) > 0 {
		var urls []struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}
		if err := json.Unmarshal(data, &urls); err != nil {
			log.Fatalf("Failed to unmarshal data from file %s: %v", s.filePath, err)
		}

		for _, u := range urls {
			s.urlMap[u.ShortURL] = u.OriginalURL
		}
	}

	s.initialized = true
}

func (s *FileURLStorage) Get(shortKey string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.initialize()

	originalURL, exists := s.urlMap[shortKey]
	if !exists {
		return "", ErrURLNotFound
	}
	return originalURL, nil
}

func (s *FileURLStorage) Set(shortKey string, originalURL string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.initialize()

	s.urlMap[shortKey] = originalURL
	return s.saveToFile()
}

func (s *FileURLStorage) saveToFile() error {
	urls := make([]struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}, 0, len(s.urlMap))

	for k, v := range s.urlMap {
		urls = append(urls, struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}{k, v})
	}

	data, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}
