package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// registrarNombresJSON hace que los errores de validación usen el nombre del campo en el JSON
// (por ejemplo "contrasena") en lugar del nombre interno de la estructura de Go ("Contrasena").
func registrarNombresJSON() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(campo reflect.StructField) string {
			nombre := strings.SplitN(campo.Tag.Get("json"), ",", 2)[0]
			if nombre == "" || nombre == "-" {
				return campo.Name
			}
			return nombre
		})
	}
}

// errorSolicitud responde 400 con un mensaje claro cuando el cuerpo de la solicitud
// no se puede leer o no cumple las reglas, sin exponer detalles internos de Go.
func errorSolicitud(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{"error": mensajeErrorSolicitud(err)})
}

func mensajeErrorSolicitud(err error) string {
	var errValidacion validator.ValidationErrors
	if errors.As(err, &errValidacion) && len(errValidacion) > 0 {
		campo := errValidacion[0]
		if campo.Tag() == "required" {
			return "El campo " + campo.Field() + " es obligatorio"
		}
		return "El campo " + campo.Field() + " no es válido"
	}

	var errTipo *json.UnmarshalTypeError
	if errors.As(err, &errTipo) {
		return "El campo " + errTipo.Field + " tiene un tipo de dato incorrecto"
	}

	var errSintaxis *json.SyntaxError
	if errors.As(err, &errSintaxis) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "El cuerpo de la solicitud no es un JSON válido"
	}

	return "La solicitud no es válida"
}
