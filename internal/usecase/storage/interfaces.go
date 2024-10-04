package storage

import "context"

type URLStorages interface {
	Get(string) (Store, bool)
	Set(string, *Store) error
	GetUserID() int
	GetUserURLS(ctx context.Context, uid int) ([]Store, error)
}
