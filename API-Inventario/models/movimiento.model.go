package models

import "time"

type Movimiento struct {
	MovimientoID int32      `json:"movimiento_id"`
	UsuarioID    int32      `json:"usuario_id"`
	ProveedorID  *int32     `json:"proveedor_id"`
	Tipo         string     `json:"tipo"`
	Fecha        time.Time  `json:"fecha"`
}