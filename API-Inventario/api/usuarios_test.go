package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rest/models"

	"github.com/gin-gonic/gin"
)

const userTestSecret = "clave-de-prueba-de-32-caracteres"

func newUsersTestServer(t *testing.T) *Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	server, err := NewServer(nil, nil, nil, userTestSecret, time.Hour)
	if err != nil {
		t.Fatalf("NewServer devolvió error: %v", err)
	}
	return server
}

func usersRequest(t *testing.T, server *Server, method, path, body string, role int32) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if role > 0 {
		token, err := server.tokenBuilder.CreateToken(50, role, "Usuario de prueba", time.Hour)
		if err != nil {
			t.Fatalf("CreateToken devolvió error: %v", err)
		}
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response := httptest.NewRecorder()
	server.router.ServeHTTP(response, request)
	return response
}

func assertJSONError(t *testing.T, response *httptest.ResponseRecorder, status int) {
	t.Helper()

	if response.Code != status {
		t.Fatalf("estado HTTP: se esperaba %d, se obtuvo %d; cuerpo: %s", status, response.Code, response.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("respuesta no es JSON válido: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("se esperaba una propiedad error no vacía, cuerpo: %s", response.Body.String())
	}
}

func TestRegister_RequiredFields(t *testing.T) {
	server := newUsersTestServer(t)
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "todos ausentes", body: `{}`, want: "El campo rolId es obligatorio"},
		{name: "rol ausente", body: `{"nombre":"Ana","contrasena":"clave-segura-123"}`, want: "El campo rolId es obligatorio"},
		{name: "nombre ausente", body: `{"rolId":1,"contrasena":"clave-segura-123"}`, want: "El campo nombre es obligatorio"},
		{name: "contrasena ausente", body: `{"rolId":1,"nombre":"Ana"}`, want: "El campo contrasena es obligatorio"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := usersRequest(t, server, http.MethodPost, "/api/v1/register", test.body, 0)
			assertJSONError(t, response, http.StatusBadRequest)

			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("respuesta no es JSON válido: %v", err)
			}
			if body["error"] != test.want {
				t.Errorf("error: se esperaba %q, se obtuvo %q", test.want, body["error"])
			}
		})
	}
}

func TestRegister_RejectsInvalidUserData(t *testing.T) {
	server := newUsersTestServer(t)
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "nombre vacío",
			body: `{"rolId":1,"nombre":"   ","contrasena":"clave-segura-123"}`,
			want: "El nombre no puede estar vacío",
		},
		{
			name: "correo inválido",
			body: `{"rolId":1,"nombre":"Ana","contrasena":"clave-segura-123","email":"correo-invalido"}`,
			want: "El correo no tiene un formato válido",
		},
		{
			name: "JSON malformado",
			body: `{"rolId":1`,
			want: "El cuerpo de la solicitud no es un JSON válido",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := usersRequest(t, server, http.MethodPost, "/api/v1/register", test.body, 0)
			assertJSONError(t, response, http.StatusBadRequest)

			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("respuesta no es JSON válido: %v", err)
			}
			if body["error"] != test.want {
				t.Errorf("error: se esperaba %q, se obtuvo %q", test.want, body["error"])
			}
		})
	}
}

func TestUserRoutes_RequireAuthorization(t *testing.T) {
	server := newUsersTestServer(t)
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/api/v1/users"},
		{method: http.MethodPut, path: "/api/v1/users/12", body: `{"nombre":"Ana"}`},
		{method: http.MethodDelete, path: "/api/v1/users/12"},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			response := usersRequest(t, server, test.method, test.path, test.body, 0)
			assertJSONError(t, response, http.StatusUnauthorized)
		})
	}
}

func TestUserRoutes_RejectNonAdmin(t *testing.T) {
	server := newUsersTestServer(t)
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/api/v1/users"},
		{method: http.MethodPut, path: "/api/v1/users/12", body: `{"nombre":"Ana"}`},
		{method: http.MethodDelete, path: "/api/v1/users/12"},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			response := usersRequest(t, server, test.method, test.path, test.body, 2)
			assertJSONError(t, response, http.StatusForbidden)
		})
	}
}

func TestUserRoutes_RejectInvalidIdentifiers(t *testing.T) {
	server := newUsersTestServer(t)
	for _, id := range []string{"abc", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			for _, operation := range []struct {
				method string
				body   string
			}{
				{method: http.MethodPut, body: `{"nombre":"Ana"}`},
				{method: http.MethodDelete},
			} {
				response := usersRequest(
					t,
					server,
					operation.method,
					"/api/v1/users/"+id,
					operation.body,
					1,
				)
				assertJSONError(t, response, http.StatusBadRequest)
			}
		})
	}
}

func TestUsuarioResponse_DoesNotExposePassword(t *testing.T) {
	email := "ana@example.com"
	response, err := json.Marshal(nuevoUsuarioResponse(&models.Usuario{
		UsuarioID:  12,
		RolID:      2,
		Nombre:     "Ana",
		Contrasena: "hash-no-debe-exponerse",
		Email:      &email,
	}))
	if err != nil {
		t.Fatalf("no se pudo serializar la respuesta: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(response, &body); err != nil {
		t.Fatalf("respuesta no es JSON válido: %v", err)
	}
	for _, field := range []string{"usuarioId", "rolId", "nombre", "email"} {
		if _, exists := body[field]; !exists {
			t.Errorf("falta la propiedad JSON %q: %s", field, response)
		}
	}
	if _, exists := body["contrasena"]; exists {
		t.Errorf("la respuesta no debe exponer contrasena: %s", response)
	}
}
