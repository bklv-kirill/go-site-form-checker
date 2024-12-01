package storage

import (
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/jmoiron/sqlx"
)

type SqlStorage struct {
	Db *sqlx.DB
}

func NewSqlStorage(cfg *config.Config) (*SqlStorage, error) {
	db, err := sqlx.Connect(cfg.DbCon, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.DbUser, cfg.DbPass, cfg.DbHost, cfg.DbPort, cfg.DbName))
	if err != nil {
		return nil, err
	}

	return &SqlStorage{
		Db: db,
	}, nil
}
