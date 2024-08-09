package handlers

import (
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/storage"
	"github.com/Yasuhiro-gh/url-shortener/internal/utils"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

func URLRouter(us *storage.URLStore) chi.Router {
	r := chi.NewRouter()
	r.HandleFunc("/", ShortURL(us))
	r.HandleFunc("/{id}", GetShortURL(us))
	return r
}

func ShortURL(us *storage.URLStore) http.HandlerFunc {
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
		if !utils.IsHashExist(urlHash, us.Urls) {
			us.Urls[urlHash] = urlString
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		_, _ = w.Write([]byte(config.Options.BaseAddr + "/" + urlHash))
	}
}

func GetShortURL(us *storage.URLStore) http.HandlerFunc {
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

		if !utils.IsHashExist(shortURL, us.Urls) {
			http.Error(w, "Invalid URL.", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", us.Urls[shortURL])
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
