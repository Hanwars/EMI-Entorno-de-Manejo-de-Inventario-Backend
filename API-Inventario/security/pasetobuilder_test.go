package security

import (
	"strings"
	"testing"
	"time"
)

// Clave de prueba de 32 caracteres, el tamaño que exige PASETO v2.
const claveDePrueba = "clave-de-prueba-de-32-caracteres"

func nuevoBuilderDePrueba(t *testing.T) Builder {
	t.Helper()

	builder, err := NewPasetoBuilder(claveDePrueba)
	if err != nil {
		t.Fatalf("NewPasetoBuilder devolvió error: %v", err)
	}
	return builder
}

func TestNewPasetoBuilder_ClaveInvalida(t *testing.T) {
	casos := []string{"", "corta", strings.Repeat("a", 31), strings.Repeat("a", 33)}

	for _, clave := range casos {
		if _, err := NewPasetoBuilder(clave); err == nil {
			t.Errorf("se esperaba error para una clave de %d caracteres", len(clave))
		}
	}
}

func TestCreateToken_DatosDelUsuario(t *testing.T) {
	builder := nuevoBuilderDePrueba(t)

	token, err := builder.CreateToken(7, 1, "Usuario Prueba", time.Hour)
	if err != nil {
		t.Fatalf("CreateToken devolvió error: %v", err)
	}
	if token == "" {
		t.Fatal("se esperaba un token no vacío")
	}

	payload, err := builder.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken devolvió error para un token válido: %v", err)
	}
	if payload.UserID != 7 || payload.Role != 1 || payload.Name != "Usuario Prueba" {
		t.Errorf("el token no conserva los datos del usuario: %+v", payload)
	}
}

func TestCreateToken_Vigencia(t *testing.T) {
	builder := nuevoBuilderDePrueba(t)
	duracion := 30 * time.Minute

	token, _ := builder.CreateToken(1, 1, "Usuario Prueba", duracion)
	payload, err := builder.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken devolvió error: %v", err)
	}

	vigencia := payload.ExpiredAt.Sub(payload.IssuedAt)
	if vigencia < duracion-time.Second || vigencia > duracion+time.Second {
		t.Errorf("se esperaba una vigencia de %v, se obtuvo %v", duracion, vigencia)
	}
}

func TestCreateToken_IDUnico(t *testing.T) {
	builder := nuevoBuilderDePrueba(t)

	t1, _ := builder.CreateToken(1, 1, "Usuario Prueba", time.Hour)
	t2, _ := builder.CreateToken(1, 1, "Usuario Prueba", time.Hour)
	p1, _ := builder.VerifyToken(t1)
	p2, _ := builder.VerifyToken(t2)

	if p1.ID == p2.ID {
		t.Error("cada token debe tener un identificador único")
	}
}

func TestVerifyToken_Expirado(t *testing.T) {
	builder := nuevoBuilderDePrueba(t)

	// Una duración negativa produce un token que ya venció.
	token, _ := builder.CreateToken(1, 1, "Usuario Prueba", -time.Minute)

	if _, err := builder.VerifyToken(token); err != ErrorExpiredToken {
		t.Errorf("se esperaba ErrorExpiredToken, se obtuvo %v", err)
	}
}

func TestVerifyToken_Alterado(t *testing.T) {
	builder := nuevoBuilderDePrueba(t)
	token, _ := builder.CreateToken(1, 1, "Usuario Prueba", time.Hour)

	// Se cambia un carácter del contenido cifrado para simular un token manipulado.
	// No se usa el último carácter porque en base64 puede contener solo bits de relleno.
	i := len(token) / 2
	reemplazo := byte('A')
	if token[i] == 'A' {
		reemplazo = 'B'
	}
	alterado := token[:i] + string(reemplazo) + token[i+1:]

	if _, err := builder.VerifyToken(alterado); err != ErrorInvalidToken {
		t.Errorf("se esperaba ErrorInvalidToken para un token alterado, se obtuvo %v", err)
	}
}

func TestVerifyToken_OtraClave(t *testing.T) {
	builder := nuevoBuilderDePrueba(t)
	token, _ := builder.CreateToken(1, 1, "Usuario Prueba", time.Hour)

	otroBuilder, _ := NewPasetoBuilder(strings.Repeat("x", 32))

	if _, err := otroBuilder.VerifyToken(token); err != ErrorInvalidToken {
		t.Errorf("un token firmado con otra clave debe ser rechazado, se obtuvo %v", err)
	}
}

func TestVerifyToken_TextoInvalido(t *testing.T) {
	builder := nuevoBuilderDePrueba(t)
	casos := []string{"", "texto-plano", "v2.local.abc"}

	for _, token := range casos {
		if _, err := builder.VerifyToken(token); err != ErrorInvalidToken {
			t.Errorf("se esperaba ErrorInvalidToken para %q, se obtuvo %v", token, err)
		}
	}
}
