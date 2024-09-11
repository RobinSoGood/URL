package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var urlMap = map[string]string{}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func saveURL(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	bytes, _ := io.ReadAll(r.Body)
	urlStr := string(bytes)
	randomPath := RandStringBytes(8)
	urlMap[randomPath] = urlStr
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v/%v", flagRunAddrB, randomPath)
}

func getURLByID(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodGet {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// shortURL := r.URL.Path[1:]
	// value, ok := urlMap[shortURL]

	shortURL := chi.URLParam(r, "shortURL")
	value, ok := urlMap[shortURL]
	if ok {
		w.Header().Set("Location", value)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// func URLShortener(w http.ResponseWriter, r *http.Request) {
func URLShortener() chi.Router {
	r := chi.NewRouter()
	r.Post("/", saveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", getURLByID)
	return r
}
func run() error {

	err := http.ListenAndServe(flagRunAddrA, URLShortener())
	if err != nil {
		panic(err)
	}
	return nil
}
func main() {
	// mux := http.NewServeMux()
	// mux.HandleFunc(`/`, URLShortener)

	parseFlags()

	if err := run(); err != nil {
		panic(err)
	}

	// err := http.ListenAndServe(`:8080`, URLShortener())
	// if err != nil {
	// 	panic(err)
	// }

	// r := chi.NewRouter()
	// r.Use(middleware.Logger)
	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Hello World!"))
	// })
	// http.ListenAndServe(":3000", r)
}
