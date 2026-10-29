package api

import (
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
	nombre := strings.TrimSpace(req.Nombre)
	if nombre == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El nombre del rol no puede estar vacío"})
		return
	}
	if utf8.RuneCountInString(nombre) > maxNombreRol {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("El nombre del rol no puede superar los %d caracteres", maxNombreRol),
		})
		return
	}

	// No se permiten dos roles con el mismo nombre. Se comprueba aquí porque la base de datos
	// no tiene un índice único sobre el nombre.
	existe, err := server.roleRepository.NombreRolExiste(ctx, nombre)
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
