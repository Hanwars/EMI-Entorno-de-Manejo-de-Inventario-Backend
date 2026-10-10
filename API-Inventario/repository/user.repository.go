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
