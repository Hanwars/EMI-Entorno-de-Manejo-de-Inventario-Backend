package models

type Usuario struct {
	UsuarioID  int32   `json:"usuario_id"`
	RolID      int32   `json:"rol_id"`
	Nombre     string  `json:"nombre"`
	Contrasena string  `json:"contrasena"`
	Email      *string `json:"email"`
}