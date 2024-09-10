package config

import (
	"flag"
	"os"
)

var Options struct {
	Addr            string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func Run() {
	flag.StringVar(&Options.Addr, "a", "localhost:8080", "http server address")
	flag.StringVar(&Options.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&Options.FileStoragePath, "f", "temp", "file storage path")
	flag.StringVar(&Options.DatabaseDSN, "d", "localhost:5432", "database dsn")

	flag.Parse()

	if servAddr := os.Getenv("SERVER_ADDRESS"); servAddr != "" {
		Options.Addr = servAddr
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		Options.BaseURL = baseURL
	}
	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		Options.FileStoragePath = fileStoragePath
	}
	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		Options.DatabaseDSN = databaseDSN
	}
}
