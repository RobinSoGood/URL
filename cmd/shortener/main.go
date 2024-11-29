package main

import (
	"bytes"
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

var urlStorage = storage.NewDiskURLStorage(fileStoragePath)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var logger, _ = zap.NewProduction()

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonRes)
}

func URLShortener() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(decompressGzipRequest)
	r.Use(gzipMiddleware)
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

type gzipResponseWriter struct {
	http.ResponseWriter
	gzwriter *gzip.Writer
}

func (grw gzipResponseWriter) Write(p []byte) (int, error) {
	return grw.gzwriter.Write(p)
}

func (grw gzipResponseWriter) Header() http.Header {
	return grw.ResponseWriter.Header()
}

func (grw gzipResponseWriter) WriteHeader(status int) {
	grw.ResponseWriter.WriteHeader(status)
}

func (grw *gzipResponseWriter) Flush() {
	_ = grw.gzwriter.Flush()
	if flusher, ok := grw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (grw *gzipResponseWriter) Close() {
	_ = grw.gzwriter.Close()
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gzw := gzip.NewWriter(w)
			defer gzw.Close()

			next.ServeHTTP(gzipResponseWriter{w, gzw}, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func decompressGzipRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Не удалось распаковать gzip", http.StatusBadRequest)
				return
			}
			defer gzr.Close()

			body, err := io.ReadAll(gzr)
			if err != nil {
				http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		next.ServeHTTP(w, r)
	})
}
