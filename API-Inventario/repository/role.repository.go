package repository

import (
	"context"
	"database/sql"

	"rest/models"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

// CrearRol registra un nuevo rol y lo devuelve con el identificador generado por la base de datos.
func (r *RoleRepository) CrearRol(ctx context.Context, nombre string) (*models.Rol, error) {

	query := `
		INSERT INTO dbo.roles (nombre)
		OUTPUT INSERTED.role_id, INSERTED.nombre
		VALUES (@p1)
	`

	var rol models.Rol

	err := r.db.QueryRowContext(ctx, query, nombre).Scan(
		&rol.RoleID,
		&rol.Nombre,
	)
	if err != nil {
		return nil, err
	}

	return &rol, nil
}

// NombreRolExiste indica si ya existe un rol con ese nombre.
// La comparación no distingue mayúsculas por la intercalación de la base de datos (Modern_Spanish_CI_AS).
func (r *RoleRepository) NombreRolExiste(ctx context.Context, nombre string) (bool, error) {

	var total int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM dbo.roles WHERE nombre = @p1",
		nombre,
	).Scan(&total)

	return total > 0, err
}
