package api

import (
	"database/sql"
	"net/http"
	"rest/models"

	"rest/security"
	"time"

	"github.com/gin-gonic/gin"
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
