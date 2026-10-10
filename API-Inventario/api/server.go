package api

/*
Puse algunas importaciones comentadas porque no se están usando, pero las dejo ahí
por si las necesitamos después.
Vamos a necesitarlas despues, pero solo las descomentamos cuando las necesitemos,
porque go no compila si hay importaciones que no se usan.
De momento para hacer los handlers y probarlos con postman,
no las necesitamos, pero luego si vamos a necesitar.
Y con ese codigo es suficiente para hacer los handlers y luego hacer la URL.
*/

import (
	"database/sql"

	"rest/repository"
	"rest/security"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	cors "github.com/itsjamie/gin-cors"
)

type Server struct {
	conn               *sql.DB
	userRepository     *repository.UserRepository
	router             *gin.Engine
	tokenBuilder       security.Builder
}

func NewServer(
	userRepository *repository.UserRepository,
	conn *sql.DB, secret string) (*Server, error) {

	builder, err := security.NewPasetoBuilder(secret)
	if err != nil {
		return nil, err
	}

	server := &Server{
		conn:               conn,
		userRepository:     userRepository,
		tokenBuilder:       builder,
	}
	router := gin.Default()

	router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET,POST,PUT,DELETE",
		RequestHeaders:  "Origin,Authorization,Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     false,
		ValidateHeaders: false,
	}))

	//RUTAS SIN MIDDLEWARE
	router.POST("api/v1/login", server.login)
	router.POST("api/v1/users", server.createUser)

	//RUTAS CON MIDDLEWARE
	//authRoutes := router.Group("/").Use(authMiddleware(server.tokenBuilder))

	//ROUTES ROLES

	//ROUTES CATEGORIES


	// ROUTES SUPPLIES

	//ROUTES SUPPLIERS

	//ROUTES USERS
	//authRoutes.POST("api/v1/users", server.createUser)

	server.router = router
	return server, nil
}

func (server *Server) Start(url string) error {
	return server.router.Run(url)
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}
