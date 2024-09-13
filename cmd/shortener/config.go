package main

import (
	"flag"
	"os"
)

// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера
// var flagRunAddrA string
// var flagRunAddrB string

var (
	defaultServerAddress = ":8080"                 // Значение по умолчанию для адреса сервера
	defaultBaseURL       = "http://localhost:8080" // Базовый URL по умолчанию
)

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
// func parseFlags() {
// 	// регистрируем переменную flagRunAddr
// 	// как аргумент -a со значением :8080 по умолчанию
// 	flag.StringVar(&flagRunAddrA, "a", ":8080", "address and port to run server")
// 	flag.StringVar(&flagRunAddrB, "b", "https://localhost:8080", "address and port to run server")
// 	// парсим переданные серверу аргументы в зарегистрированные переменные
// 	flag.Parse()
// }

// func GetConfig() *Config {
// 	c := new(Config)

// 	if EnvServAddr := os.Getenv("SERVER_ADDRESS"); EnvServAddr != "" {
// 		flagRunAddrA = EnvServAddr
// 	}

// 	if EnvBaseAddr := os.Getenv("BASE_URL"); EnvBaseAddr != "" {
// 		flagRunAddrB = EnvBaseAddr
// 	}

// 	return c
// }

func ParseOptions() {
	// Определение флагов командной строки
	var serverAddress string
	var baseURL string

	flag.StringVar(&serverAddress, "a", ":8080", "Адрес запуска HTTP-сервера")
	flag.StringVar(&baseURL, "b", "https://localhost:8080", "Базовый адрес результирующего сокращённого URL")

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
}
