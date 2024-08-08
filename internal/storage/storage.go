package storage

type URLStore struct {
	Urls map[string]string
}

func NewURLStore() *URLStore {
	return &URLStore{Urls: make(map[string]string)}
}
