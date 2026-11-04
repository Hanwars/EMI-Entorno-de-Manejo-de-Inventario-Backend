# Registro de pruebas del módulo de usuarios


## Ejecución

```powershell
Set-Location API-Inventario
go test ./...
$env:RUN_USERS_DB_INTEGRATION = '1'
go test ./api -run '^TestUsersSQLServerCRUD$' -v
```

Resultado: **correcto**. Los paquetes compilan; `rest/security` pasa sus pruebas
existentes, `rest/api` pasa los casos unitarios y la prueba de integración
`TestUsersSQLServerCRUD` finaliza con `PASS` contra SQL Server. La variable de
entorno activa únicamente la prueba de integración; sin ella, el caso se omite.
La prueba usa correos aleatorios y elimina las filas creadas al terminar.

## Casos ejecutados

| Operación/caso | Resultado esperado | Resultado |
|---|---|---|
| Registro sin todos los campos | HTTP 400, JSON con `error` y campo obligatorio | Aprobado: `rolId` |
| Registro sin `rolId`, `nombre` o `contrasena` | HTTP 400 con el nombre JSON del campo | Aprobado |
| Registro con nombre vacío | HTTP 400 y mensaje de validación | Aprobado |
| Registro con correo mal formado | HTTP 400 y mensaje de validación | Aprobado |
| Registro con JSON malformado | HTTP 400 y respuesta JSON de error | Aprobado |
| Consulta/actualización/eliminación sin token | HTTP 401 y respuesta JSON de error | Aprobado |
| Consulta/actualización/eliminación con token no administrador | HTTP 403 y respuesta JSON de error | Aprobado |
| Actualización/eliminación con ID no numérico, cero o negativo | HTTP 400 y respuesta JSON de error | Aprobado |
| Serialización de usuario | Incluye `usuarioId`, `rolId`, `nombre`, `email`; excluye contraseña | Aprobado |
| Registro válido | HTTP 200, `idGenerado` y fila persistida con contraseña hasheada | Aprobado en SQL Server |
| Registro duplicado | HTTP 409 y no crea una fila duplicada | Aprobado en SQL Server |
| Listado con administrador | HTTP 200, arreglo JSON que contiene al usuario creado y no expone contraseña | Aprobado en SQL Server |
| Actualización válida | HTTP 200; respuesta y valores de SQL contienen nombre/correo actualizados | Aprobado en SQL Server |
| Actualización a correo duplicado | HTTP 409; el correo previo permanece en la fila | Aprobado en SQL Server |
| Actualización de ID inexistente | HTTP 404 | Aprobado en SQL Server |
| Eliminación de ID inexistente | HTTP 404 | Aprobado en SQL Server |
| Eliminación válida | HTTP 200 y fila ausente en SQL | Aprobado en SQL Server |

Los casos unitarios usan un servidor HTTP en memoria. Los casos de integración
se ejecutaron con la autorización para usar la base de pruebas; la prueba
limpió los registros temporales y no informó errores de limpieza.

## Operaciones pendientes de integración adicional

Queda por ampliar la integración para cubrir la respuesta de listado vacío y la
protección al eliminar usuarios con movimientos. No se generaron movimientos
de prueba, para no insertar historial de inventario innecesario.

| Caso de integración pendiente | Comprobación necesaria |
|---|---|
| Eliminación de usuario con movimientos | HTTP 409 y fila/historial conservados |
| Listado vacío | HTTP 200 y arreglo JSON vacío |

## Hallazgos y limitaciones

- Antes de agregar estas pruebas, `api` y `repository` no tenían casos de prueba
  automatizados. La suite existente se concentraba en `security`.
- La ruta de registro disponible es `POST /api/v1/register` y es pública. La
  consulta, actualización y eliminación de usuarios requieren token; la ruta
  `POST /api/v1/users` no está habilitada.
- No se encontró una colección Postman con solicitudes almacenadas en el
  directorio del proyecto.
- No se detectaron fallos en los casos automatizados ejecutados. La
  persistencia, los códigos de respuesta verificados y la unicidad de correo
  pasaron en SQL Server. No se reportaron errores funcionales en la ejecución.
