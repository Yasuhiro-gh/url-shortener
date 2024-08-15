package handlers

import (
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/storage"
	"github.com/Yasuhiro-gh/url-shortener/internal/utils"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

func URLRouter(us *storage.URLS) chi.Router {
	r := chi.NewRouter()
	r.HandleFunc("/", ShortURL(us))
	r.HandleFunc("/{id}", GetShortURL(us))
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
