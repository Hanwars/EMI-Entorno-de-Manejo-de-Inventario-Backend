# Integración de la validación de credenciales en el inicio de sesión


## 1. Introducción
El presente documento describe la integración del mecanismo de comparación de credenciales en el proceso de autenticación del sistema EMI. El objetivo de esta tarea es que el inicio de sesión aplique los criterios de seguridad definidos previamente por el equipo en la *Documentación bcrypt*, en particular el uso del algoritmo bcrypt con un costo de 12, el límite máximo de 72 bytes para las contraseñas y el uso de un mensaje genérico ante un intento de autenticación fallido.

## 2. Alcance
La modificación se limita a la función `login` del archivo `api/users.handler.go`, que atiende las solicitudes dirigidas al endpoint `POST /api/v1/login`. No se realizaron cambios en la base de datos, en el repositorio de usuarios ni en el módulo de generación de tokens.

## 3. Situación previa
Antes de esta tarea, el endpoint de inicio de sesión validaba la contraseña mediante la función `VerificarPassword`, definida en el archivo `security/hashing.go`. Dicha función realiza la comparación con bcrypt, pero no aplica el límite de 72 bytes establecido para el sistema. Debido a que bcrypt ignora los bytes posteriores a ese límite, una contraseña que coincidiera en sus primeros 72 bytes con la contraseña registrada podía ser aceptada aunque el resto fuera distinto.

Asimismo, el endpoint devolvía respuestas diferentes según la causa del error. Cuando el correo electrónico no se encontraba registrado, respondía con el código **404 Not Found**; cuando la contraseña era incorrecta, respondía con el código **401 Unauthorized** y el mensaje *"Contraseña incorrecta"*. Esta diferencia permitía a un tercero determinar qué correos electrónicos están registrados en el sistema, práctica conocida como enumeración de usuarios, lo cual contradice las recomendaciones recogidas en el documento *Buenas prácticas para el almacenamiento de credenciales*.

## 4. Descripción de los cambios
La validación de la contraseña se realiza ahora mediante la función `CheckPassword`, definida en el archivo `security/password.go`. Esta función verifica, antes de efectuar la comparación, que la contraseña ingresada no supere los 72 bytes, y delega la comparación en `bcrypt.CompareHashAndPassword`, que obtiene el costo y el salt del propio hash almacenado y realiza la comparación en tiempo constante. La función devuelve dos valores: un indicador que señala si la contraseña es correcta y un error que solo se produce cuando el hash almacenado no tiene un formato válido.

Con el fin de evitar la enumeración de usuarios, la respuesta ante un correo electrónico no registrado se modificó de **404** a **401**. De este modo, tanto en ese caso como en el de una contraseña incorrecta, el sistema responde con el código **401 Unauthorized** y el mensaje genérico `{"message": "Credenciales inválidas"}`, sin indicar cuál de los dos datos es erróneo.

En el caso de que el hash almacenado en la base de datos no tenga un formato válido, el sistema responde con el código **500 Internal Server Error**, dado que se trata de una inconsistencia en los datos del servidor y no de un error atribuible al usuario.

Es importante señalar que los usuarios registrados con anterioridad, cuyas contraseñas fueron almacenadas con un costo de 10, conservan la posibilidad de iniciar sesión. Esto se debe a que bcrypt incluye el costo dentro del propio hash y lo utiliza automáticamente al realizar la comparación.

## 5. Flujo de autenticación
A partir de los cambios descritos, el proceso de inicio de sesión se desarrolla de la siguiente manera:

1. El frontend envía el correo electrónico y la contraseña del usuario al endpoint `POST /api/v1/login`.
2. El backend verifica que ambos campos estén presentes en la solicitud. En caso contrario, responde con el código **400 Bad Request**.
3. El backend consulta la tabla `usuarios` para obtener el registro asociado al correo electrónico. Si no existe, responde con el código **401** y el mensaje *"Credenciales inválidas"*.
4. El backend compara la contraseña ingresada con el hash almacenado mediante la función `CheckPassword`. Si no coinciden, responde con el código **401** y el mismo mensaje genérico.
5. Si la comparación es exitosa, el backend genera un token PASETO con una vigencia de una hora y responde con el código **200 OK**, incluyendo el token de acceso y los datos básicos del usuario: identificador, nombre y rol.

## 6. Verificación
Para comprobar el correcto funcionamiento de los cambios se realizaron las siguientes verificaciones:

- **Análisis estático:** la ejecución de `go vet ./...` no reportó errores.
- **Pruebas unitarias:** la ejecución de `go test ./security/` completó satisfactoriamente las siete pruebas del módulo de seguridad.
- **Pruebas funcionales:** se ejecutó la API de forma local y se enviaron solicitudes al endpoint de inicio de sesión utilizando un rol y un usuario de prueba, los cuales fueron eliminados de la base de datos al concluir las pruebas.

En las pruebas funcionales se evaluaron cinco casos, y en todos ellos el sistema respondió según lo esperado.

Al enviar el correo electrónico y la contraseña correctos del usuario de prueba, la API respondió con el código 200 y devolvió el token de acceso junto con el identificador, el nombre y el rol del usuario.

Posteriormente se envió el mismo correo con una contraseña que solo difería en una letra mayúscula. La API rechazó el acceso con el código 401 y el mensaje "Credenciales inválidas". Esa misma respuesta se obtuvo al utilizar un correo electrónico que no está registrado en el sistema, con lo que se comprobó que ya no es posible distinguir entre un usuario inexistente y una contraseña incorrecta.

También se probó una contraseña de más de 72 bytes cuyos primeros caracteres coincidían con la contraseña registrada. La API la rechazó con el código 401, lo que confirma que el límite de longitud se aplica antes de realizar la comparación.

Por último, se envió una solicitud sin el campo de contraseña, y la API respondió con el código 400, ya que ambos campos son obligatorios.
