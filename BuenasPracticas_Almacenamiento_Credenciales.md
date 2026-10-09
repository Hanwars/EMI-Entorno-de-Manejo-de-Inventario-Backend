BUENAS PRÁCTICAS PARA EL ALMACENAMIENTO DE CREDENCIALES
HU 3 - Hashing y protección de credenciales
Proyecto EMI - Coconutz BrewHouse

El sistema EMI guarda las credenciales de sus usuarios en SQL Server, y el objetivo es que, aunque alguien obtenga una copia de la base de datos, no pueda conocer ni usar esas contraseñas. La elección del algoritmo de hashing la cubre otro documento del equipo; aquí se resumen las prácticas que deben acompañarlo.

Lo primero es que las contraseñas nunca se guardan en texto plano ni encriptadas, sino hasheadas. A diferencia de la encriptación, el hash no se puede revertir: para validar un inicio de sesión, el sistema hashea lo que escribió el usuario y lo compara con el hash guardado. Esa comparación debe hacerse en tiempo constante (en Go, con crypto/subtle), para no dar pistas a un atacante por diferencias de tiempo.

Cada contraseña debe llevar un salt, que es un valor aleatorio y distinto para cada usuario. Así, dos personas con la misma contraseña tienen hashes diferentes y no sirven las tablas de hashes precalculados. El hash se guarda junto con el algoritmo, sus parámetros y el salt en un solo texto, en una columna VARCHAR(255) NOT NULL. Esto además permite actualizar los parámetros en el futuro sin perder información.

Los secretos del sistema, como la cadena de conexión a SQL Server y la clave de los tokens de sesión, no deben estar en el código ni en GitHub. Se guardan en variables de entorno o en un archivo .env incluido en el .gitignore. Si alguno se sube por error, hay que cambiarlo, porque queda en el historial.

Al fallar un inicio de sesión, el mensaje debe ser genérico ("Usuario o contraseña incorrectos") para no revelar qué cuentas existen, y conviene limitar los intentos fallidos para frenar ataques de fuerza bruta. Las contraseñas y los tokens tampoco deben aparecer nunca en los logs.

También importa qué contraseñas se aceptan. En lugar de reglas complicadas de mayúsculas y símbolos, que suelen llevar a contraseñas como "Clave123!", se recomienda exigir una longitud mínima (por ejemplo, 8 caracteres o más), permitir contraseñas largas y frases, y rechazar las contraseñas más comunes o que ya han aparecido en filtraciones conocidas, como "123456" o "password". Una contraseña débil se adivina rápido aunque esté bien hasheada.

La base de datos también se protege: el backend debe conectarse con un usuario de permisos mínimos (no con "sa"), usar consultas parametrizadas para evitar inyecciones SQL y guardar los respaldos en un lugar con acceso restringido.


Referencias
- OWASP. Password Storage Cheat Sheet. https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- OWASP. Authentication Cheat Sheet. https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- OWASP. Secrets Management Cheat Sheet. https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html