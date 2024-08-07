package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type URLStore struct {
	urls map[string]string
}

func NewURLStore() *URLStore {
	return &URLStore{urls: make(map[string]string)}
}

func isValidURL(urlToValid string) bool {
	_, err := url.ParseRequestURI(urlToValid)
	return err == nil
}

func isValidHash(hash string, urls map[string]string) bool {
	_, ok := urls[hash]
	return ok
}

func hashURL(url string) string {
	hash := sha256.New()
	hash.Write([]byte(url))
	return fmt.Sprintf("%x", hash.Sum(nil))[:8]
}

func (u *URLStore) shortURL(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported.", http.StatusBadRequest)
		return
	}

	body, _ := io.ReadAll(req.Body)

	urlString := string(body)
	if urlString == "" {
		http.Error(w, "Please provide a URL.", http.StatusBadRequest)
		return
	}

	if !isValidURL(urlString) {
		http.Error(w, "Invalid URL.", http.StatusBadRequest)
		return
	}

	urlHash := hashURL(urlString)
	if _, ok := u.urls[urlString]; !ok {
		u.urls[urlHash] = urlString
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, _ = w.Write([]byte("http://localhost:8080/" + string(urlHash)))
}

func (u *URLStore) getShortURL(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Only GET method is supported.", http.StatusBadRequest)
		return
	}

	urlHash := req.PathValue("id")

	if !isValidHash(urlHash, u.urls) {
		http.Error(w, "Invalid URL.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", u.urls[urlHash])
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	urlStore := NewURLStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/", urlStore.shortURL)
	mux.HandleFunc("/{id}", urlStore.getShortURL)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
