package utils

import (
	"crypto/sha256"
	"fmt"
	"net/url"
)

func IsValidURL(urlToValid string) bool {
	_, err := url.ParseRequestURI(urlToValid)
	return err == nil
}

func IsHashExist(hash string, urls map[string]string) bool {
	_, ok := urls[hash]
	return ok
}

func HashURL(url string) string {
	hash := sha256.New()
	hash.Write([]byte(url))
	return fmt.Sprintf("%x", hash.Sum(nil))[:8]
}
