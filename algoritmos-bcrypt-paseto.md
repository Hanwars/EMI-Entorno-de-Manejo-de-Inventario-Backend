# Seguridad en el Sistema EMI: bcrypt y PASETO v2

**Proyecto:** Sistema EMI (Entorno de Manejo de Inventario) — Coconutz BrewHouse
**Tema:** Documentación de algoritmos de seguridad para autenticación

---

## Introducción

Este documento describe dos mecanismos de seguridad evaluados para el sistema EMI: **bcrypt** y **PASETO v2**. Cumplen funciones distintas dentro de la autenticación:

| Mecanismo | Función | Momento de uso |
|---|---|---|
| bcrypt | Protege la contraseña almacenada en la base de datos | Al registrar un usuario y al iniciar sesión |
| PASETO | Mantiene la sesión del usuario mediante un token | Después de iniciar sesión, en cada petición |

PASETO **no** es un algoritmo para guardar contraseñas. bcrypt tampoco sirve para mantener sesiones.

---

## 1. bcrypt

### Qué es

bcrypt es una función de hash adaptativa diseñada para proteger contraseñas, basada en el cifrado simétrico Blowfish. Fue creada en 1999 por Niels Provos y David Mazières. Transforma la contraseña en una cadena ilegible e irreversible.

### Cómo funciona

- **Hash unidireccional:** no se puede recuperar la contraseña original a partir del hash. Para verificar una cuenta, el sistema recalcula el hash de la contraseña ingresada y lo compara con el almacenado.
- **Salt automático de 128 bits:** se agrega un valor aleatorio a cada contraseña antes de aplicar el hash. Dos usuarios con la misma contraseña obtienen hashes distintos, lo que neutraliza los ataques con tablas arcoíris (*rainbow tables*).
- **Factor de costo configurable (*work factor*):** el algoritmo es intencionalmente lento. Al aumentar el costo se realizan más rondas de cifrado, lo que mantiene el proceso costoso para un atacante aunque el hardware mejore. En la implementación de Go los valores van de 4 a 31, con 10 por defecto.

### Ventajas

- Muy usado y probado.
- Disponible en casi todos los lenguajes de programación.
- Protege contra tablas arcoíris mediante el salt.
- Fácil de implementar: el salt va incluido en el propio hash y no hay que almacenarlo aparte.

### Desventajas

- No es *memory-hard*: usa solo unos 4 KB fijos, por lo que las GPU y los ASIC pueden acelerarlo.
- Trunca las contraseñas a 72 bytes (en UTF-8). Los caracteres adicionales se ignoran.
- Algoritmos más recientes, como Argon2id, ofrecen mayor protección.

### Aplicación al sistema EMI

Se utilizaría en los casos de uso **Iniciar sesión** y **Gestionar usuarios**. En la tabla de usuarios de MySQL solo se almacenaría el hash, nunca la contraseña en texto plano.

Dado que el sistema tendrá pocos usuarios (administradora, supervisora, personal y TI) y se desplegará en una computadora de capacidad desconocida proporcionada por la empresa, el costo de cálculo de bcrypt no representa un problema. Además, es la alternativa de respaldo si Argon2id resulta demasiado exigente en memoria para ese equipo.

### Ejemplo en Python

```python
import bcrypt

# Al registrar o cambiar la contraseña
hash_guardado = bcrypt.hashpw(clave.encode("utf-8"), bcrypt.gensalt(rounds=12))

# Al iniciar sesión
es_valida = bcrypt.checkpw(clave_ingresada.encode("utf-8"), hash_guardado)
```

---

## 2. PASETO v2

### Qué es

PASETO (*Platform-Agnostic Security Tokens*) es una alternativa a JWT para crear tokens de sesión seguros. Su diseño busca evitar los errores de configuración típicos de JWT: cada versión de PASETO fija un único algoritmo criptográfico, por lo que el desarrollador no puede elegir una combinación insegura.

### Cómo funciona

Un token PASETO se compone de una versión, un propósito y un contenido, por ejemplo `v2.local.…`. El propósito puede ser:

- **local:** cifrado simétrico. En v2 usa XChaCha20-Poly1305. Emisor y verificador comparten la misma clave.
- **public:** firma asimétrica. En v2 usa Ed25519. El verificador solo necesita la clave pública.

El verificador rechaza cualquier token cuya versión o propósito no coincida con lo configurado, sin siquiera procesarlo.

### Versiones de PASETO

| Versión | Local | Public | Estado |
|---|---|---|---|
| v1 | AES-256-CTR + HMAC-SHA384 | RSA-PSS | Obsoleta; v1 public se considera insegura |
| **v2** | **XChaCha20-Poly1305** | **Ed25519** | **Heredada: segura, pero se prefiere v4** |
| v3 | AES-256-CTR + HMAC-SHA384 | ECDSA P-384 | Para cumplimiento NIST |
| v4 | XChaCha20 + BLAKE2b | Ed25519 | **Recomendada** |

### Ventajas

- Elimina la elección de algoritmos, principal fuente de vulnerabilidades en JWT.
- Los tokens *local* van cifrados, no solo firmados, por lo que su contenido no queda legible.
- Existen librerías para varios lenguajes; en Python se dispone de `pyseto`.

### Desventajas

- **Es una versión heredada:** v4 es la versión recomendada para proyectos nuevos.
- v2 usa un modo AEAD basado en Poly1305, que no ofrece compromiso de mensaje ni de clave: un mismo token podría descifrarse a dos contenidos distintos bajo claves diferentes. v4 elimina Poly1305 en los tokens locales por esta razón.
- Migrar de v2 a v4 más adelante obliga a manejar ambas versiones durante la transición.

### Aplicación al sistema EMI

Después del inicio de sesión, el servidor emitiría un token con el usuario, su rol y una fecha de expiración. Las tabletas lo enviarían en cada petición para acceder a las pantallas según el rol. Un token **local** es suficiente: el sistema se despliega en LAN, con un único servidor que emite y verifica con la misma clave.

**Consideraciones de diseño:**

1. Si el proyecto adopta PASETO, se recomienda usar **v4**, o dejar documentado que v2 se evaluó y se descartó por ser una versión heredada.
2. Un sistema pequeño en LAN puede funcionar con **sesiones del lado del servidor**, que son más simples. Los tokens aportan más valor cuando hay varios servicios o clientes externos. Debe justificarse por qué el proyecto necesita un token.

### Ejemplo en Python

> Ejemplo ilustrativo. Verificar la sintaxis en la documentación oficial de `pyseto` antes de usarlo.

```python
import os
import pyseto
from pyseto import Key

clave = Key.new(version=2, purpose="local", key=os.urandom(32))
token = pyseto.encode(clave, b'{"usuario": "supervisora", "rol": "supervisor"}')
datos = pyseto.decode(clave, token).payload
```

---

## Cómo encajan en el sistema EMI

| Etapa | Mecanismo |
|---|---|
| Guardar la contraseña en MySQL | bcrypt (o Argon2id como opción principal) |
| Mantener la sesión tras el inicio de sesión | Token PASETO (idealmente v4) o sesión del servidor |

---

## Referencias

- Investigación del equipo: *Investigación sobre algoritmos de hashing para contraseñas* (bcrypt).
- Especificación de PASETO: <https://github.com/paseto-standard/paseto-spec>
- Paragon Initiative Enterprises, *PASETO is an Even More Secure Alternative to the JOSE Standards (JWT, etc.)*: <https://paragonie.com/blog/2021/08/paseto-is-even-more-secure-alternative-jose-standards-jwt-etc>
- *PASETO explained* (versiones y propósitos): <https://guptadeepak.com/ciam-compass/guides/paseto-explained/>
