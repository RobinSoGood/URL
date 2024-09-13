package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func Test_saveURL(t *testing.T) {
	ts := httptest.NewServer(URLShortener())
	defer ts.Close()
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
		emptyBody    bool
	}{
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, emptyBody: true},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, emptyBody: true},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, emptyBody: true},
		{method: http.MethodPost, expectedCode: http.StatusCreated, emptyBody: false},
	}

	for _, tc := range testCases {
		// t.Run(tc.method, func(t *testing.T) {
		// 	r := httptest.NewRequest(tc.method, "/", nil)
		// 	w := httptest.NewRecorder()
		// 	// вызовем хендлер как обычную функцию, без запуска самого сервера
		// 	saveURL(w, r)

		// 	assert.Equal(t, tc.expectedCode, w.Code)
		// 	// проверим корректность полученного тела ответа, если мы его ожидаем
		// 	if !tc.expectedBody {
		// 		// assert.JSONEq помогает сравнить две JSON-строки
		// 		// assert.JSONEq(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
		// 		assert.NotEqual(t, w.Body.String(), "")
		// 	}
		// })
		resp, body := testRequest(t, ts, tc.method, "/", nil)
		defer resp.Body.Close()

		assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

		if !tc.emptyBody {
			assert.NotEqual(t, body, "", "Тело ответа пустое")
		}
	}
}

func Test_getURLByID(t *testing.T) {
	ts := httptest.NewServer(URLShortener())
	defer ts.Close()

	linkToSave := "https://practicum.yandex.ru/"
	expectedHostName := "practicum.yandex.ru"
	resp, shortLink := testRequest(t, ts, http.MethodPost, "/", strings.NewReader(linkToSave))
	defer resp.Body.Close()
	shortLinkID := strings.Split(shortLink, "/")[len(strings.Split(shortLink, "/"))-1]

	testCases := []struct {
		method       string
		expectedCode int
	}{
		{method: http.MethodGet, expectedCode: http.StatusOK},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPost, expectedCode: http.StatusMethodNotAllowed},
	}

	for _, tc := range testCases {
		// t.Run(tc.method, func(t *testing.T) {
		// 	r := httptest.NewRequest(tc.method, string(shortLink), nil)
		// 	w := httptest.NewRecorder()

		// 	getURLByID(w, r)

		// 	if tc.method == http.MethodGet {
		// 		assert.Equal(t, w.Header().Get("Location"), linkToSave)
		// 	}
		// 	assert.Equal(t, tc.expectedCode, w.Code)
		// })
		resp, _ := testRequest(t, ts, tc.method, "/"+string(shortLinkID), nil)

		defer resp.Body.Close()

		if tc.method == http.MethodGet {
			assert.Equal(t, resp.Request.URL.Hostname(), expectedHostName, "Redirect не был выполнен успешно")
		}
		assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
	}

}
