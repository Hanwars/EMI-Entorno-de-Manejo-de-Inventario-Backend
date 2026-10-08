# Decisión: algoritmo de hashing de contraseñas del sistema EMI

**Proyecto:** Sistema EMI (Entorno de Manejo de Inventario) — Coconutz BrewHouse
**Decisión:** se utilizará **bcrypt** para almacenar las contraseñas de los usuarios.
**Documento relacionado:** *Investigación sobre algoritmos de hashing para contraseñas*

---

## 1. Resumen

| Aspecto | Definición |
|---|---|
| Algoritmo | bcrypt |
| Librería | `golang.org/x/crypto/bcrypt` (oficial de Go) |
| Costo (*cost factor*) | 12 |
| Salt | 128 bits, generado automáticamente y distinto para cada contraseña |
| Longitud del hash | 60 caracteres |
| Contraseña permitida | Mínimo 8 caracteres y máximo 72 bytes |
| Almacenamiento | Columna `contraseña` de la tabla `usuarios` |

---

## 2. Contexto

El sistema EMI maneja usuarios con distintos roles (administradora, supervisora, personal y encargado de TI) y requiere los casos de uso **Iniciar sesión** y **Gestionar usuarios**. Las contraseñas nunca deben almacenarse en texto plano, por lo que se investigaron cinco algoritmos de hashing: bcrypt, scrypt, Argon2id, PBKDF2 y yescrypt.

---

## 3. Alternativas evaluadas

Los algoritmos se compararon según su nivel de seguridad, su facilidad de implementación en Go y su consumo de recursos en el servidor.

| Algoritmo | *Memory-hard* | Soporte en Go | Complejidad de configuración |
|---|---|---|---|
| **bcrypt** | No | Oficial (`golang.org/x/crypto/bcrypt`) | Baja (un solo parámetro) |
| scrypt | Sí | Oficial (`golang.org/x/crypto/scrypt`) | Media (N, r, p) |
| Argon2id | Sí | Oficial (`golang.org/x/crypto/argon2`) | Media (memoria, tiempo, paralelismo) |
| PBKDF2 | No | Oficial (`crypto/pbkdf2`) | Baja, pero requiere cientos de miles de iteraciones |
| yescrypt | Sí | Sin soporte oficial | Alta |

### Motivos de descarte

- **yescrypt:** su configuración es muy compleja y no cuenta con una librería oficial en Go.
- **PBKDF2:** es el algoritmo más fácil de acelerar con GPU, ya que no depende de la memoria.
- **scrypt y Argon2id:** ofrecen mayor resistencia por ser *memory-hard*, pero su librería en Go solo genera el hash. El salt, el formato de almacenamiento y la comparación segura deben programarse manualmente, lo que aumenta el riesgo de cometer errores. Además, cada inicio de sesión reservaría varios MiB de memoria en el servidor.

---

## 4. Justificación de la elección

Se eligió bcrypt por las siguientes razones:

1. **Ampliamente probado y aceptado por OWASP.**
2. **La librería resuelve por sí sola lo delicado:** generación del salt, formato del hash y comparación. Esto reduce el riesgo de errores de implementación frente a scrypt y Argon2id.
3. **Soporte oficial en Go.** El proyecto ya utiliza `golang.org/x/crypto` como dependencia, por lo que no se agregan librerías nuevas.
4. **Nivel de seguridad adecuado** para el tamaño y el tipo de uso del sistema, que tiene un número reducido de usuarios.
5. **Un único parámetro de configuración** (el costo), lo que facilita su mantenimiento y su documentación.

---

## 5. Parámetros definidos

### Costo: 12

Es mayor al valor por defecto de Go (`DefaultCost = 10`). Con este valor cada hash tarda aproximadamente 250 ms. Ese tiempo no afecta la experiencia del usuario al iniciar sesión, pero vuelve muy lentos los ataques de fuerza bruta. Al ser configurable, podrá aumentarse en el futuro conforme mejore el hardware.

> El tiempo de 250 ms es una estimación. Conviene medirlo en el servidor donde se despliegue el sistema y ajustar el costo si es necesario.

### Salt

Lo genera bcrypt automáticamente. Es de 128 bits y distinto para cada contraseña, por lo que dos usuarios con la misma contraseña obtienen hashes diferentes. Se guarda dentro del propio hash, por lo que **no se necesita una columna adicional**.

### Longitud del hash

60 caracteres. La columna `contraseña`, definida como `nvarchar(255)`, tiene espacio suficiente.

### Longitud de la contraseña

- **Mínimo 8 caracteres**, según la recomendación de OWASP.
- **Máximo 72 bytes.** bcrypt ignora todo lo que exceda ese límite. Por eso el backend **rechazará** las contraseñas más largas en lugar de truncarlas sin avisar.

---

## 6. Uso en el sistema

Las contraseñas nunca se almacenan en texto plano. El backend genera el hash antes de guardarlas, por lo que el frontend solo se encarga de enviar la contraseña al servidor.

### Registro de usuario

1. El backend valida la longitud de la contraseña.
2. Genera el hash con `bcrypt.GenerateFromPassword` usando costo 12.
3. Guarda el resultado en la columna `contraseña` de la tabla `usuarios`.

### Inicio de sesión

1. El backend obtiene el hash almacenado del usuario.
2. Compara la contraseña ingresada con `bcrypt.CompareHashAndPassword`.
3. Si no coinciden, responde con un mensaje genérico de credenciales inválidas, **sin indicar** si el error está en el usuario o en la contraseña.

### Ejemplo en Go

> Ejemplo ilustrativo. Debe adaptarse a la estructura del proyecto.

```go
package auth

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost       = 12
	minPasswordChars = 8
	maxPasswordBytes = 72
)

var ErrContrasenaInvalida = errors.New(
	"la contraseña debe tener al menos 8 caracteres y no superar los 72 bytes",
)

// HashPassword valida la contraseña y devuelve su hash bcrypt.
func HashPassword(password string) (string, error) {
	if utf8.RuneCountInString(password) < minPasswordChars || len(password) > maxPasswordBytes {
		return "", ErrContrasenaInvalida
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compara la contraseña ingresada con el hash almacenado.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

---

## 7. Limitaciones y mitigaciones

| Limitación | Mitigación |
|---|---|
| No es *memory-hard* (usa unos 4 KB fijos), por lo que las GPU y los ASIC pueden acelerarlo. | Costo 12, que se puede aumentar en el futuro. |
| Ignora los bytes posteriores al 72. | El backend rechaza las contraseñas de más de 72 bytes. |
| Existen algoritmos más recientes con mayor resistencia, como Argon2id. | El hash bcrypt incluye un prefijo que identifica el algoritmo y el costo, lo que permitiría migrar a otro algoritmo de forma gradual: al iniciar sesión correctamente, se vuelve a generar el hash con el nuevo método. |

---

## 8. Referencias

- *Investigación sobre algoritmos de hashing para contraseñas* (documento del equipo).
- Paquete `golang.org/x/crypto/bcrypt`: <https://pkg.go.dev/golang.org/x/crypto/bcrypt>
- OWASP Password Storage Cheat Sheet: <https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html>
