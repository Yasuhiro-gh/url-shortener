package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/logger"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/storage"
	"github.com/Yasuhiro-gh/url-shortener/internal/utils"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strings"
)

func URLRouter(us *storage.URLS) chi.Router {
	r := chi.NewRouter()
	r.Handle("/", logger.Logging(ShortURL(us)))
	r.Handle("/{id}", logger.Logging(GetShortURL(us)))
	r.Handle("/api/shorten", logger.Logging(ShortURLJSON(us)))
	return r
}

func ShortURL(us *storage.URLS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported.", http.StatusBadRequest)
			return
		}

		body, _ := io.ReadAll(r.Body)

		urlString := string(body)
		if urlString == "" {
			http.Error(w, "Please provide a URL.", http.StatusBadRequest)
			return
		}

		if !utils.IsValidURL(urlString) {
			http.Error(w, "Invalid URL.", http.StatusBadRequest)
			return
		}

		urlHash := utils.HashURL(urlString)
		if _, exist := us.Get(urlHash); !exist {
			us.Set(urlHash, urlString)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		_, _ = w.Write([]byte(config.Options.BaseURL + "/" + urlHash))
	}
}

func GetShortURL(us *storage.URLS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET method is supported.", http.StatusBadRequest)
			return
		}

		shortURL := r.PathValue("id")

		if shortURL == "" {
			http.Error(w, "Please provide a URL.", http.StatusBadRequest)
			return
		}

		url, exist := us.Get(shortURL)
		if !exist {
			http.Error(w, "Invalid URL.", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func ShortURLJSON(us *storage.URLS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported.", http.StatusBadRequest)
			return
		}

		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Only JSON content type is supported.", http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer

		type ShortenJSON struct {
			URL string `json:"url"`
		}

		var shortenRequest ShortenJSON

		type ResponseJSON struct {
			Result string `json:"result"`
		}

		var shortenResponse ResponseJSON

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &shortenRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if shortenRequest.URL == "" {
			http.Error(w, "Please provide a URL.", http.StatusBadRequest)
			return
		}

		if !utils.IsValidURL(shortenRequest.URL) {
			http.Error(w, "Invalid URL.", http.StatusBadRequest)
			return
		}

		urlHash := utils.HashURL(shortenRequest.URL)
		if _, exist := us.Get(urlHash); !exist {
			us.Set(urlHash, shortenRequest.URL)
		}
		shortenResponse.Result = config.Options.BaseURL + "/" + urlHash

		resp, err := json.Marshal(shortenResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(resp)
	}
}
