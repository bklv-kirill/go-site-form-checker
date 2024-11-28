package storage

import (
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type FormSQLStorage struct {
	Db *sqlx.DB
}

func New(cfg *config.Config) (*FormSQLStorage, error) {
	db, err := sqlx.Connect(cfg.DbCon, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.DbUser, cfg.DbPass, cfg.DbHost, cfg.DbPort, cfg.DbName))
	if err != nil {
		return nil, err
	}

	return &FormSQLStorage{
		Db: db,
	}, nil
}

func (strg *FormSQLStorage) GetAll() ([]form.Form, error) {
	var fs []form.Form

	var q string = "SELECT * FROM forms"
	if err := strg.Db.Select(&fs, q); err != nil {
		return nil, err
	}

	return fs, nil
}

func (strg *FormSQLStorage) GetAllWithInputs() ([]form.Form, error) {
	fs, err := strg.GetAll()
	if err != nil {
		return fs, err
	}

	for i := range fs {
		var q string = fmt.Sprintf("SELECT * FROM inputs WHERE form_id = '%d'", fs[i].Id)
		if err = strg.Db.Select(&fs[i].Inputs, q); err != nil {
			return fs, err
		}
	}

	return fs, nil
}
