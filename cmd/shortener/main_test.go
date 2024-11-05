package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortenHandler(t *testing.T) {
	url := "https://practicum.yandex.ru"
	reqBody, _ := json.Marshal(URLRequest{URL: url})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(shortenHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Unexpected status code: got %v, want %v", status, http.StatusCreated)
	}

	var res URLResponse
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Fatal(err)
	}

	if res.Result == "" {
		t.Error("Expected result to be non-empty")
	}
}
