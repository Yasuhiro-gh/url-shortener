package storage

type URLStorages interface {
	Get(string) (string, bool)
	Set(string, string) error
}
