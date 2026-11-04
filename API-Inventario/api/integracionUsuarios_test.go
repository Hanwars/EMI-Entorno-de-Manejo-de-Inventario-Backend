package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"rest/repository"
	"rest/security"
	"rest/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestUsersSQLServerCRUD(t *testing.T) {
	if os.Getenv("RUN_USERS_DB_INTEGRATION") != "1" {
		t.Skip("defina RUN_USERS_DB_INTEGRATION=1 para usar el SQL Server de pruebas configurado")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	config, err := utils.LoadConfig("..")
	if err != nil {
		t.Fatal("no se pudo cargar la configuración local para la prueba")
	}
	db, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		t.Fatal("no se pudo abrir la conexión a la base de pruebas")
	}
	defer db.Close()
	db.SetMaxOpenConns(4)

	if err := db.PingContext(ctx); err != nil {
		t.Fatal("no se pudo conectar con la base de pruebas")
	}

	var adminRoleCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(1) FROM dbo.roles WHERE role_id = @p1", 1).Scan(&adminRoleCount); err != nil {
		t.Fatal("no se pudo comprobar el rol administrador en la base de pruebas")
	}
	if adminRoleCount == 0 {
		t.Fatal("la base de pruebas no contiene el rol administrador con ID 1")
	}

	gin.SetMode(gin.TestMode)
	server, err := NewServer(
		repository.NewUserRepository(db),
		repository.NewRolRepository(db),
		db,
		config.SECRET,
		config.TokenDuration,
	)
	if err != nil {
		t.Fatal("no se pudo crear el servidor de prueba")
	}

	suffix := uuid.NewString()
	adminEmail := "users-test-admin-" + suffix + "@example.test"
	userEmail := "users-test-user-" + suffix + "@example.test"
	updatedEmail := "users-test-updated-" + suffix + "@example.test"
	adminPassword := "Test-Admin-Password-123"
	userPassword := "Test-User-Password-123"

	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if _, err := db.ExecContext(
			cleanupCtx,
			"DELETE FROM dbo.usuarios WHERE email IN (@p1, @p2, @p3)",
			adminEmail,
			userEmail,
			updatedEmail,
		); err != nil {
			t.Error("no se pudieron limpiar las filas de prueba")
		}
	}()

	doRequest := func(method, path, body, token string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		server.router.ServeHTTP(response, request)
		return response
	}

	type createdUser struct {
		ID int32 `json:"idGenerado"`
	}
	register := func(email, name, password string) int32 {
		t.Helper()
		body, err := json.Marshal(map[string]any{
			"rolId":      1,
			"nombre":     name,
			"contrasena": password,
			"email":      email,
		})
		if err != nil {
			t.Fatal("no se pudo preparar la solicitud de registro")
		}
		response := doRequest(http.MethodPost, "/api/v1/register", string(body), "")
		if response.Code != http.StatusOK {
			t.Fatalf("registro: se esperaba HTTP 200, se obtuvo %d", response.Code)
		}
		var result createdUser
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.ID <= 0 {
			t.Fatal("registro no devolvió un idGenerado válido")
		}
		return result.ID
	}

	register(adminEmail, "Usuario de prueba administrador", adminPassword)

	loginBody, err := json.Marshal(map[string]string{
		"email":      adminEmail,
		"contrasena": adminPassword,
	})
	if err != nil {
		t.Fatal("no se pudo preparar la solicitud de login")
	}
	loginResponse := doRequest(http.MethodPost, "/api/v1/login", string(loginBody), "")
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login de prueba: se esperaba HTTP 200, se obtuvo %d", loginResponse.Code)
	}
	var loginResult struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &loginResult); err != nil || loginResult.AccessToken == "" {
		t.Fatal("login no devolvió un access_token válido")
	}

	duplicateRegisterBody, err := json.Marshal(map[string]any{
		"rolId":      1,
		"nombre":     "Usuario duplicado",
		"contrasena": userPassword,
		"email":      adminEmail,
	})
	if err != nil {
		t.Fatal("no se pudo preparar la solicitud duplicada")
	}
	duplicateRegister := doRequest(http.MethodPost, "/api/v1/register", string(duplicateRegisterBody), "")
	assertJSONError(t, duplicateRegister, http.StatusConflict)

	userID := register(userEmail, "Usuario de prueba", userPassword)
	var storedPassword string
	if err := db.QueryRowContext(ctx, "SELECT contraseña FROM dbo.usuarios WHERE usuario_id = @p1", userID).Scan(&storedPassword); err != nil {
		t.Fatal("no se encontró el usuario registrado en SQL Server")
	}
	if matches, err := security.CheckPassword(storedPassword, userPassword); err != nil || !matches {
		t.Fatal("la contraseña persistida no corresponde al hash esperado")
	}

	listResponse := doRequest(http.MethodGet, "/api/v1/users", "", loginResult.AccessToken)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("listado: se esperaba HTTP 200, se obtuvo %d", listResponse.Code)
	}
	var users []map[string]any
	if err := json.Unmarshal(listResponse.Body.Bytes(), &users); err != nil {
		t.Fatal("el listado no devolvió un arreglo JSON")
	}
	foundUser := false
	for _, user := range users {
		if user["usuarioId"] == float64(userID) {
			foundUser = true
			if _, exposesPassword := user["contrasena"]; exposesPassword {
				t.Error("el listado expuso la contraseña")
			}
		}
	}
	if !foundUser {
		t.Fatal("el listado no incluyó el usuario recién registrado")
	}

	updateBody, err := json.Marshal(map[string]string{
		"nombre": "Usuario actualizado",
		"email":  updatedEmail,
	})
	if err != nil {
		t.Fatal("no se pudo preparar la solicitud de actualización")
	}
	updateResponse := doRequest(
		http.MethodPut,
		"/api/v1/users/"+strconv.FormatInt(int64(userID), 10),
		string(updateBody),
		loginResult.AccessToken,
	)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("actualización: se esperaba HTTP 200, se obtuvo %d", updateResponse.Code)
	}
	var updatedUser map[string]any
	if err := json.Unmarshal(updateResponse.Body.Bytes(), &updatedUser); err != nil {
		t.Fatal("la actualización no devolvió un objeto JSON")
	}
	if updatedUser["nombre"] != "Usuario actualizado" || updatedUser["email"] != updatedEmail {
		t.Error("la respuesta no contiene los campos actualizados")
	}
	if _, exposesPassword := updatedUser["contrasena"]; exposesPassword {
		t.Error("la actualización expuso la contraseña")
	}
	var storedName, persistedEmail string
	if err := db.QueryRowContext(
		ctx,
		"SELECT nombre, email FROM dbo.usuarios WHERE usuario_id = @p1",
		userID,
	).Scan(&storedName, &persistedEmail); err != nil {
		t.Fatal("no se pudo leer el usuario actualizado de SQL Server")
	}
	if storedName != "Usuario actualizado" || persistedEmail != updatedEmail {
		t.Error("los cambios de la actualización no se persistieron en SQL Server")
	}

	duplicateEmailUpdate := doRequest(
		http.MethodPut,
		"/api/v1/users/"+strconv.FormatInt(int64(userID), 10),
		`{"email":"`+adminEmail+`"}`,
		loginResult.AccessToken,
	)
	assertJSONError(t, duplicateEmailUpdate, http.StatusConflict)
	if err := db.QueryRowContext(
		ctx,
		"SELECT email FROM dbo.usuarios WHERE usuario_id = @p1",
		userID,
	).Scan(&persistedEmail); err != nil {
		t.Fatal("no se pudo comprobar el correo tras rechazar el duplicado")
	}
	if persistedEmail != updatedEmail {
		t.Error("el intento de actualizar a correo duplicado modificó el registro")
	}

	missingID := int32(2147483647)
	missingUpdate := doRequest(
		http.MethodPut,
		"/api/v1/users/"+strconv.FormatInt(int64(missingID), 10),
		`{"nombre":"No existe"}`,
		loginResult.AccessToken,
	)
	assertJSONError(t, missingUpdate, http.StatusNotFound)
	missingDelete := doRequest(
		http.MethodDelete,
		"/api/v1/users/"+strconv.FormatInt(int64(missingID), 10),
		"",
		loginResult.AccessToken,
	)
	assertJSONError(t, missingDelete, http.StatusNotFound)

	deleteResponse := doRequest(
		http.MethodDelete,
		"/api/v1/users/"+strconv.FormatInt(int64(userID), 10),
		"",
		loginResult.AccessToken,
	)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("eliminación: se esperaba HTTP 200, se obtuvo %d", deleteResponse.Code)
	}
	var remaining int
	if err := db.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM dbo.usuarios WHERE usuario_id = @p1",
		userID,
	).Scan(&remaining); err != nil {
		t.Fatal("no se pudo comprobar la eliminación en SQL Server")
	}
	if remaining != 0 {
		t.Error("el usuario eliminado todavía existe en SQL Server")
	}

}
