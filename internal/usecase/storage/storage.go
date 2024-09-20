package storage

type URLStorage struct {
	urls map[string]string
}

func NewURLStorage() *URLStorage {
	return &URLStorage{urls: make(map[string]string)}
}

func (us *URLStorage) Get(key string) (string, bool) {
	value, ok := us.urls[key]
	return value, ok
}

func (us *URLStorage) Set(key string, value string) error {
	us.urls[key] = value
	return nil
}

type URLS struct {
	storage URLStorages
}

func NewURLS(us URLStorages) *URLS {
	return &URLS{storage: us}
}

func (us *URLS) Get(url string) (string, bool) {
	val, exist := us.storage.Get(url)
	return val, exist
}
func (us *URLS) Set(shortURL, originalURL string) error {
	return us.storage.Set(shortURL, originalURL)
}
