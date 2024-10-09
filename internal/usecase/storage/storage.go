package storage

import (
	"context"
	"errors"
)

type Store struct {
	OriginalURL string
	ShortURL    string
	UserID      int  `json:"-"`
	DeletedFlag bool `json:"is_deleted"`
}

type URLStorage struct {
	urls map[string]Store
}

func NewURLStorage() *URLStorage {
	return &URLStorage{urls: make(map[string]Store)}
}

func (us *URLStorage) Get(key string) (Store, bool) {
	value, ok := us.urls[key]
	return value, ok
}

func (us *URLStorage) GetUserID() int {
	uid := 0
	for _, store := range us.urls {
		if store.UserID > uid {
			uid = store.UserID
		}
	}
	return uid
}

func (us *URLStorage) GetUserURLS(ctx context.Context, uid int) ([]Store, error) {
	urlStores := make([]Store, 0)
	for _, store := range us.urls {
		if store.UserID == uid {
			urlStores = append(urlStores, store)
		}
	}
	return urlStores, nil
}

func (us *URLStorage) Set(key string, value *Store) error {
	us.urls[key] = *value
	return nil
}

func (us *URLStorage) Delete(key string, userID int) error {
	if url, ok := us.urls[key]; ok && url.UserID == userID {
		url.DeletedFlag = true
		us.urls[key] = url
		return nil
	}
	return errors.New("wrong user")
}

type URLS struct {
	storage URLStorages
}

func NewURLS(us URLStorages) *URLS {
	return &URLS{storage: us}
}

func (us *URLS) Get(url string) (Store, bool) {
	value, exist := us.storage.Get(url)
	return value, exist
}

func (us *URLS) GetUserID() int {
	return us.storage.GetUserID()
}

func (us *URLS) GetUserURLS(ctx context.Context, uid int) ([]Store, error) {
	return us.storage.GetUserURLS(ctx, uid)
}

func (us *URLS) Set(shortURL string, value *Store) error {
	return us.storage.Set(shortURL, value)
}

func (us *URLS) Delete(shortURL string, userID int) error {
	return us.storage.Delete(shortURL, userID)
}
