package config

import "flag"

var Options struct {
	Addr     string
	BaseAddr string
}

func init() {
	flag.StringVar(&Options.Addr, "a", "localhost:8080", "http server address")
	flag.StringVar(&Options.BaseAddr, "b", "http://localhost:8080", "base server address")
}
