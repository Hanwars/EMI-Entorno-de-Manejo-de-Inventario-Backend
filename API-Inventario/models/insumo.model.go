package models

type Insumo struct {
	InsumoID    int32  `json:"insumo_id"`
	Nombre      string `json:"nombre"`
	Tipo        string `json:"tipo"`
	Cantidad    int32  `json:"cantidad"`
	CategoriaID *int32 `json:"categoria_id"`
}