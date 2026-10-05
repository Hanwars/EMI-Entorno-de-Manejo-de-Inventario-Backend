package models

type Proveedor struct {
	ProveedorID int32   `json:"proveedor_id"`
	Nombre     string  `json:"nombre"`
	Telefono   *string `json:"telefono"`
	Email      *string `json:"email"`
}