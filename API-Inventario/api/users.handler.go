package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"rest/models"
	"rest/security"

	"github.com/gin-gonic/gin"
	mssql "github.com/microsoft/go-mssqldb"
)

// Cosas del Login
type loginRequest struct {
	Email      string `json:"email" binding:"required"`
	Contrasena string `json:"contrasena" binding:"required"`
}

type LoginResponse struct {
	AccessToken string  `json:"access_token"`
	Payload     payload `json:"payload"`
}

type payload struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Role int32  `json:"role"`
}

// updateUserURI identifica al usuario a actualizar: PUT api/v1/users/:usuarioId
type updateUserURI struct {
	UsuarioID int32 `uri:"usuarioId" binding:"required,min=1"`
}

// updateUserRequest contiene los campos a actualizar. Todos son opcionales:
// solo se modifican los que se envían, el resto conserva su valor actual.
type updateUserRequest struct {
	RolID  *int32  `json:"rolId"`
	Nombre *string `json:"nombre"`
	Email  *string `json:"email"`
}

// usuarioResponse es la representación pública de un usuario (sin la contraseña).
type usuarioResponse struct {
	UsuarioID int32   `json:"usuarioId"`
	RolID     int32   `json:"rolId"`
	Nombre    string  `json:"nombre"`
	Email     *string `json:"email"`
}

// Longitudes máximas definidas en la tabla usuarios.
const (
	maxNombreUsuario = 300
	maxEmailUsuario  = 150
)

type deleteUserRequest struct {
	UsuarioID int32 `json:"usuarioId" binding:"required"`
}

type deleteUserRequestURI struct {
	UsuarioID int32 `uri:"usuarioId" binding:"required"`
}

// Requests para usuarios
type createUserRequest struct {
	RolID      int32   `json:"rolId" binding:"required"`
	Nombre     string  `json:"nombre" binding:"required"`
	Contrasena string  `json:"contrasena" binding:"required"`
	Email      *string `json:"email"`
}

/*Funcion de Login*/
func (server *Server) login(ctx *gin.Context) {
	var req loginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.userRepository.Login(ctx, req.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "Credenciales inválidas",
			})
			return
		}

		internalServerError(ctx, err)
		return
	}

	ok, err := security.CheckPassword(user.Contrasena, req.Contrasena)
	if err != nil {
		internalServerError(ctx, err)
		return
	}
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas"})
		return
	}

	//GENERACION DEL TOKEN
	accessToken, err := server.tokenBuilder.CreateToken(user.UsuarioID, user.RolID, user.Nombre, server.tokenDuration)
	if err != nil {
		internalServerError(ctx, err)
		return
	}
	resp := LoginResponse{
		AccessToken: accessToken,
		Payload: payload{
			ID:   user.UsuarioID,
			Name: user.Nombre,
			Role: user.RolID,
		},
	}
	ctx.JSON(http.StatusOK, resp)

}

// getAllUsers es el handler para obtener todos los usuarios.
func (server *Server) getAllUsers(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	users, err := server.userRepository.ObtenerUsuarios(ctx)
	if err != nil {
		internalServerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, users)
}

// createUser es el handler para crear un nuevo usuario.
func (server *Server) createUser(ctx *gin.Context) {

	var req createUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if len(req.Contrasena) > 72 {
    ctx.JSON(http.StatusBadRequest, gin.H{
        "error": "La contraseña no puede superar los 72 bytes",
    })
    return
}

	contraseña_hasheada, err := security.GenerarHash(req.Contrasena)

	if err != nil {
		internalServerError(ctx, err)
		return
	}

	user := models.Usuario{
		RolID:      req.RolID,
		Nombre:     req.Nombre,
		Contrasena: contraseña_hasheada,
		Email:      req.Email,
	}

	result, err := server.userRepository.CrearUsuario(ctx, &user)

	if err != nil {
		var sqlErr mssql.Error

		if errors.As(err, &sqlErr) {
			if sqlErr.Number == 2601 || sqlErr.Number == 2627 {
				ctx.JSON(http.StatusConflict, gin.H{
					"error": "El correo ya está registrado",
				})
				return
			}
		}

		internalServerError(ctx, err)
		return
	}

	id := result.UsuarioID

	ctx.JSON(http.StatusOK, gin.H{
		"idGenerado": id,
	})
}

// updateUser es el handler para actualizar un usuario.
func (server *Server) updateUser(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	var uri updateUserURI
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Identificador de usuario inválido"})
		return
	}

	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validación de la información recibida.
	if mensaje := validarActualizacionUsuario(&req); mensaje != "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": mensaje})
		return
	}

	if authPayload.UserID == uri.UsuarioID && req.RolID != nil && *req.RolID != authPayload.Role {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "No puedes cambiar tu propio rol",
		})
		return
	}

	// El usuario debe existir.
	if _, err := server.userRepository.ObtenerUsuario(ctx, uri.UsuarioID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
			return
		}
		internalServerError(ctx, err)
		return
	}

	if req.RolID != nil {
		existe, err := server.userRepository.RolExiste(ctx, *req.RolID)
		if err != nil {
			internalServerError(ctx, err)
			return
		}
		if !existe {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "El rol no existe"})
			return
		}
	}

	// El correo es único: no puede pertenecer a otro usuario.
	if req.Email != nil {
		enUso, err := server.userRepository.EmailEnUso(ctx, *req.Email, uri.UsuarioID)
		if err != nil {
			internalServerError(ctx, err)
			return
		}
		if enUso {
			ctx.JSON(http.StatusConflict, gin.H{"error": "El correo ya está registrado"})
			return
		}
	}

	usuario, err := server.userRepository.ActualizarUsuario(ctx, uri.UsuarioID, req.RolID, req.Nombre, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
			return
		}

		var sqlErr mssql.Error
		if errors.As(err, &sqlErr) {
			// Respaldo por si la base de datos tiene un índice único sobre el correo.
			if sqlErr.Number == 2601 || sqlErr.Number == 2627 {
				ctx.JSON(http.StatusConflict, gin.H{
					"error": "El correo ya está registrado",
				})
				return
			}
		}

		internalServerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, usuarioResponse{
		UsuarioID: usuario.UsuarioID,
		RolID:     usuario.RolID,
		Nombre:    usuario.Nombre,
		Email:     usuario.Email,
	})
}

// validarActualizacionUsuario limpia los datos recibidos y devuelve un mensaje de error,
// o una cadena vacía si son válidos.
func validarActualizacionUsuario(req *updateUserRequest) string {
	if req.RolID == nil && req.Nombre == nil && req.Email == nil {
		return "Debe enviar al menos un campo para actualizar"
	}

	if req.RolID != nil && *req.RolID <= 0 {
		return "El rol no es válido"
	}

	if req.Nombre != nil {
		nombre := strings.TrimSpace(*req.Nombre)
		if nombre == "" {
			return "El nombre no puede estar vacío"
		}
		if utf8.RuneCountInString(nombre) > maxNombreUsuario {
			return fmt.Sprintf("El nombre no puede superar los %d caracteres", maxNombreUsuario)
		}
		req.Nombre = &nombre
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if utf8.RuneCountInString(email) > maxEmailUsuario {
			return fmt.Sprintf("El correo no puede superar los %d caracteres", maxEmailUsuario)
		}
		// ParseAddress acepta también "Nombre <correo>", por eso se exige que el resultado sea idéntico.
		direccion, err := mail.ParseAddress(email)
		if err != nil || direccion.Address != email {
			return "El correo no tiene un formato válido"
		}
		req.Email = &email
	}

	return ""
}

// deleteUser elimina un usuario por JSON.
func (server *Server) deleteUserJSON(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	var req deleteUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if authPayload.UserID == req.UsuarioID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "No puedes eliminar tu propia cuenta",
		})
		return
	}

	err := server.userRepository.EliminarUsuario(ctx, req.UsuarioID)

	if err != nil {
		internalServerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Usuario eliminado correctamente",
	})
}

// PETICION DE COCA COLA MASTER
// deleteUserURI LO que hace es que elimina un usuario por medio de la URI.
func (server *Server) deleteUserURI(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	var req deleteUserRequestURI

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if authPayload.UserID == req.UsuarioID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "No puedes eliminar tu propia cuenta",
		})
		return
	}

	err := server.userRepository.EliminarUsuario(ctx, req.UsuarioID)

	if err != nil {
		internalServerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Usuario eliminado correctamente",
	})
}
