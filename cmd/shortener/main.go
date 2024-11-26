package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/RobinSoGood/URL/internal/app/middleware"
	"github.com/RobinSoGood/URL/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var urlStorage = storage.NewInMemoryURLStorage()

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var logger, _ = zap.NewProduction()

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	w.Header().Del("Content-Length") // Сжатые ответы не должны содержать этот заголовок
	w.ResponseWriter.WriteHeader(status)
}

func newGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	gz, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
	return &gzipResponseWriter{
		ResponseWriter: w,
		gzipWriter:     gz,
	}
}

// Middleware для сжатия ответа
func gzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := newGzipResponseWriter(w)
			defer gz.gzipWriter.Close()
			next.ServeHTTP(gz, r)
		} else {
			next.ServeHTTP(w, r) // Если клиент не поддерживает gzip, просто передаем дальше
		}
	})
}

func saveURL(w http.ResponseWriter, r *http.Request) {
	bytes, _ := io.ReadAll(r.Body)
	urlStr := string(bytes)
	randomPath := RandStringBytes(8)

	err := urlStorage.Set(randomPath, urlStr)
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v/%v", baseURL, randomPath)
}

func getURLByID(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")
	value, err := urlStorage.Get(shortURL)
	if err == nil {
		w.Header().Set("Location", value)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

type requestBody struct {
	URL string `json:"url"`
}

type responseBody struct {
	Result string `json:"result"`
}

func createShortURL(w http.ResponseWriter, r *http.Request) {
	var req requestBody

	gz, ok := w.(*gzipResponseWriter)
	if !ok {
		gz = newGzipResponseWriter(w) // Создаем gzipWriter, если его ещё нет
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "Поле 'url' обязательно", http.StatusUnprocessableEntity)
		return
	}

	randomPath := RandStringBytes(8)
	err = urlStorage.Set(randomPath, req.URL)
	if err != nil {
		http.Error(w, "Ошиба сохранения", http.StatusInternalServerError)
		return
	}

	res := responseBody{
		Result: fmt.Sprintf("%s/%s", baseURL, randomPath),
	}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gz.WriteHeader(http.StatusCreated)
	_, err = gz.gzipWriter.Write(jsonRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func URLShortener() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(gzipHandler) // Добавляем middleware для сжатия ответов
	r.Post("/api/shorten", createShortURL)
	r.Post("/", saveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", getURLByID)
	return r
}

func main() {
	defer logger.Sync()
	ParseOptions()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	err := http.ListenAndServe(serverAddress, URLShortener())
	if err != nil {
		return err
	}
	return nil
}
