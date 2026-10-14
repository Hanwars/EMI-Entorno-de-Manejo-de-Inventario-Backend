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

// ListarUsuarios devuelve todos los usuarios (sin la contraseña).
func (r *UserRepository) ListarUsuarios(ctx context.Context) ([]models.Usuario, error) {
	query := `
		SELECT usuario_id, rol_id, nombre, email
		FROM dbo.usuarios
		ORDER BY usuario_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usuarios := []models.Usuario{}
	for rows.Next() {
		var u models.Usuario
		if err := rows.Scan(&u.UsuarioID, &u.RolID, &u.Nombre, &u.Email); err != nil {
			return nil, err
		}
		usuarios = append(usuarios, u)
	}

	return usuarios, rows.Err()
}

// ObtenerUsuarioPorID devuelve un usuario (sin la contraseña).

func (r *UserRepository) ObtenerUsuarioPorID(ctx context.Context, id int32) (*models.Usuario, error) {
	query := `
		SELECT usuario_id, rol_id, nombre, email
		FROM dbo.usuarios
		WHERE usuario_id = @p1
	`

	var u models.Usuario
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.UsuarioID, &u.RolID, &u.Nombre, &u.Email)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// ActualizarUsuario modifica rol, nombre y email. Si actualizarContrasena es true,

func (r *UserRepository) ActualizarUsuario(ctx context.Context, user *models.Usuario, actualizarContrasena bool) error {
	var (
		result sql.Result
		err    error
	)

	if actualizarContrasena {
		query := `
			UPDATE dbo.usuarios
			SET rol_id = @p1, nombre = @p2, email = @p3, contraseña = @p4
			WHERE usuario_id = @p5
		`
		result, err = r.db.ExecContext(ctx, query, user.RolID, user.Nombre, user.Email, user.Contrasena, user.UsuarioID)
	} else {
		query := `
			UPDATE dbo.usuarios
			SET rol_id = @p1, nombre = @p2, email = @p3
			WHERE usuario_id = @p4
		`
		result, err = r.db.ExecContext(ctx, query, user.RolID, user.Nombre, user.Email, user.UsuarioID)
	}
	if err != nil {
		return err
	}

	return verificarFilaAfectada(result)
}

// EliminarUsuario borra un usuario por su id.
func (r *UserRepository) EliminarUsuario(ctx context.Context, id int32) error {
	query := `DELETE FROM dbo.usuarios WHERE usuario_id = @p1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return verificarFilaAfectada(result)
}


func verificarFilaAfectada(result sql.Result) error {
	filas, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if filas == 0 {
		return sql.ErrNoRows
	}
	return nil
}
