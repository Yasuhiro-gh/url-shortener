package db

import (
	"database/sql"
	"github.com/Yasuhiro-gh/url-shortener/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB() *PostgresDB {
	return &PostgresDB{}
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
