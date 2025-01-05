package models

type Input struct {
	Id        int    `db:"id" json:"id"`
	FormId    int    `db:"form_id" json:"form_id"`
	Selector  string `db:"selector" json:"selector"`
	Value     string `db:"value" json:"value"`
	ForUuid   bool   `db:"for_uuid" json:"for_uuid"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
}
