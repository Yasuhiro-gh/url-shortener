package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var urlHashMap = make(map[string]string)

func isValidUrl(urlToValid string) bool {
	_, err := url.ParseRequestURI(urlToValid)
	return err == nil
}

func isValidHash(hash string) bool {
	_, ok := urlHashMap[hash]
	return ok
}

func hashUrl(url string) string {
	hash := sha256.New()
	hash.Write([]byte(url))
	return fmt.Sprintf("%x", hash.Sum(nil))[:8]
}

func shortUrl(w http.ResponseWriter, req *http.Request) {
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

	if !isValidUrl(urlString) {
		http.Error(w, "Invalid URL.", http.StatusBadRequest)
		return
	}

	urlHash := hashUrl(urlString)
	if _, ok := urlHashMap[urlString]; !ok {
		urlHashMap[urlHash] = urlString
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, _ = w.Write([]byte("http://localhost:8080/" + string(urlHash)))
}

func getShortUrl(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Only GET method is supported.", http.StatusBadRequest)
		return
	}

	urlHash := req.PathValue("id")

	if !isValidHash(urlHash) {
		http.Error(w, "Invalid URL.", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", urlHashMap[urlHash])
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", shortUrl)
	mux.HandleFunc("/{id}/", getShortUrl)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
