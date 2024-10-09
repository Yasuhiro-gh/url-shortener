package storage

import "context"

type URLStorages interface {
	Get(string) (Store, bool)
	GetUserID() int
	GetUserURLS(ctx context.Context, uid int) ([]Store, error)
	Set(string, *Store) error
	Delete(string, int) error
}
