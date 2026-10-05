
package main
/*
import (
	"database/sql"
	"log"

	"rest/api"
	"rest/repository"
	"rest/utils"

	_ "github.com/microsoft/go-mssqldb"
)

func main() {

	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("Error, no se puede cargar la configuración: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Error, no se pudo conectar a la base de datos: ", err)
	}

	// Verificar que la conexión con SQL Server funcione.
	if err := conn.Ping(); err != nil {
		log.Fatal("Error al conectarse a SQL Server: ", err)
	}

	log.Println("Conexión a SQL Server exitosa")

	// Crear los repostorios.


	// Crear el servidor.
	server, err := api.NewServer(
		conn,
		config.SECRET,
	)
	if err != nil {
		log.Fatal("No se puede iniciar el servidor: ", err)
	}

	// Iniciar la API.
	err = server.Start(config.ServerURL)
	if err != nil {
		log.Fatal("Error al iniciar el servidor: ", err)
	}
}
	
*/