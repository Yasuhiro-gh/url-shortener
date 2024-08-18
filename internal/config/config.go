package config

import (
	"flag"
	"os"
)

var Options struct {
	Addr    string
	BaseURL string
}

func Run() {
	flag.StringVar(&Options.Addr, "a", "localhost:8080", "http server address")
	flag.StringVar(&Options.BaseURL, "b", "http://localhost:8080", "base url")

	flag.Parse()

	if servAddr := os.Getenv("SERVER_ADDRESS"); servAddr != "" {
		Options.Addr = servAddr
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		Options.BaseURL = baseURL
	}
}
