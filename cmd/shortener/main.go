package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/RobinSoGood/URL/internal/app/middleware"
	"github.com/RobinSoGood/URL/internal/app/storage"
	"go.uber.org/zap"
)

type URLRequest struct {
	URL string `json:"url"`
}

type URLResponse struct {
	Result string `json:"result"`
}

var (
	urlStorage      = storage.NewInMemoryURLStorage()
	shortKeyCounter int
	mutex           sync.Mutex
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	http.HandleFunc("/api/shorten", middleware.LoggerMiddleware(logger)(http.HandlerFunc(shortenHandler)).ServeHTTP)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req URLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	shortKeyCounter++
	shortKey := generateShortKey(shortKeyCounter)
	urlStorage.Set(shortKey, req.URL)

	shortURL := "http://localhost:8080/" + shortKey
	res := URLResponse{Result: shortURL}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func generateShortKey(counter int) string {
	return fmt.Sprintf("%X", counter)
}
