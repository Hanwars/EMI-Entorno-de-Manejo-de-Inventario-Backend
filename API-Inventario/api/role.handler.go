
package api

import (
	"database/sql"
	"errors"
	"net/http"

	"rest/models"
	"rest/security"

	"github.com/gin-gonic/gin"
	mssql "github.com/microsoft/go-mssqldb"
)

// Requests para roles
type createRoleRequest struct {
	Nombre string `json:"nombre" binding:"required"`
}

type updateRoleRequest struct {
	RoleID int32  `json:"roleId" binding:"required"`
	Nombre string `json:"nombre" binding:"required"`
}

type deleteRoleRequest struct {
	RoleID int32 `json:"roleId" binding:"required"`
}

type deleteRoleRequestURI struct {
	RoleID int32 `uri:"roleId" binding:"required"`
}

// getAllRoles es el handler para obtener todos los roles.
func (server *Server) getAllRoles(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	roles, err := server.roleRepository.ObtenerRoles(ctx)
	if err != nil {
		internalServerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, roles)
}

// createRole es el handler para crear un nuevo rol.
func (server *Server) createRole(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	var req createRoleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	rol := models.Rol{
		Nombre: req.Nombre,
	}

	result, err := server.roleRepository.CrearRol(ctx, &rol)

	if err != nil {
		internalServerError(ctx, err)
		return
	}

	id := result.RoleID

	ctx.JSON(http.StatusOK, gin.H{
		"idGenerado": id,
	})
}

// updateRole es el handler para actualizar un rol.
func (server *Server) updateRole(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	var req updateRoleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, err := server.roleRepository.ActualizarRol(
		ctx,
		req.RoleID,
		req.Nombre,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Rol no encontrado",
			})
			return
		}

		internalServerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Rol actualizado correctamente",
	})
}

// deleteRole elimina un rol por JSON.
func (server *Server) deleteRoleJSON(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	var req deleteRoleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.roleRepository.EliminarRol(ctx, req.RoleID)

	if err != nil {
		var sqlErr mssql.Error

		if errors.As(err, &sqlErr) {
			if sqlErr.Number == 547 {
				ctx.JSON(http.StatusConflict, gin.H{
					"error": "No se puede eliminar el rol porque está asignado a uno o más usuarios",
				})
				return
			}
		}

		internalServerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Rol eliminado correctamente",
	})
}

// deleteRoleURI elimina un rol por medio de la URI.
func (server *Server) deleteRoleURI(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	var req deleteRoleRequestURI

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.roleRepository.EliminarRol(ctx, req.RoleID)

	if err != nil {
		var sqlErr mssql.Error

		if errors.As(err, &sqlErr) {
			if sqlErr.Number == 547 {
				ctx.JSON(http.StatusConflict, gin.H{
					"error": "No se puede eliminar el rol porque está asignado a uno o más usuarios",
				})
				return
			}
		}

		internalServerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Rol eliminado correctamente",
	})
}
