package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
    "bytes"
    "compress/gzip"
    "io/ioutil"
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

func saveURL(w http.ResponseWriter, r *http.Request) {
    var buf bytes.Buffer
    gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
    if err != nil {
        http.Error(w, "Не удалось создать gzip writer", http.StatusInternalServerError)
        return
    }
    defer gz.Close()

    _, err = io.Copy(gz, r.Body)
    if err != nil {
        http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
        return
    }

    err = gz.Flush()
    if err != nil {
        http.Error(w, "Не удалось завершить сжатие", http.StatusInternalServerError)
        return
    }

    urlStr := buf.String()
    randomPath := RandStringBytes(8)

    err = urlStorage.Set(randomPath, urlStr)
    if err != nil {
        http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
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

    if r.Header.Get("Content-Encoding") == "gzip" {
        gzr, err := gzip.NewReader(r.Body)
        if err != nil {
            http.Error(w, "Ошибка распаковки gzip", http.StatusBadRequest)
            return
        }
        defer gzr.Close()

        body, err := ioutil.ReadAll(gzr)
        if err != nil {
            http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
            return
        }

        err = json.Unmarshal(body, &req)
        if err != nil {
            http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
            return
        }
    } else {
        err := json.NewDecoder(r.Body).Decode(&req)
        if err != nil {
            http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
            return
        }
    }

    if req.URL == "" {
        http.Error(w, "Поле 'url' обязательно", http.StatusUnprocessableEntity)
        return
    }

    randomPath := RandStringBytes(8)
    err := urlStorage.Set(randomPath, req.URL)
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

    if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
        w.Header().Set("Content-Encoding", "gzip")
        var buf bytes.Buffer
        gz := gzip.NewWriter(&buf)
        defer gz.Close()

        _, err = gz.Write(jsonRes)
        if err != nil {
            http.Error(w, "Ошибка сжатия ответа", http.StatusInternalServerError)
            return
        }

        err = gz.Flush()
        if err != nil {
            http.Error(w, "Ошибка завершения сжатия", http.StatusInternalServerError)
            return
        }

        w.Write(buf.Bytes())
    } else {
        w.WriteHeader(http.StatusCreated)
        w.Write(jsonRes)
    }
}

func URLShortener() chi.Router {
    r := chi.NewRouter()
    r.Use(middleware.LoggerMiddleware(logger))
    r.Use(CompressResponseMiddleware)
    r.Post("/api/shorten", createShortURL)
    r.Post("/", saveURL)
    r.Get("/{shortURL:[A-Za-z]{8}}", getURLByID)
    return r
}

func CompressResponseMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            w.Header().Set("Content-Encoding", "gzip")
            gz := gzip.NewWriter(w)
            defer gz.Close()

            next.ServeHTTP(gzipResponseWriter{gz}, r)
        } else {
            next.ServeHTTP(w, r)
        }
    })
}

type gzipResponseWriter struct {
    io.Writer
}

func (w gzipResponseWriter) WriteHeader(status int) {
    w.Writer.(http.ResponseWriter).WriteHeader(status)
}

func (w gzipResponseWriter) Header() http.Header {
    return w.Writer.(http.ResponseWriter).Header()
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