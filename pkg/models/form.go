package models

import (
	"fmt"
)

type Form struct {
	Id           int     `db:"id" json:"id"`
	Name         string  `db:"name" json:"name"`
	Url          string  `db:"url" json:"url"`
	ElemForClick string  `db:"element_for_click" json:"element_for_click"`
	ExpElem      string  `db:"expected_element" json:"expected_element"`
	SubmitElem   string  `db:"submit_element" json:"submit_element"`
	ResElem      string  `db:"result_element" json:"result_element"`
	Inputs       []Input `db:"inputs" json:"inputs"`
	CreatedAt    string  `db:"created_at" json:"created_at"`
	UpdatedAt    string  `db:"updated_at" json:"updated_at"`
}

func (f *Form) GetPrevMsg() string {
	return fmt.Sprintf("Название: %s | Ссылка: %s\n", f.Name, f.Url)
}
