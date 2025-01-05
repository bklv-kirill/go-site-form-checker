package repo

import (
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type SqlRepo struct {
	Db *sqlx.DB
}

func NewSqlRepo(cfg *config.Config) (*SqlRepo, error) {
	db, err := sqlx.Connect(cfg.DbCon, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.DbUser, cfg.DbPass, cfg.DbHost, cfg.DbPort, cfg.DbName))
	if err != nil {
		return nil, err
	}

	return &SqlRepo{
		Db: db,
	}, nil
}
