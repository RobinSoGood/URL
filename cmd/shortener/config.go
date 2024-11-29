package main

import (
	"flag"
	"os"
)

var (
	defaultServerAddress   = ":8080"                 // Значение по умолчанию для адреса сервера
	defaultBaseURL         = "http://localhost:8080" // Базовый URL по умолчанию
	defaultFileStoragePath = "./urls.json"           // Путь к файлу по умолчанию
)

var serverAddress string
var baseURL string
var fileStoragePath string

func ParseOptions() {
	// Определение флагов командной строки
	flag.StringVar(&serverAddress, "a", defaultServerAddress, "Адрес запуска HTTP-сервера")
	flag.StringVar(&baseURL, "b", defaultBaseURL, "Базовый адрес результирующего сокращённого URL")
	flag.StringVar(&fileStoragePath, "f", defaultFileStoragePath, "Путь до файла для хранения данных")

	// Парсим флаги
	flag.Parse()

	// Получаем значения из переменных окружения
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		serverAddress = envServerAddress
	} else if serverAddress == "" {
		serverAddress = defaultServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	} else if baseURL == "" {
		baseURL = defaultBaseURL
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = envFileStoragePath
	} else if fileStoragePath == "" {
		fileStoragePath = defaultFileStoragePath
	}
}
