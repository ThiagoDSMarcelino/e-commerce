package catalog

type Category struct {
	ID    string
	Label string
}

var Categories = []Category{
	{ID: "papelaria", Label: "Papelaria"},
	{ID: "escritorio", Label: "Escritório"},
	{ID: "escolar", Label: "Material escolar"},
	{ID: "informatica", Label: "Informática"},
	{ID: "arte", Label: "Arte e desenho"},
}
