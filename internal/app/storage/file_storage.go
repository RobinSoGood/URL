package storage

import (
	"encoding/json"
	"io/ioutil"
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

func (s *FileURLStorage) loadFromFile() error {
	if s.initialized {
		return nil
	}

	file, err := ioutil.ReadFile(s.filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to read from file %s: %v\n", s.filePath, err)
		return err
	}

	if len(file) > 0 {
		err = json.Unmarshal(file, &s.urlMap)
		if err != nil {
			log.Printf("Failed to unmarshal data from file %s: %v\n", s.filePath, err)
			return err
		}
	}

	s.initialized = true
	return nil
}

func (s *FileURLStorage) SaveToFile() error {
	data, err := json.MarshalIndent(s.urlMap, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal data for file %s: %v\n", s.filePath, err)
		return err
	}

	err = ioutil.WriteFile(s.filePath, data, 0644)
	if err != nil {
		log.Printf("Failed to write to file %s: %v\n", s.filePath, err)
		return err
	}

	return nil
}

func (s *FileURLStorage) Get(shortKey string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if err := s.loadFromFile(); err != nil {
		return "", err
	}

	originalURL, exists := s.urlMap[shortKey]
	if !exists {
		return "", ErrNotFound
	}
	return originalURL, nil
}

func (s *FileURLStorage) Set(shortKey string, originalURL string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.loadFromFile(); err != nil {
		return err
	}

	s.urlMap[shortKey] = originalURL
	return s.SaveToFile()
}
