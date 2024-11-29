package storage

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type DiskURLStorage struct {
	filePath string
	urlMap   map[string]string
	mutex    sync.RWMutex
}

type StorageEntry struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewDiskURLStorage(filePath string) *DiskURLStorage {
	storage := &DiskURLStorage{
		filePath: filePath,
		urlMap:   make(map[string]string),
	}

	storage.loadFromDisk()

	return storage
}

func (s *DiskURLStorage) loadFromDisk() {
	data, err := os.ReadFile(s.filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Ошибка чтения файла %s: %v", s.filePath, err)
	}

	if len(data) > 0 {
		var entries []StorageEntry
		if err := json.Unmarshal(data, &entries); err != nil {
			log.Fatalf("Ошибка разбора JSON из файла %s: %v", s.filePath, err)
		}

		for _, entry := range entries {
			s.urlMap[entry.ShortURL] = entry.OriginalURL
		}
	}
}

func (s *DiskURLStorage) persistToDisk() {
	entries := make([]StorageEntry, 0, len(s.urlMap))
	for key, value := range s.urlMap {
		entries = append(entries, StorageEntry{
			ShortURL:    key,
			OriginalURL: value,
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		log.Fatalf("Ошибка сериализации данных в JSON: %v", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		log.Fatalf("Ошибка записи в файл %s:= %v", s.filePath, err)
	}
}

func (s *DiskURLStorage) Get(shortKey string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	originalURL, exists := s.urlMap[shortKey]
	if !exists {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (s *DiskURLStorage) Set(shortKey string, originalURL string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.urlMap[shortKey] = originalURL
	s.persistToDisk()
	return nil
}
