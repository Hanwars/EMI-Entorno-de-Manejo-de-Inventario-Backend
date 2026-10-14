package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
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


type usuarioResponse struct {
	UsuarioID int32   `json:"usuarioId"`
	RolID     int32   `json:"rolId"`
	Nombre    string  `json:"nombre"`
	Email     *string `json:"email"`
}

// updateUserRequest: la contraseña es opcional; si no se envía, se conserva la actual.
type updateUserRequest struct {
	RolID      int32   `json:"rolId" binding:"required"`
	Nombre     string  `json:"nombre" binding:"required"`
	Email      *string `json:"email" binding:"omitempty,email"`
	Contrasena *string `json:"contrasena"`
}

func nuevoUsuarioResponse(u models.Usuario) usuarioResponse {
	return usuarioResponse{
		UsuarioID: u.UsuarioID,
		RolID:     u.RolID,
		Nombre:    u.Nombre,
		Email:     u.Email,
	}
}


func parseUserID(ctx *gin.Context) (int32, bool) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id de usuario inválido"})
		return 0, false
	}
	return int32(id), true
}


func errorInterno(ctx *gin.Context, err error) {
	log.Printf("error interno en %s %s: %v", ctx.Request.Method, ctx.FullPath(), err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
}

// GET /api/v1/users
func (server *Server) getUsers(ctx *gin.Context) {
	usuarios, err := server.userRepository.ListarUsuarios(ctx)
	if err != nil {
		errorInterno(ctx, err)
		return
	}

	resp := make([]usuarioResponse, 0, len(usuarios))
	for _, u := range usuarios {
		resp = append(resp, nuevoUsuarioResponse(u))
	}

	ctx.JSON(http.StatusOK, resp)
}

// GET /api/v1/users/:id
func (server *Server) getUser(ctx *gin.Context) {
	id, ok := parseUserID(ctx)
	if !ok {
		return
	}

	usuario, err := server.userRepository.ObtenerUsuarioPorID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
			return
		}
		errorInterno(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, nuevoUsuarioResponse(*usuario))
}

// PUT /api/v1/users/:id
func (server *Server) updateUser(ctx *gin.Context) {
	id, ok := parseUserID(ctx)
	if !ok {
		return
	}

	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	usuario := models.Usuario{
		UsuarioID: id,
		RolID:     req.RolID,
		Nombre:    req.Nombre,
		Email:     req.Email,
	}

	// Solo se cambia la contraseña si viene en la petición y no está vacía.
	actualizarContrasena := req.Contrasena != nil && *req.Contrasena != ""
	if actualizarContrasena {
		hash, err := security.GenerarHash(*req.Contrasena)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		usuario.Contrasena = hash
	}

	if err := server.userRepository.ActualizarUsuario(ctx, &usuario, actualizarContrasena); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
			return
		}
		errorInterno(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, nuevoUsuarioResponse(usuario))
}

// DELETE /api/v1/users/:id
func (server *Server) deleteUser(ctx *gin.Context) {
	id, ok := parseUserID(ctx)
	if !ok {
		return
	}

	if err := server.userRepository.EliminarUsuario(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
			return
		}

		
		var sqlErr mssql.Error
		if errors.As(err, &sqlErr) && sqlErr.Number == 547 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "no se puede eliminar: el usuario tiene registros asociados"})
			return
		}

		errorInterno(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
