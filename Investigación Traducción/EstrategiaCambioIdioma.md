# HU 4	Investigación - Implementación de traducción y soporte multilingüe


### 1. Idiomas soportados
La aplicación se ofrecerá en dos idiomas: **español**, identificado con el código `es`, e **inglés**, identificado con el código `en`. Se utilizan los códigos de dos letras del estándar ISO 639-1 porque son los mismos que emplean el navegador (`navigator.language`), el atributo `lang` de HTML y los nombres de los archivos de traducción, lo que permite usar un único identificador en todas las capas de la aplicación.

El español será el idioma por defecto, ya que es la lengua principal del personal que utilizará el sistema. La estrategia es extensible: para incorporar un idioma nuevo basta con crear su archivo JSON con las mismas claves y añadir su código a la lista de idiomas permitidos del servicio de idioma, sin modificar los componentes de la aplicación.

### 2. Arquitectura de la solución
El cambio de idioma se resolverá con tres piezas, cada una con una responsabilidad única.

La primera es la **configuración global de ngx-translate**, que se registra en `app.config.ts` junto con los demás proveedores de la aplicación. En ella se indica el idioma de respaldo y el cargador HTTP que obtendrá los archivos JSON. Como el proyecto utiliza componentes standalone, se emplean las funciones de proveedor de la librería en lugar de módulos:

```typescript
provideTranslateService({
  fallbackLang: 'es',
  loader: provideTranslateHttpLoader({ prefix: '/i18n/', suffix: '.json' })
})
```

La segunda es un **servicio de idioma** (`IdiomaService`), ubicado en la carpeta `core/` y registrado como *singleton* (`providedIn: 'root'`). Este servicio será el único punto de la aplicación autorizado para cambiar el idioma: determina el idioma inicial, ejecuta el cambio, guarda la preferencia y actualiza el atributo `lang` del documento. Centralizar esta lógica evita que cada componente invoque `TranslateService` por su cuenta y garantiza que el cambio se comporte siempre de la misma forma.

La tercera es un **componente selector de idioma** (`SelectorIdiomaComponent`), un componente standalone de presentación que solo dibuja el control visual y delega la acción en `IdiomaService`. Al no contener lógica propia, puede reutilizarse sin cambios en cualquier pantalla.

### 3. Interfaz de selección
El selector se ubicará en la **barra superior** de la aplicación, de modo que esté disponible en todas las pantallas, y también en la **pantalla de inicio de sesión**, para que el usuario pueda elegir su idioma antes de autenticarse.

Se implementará como un botón segmentado con las opciones **ES** y **EN**, donde la opción activa aparece resaltada y la otra se puede elegir con un solo toque. Las abreviaturas mantienen el control compacto dentro de la barra, pero cada opción tendrá como etiqueta accesible (`aria-label`) el nombre del idioma escrito en su propia lengua, **"Español"** y **"English"**, para que un lector de pantalla las anuncie correctamente y cualquier persona reconozca su idioma aunque no entienda el que está activo. El estado seleccionado se expondrá con `aria-pressed`. Dado que el sistema se utilizará principalmente en tabletas, cada opción tendrá un área táctil mínima de 44 × 44 píxeles, que es el tamaño recomendado por las pautas de accesibilidad para controles táctiles.

### 4. Flujo de cambio de idioma
Cuando el usuario elige un idioma, el componente selector llama al método `cambiarIdioma()` de `IdiomaService`. El servicio valida que el código pertenezca a la lista de idiomas permitidos y ejecuta `TranslateService.use()` con ese código.

Si es la primera vez que se solicita ese idioma durante la sesión, ngx-translate realiza una petición HTTP `GET` a `/i18n/en.json` (o `/i18n/es.json`) mediante el cargador configurado. Una vez descargado, el contenido queda almacenado en memoria, por lo que los cambios posteriores entre idiomas ya cargados son inmediatos y no generan nuevas peticiones.

Al completarse la carga, ngx-translate emite el evento `onLangChange`. El pipe `translate`, que se usa en las plantillas con la sintaxis `{{ 'auth.login' | translate }}`, está suscrito a ese evento y vuelve a resolver sus claves, por lo que todos los textos visibles se actualizan sin recargar la página y sin perder el estado de la aplicación ni la sesión del usuario.

Finalmente, el servicio guarda el código elegido en `localStorage` y actualiza el atributo `lang` del elemento raíz con `document.documentElement.lang = codigo`. Este atributo permite que los lectores de pantalla pronuncien el contenido con el idioma correcto y que el navegador no ofrezca traducciones automáticas innecesarias.

### 5. Resolución del idioma inicial
Al arrancar la aplicación, `IdiomaService` determina qué idioma mostrar siguiendo un orden de prioridad de tres niveles.

En primer lugar, consulta si existe una **preferencia guardada** en `localStorage` y si su valor corresponde a un idioma permitido; de ser así, la utiliza. Si no existe una preferencia válida, toma el **idioma del navegador** mediante `navigator.language`, conserva solo las dos primeras letras (de modo que `es-CR` o `en-US` se interpreten como `es` y `en`) y lo usa si es uno de los idiomas soportados. Si ninguna de las dos condiciones se cumple, aplica el **idioma por defecto**, que es el español.

```typescript
const guardado = leerPreferencia();                                // 'es' | 'en' | null
const navegador = navigator.language.slice(0, 2).toLowerCase();
const inicial = guardado ?? (IDIOMAS.includes(navegador) ? navegador : 'es');
```

Esta resolución se ejecutará durante el arranque de la aplicación, por ejemplo con `provideAppInitializer`, para que la primera pantalla se muestre directamente en el idioma correcto y no se produzca un parpadeo de textos en otro idioma.

### 6. Persistencia de la preferencia
La preferencia se almacenará en el `localStorage` del navegador bajo la clave `emi.idioma`, con el valor `es` o `en`. Se eligió este mecanismo porque conserva el dato entre sesiones, incluso después de cerrar la aplicación o reiniciar la tableta, y no requiere cambios en la base de datos ni en la API.

El acceso a `localStorage` se hará dentro de bloques `try/catch`, porque algunos navegadores lo bloquean en modo privado o con restricciones de almacenamiento. Si la lectura o la escritura fallan, la aplicación seguirá funcionando con el idioma resuelto en ese momento, aunque la preferencia no se conserve para la siguiente apertura.

Es importante señalar que, con este enfoque, la preferencia queda asociada al **dispositivo** y no al usuario: si varias personas comparten la misma tableta, todas verán el último idioma elegido en ella. Si en el futuro se requiere una preferencia por usuario, se podría agregar una columna `idioma` a la tabla `usuarios` y aplicarla al iniciar sesión, pero esa ampliación no forma parte del alcance actual.

### 7. Manejo de traducciones faltantes
Si una clave existe en `es.json` pero no en `en.json`, ngx-translate utilizará el texto del idioma de respaldo configurado (`fallbackLang: 'es'`), de modo que el usuario verá el texto en español en lugar de un espacio vacío o de la clave interna, como `auth.login`. Si la clave tampoco existe en el idioma de respaldo, la librería mostrará la propia clave, lo que facilita detectar el error durante el desarrollo.

Para prevenir estos casos, el equipo adoptará como regla que todo texto nuevo se agregue en **ambos archivos** con la misma clave y la misma jerarquía, siguiendo la estructura por módulos definida en *EstructuraArchivosIdioma.md*.

### 8. Formato de fechas y números
La traducción de textos no modifica por sí sola el formato de fechas y números. Los pipes nativos de Angular, como `date` y `number`, utilizan el `LOCALE_ID` de la aplicación, que se fija al arrancar y no cambia en tiempo de ejecución. Por ello, se registrarán los datos regionales del español con `registerLocaleData(localeEs)` y los pipes recibirán el idioma activo como parámetro, por ejemplo `{{ movimiento.fecha | date:'shortDate':undefined:idiomaActual }}`. Así, la fecha del 9 de octubre de 2026 se mostrará como `9/10/26` en español (día/mes) y como `10/9/26` en inglés (mes/día), de forma coherente con el idioma elegido.

### 9. Integración con el backend
La traducción será responsabilidad exclusiva del frontend; el backend no necesita conocer el idioma seleccionado por el usuario.

Actualmente, la API desarrollada en Go devuelve los errores con dos formatos: la función `errorResponse()` envía el texto técnico del error en el campo `error`, y algunos manejadores envían mensajes fijos en español en el campo `message`, como `"Contraseña incorrecta"` en el inicio de sesión. Esos textos no deben mostrarse directamente al usuario, porque no cambiarían al seleccionar inglés.

Como estrategia inicial, el frontend determinará el mensaje a partir del **código de estado HTTP** de la respuesta, que es independiente del idioma: un `401` se traducirá con la clave `errors.unauthorized`, un `404` con `errors.notFound`, un `400` con `errors.badRequest` y un `500` con `errors.server`. Esta solución funciona con el backend actual, sin modificarlo.

Como mejora posterior, se propone que la API agregue a sus respuestas de error un campo `code` con un identificador estable, por ejemplo `CREDENCIALES_INVALIDAS`, que el frontend usaría como clave de traducción para mostrar mensajes más específicos. Este cambio corresponde al equipo de backend y deberá acordarse con él.

### 10. Ubicación de los archivos de traducción
En la versión de Angular que utiliza el proyecto (v22), los archivos estáticos ya no se colocan en `src/assets/`, sino en la carpeta `public/`, tal como lo define la configuración `assets` de `angular.json`. Por lo tanto, los archivos de idioma se ubicarán en `public/i18n/es.json` y `public/i18n/en.json`, y el servidor los publicará en la ruta `/i18n/`, que es la que utiliza el cargador HTTP configurado en la sección 2. La estructura y el contenido de los archivos se mantienen tal como se definieron en *EstructuraArchivosIdioma.md*.

En cuanto a la versión de la librería, la versión estable actual de `@ngx-translate/core` y `@ngx-translate/http-loader` es la **18**, que es compatible con Angular 18 o superior y, por lo tanto, con el proyecto.
