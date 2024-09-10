package app

import (
	"context"
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/db"
	"github.com/Yasuhiro-gh/url-shortener/internal/handlers"
	"github.com/Yasuhiro-gh/url-shortener/internal/logger"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/storage"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/storage/filestore"
	"net/http"
)

func Run() {
	config.Run()
	logger.Run()

	us := storage.NewURLStorage()
	urls := storage.NewURLS(us)

	pdb := db.NewPostgresDB()
	err := pdb.OpenConnection()
	if err != nil {
		panic(err)
	}
	defer pdb.CloseConnection()

	err = filestore.Restore(urls)
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(config.Options.Addr, handlers.URLRouter(context.Background(), urls, pdb))
	if err != nil {
		panic(err)
	}
}
