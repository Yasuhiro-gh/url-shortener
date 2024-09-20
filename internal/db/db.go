package db

import (
	"database/sql"
	"errors"
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
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

func (pdb *PostgresDB) Get(shortURL string) (string, bool) {
	qr := pdb.DB.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortURL)
	if qr.Err() != nil {
		return "", false
	}
	var originalURL string
	err := qr.Scan(&originalURL)
	if err != nil {
		return "", false
	}
	return originalURL, true
}

func (pdb *PostgresDB) Set(shortURL string, originalURL string) error {
	_, err := pdb.DB.Exec("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)", shortURL, originalURL)
	if err != nil && strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
		return errors.New(pgerrcode.UniqueViolation)
	}
	return err
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
	_, err := pdb.DB.Exec(`CREATE TABLE urls("original_url" TEXT UNIQUE, "short_url" TEXT)`)
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
