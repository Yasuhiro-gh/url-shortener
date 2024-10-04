package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	"github.com/Yasuhiro-gh/url-shortener/internal/usecase/storage"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
	"strings"
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB() *PostgresDB {
	return &PostgresDB{}
}

func (pdb *PostgresDB) Get(shortURL string) (storage.Store, bool) {
	qr := pdb.DB.QueryRow("SELECT original_url, user_id FROM urls WHERE short_url = $1", shortURL)
	if qr.Err() != nil {
		return storage.Store{}, false
	}
	var store storage.Store
	err := qr.Scan(&store.OriginalURL, &store.UserID)
	if err != nil {
		return storage.Store{}, false
	}
	return store, true
}

func (pdb *PostgresDB) Set(shortURL string, store *storage.Store) error {
	_, err := pdb.DB.Exec("INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		shortURL, store.OriginalURL, store.UserID)
	if err != nil && strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
		return errors.New(pgerrcode.UniqueViolation)
	}
	return err
}

func (pdb *PostgresDB) GetUserURLS(ctx context.Context, userID int) ([]storage.Store, error) {
	qc, err := pdb.DB.QueryContext(ctx, "SELECT original_url, short_url FROM urls WHERE user_id = $1", userID)
	if qc.Err() != nil {
		return nil, qc.Err()
	}
	if err != nil {
		return nil, err
	}
	defer qc.Close()
	var urls []storage.Store
	for qc.Next() {
		var store storage.Store
		err := qc.Scan(&store.OriginalURL, &store.ShortURL)
		if err != nil {
			return nil, err
		}
		urls = append(urls, store)
	}
	return urls, nil
}

func (pdb *PostgresDB) GetUserID() int {
	qr := pdb.DB.QueryRow("SELECT user_id FROM urls GROUP BY user_id ORDER BY user_id DESC LIMIT 1")
	if qr.Err() != nil {
		return 0
	}
	var userID int
	err := qr.Scan(&userID)
	if err != nil {
		return 0
	}
	return userID
}

func isTableExist(pdb *PostgresDB, table string) bool {
	var n int
	err := pdb.DB.QueryRow("SELECT 1 FROM information_schema.tables WHERE table_name = $1", table).Scan(&n)
	return err == nil
}

func CreateDatabaseTable(pdb *PostgresDB) error {
	if isTableExist(pdb, "urls") {
		return nil
	}
	_, err := pdb.DB.Exec(`CREATE TABLE urls("original_url" TEXT UNIQUE, "short_url" TEXT, "user_id" INTEGER)`)
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDB) OpenConnection() error {
	db, err := sql.Open("pgx", config.Options.DatabaseDSN)
	if err != nil {
		return err
	}
	pdb.DB = db
	return nil
}

func (pdb *PostgresDB) CloseConnection() error {
	err := pdb.DB.Close()
	return err
}
