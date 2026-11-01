package repository

import (
	"context"
	"database/sql"

	"rest/models"
)

type RolRepository struct {
	db *sql.DB
}

func NewRolRepository(db *sql.DB) *RolRepository {
	return &RolRepository{
		db: db,
	}
}

func (r *RolRepository) ObtenerRoles(ctx context.Context) ([]models.Rol, error) {

	stmt, err := r.db.PrepareContext(
		ctx,
		"SELECT role_id, nombre FROM dbo.roles",
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Rol

	for rows.Next() {

		var rol models.Rol

		err := rows.Scan(
			&rol.RoleID,
			&rol.Nombre,
		)

		if err != nil {
			return nil, err
		}

		roles = append(roles, rol)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *RolRepository) ObtenerRol(ctx context.Context, roleID int32) (*models.Rol, error) {

	stmt, err := r.db.PrepareContext(
		ctx,
		"SELECT role_id, nombre FROM dbo.roles WHERE role_id = @p1",
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var rol models.Rol

	err = stmt.QueryRowContext(ctx, roleID).Scan(
		&rol.RoleID,
		&rol.Nombre,
	)

	if err != nil {
		return nil, err
	}

	return &rol, nil
}

func (r *RolRepository) CrearRol(ctx context.Context, rol *models.Rol) (*models.Rol, error) {

	query := `
		INSERT INTO dbo.roles (
			nombre
		)
		OUTPUT INSERTED.role_id
		VALUES (@p1)
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		rol.Nombre,
	).Scan(&rol.RoleID)

	if err != nil {
		return nil, err
	}

	return rol, nil
}

// EliminarRol elimina un rol por su identificador.
func (r *RolRepository) EliminarRol(ctx context.Context, roleID int32) error {

	stmt, err := r.db.PrepareContext(
		ctx,
		"DELETE FROM dbo.roles WHERE role_id = @p1",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, roleID)

	return err
}

// ActualizarRol actualiza el nombre de un rol.
func (r *RolRepository) ActualizarRol(
	ctx context.Context,
	roleID int32,
	nombre string,
) (*models.Rol, error) {

	stmt, err := r.db.PrepareContext(
		ctx,
		`UPDATE dbo.roles
		SET nombre = @p2
		OUTPUT INSERTED.role_id, INSERTED.nombre
		WHERE role_id = @p1`,
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var rol models.Rol

	err = stmt.QueryRowContext(
		ctx,
		roleID,
		nombre,
	).Scan(
		&rol.RoleID,
		&rol.Nombre,
	)

	if err != nil {
		return nil, err
	}

	return &rol, nil
}
