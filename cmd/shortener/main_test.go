package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_saveURL(t *testing.T) {
	// описываем ожидаемое тело ответа при успешном запросе
	// successBody := `{
	//     "response": {
	//         "text": "Извините, я пока ничего не умею"
	//     },
	//     "version": "1.0"
	// }

	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		method       string
		expectedCode int
		expectedBody bool
	}{
		{method: http.MethodGet, expectedCode: http.StatusBadRequest, expectedBody: true},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest, expectedBody: true},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest, expectedBody: true},
		{method: http.MethodPost, expectedCode: http.StatusCreated, expectedBody: false},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", nil)
			w := httptest.NewRecorder()
			// вызовем хендлер как обычную функцию, без запуска самого сервера
			saveURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			// проверим корректность полученного тела ответа, если мы его ожидаем
			if !tc.expectedBody {
				// assert.JSONEq помогает сравнить две JSON-строки
				// assert.JSONEq(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
				assert.NotEqual(t, w.Body.String(), "", "Тело ответа пустое")
			}
		})
	}
}

func Test_getURLByID(t *testing.T) {

	linkToSave := "https://practicum.yandex.ru/"

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(linkToSave))
	res := httptest.NewRecorder()
	saveURL(res, req)

	shortLink, _ := io.ReadAll(res.Body)

	testCases := []struct {
		method       string
		expectedCode int
	}{
		{method: http.MethodGet, expectedCode: http.StatusTemporaryRedirect},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, string(shortLink), nil)
			w := httptest.NewRecorder()

			getURLByID(w, r)

			if tc.method == http.MethodGet {
				assert.Equal(t, w.Header().Get("Location"), linkToSave)
			}
			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}

}
