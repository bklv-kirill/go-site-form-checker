package formStorage

import (
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/models"
	"github.com/bklv-kirill/go-site-form-checker/pkg/storage"
	_ "github.com/go-sql-driver/mysql"
)

type FormSqlStorage struct {
	*storage.SqlStorage
}

func NewFormSqlStorage(cfg *config.Config) (*FormSqlStorage, error) {
	sqlStrg, err := storage.NewSqlStorage(cfg)
	if err != nil {
		return nil, err
	}

	return &FormSqlStorage{
		sqlStrg,
	}, nil
}

func (formSqlStrg *FormSqlStorage) GetAll() ([]form.Form, error) {
	var fs []form.Form

	var q string = "SELECT * FROM forms"
	if err := formSqlStrg.Db.Select(&fs, q); err != nil {
		return nil, err
	}

	return fs, nil
}

func (formSqlStrg *FormSqlStorage) GetAllWithInputs() ([]form.Form, error) {
	fs, err := formSqlStrg.GetAll()
	if err != nil {
		return fs, err
	}

	for i := range fs {
		var q string = fmt.Sprintf("SELECT * FROM inputs WHERE form_id = '%d'", fs[i].Id)
		if err = formSqlStrg.Db.Select(&fs[i].Inputs, q); err != nil {
			return fs, err
		}
	}

	return fs, nil
}
