package form

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

type Input struct {
	Id        int    `db:"id" json:"id"`
	FormId    int    `db:"form_id" json:"form_id"`
	Selector  string `db:"selector" json:"selector"`
	Value     string `db:"value" json:"value"`
	ForUuid   bool   `db:"for_uuid" json:"for_uuid"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
}
