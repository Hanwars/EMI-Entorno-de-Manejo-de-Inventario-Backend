package repository

import (
	"context"
	"database/sql"

	"rest/models"
)

type UserRepository struct {
	db *sql.DB
}


func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Login(ctx context.Context, email string) (*models.Usuario, error) {

	query := `
		SELECT usuario_id, rol_id, nombre, contraseña, email
		FROM dbo.usuarios
		WHERE email = @p1
	`

	row := r.db.QueryRowContext(ctx, query, email)

	var user models.Usuario

	err := row.Scan(
		&user.UsuarioID,
		&user.RolID,
		&user.Nombre,
		&user.Contrasena,
		&user.Email,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ObtenerRoles ejecuta el procedimiento almacenado paObtenerRoles.
func (r *UserRepository) ObtenerUsuarios(ctx context.Context) ([]models.Usuario, error) {

	stmt, err := r.db.PrepareContext(
		ctx,
		"SELECT usuario_id, rol_id, nombre, email From usuarios",
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

	var usuarios []models.Usuario

	for rows.Next() {

		var usuario models.Usuario

		err := rows.Scan(
			&usuario.UsuarioID,
			&usuario.RolID,
			&usuario.Nombre,
			&usuario.Email,
		)

		if err != nil {
			return nil, err
		}

		usuarios = append(usuarios, usuario)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return usuarios, nil
}

// ObtenerRol ejecuta el procedimiento almacenado paObtenerRol.
func (r *UserRepository) ObtenerUsuario(ctx context.Context, usuarioID int32) (*models.Usuario, error) {

	stmt, err := r.db.PrepareContext(
		ctx,
		"SELECT usuario_id, rol_id, nombre, email From usuarios where usuario_id = @p1",
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var usuario models.Usuario

	err = stmt.QueryRowContext(ctx, usuarioID).Scan(
		&usuario.UsuarioID,
		&usuario.RolID,
		&usuario.Nombre,
		&usuario.Email,
	)

	if err != nil {
		return nil, err
	}

	return &usuario, nil
}

func (r *UserRepository) CrearUsuario(ctx context.Context, user *models.Usuario) (*models.Usuario, error) {

    query := `
        INSERT INTO dbo.usuarios (
            rol_id,
            nombre,
            contraseña,
            email
        )
        OUTPUT INSERTED.usuario_id
        VALUES (@p1, @p2, @p3, @p4)
    `

    err := r.db.QueryRowContext(
        ctx,
        query,
        user.RolID,
        user.Nombre,
        user.Contrasena,
        user.Email,
    ).Scan(&user.UsuarioID)

    if err != nil {
        return nil, err
    }

    return user, nil
}

// EliminarRol ejecuta el procedimiento almacenado paEliminarRol.
func (r *UserRepository) EliminarUsuario(ctx context.Context, usuario int32) error {

	stmt, err := r.db.PrepareContext(
		ctx,
		"DELETE FROM dbo.usuarios WHERE usuario_id = @p1",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, usuario)

	return err
}


// ActualizarUsuario actualiza solo los campos recibidos: un valor nil conserva el dato actual.
// Devuelve sql.ErrNoRows si el usuario no existe.
func (r *UserRepository) ActualizarUsuario(
	ctx context.Context,
	userID int32,
	rol_id *int32,
	nombre *string,
	email *string,
) (*models.Usuario, error) {

	stmt, err := r.db.PrepareContext(
		ctx,
		`UPDATE dbo.usuarios
		SET rol_id = COALESCE(@p2, rol_id),
		    nombre = COALESCE(@p3, nombre),
		    email  = COALESCE(@p4, email)
		OUTPUT INSERTED.usuario_id, INSERTED.rol_id, INSERTED.nombre, INSERTED.email
		WHERE usuario_id = @p1`,
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var usuario models.Usuario

	err = stmt.QueryRowContext(
		ctx,
		userID,
		rol_id,
		nombre,
		email,
	).Scan(
		&usuario.UsuarioID,
		&usuario.RolID,
		&usuario.Nombre,
		&usuario.Email,
	)

	if err != nil {
		return nil, err
	}

	return &usuario, nil
}

// RolExiste indica si existe un rol con el identificador recibido.
func (r *UserRepository) RolExiste(ctx context.Context, rolID int32) (bool, error) {

	var total int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM dbo.roles WHERE role_id = @p1",
		rolID,
	).Scan(&total)

	return total > 0, err
}

// EmailEnUso indica si otro usuario, distinto de usuarioID, ya tiene registrado el correo.
func (r *UserRepository) EmailEnUso(ctx context.Context, email string, usuarioID int32) (bool, error) {

	var total int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM dbo.usuarios WHERE email = @p1 AND usuario_id <> @p2",
		email,
		usuarioID,
	).Scan(&total)

	return total > 0, err
}
