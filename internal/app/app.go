package app

import (
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/handlers"
	"github.com/Yasuhiro-gh/url-shortener/internal/logger"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/storage"
	"net/http"
)

func Run() {
	config.Run()
	logger.Run()

	us := storage.NewURLStorage()
	urls := storage.NewURLS(us)
	err := http.ListenAndServe(config.Options.Addr, handlers.URLRouter(urls))
	if err != nil {
		panic(err)
	}
}
