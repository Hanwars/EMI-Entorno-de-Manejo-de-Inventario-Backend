package security

import (
	"strings"
	"testing"
)

func TestCheckPassword_Correcta(t *testing.T) {
	hash, err := HashPassword("Coconutz2026")
	if err != nil {
		t.Fatalf("HashPassword devolvió error: %v", err)
	}

	ok, err := CheckPassword(hash, "Coconutz2026")
	if err != nil {
		t.Fatalf("CheckPassword devolvió error: %v", err)
	}
	if !ok {
		t.Error("se esperaba que la contraseña correcta fuera aceptada")
	}
}

func TestCheckPassword_Incorrecta(t *testing.T) {
	hash, _ := HashPassword("Coconutz2026")

	ok, err := CheckPassword(hash, "coconutz2026") // cambia una mayúscula
	if err != nil {
		t.Fatalf("una contraseña incorrecta no debe producir error: %v", err)
	}
	if ok {
		t.Error("se esperaba que la contraseña incorrecta fuera rechazada")
	}
}

func TestCheckPassword_MasDe72Bytes(t *testing.T) {
	base := strings.Repeat("a", 72)
	hash, err := HashPassword(base)
	if err != nil {
		t.Fatalf("HashPassword devolvió error: %v", err)
	}

	// Sin la validación, bcrypt aceptaría esto porque ignora lo que pasa de 72 bytes.
	ok, _ := CheckPassword(hash, base+"textoExtra")
	if ok {
		t.Error("una contraseña de más de 72 bytes no debe ser aceptada")
	}
}

func TestCheckPassword_HashInvalido(t *testing.T) {
	casos := []string{"", "texto-plano", "$2a$12$corto"}

	for _, h := range casos {
		ok, err := CheckPassword(h, "Coconutz2026")
		if err == nil || ok {
			t.Errorf("se esperaba error para el hash %q", h)
		}
	}
}

func TestHashPassword_Formato(t *testing.T) {
	hash, _ := HashPassword("Coconutz2026")

	if len(hash) != 60 {
		t.Errorf("se esperaban 60 caracteres, se obtuvieron %d", len(hash))
	}
	if !strings.HasPrefix(hash, "$2a$12$") {
		t.Errorf("se esperaba el prefijo $2a$12$, se obtuvo %s", hash[:7])
	}
}

func TestHashPassword_SaltDistinto(t *testing.T) {
	h1, _ := HashPassword("misma-clave")
	h2, _ := HashPassword("misma-clave")

	if h1 == h2 {
		t.Error("la misma contraseña debe producir hashes distintos gracias al salt")
	}
}

func TestHashPassword_LongitudInvalida(t *testing.T) {
	casos := []string{"corta", strings.Repeat("a", 73)}

	for _, c := range casos {
		if _, err := HashPassword(c); err != ErrContrasenaInvalida {
			t.Errorf("se esperaba ErrContrasenaInvalida para una contraseña de %d bytes", len(c))
		}
	}
}