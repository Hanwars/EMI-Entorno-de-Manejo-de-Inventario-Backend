package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

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

type updateUserRequest struct {
	UsuarioID  int32   `json:"usuarioIdId" binding:"required"`
	RolID      int32   `json:"rolId" binding:"required"`
	Nombre     string  `json:"nombre" binding:"required"`
	Email      *string `json:"email" binding:"required"`
}

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
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Credenciales inválidas"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ok, err := security.CheckPassword(user.Contrasena, req.Contrasena)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Credenciales inválidas"})
		return
	}

	//GENERACION DEL TOKEN
	accessToken, err := server.tokenBuilder.CreateToken(user.UsuarioID, user.RolID, user.Nombre, time.Hour*1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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

	contraseña_hasheada, err := security.GenerarHash(req.Contrasena)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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

	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if authPayload.UserID == req.UsuarioID && authPayload.Role != req.RolID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "No puedes cambiar tu propio rol",
		})
		return
	}



	_, err := server.userRepository.ActualizarUsuario(ctx, req.UsuarioID, req.RolID, req.Nombre, *req.Email)
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

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Usuario actualizado correctamente",
	})
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
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Usuario eliminado correctamente",
	})
}
