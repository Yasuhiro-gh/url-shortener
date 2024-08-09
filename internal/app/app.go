package app

import (
	"github.com/Yasuhiro-gh/url-shortener/internal/handlers"
	"github.com/Yasuhiro-gh/url-shortener/internal/storage"
	"net/http"
)

func Run() {
	urlStore := storage.NewURLStore()

	err := http.ListenAndServe(":8080", handlers.URLRouter(urlStore))
	if err != nil {
		panic(err)
	}
}
