package app

import (
	"flag"
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/handlers"
	"github.com/Yasuhiro-gh/url-shortener/internal/storage"
	"net/http"
)

func Run() {
	urlStore := storage.NewURLStore()

	flag.Parse()

	err := http.ListenAndServe(config.Options.Addr, handlers.URLRouter(urlStore))
	if err != nil {
		panic(err)
	}
}
