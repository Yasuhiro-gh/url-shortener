package handlers

import (
	"github.com/Yasuhiro-gh/url-shortener/internal/storage"
	"github.com/Yasuhiro-gh/url-shortener/internal/utils"
	"io"
	"net/http"
)

func ShortURL(us *storage.URLStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		_, _ = w.Write([]byte("http://localhost:8080/" + string(urlHash)))
	})
}

func GetShortURL(us *storage.URLStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET method is supported.", http.StatusBadRequest)
			return
		}

		urlHash := r.PathValue("id")

		if urlHash == "" {
			http.Error(w, "Please provide a URL.", http.StatusBadRequest)
			return
		}

		if !utils.IsHashExist(urlHash, us.Urls) {
			http.Error(w, "Invalid URL.", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", us.Urls[urlHash])
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}
