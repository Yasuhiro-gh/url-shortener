package handlers

import (
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/storage"
	"github.com/Yasuhiro-gh/url-shortener/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type mockURLS struct {
	shortURL string
	fullURL  string
}

func TestMain(m *testing.M) {
	config.Run()
	os.Exit(m.Run())
}

func NewMockMapURLS(urls ...mockURLS) *storage.URLS {
	us := storage.NewURLStorage()
	for _, url := range urls {
		us.Set(url.shortURL, url.fullURL)
	}
	return storage.NewURLS(us)
}

func TestShortURLMethods(t *testing.T) {
	tests := []struct {
		storage      *storage.URLS
		method       string
		expectedCode int
		expectedBody string
	}{
		{
			storage:      NewMockMapURLS(),
			method:       http.MethodGet,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			storage:      NewMockMapURLS(),
			method:       http.MethodDelete,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			storage:      NewMockMapURLS(),
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			storage:      NewMockMapURLS(),
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}
	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "http://localhost:8080/", nil)
			w := httptest.NewRecorder()

			NewURLHandler(test.storage).ShortURL().ServeHTTP(w, r)

			assert.Equal(t, test.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestShortURL(t *testing.T) {
	tests := []struct {
		name                string
		storage             *storage.URLS
		body                string
		expectedCode        int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "empty url",
			storage:             NewMockMapURLS(),
			body:                "",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Please provide a URL.\n",
		},
		{
			name:                "invalid url",
			storage:             NewMockMapURLS(),
			body:                "yandex",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Invalid URL.\n",
		},
		{
			name:                "valid url",
			storage:             NewMockMapURLS(),
			body:                "https://yandex.com",
			expectedCode:        http.StatusCreated,
			expectedContentType: "text/plain",
			expectedBody:        "http://localhost:8080/" + utils.HashURL("https://yandex.com"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", strings.NewReader(test.body))
			w := httptest.NewRecorder()

			NewURLHandler(test.storage).ShortURL().ServeHTTP(w, r)

			res := w.Result()

			assert.Equal(t, test.expectedCode, res.StatusCode, "Wrong response code status")

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.expectedContentType, res.Header.Get("Content-Type"), "Wrong content type")
			assert.Equal(t, test.expectedBody, string(resBody), "Wrong response body")
		})
	}
}

func TestShortURLJSON(t *testing.T) {
	tests := []struct {
		name                string
		storage             *storage.URLS
		body                string
		expectedCode        int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "empty url",
			storage:             NewMockMapURLS(),
			body:                "{}",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Please provide a URL.\n",
		},
		{
			name:                "invalid url",
			storage:             NewMockMapURLS(),
			body:                `{"url": "yandex"}`,
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        "Invalid URL.\n",
		},
		{
			name:                "valid url",
			storage:             NewMockMapURLS(),
			body:                `{"url": "https://yandex.com"}`,
			expectedCode:        http.StatusCreated,
			expectedContentType: "application/json",
			expectedBody:        `{"result":` + "\"" + "http://localhost:8080/" + utils.HashURL("https://yandex.com") + "\"" + "}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", strings.NewReader(test.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			NewURLHandler(test.storage).ShortURLJSON().ServeHTTP(w, r)

			res := w.Result()

			assert.Equal(t, test.expectedCode, res.StatusCode, "Wrong response code status")

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.expectedContentType, res.Header.Get("Content-Type"), "Wrong content type")
			assert.Equal(t, test.expectedBody, string(resBody), "Wrong response body")
		})
	}
}

func TestGetShortURLMethods(t *testing.T) {
	tests := []struct {
		storage      *storage.URLS
		method       string
		expectedCode int
		expectedBody string
	}{
		{
			storage:      NewMockMapURLS(),
			method:       http.MethodGet,
			expectedCode: http.StatusBadRequest,
			expectedBody: "Please provide a URL.\n",
		},
		{
			storage:      NewMockMapURLS(),
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			storage:      NewMockMapURLS(),
			method:       http.MethodDelete,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			storage:      NewMockMapURLS(),
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			storage:      NewMockMapURLS(),
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}
	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "http://localhost:8080/", nil)
			w := httptest.NewRecorder()

			NewURLHandler(test.storage).GetShortURL().ServeHTTP(w, r)

			assert.Equal(t, test.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGetShortURL(t *testing.T) {
	type header struct {
		contentType string
		location    string
	}
	tests := []struct {
		name           string
		storage        *storage.URLS
		expectedCode   int
		expectedHeader header
		expectedBody   string
		shortURL       string
	}{
		{
			name:           "non-existent short url",
			storage:        NewMockMapURLS(),
			expectedCode:   http.StatusBadRequest,
			expectedHeader: header{contentType: "text/plain; charset=utf-8"},
			expectedBody:   "Invalid URL.\n",
			shortURL:       "12345678",
		},
		{
			name:           "existent short url",
			storage:        NewMockMapURLS(mockURLS{utils.HashURL("https://yandex.com"), "https://yandex.com"}),
			expectedCode:   http.StatusTemporaryRedirect,
			expectedHeader: header{contentType: "text/plain", location: "https://yandex.com"},
			shortURL:       utils.HashURL("https://yandex.com"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "http://localhost:8080", nil)
			r.SetPathValue("id", test.shortURL)
			w := httptest.NewRecorder()

			NewURLHandler(test.storage).GetShortURL().ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.expectedCode, res.StatusCode, "Wrong response code status")
			assert.Equal(t, test.expectedHeader.contentType, res.Header.Get("Content-Type"), "Wrong content type")
			assert.Equal(t, test.expectedHeader.location, res.Header.Get("Location"), "Wrong location")
		})
	}
}
