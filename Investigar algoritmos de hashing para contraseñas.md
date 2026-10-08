# Investigación sobre algoritmos de hashing para contraseñas

## [bcrypt](https://www.webempresa.com/blog/bcrypt-que-es-y-para-que-funciona.html) 
Bcrypt es una función de hash adaptativa diseñada para proteger las contraseñas de los usuarios transformándolas en cadenas de caracteres ilegibles e irreversibles.
### Características principales
- **Función unidireccional (Hashing):** No se puede "descifrar" un hash de bcrypt para recuperar la contraseña original. Para verificar una cuenta, el sistema vuelve a calcular el hash de la contraseña ingresada y lo compara con el almacenado.
- **Uso de Salt (Sal):** Agrega de manera automática un valor aleatorio (salt) a cada contraseña antes de aplicar el hash. Esto garantiza que dos usuarios con la misma contraseña obtengan hashes totalmente diferentes, protegiendo contra ataques de tablas arcoíris (rainbow tables).
- **Costo adaptable (Work Factor):** Es intencionalmente lento. Su velocidad se puede configurar mediante un parámetro de costo para que, aunque la tecnología y el hardware avancen, el proceso siga siendo difícil y costoso de atacar mediante fuerza bruta.
-  **Base matemática:** Está basado en el algoritmo de cifrado simétrico Blowfish. Fue creado en 1999 por Niels Provos y David Mazières.
  
Basado en el cifrado Blowfish, bcrypt genera un hash salado y ajustable mediante un parámetro de costo (cost o work factor). Incluye un salt interno de 128 bits. Al aumentar el cost, el algoritmo realiza más rondas de cifrado, ralentizándolo. Los valores permitidos en Go son de MinCost=4 a MaxCost=31, con DefaultCost=10.
### Ventajas:
Muy usado y probado, disponibilidad en casi todos los lenguajes, y protección frente a tablas arcoíris mediante el salt.
### Desventajas:
No es memory-hard (usa sólo ~4 KB fijos), por lo que GPUs y ASICs pueden acelerar su ejecución; además, trunca contraseñas a 72 bytes (convertidas en UTF-8), por lo que cadenas más largas se ignoran. Aun así, al poder elevar el cost, permite ajustar su resistencia a la potencia de cómputo creciente. Si bien sigue siendo aceptable, bcrypt ha sido superado en protección de nichos muy especializados por algoritmos más nuevos.

## [sCrypt](https://www.cryptominerbros.com/es/blog/what-is-scrypt-algorithm/)
sCrypt es una función de derivación de claves basada en contraseñas y un algoritmo de hash diseñado para proteger contraseñas y datos confidenciales.
### Características principales
- Resistente a ataques de fuerza bruta: Fue creado por Colin Percival en 2009 (y estandarizado como RFC 7914) para dificultar los ataques informáticos a gran escala.
- Consumo intensivo de memoria: A diferencia de otros algoritmos, scrypt exige una gran cantidad de memoria RAM para calcularse, lo que encarece y frena el uso de hardware especializado (como circuitos ASIC o granjas de GPU) para descifrar claves.
- Uso con "sal" (salt): Combina la contraseña con un valor aleatorio único antes de generar el hash, evitando que contraseñas idénticas generen el mismo resultado.
- Aplicaciones en criptomonedas: Se utiliza como algoritmo de consenso de Prueba de Trabajo (PoW) en diversas criptomonedas, como Litecoin y Dogecoin.

Es un KDF memory-hard: además de la iteración de CPU (parámetros N, r, p), requiere consumir una cantidad grande de memoria para calcular el hash. Sus parámetros Go son N (power of two), r, p (paralelismo), y longitud de clave derivada. Por ejemplo, la librería oficial Go recomienda en 2017 usar N=32768, r=8, p=1 para logins interactivos, y ajustar N al máximo viable en ~100 ms. 
### Ventajas: 
Consagra la memoria como factor de costo, ofreciendo fuerte resistencia ante GPUs (que suelen tener mucho CPU pero pueden estar limitadas por memoria). 
### Desventajas: 
Para entornos con memoria muy limitada puede requerir compensar con más iteraciones; además, a muy bajos valores de memoria puede ser menos resistente que bcrypt. Scrypt no tiene longitud de contraseña máxima documentada, pero al requerir un salt único (recomendado ~8 bytes o más) debe almacenarse junto al hash.

## [Argon2](https://es.wikipedia.org/wiki/Argon2) 
[Revisar tambien](https://specopssoft.com/es/blog/significa-argon2-que-la-contrasena-es-imposible-de-descifrar/)  

Argon2 es un algoritmo criptográfico ganador de la Password Hashing Competition en 2015, diseñado específicamente para el almacenamiento seguro de contraseñas y la derivación de claves.

### ¿Cómo funciona?
- **Consumo intensivo de memoria y tiempo:** Está configurado para requerir grandes cantidades de memoria RAM y tiempo de procesamiento, lo que vuelve muy costosos y lentos los ataques de fuerza bruta o el uso de procesadores gráficos (GPU / ASIC).
- **Uso de "salt" (sal):** Añade un valor aleatorio único a cada contraseña antes de procesarla, evitando que los atacantes utilicen tablas precalculadas (como rainbow tables).
- **Estandarización:** Fue publicado oficialmente por el IETF en septiembre de 2021 bajo el estándar RFC 9106 y es recomendado por el Proyecto OWASP y plataformas como JumpCloud.

### Variantes de Argon2
Argon2 cuenta con tres versiones principales según el tipo de amenaza que se desea mitigar:
1. **Argon2d:** Maximiza la resistencia a ataques con GPU al acceder a la memoria en un orden que depende de la contraseña; es ideal para aplicaciones de criptomonedas o Proof-of-Work.
2. **Argon2i:** Optimizado para la resistencia contra ataques de canal lateral (que miden tiempos o consumo de energía del hardware); es seguro para derivación de claves.
3. **Argon2id:** Es una combinación híbrida de Argon2d y Argon2i. Es la opción recomendada y el estándar actual por defecto para el hash de contraseñas en aplicaciones web y bases de datos.
   
### Parámetros
Argon2 permite configurar principalmente tres parámetros:
- **Memoria:** Cantidad de memoria RAM utilizada durante el cálculo, expresada normalmente en KiB.
- **Tiempo:** Número de iteraciones realizadas durante el proceso.
- **Paralelismo:** Cantidad de hilos utilizados para realizar el cálculo.
  
El RFC 9106 presenta diferentes configuraciones dependiendo de los recursos disponibles. Entre ellas se encuentra una configuración de mayor seguridad con 2 GiB de memoria, 1 iteración y 4 hilos, y otra orientada a sistemas con menos memoria que utiliza 64 MiB, 3 iteraciones y 4 hilos.
Estos valores no deben interpretarse como valores universales para cualquier sistema. Los parámetros deben probarse en el hardware donde se ejecutará la aplicación, buscando que el proceso de autenticación tenga un costo suficientemente alto para un atacante sin afectar excesivamente la experiencia del usuario.

### Ventajas:
Ofrece una alta resistencia frente a ataques de fuerza bruta y ataques realizados mediante GPU o hardware especializado gracias a su consumo configurable de memoria. Además, fue diseñado específicamente para el almacenamiento seguro de contraseñas y permite ajustar su nivel de seguridad conforme aumente la capacidad del hardware.
### Desventajas:
Su principal desventaja es que puede requerir una cantidad considerable de memoria y recursos de procesamiento, especialmente cuando se utilizan configuraciones de seguridad elevadas. Esto puede ser un inconveniente en servidores con recursos limitados.

## [PBKDF2](https://en.wikipedia.org/wiki/PBKDF2) 
[Revisar tambien](https://www-hackerone-com.translate.goog/blog/understanding-benefits-key-derivation-functions-deep-dive-pbkdf2?_x_tr_sl=en&_x_tr_tl=es&_x_tr_hl=es&_x_tr_pto=tc)  

PBKDF2 (Password-Based Key Derivation Function 2) es una función de derivación de claves basada en contraseñas. Utiliza una función HMAC, como HMAC-SHA256, y aumenta el costo del cálculo mediante un número configurable de iteraciones.

### ¿Cómo funciona?
PBKDF2 protege las contraseñas mediante dos mecanismos principales:
- **Sal (Salting):** Se añade una secuencia de bits aleatorios (la sal) a la contraseña antes de aplicar la función hash. Esto garantiza que dos usuarios con la misma contraseña generen hashes completamente diferentes, neutralizando los ataques por tablas arcoíris.
- **Iteraciones (Iteración):** La función hash interna (como HMAC-SHA256) se repite miles o millones de veces (c). Esto ralentiza el proceso de forma intencionada, lo que encarece y dificulta enormemente los ataques de fuerza bruta o de diccionario.

A diferencia de algoritmos memory-hard como scrypt o Argon2, PBKDF2 no requiere grandes cantidades de memoria. Su principal mecanismo de protección consiste en aumentar el número de iteraciones necesarias para obtener cada resultado. Esto hace que el proceso sea más lento, pero las funciones utilizadas, como SHA-256, son muy eficientes en hardware moderno, especialmente en GPU, por lo que los atacantes pueden realizar una gran cantidad de intentos por segundo.

En cuanto a la configuración, las recomendaciones actuales utilizan cientos de miles de iteraciones. Por ejemplo, para HMAC-SHA256 pueden utilizarse valores superiores a 300 000 iteraciones, aunque el número adecuado debe ajustarse según el hardware y los requisitos del sistema.

En Go puede utilizarse mediante el paquete crypto/pbkdf2 en las versiones que lo incluyen, o mediante golang.org/x/crypto/pbkdf2 en otros entornos. Su principal ventaja es la compatibilidad e interoperabilidad con sistemas existentes y estándares ampliamente utilizados. Sin embargo, frente a algoritmos modernos memory-hard, su resistencia ante ataques realizados con hardware especializado es menor.

## [Yescrypt](https://en-wikipedia-org.translate.goog/wiki/Yescrypt?_x_tr_sl=en&_x_tr_tl=es&_x_tr_hl=es&_x_tr_pto=tc) 
[Revisar tambien](https://prezi.com/p/vv53rbo0kx8u/introduccion-a-yescrypt/) 

Yescrypt es una función de derivación de claves basada en contraseñas (KDF) y un esquema de hash de contraseñas, diseñado por Alexander Peslyak (Solar Designer) como una evolución moderna del algoritmo scrypt.

### Características principales
- **Algoritmo memory-hard:** Requiere el uso combinado de la memoria RAM y el procesador (CPU) para calcular los hashes, lo que encarece y dificulta los ataques a gran escala.
- **Resistencia a hardware especializado:** Está diseñado para ser muy difícil de acelerar mediante tarjetas gráficas (GPU), circuitos integrados (ASIC) o matrices de puertas (FPGA).
- **Uso en Linux:** Es el algoritmo de hash de contraseñas predeterminado en diversas distribuciones modernas de Linux (como Fedora, Debian, Ubuntu, Arch Linux y RHEL), gestionado habitualmente a través de libxcrypt.
- **Funcionamiento seguro:** Convierte las contraseñas en valores irreversibles e incluye mecanismos de salt (datos aleatorios añadidos a la contraseña) para proteger la base de datos contra ataques de fuerza bruta y de diccionario.

### Ventajas
- **Alta escalabilidad:** Está diseñado para mantener una resistencia cercana a la óptima en un amplio rango de memoria, desde KB hasta TB.
- **Resistencia frente a hardware especializado:** Puede utilizar cachés L1/L2 incluso cuando se configura con poca memoria, aumentando la dificultad para GPUs y ASICs.
- **Adecuado para despliegues grandes:** Puede utilizarse en servidores de autenticación masiva, donde se necesita una alta resistencia frente a ataques.
- **Adopción en sistemas modernos:** En sistemas Linux actuales se ha adoptado como KDF por defecto. Arch Linux, Debian 11+ y Fedora 35+ utilizan yescrypt en shadow por defecto.

### Desventajas
- **Complejidad:** Es muy complejo de configurar debido a la cantidad de parámetros disponibles.
- **Menor difusión:** Es menos utilizado que alternativas como bcrypt, scrypt o Argon2.
- **Soporte limitado:** Pocos proyectos lo soportan de forma nativa.

yescrypt es una evolución de scrypt diseñada para máxima escalabilidad. Introduce un ROM compartido de gran tamaño, que puede alcanzar decenas de GB, además de otros parámetros avanzados.  
También está diseñado para resistir ataques en entornos masivos. Incluso con poca memoria interna, utiliza parte de las cachés L1/L2 para aumentar la dificultad de los ataques realizados mediante GPUs y ASICs.

# Definición del algoritmo y parámetros de seguridad a utilizar

## Comparación de los algoritmos investigados
A partir de la investigación anterior, se compararon los cinco algoritmos tomando en cuenta su nivel de seguridad, su facilidad de implementación en Go y su consumo de recursos en el servidor.

| Algoritmo | Memory-hard | Soporte en Go                          | Complejidad de configuración                        |
|-----------|-------------|----------------------------------------|-----------------------------------------------------|
| bcrypt    | No          | Oficial (`golang.org/x/crypto/bcrypt`) | Baja (un solo parámetro)                            |
| scrypt    | Sí          | Oficial (`golang.org/x/crypto/scrypt`) | Media (N, r, p)                                     |
| Argon2id  | Sí          | Oficial (`golang.org/x/crypto/argon2`) | Media (memoria, tiempo, paralelismo)                |
| PBKDF2    | No          | Oficial (`crypto/pbkdf2`)              | Baja, pero requiere cientos de miles de iteraciones |
| yescrypt  | Sí          | Sin soporte oficial                    | Alta                                                |

Se descartó **yescrypt** por su complejidad y por no contar con una librería oficial en Go. Se descartó **PBKDF2** porque, como se indicó en la investigación, es el más fácil de acelerar con GPU. **scrypt** y **Argon2id** ofrecen mayor resistencia por ser memory-hard, pero su librería en Go solo genera el hash: el salt, el formato de almacenamiento y la comparación segura deben programarse de forma manual, lo que aumenta la posibilidad de cometer errores. Además, cada inicio de sesión reservaría varios MiB de memoria en el servidor.

## Algoritmo seleccionado: bcrypt
Se eligió **bcrypt** para el almacenamiento de las contraseñas del sistema EMI. Es un algoritmo ampliamente probado, aceptado por OWASP y su librería oficial en Go resuelve por sí sola la generación del salt, el formato del hash y la comparación, lo que reduce el riesgo de errores en la implementación. Para el tamaño y el tipo de uso de este sistema, ofrece un nivel de seguridad adecuado.

Las contraseñas nunca se almacenarán en texto plano. El backend generará el hash antes de guardarlas en la columna `contraseña` de la tabla `usuarios`, por lo que el frontend solo se encarga de enviar la contraseña al servidor.

## Parámetros definidos
- **Librería:** `golang.org/x/crypto/bcrypt`, mantenida por el equipo oficial de Go. El proyecto ya utiliza `golang.org/x/crypto` como dependencia.
- **Costo (cost factor):** 12. Es mayor al valor por defecto de Go (`DefaultCost = 10`) y hace que cada hash tarde aproximadamente 250 ms. Este tiempo no afecta la experiencia del usuario al iniciar sesión, pero vuelve muy lentos los ataques de fuerza bruta. Al ser configurable, podrá aumentarse en el futuro conforme mejore el hardware.
- **Salt:** generado automáticamente por bcrypt, de 128 bits y distinto para cada contraseña. Se guarda dentro del propio hash, por lo que no se necesita una columna adicional.
- **Longitud del hash:** 60 caracteres. La columna `contraseña`, definida como `nvarchar(255)`, tiene espacio suficiente.
- **Longitud de la contraseña:** mínimo 8 caracteres, según la recomendación de OWASP, y máximo 72 bytes. Este límite responde a la desventaja señalada en la investigación: bcrypt ignora lo que pase de 72 bytes, por lo que el backend rechazará las contraseñas más largas en lugar de truncarlas sin avisar.

## Uso en el sistema
- **Registro de usuario:** el backend generará el hash con `bcrypt.GenerateFromPassword(contraseña, 12)` y guardará el resultado en la base de datos.
- **Inicio de sesión:** el backend comparará la contraseña ingresada con el hash almacenado mediante `bcrypt.CompareHashAndPassword`. Si no coinciden, se responderá con un mensaje genérico de credenciales inválidas, sin indicar si el error está en el usuario o en la contraseña.

