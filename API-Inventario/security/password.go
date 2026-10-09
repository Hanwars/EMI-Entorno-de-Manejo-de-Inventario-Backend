
package security
 
import (
	"errors"
	"unicode/utf8"
 
	"golang.org/x/crypto/bcrypt"
)
 
const (
	BcryptCost       = 12
	MinPasswordChars = 8
	MaxPasswordBytes = 72
)
 
// ErrContrasenaInvalida se devuelve cuando la contraseña no cumple la longitud permitida.
var ErrContrasenaInvalida = errors.New(
	"la contraseña debe tener al menos 8 caracteres y no superar los 72 bytes",
)
 
// HashPassword valida la contraseña y devuelve su hash bcrypt (60 caracteres).
// Se usa al registrar un usuario o al cambiar su contraseña.
func HashPassword(password string) (string, error) {
	if utf8.RuneCountInString(password) < MinPasswordChars || len(password) > MaxPasswordBytes {
		return "", ErrContrasenaInvalida
	}
 
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
 
func CheckPassword(hash, password string) (bool, error) {
	// bcrypt ignora lo que pasa de 72 bytes. Si no se revisa, una contraseña
	// de 72 bytes seguida de cualquier texto sería aceptada como válida.
	if len(password) > MaxPasswordBytes {
		return false, nil
	}
 
	// CompareHashAndPassword lee el costo y el salt del propio hash,
	// recalcula el hash y lo compara en tiempo constante.
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return false, err
}