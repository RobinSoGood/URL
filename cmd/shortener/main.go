package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/RobinSoGood/URL.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

var urlStorage = storage.NewInMemoryURLStorage()

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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

func URLShortener() chi.Router {
	r := chi.NewRouter()
	r.Post("/", saveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", getURLByID)
	return r
}

func main() {

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
