package models

type Usuario struct {
	UsuarioID  int32   `json:"usuarioId"`
	RolID      int32   `json:"rolId"`
	Nombre     string  `json:"nombre"`
	Contrasena string  `json:"contrasena"`
	Email      *string `json:"email"`
}