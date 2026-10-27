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

// deleteUserURI identifica al usuario a eliminar: DELETE api/v1/users/:usuarioId
type deleteUserURI struct {
	UsuarioID int32 `uri:"usuarioId" binding:"required,min=1"`
}

// Código de error de SQL Server cuando una operación viola una llave foránea.
const errorSQLLlaveForanea = 547

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
		errorSolicitud(ctx, err)
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

	// Se devuelve la representación pública (sin contraseña) y una lista vacía si no hay usuarios.
	respuesta := make([]usuarioResponse, 0, len(users))
	for _, u := range users {
		respuesta = append(respuesta, nuevoUsuarioResponse(&u))
	}

	ctx.JSON(http.StatusOK, respuesta)
}

// createUser es el handler para crear un nuevo usuario.
func (server *Server) createUser(ctx *gin.Context) {

	var req createUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorSolicitud(ctx, err)
		return
	}

	// Validación de la información recibida, con las mismas reglas que la actualización.
	if req.RolID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El rol no es válido"})
		return
	}

	nombre, mensaje := validarNombreUsuario(req.Nombre)
	if mensaje != "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": mensaje})
		return
	}
	req.Nombre = nombre

	if req.Email != nil {
		email, mensaje := validarEmailUsuario(*req.Email)
		if mensaje != "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": mensaje})
			return
		}
		req.Email = &email
	}

	existe, err := server.userRepository.RolExiste(ctx, req.RolID)
	if err != nil {
		internalServerError(ctx, err)
		return
	}
	if !existe {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El rol no existe"})
		return
	}

	// El correo es único. Se comprueba aquí porque la base de datos no tiene un índice único sobre él.
	if req.Email != nil {
		enUso, err := server.userRepository.EmailEnUso(ctx, *req.Email, 0)
		if err != nil {
			internalServerError(ctx, err)
			return
		}
		if enUso {
			ctx.JSON(http.StatusConflict, gin.H{"error": "El correo ya está registrado"})
			return
		}
	}

	// HashPassword aplica las reglas definidas para bcrypt: costo 12 y contraseña de 8 caracteres a 72 bytes.
	contraseña_hasheada, err := security.HashPassword(req.Contrasena)
	if err != nil {
		if errors.Is(err, security.ErrContrasenaInvalida) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "La contraseña debe tener al menos 8 caracteres y no superar los 72 bytes",
			})
			return
		}
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
			// Respaldo por si la base de datos tiene un índice único sobre el correo.
			if sqlErr.Number == 2601 || sqlErr.Number == 2627 {
				ctx.JSON(http.StatusConflict, gin.H{
					"error": "El correo ya está registrado",
				})
				return
			}
			// Respaldo: el rol dejó de existir entre la comprobación y el registro.
			if sqlErr.Number == errorSQLLlaveForanea {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "El rol no existe"})
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
		errorSolicitud(ctx, err)
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

	ctx.JSON(http.StatusOK, nuevoUsuarioResponse(usuario))
}

// nuevoUsuarioResponse convierte un usuario en su representación pública, sin la contraseña.
func nuevoUsuarioResponse(u *models.Usuario) usuarioResponse {
	return usuarioResponse{
		UsuarioID: u.UsuarioID,
		RolID:     u.RolID,
		Nombre:    u.Nombre,
		Email:     u.Email,
	}
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
		nombre, mensaje := validarNombreUsuario(*req.Nombre)
		if mensaje != "" {
			return mensaje
		}
		req.Nombre = &nombre
	}

	if req.Email != nil {
		email, mensaje := validarEmailUsuario(*req.Email)
		if mensaje != "" {
			return mensaje
		}
		req.Email = &email
	}

	return ""
}

// validarNombreUsuario quita los espacios sobrantes y comprueba la longitud del nombre.
// Devuelve el nombre limpio y un mensaje de error, o una cadena vacía si es válido.
func validarNombreUsuario(nombre string) (string, string) {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return "", "El nombre no puede estar vacío"
	}
	if utf8.RuneCountInString(nombre) > maxNombreUsuario {
		return "", fmt.Sprintf("El nombre no puede superar los %d caracteres", maxNombreUsuario)
	}
	return nombre, ""
}

// validarEmailUsuario quita los espacios sobrantes y comprueba la longitud y el formato del correo.
// Devuelve el correo limpio y un mensaje de error, o una cadena vacía si es válido.
func validarEmailUsuario(email string) (string, string) {
	email = strings.TrimSpace(email)
	if utf8.RuneCountInString(email) > maxEmailUsuario {
		return "", fmt.Sprintf("El correo no puede superar los %d caracteres", maxEmailUsuario)
	}
	// ParseAddress acepta también "Nombre <correo>", por eso se exige que el resultado sea idéntico.
	direccion, err := mail.ParseAddress(email)
	if err != nil || direccion.Address != email {
		return "", "El correo no tiene un formato válido"
	}
	return email, ""
}

// deleteUser elimina un usuario identificado en la URI: DELETE api/v1/users/:usuarioId
func (server *Server) deleteUser(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	var uri deleteUserURI
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Identificador de usuario inválido"})
		return
	}

	if authPayload.UserID == uri.UsuarioID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "No puedes eliminar tu propia cuenta",
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

	// Integridad: un usuario con movimientos no se elimina para no perder el historial del inventario.
	tieneMovimientos, err := server.userRepository.TieneMovimientos(ctx, uri.UsuarioID)
	if err != nil {
		internalServerError(ctx, err)
		return
	}
	if tieneMovimientos {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "El usuario tiene movimientos registrados y no se puede eliminar",
		})
		return
	}

	eliminado, err := server.userRepository.EliminarUsuario(ctx, uri.UsuarioID)
	if err != nil {
		var sqlErr mssql.Error
		// Respaldo: la base de datos rechaza la eliminación si quedan registros relacionados.
		if errors.As(err, &sqlErr) && sqlErr.Number == errorSQLLlaveForanea {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "El usuario tiene registros relacionados y no se puede eliminar",
			})
			return
		}

		internalServerError(ctx, err)
		return
	}
	if !eliminado {
		// El usuario fue eliminado por otra petición entre la comprobación y el borrado.
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Usuario eliminado correctamente",
	})
}
