package app

import (
	"github.com/Yasuhiro-gh/url-shortener/internal/handlers"
	"github.com/Yasuhiro-gh/url-shortener/internal/storage"
	"net/http"
)

func Run() {
	urlStore := storage.NewURLStore()

	mux := http.NewServeMux()
	mux.Handle("/", handlers.ShortURL(urlStore))
	mux.Handle("/{id}", handlers.GetShortURL(urlStore))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
