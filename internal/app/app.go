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

	pdb := db.NewPostgresDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var urls *storage.URLS
	if config.Options.DatabaseDSN != "" {
		err := pdb.OpenConnection()
		if err != nil {
			panic(err)
		}
		err = db.CreateDatabaseTable(pdb)
		if err != nil {
			panic(err)
		}
		defer pdb.CloseConnection()
		urls = storage.NewURLS(pdb)
	} else {
		us := storage.NewURLStorage()
		urls = storage.NewURLS(us)
	}

	err := filestore.Restore(urls)
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(config.Options.Addr, handlers.URLRouter(ctx, urls, pdb))
	if err != nil {
		panic(err)
	}
}
