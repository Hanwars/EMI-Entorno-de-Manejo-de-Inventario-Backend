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

// NombreRolExiste indica si otro rol, distinto de excluirID, ya tiene ese nombre.
// Al crear se pasa 0 para comparar con todos los roles.
// La comparación no distingue mayúsculas por la intercalación de la base de datos (Modern_Spanish_CI_AS).
func (r *RoleRepository) NombreRolExiste(ctx context.Context, nombre string, excluirID int32) (bool, error) {

	var total int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM dbo.roles WHERE nombre = @p1 AND role_id <> @p2",
		nombre,
		excluirID,
	).Scan(&total)

	return total > 0, err
}

// ExisteRol indica si existe un rol con el identificador recibido.
func (r *RoleRepository) ExisteRol(ctx context.Context, roleID int32) (bool, error) {

	var total int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM dbo.roles WHERE role_id = @p1",
		roleID,
	).Scan(&total)

	return total > 0, err
}

// ActualizarRol cambia el nombre de un rol. Solo modifica el nombre: el role_id no cambia,
// por lo que los usuarios asociados al rol se mantienen.
// Devuelve sql.ErrNoRows si el rol no existe.
func (r *RoleRepository) ActualizarRol(ctx context.Context, roleID int32, nombre string) (*models.Rol, error) {

	query := `
		UPDATE dbo.roles
		SET nombre = @p2
		OUTPUT INSERTED.role_id, INSERTED.nombre
		WHERE role_id = @p1
	`

	var rol models.Rol

	err := r.db.QueryRowContext(ctx, query, roleID, nombre).Scan(
		&rol.RoleID,
		&rol.Nombre,
	)
	if err != nil {
		return nil, err
	}

	return &rol, nil
}
