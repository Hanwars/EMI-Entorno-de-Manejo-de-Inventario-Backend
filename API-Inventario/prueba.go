
package main

/*
import (
	"database/sql"
	"log"

	"rest/utils"

	_ "github.com/microsoft/go-mssqldb"
)

func main() {

	// Cargar la configuración desde el archivo .env / configuración.
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("Error, no se puede cargar la configuración: ", err)
	}

	// Abrir la conexión con SQL Server.
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Error, no se pudo abrir la conexión a la base de datos: ", err)
	}
	defer conn.Close()

	// Verificar que realmente podamos conectarnos a SQL Server.
	if err := conn.Ping(); err != nil {
		log.Fatal("Error al conectarse a SQL Server: ", err)
	}

	log.Println("Conexión a SQL Server exitosa")
}
*/