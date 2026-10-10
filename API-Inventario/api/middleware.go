package api

import (
	"errors"
	"net/http"
	"rest/security"
	"strings"

	"github.com/gin-gonic/gin"
)

func authMiddleware(token security.Builder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("authorization")
		if len(authHeader) == 0 {
			err := errors.New("Token no suministrado en la cabecera")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			err := errors.New("Formato del token invalido")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		if strings.ToLower(fields[0]) != "bearer" {
			err := errors.New("tipo de autorizacion no soportado")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		accessToken := fields[1]
		payload, err := token.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		ctx.Set("authorized", payload)
		ctx.Next()
	}
}
