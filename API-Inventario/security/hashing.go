package security

import "golang.org/x/crypto/bcrypt"

// Genera un hash a partir de una contraseña.
func GenerarHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// Comprueba si una contraseña coincide con un hash.
func VerificarPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)

	return err == nil
}