package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"rest/security"

	"github.com/gin-gonic/gin"
	mssql "github.com/microsoft/go-mssqldb"
)

// createRoleRequest contiene los datos para crear un rol. El role_id lo genera la base de datos.
type createRoleRequest struct {
	Nombre string `json:"nombre" binding:"required"`
}

// updateRoleURI identifica al rol a actualizar: PUT api/v1/roles/:roleId
type updateRoleURI struct {
	RoleID int32 `uri:"roleId" binding:"required,min=1"`
}

// updateRoleRequest contiene el único campo modificable de un rol. Si se envía role_id
// en el cuerpo, se ignora: el identificador se toma de la URI.
type updateRoleRequest struct {
	Nombre string `json:"nombre" binding:"required"`
}

// Longitud máxima definida en la tabla roles.
const maxNombreRol = 300

// createRole es el handler para crear un nuevo rol: POST api/v1/roles
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
		errorSolicitud(ctx, err)
		return
	}

	// Validación según las restricciones de la tabla roles.
	nombre, mensaje := validarNombreRol(req.Nombre)
	if mensaje != "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": mensaje})
		return
	}

	// No se permiten dos roles con el mismo nombre. Se comprueba aquí porque la base de datos
	// no tiene un índice único sobre el nombre.
	existe, err := server.roleRepository.NombreRolExiste(ctx, nombre, 0)
	if err != nil {
		internalServerError(ctx, err)
		return
	}
	if existe {
		ctx.JSON(http.StatusConflict, gin.H{"error": "El rol ya existe"})
		return
	}

	rol, err := server.roleRepository.CrearRol(ctx, nombre)
	if err != nil {
		var sqlErr mssql.Error
		// Respaldo por si la base de datos tiene un índice único sobre el nombre.
		if errors.As(err, &sqlErr) && (sqlErr.Number == 2601 || sqlErr.Number == 2627) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "El rol ya existe"})
			return
		}

		internalServerError(ctx, err)
		return
	}

	// 201 Created con la ubicación del nuevo recurso, según el estándar HTTP.
	ctx.Header("Location", fmt.Sprintf("/api/v1/roles/%d", rol.RoleID))
	ctx.JSON(http.StatusCreated, rol)
}

// updateRole es el handler para actualizar un rol: PUT api/v1/roles/:roleId
// Solo se modifica el nombre; el identificador del rol y sus asociaciones no cambian.
func (server *Server) updateRole(ctx *gin.Context) {

	authPayload := ctx.MustGet("authorized").(*security.Payload)

	if authPayload.Role != 1 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Solo administradores",
		})
		return
	}

	var uri updateRoleURI
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Identificador de rol inválido"})
		return
	}

	var req updateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorSolicitud(ctx, err)
		return
	}

	nombre, mensaje := validarNombreRol(req.Nombre)
	if mensaje != "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": mensaje})
		return
	}

	// El rol debe existir.
	existeRol, err := server.roleRepository.ExisteRol(ctx, uri.RoleID)
	if err != nil {
		internalServerError(ctx, err)
		return
	}
	if !existeRol {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Rol no encontrado"})
		return
	}

	// El nombre no puede pertenecer a otro rol. Se excluye el propio rol para permitir,
	// por ejemplo, cambiar solo las mayúsculas de su nombre.
	nombreEnUso, err := server.roleRepository.NombreRolExiste(ctx, nombre, uri.RoleID)
	if err != nil {
		internalServerError(ctx, err)
		return
	}
	if nombreEnUso {
		ctx.JSON(http.StatusConflict, gin.H{"error": "El rol ya existe"})
		return
	}

	// Solo se responde con éxito si la base de datos confirma el cambio.
	rol, err := server.roleRepository.ActualizarRol(ctx, uri.RoleID, nombre)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// El rol fue eliminado por otra petición entre la comprobación y la actualización.
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Rol no encontrado"})
			return
		}

		var sqlErr mssql.Error
		// Respaldo por si la base de datos tiene un índice único sobre el nombre.
		if errors.As(err, &sqlErr) && (sqlErr.Number == 2601 || sqlErr.Number == 2627) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "El rol ya existe"})
			return
		}

		internalServerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, rol)
}

// validarNombreRol quita los espacios sobrantes y comprueba las restricciones de la tabla roles.
// Devuelve el nombre limpio y un mensaje de error, o una cadena vacía si es válido.
func validarNombreRol(nombre string) (string, string) {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return "", "El nombre del rol no puede estar vacío"
	}
	if utf8.RuneCountInString(nombre) > maxNombreRol {
		return "", fmt.Sprintf("El nombre del rol no puede superar los %d caracteres", maxNombreRol)
	}
	return nombre, ""
}
