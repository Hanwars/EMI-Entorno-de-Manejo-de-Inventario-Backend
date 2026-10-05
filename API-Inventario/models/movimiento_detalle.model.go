package models

type MovimientoDetalle struct {
	DetalleID    int32 `json:"detalle_id"`
	MovimientoID int32 `json:"movimiento_id"`
	InsumoID     int32 `json:"insumo_id"`
	Cantidad     int32 `json:"cantidad"`
}