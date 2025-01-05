package formRepo

import (
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/models"
	"github.com/bklv-kirill/go-site-form-checker/pkg/repo"
)

type FormSqlRepo struct {
	*repo.SqlRepo
}

func NewSqlRepo(cfg *config.Config) (*FormSqlRepo, error) {
	sqlRepo, err := repo.NewSqlRepo(cfg)
	if err != nil {
		return nil, err
	}

	return &FormSqlRepo{
		sqlRepo,
	}, nil
}

func (formSqlRepo *FormSqlRepo) GetAll() ([]models.Form, error) {
	var fs []models.Form

	var q string = "SELECT * FROM forms"
	if err := formSqlRepo.Db.Select(&fs, q); err != nil {
		return nil, err
	}

	return fs, nil
}

func (formSqlRepo *FormSqlRepo) GetAllWithInputs() ([]models.Form, error) {
	f, err := formSqlRepo.GetAll()
	if err != nil {
		return nil, err
	}

	for i := range f {
		var q string = fmt.Sprintf("SELECT * FROM inputs WHERE form_id = '%d'", f[i].Id)
		if err = formSqlRepo.Db.Select(&f[i].Inputs, q); err != nil {
			return nil, err
		}
	}

	return f, nil
}
