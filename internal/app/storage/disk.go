package storage

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type FileURLStorage struct {
	filePath string
	urlMap   map[string]string
	mutex    sync.RWMutex
}

func NewFileURLStorage(filePath string) *FileURLStorage {
	return &FileURLStorage{
		filePath: filePath,
		urlMap:   make(map[string]string),
	}
}

func (s *FileURLStorage) LoadFromFile() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Чтение данных из файла
	data, err := os.ReadFile(s.filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Ошибка чтения файла %s: %v\n", s.filePath, err)
		return err
	}

	// Разбор данных в карту
	if len(data) > 0 {
		var entries []map[string]string
		if err := json.Unmarshal(data, &entries); err != nil {
			log.Printf("Ошибка разбора JSON из файла %s: %v\n", s.filePath, err)
			return err
		}

		for _, entry := range entries {
			for k, v := range entry {
				s.urlMap[k] = v
			}
		}
	}

	return nil
}

func (s *FileURLStorage) SaveToFile() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Преобразование карты в массив структур
	entries := make([]map[string]string, 0, len(s.urlMap))
	for k, v := range s.urlMap {
		entry := map[string]string{k: v}
		entries = append(entries, entry)
	}

	// Сериализация массива в JSON
	data, err := json.MarshalIndent(entries, "", "\t")
	if err != nil {
		log.Printf("Ошибка сериализации данных в JSON: %v\n", err)
		return err
	}

	// Запись данных в файл
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		log.Printf("Ошибка записи в файл %s: %v\n", s.filePath, err)
		return err
	}

	return nil
}

func (s *FileURLStorage) Get(shortKey string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	originalURL, exists := s.urlMap[shortKey]
	if !exists {
		return "", ErrURLNotFound
	}
	return originalURL, nil
}

func (s *FileURLStorage) Set(shortKey string, originalURL string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.urlMap[shortKey] = originalURL
	return nil
}
