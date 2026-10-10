# HU 4 Documentación - Implementación de traducción y soporte multilingüe

### 1. Modelo elegido
Para la traducción de la aplicación se eligió la librería **ngx-translate**. Esta funciona con archivos JSON (uno por idioma) que se cargan mientras la aplicación está corriendo, por lo que el usuario puede cambiar de idioma con un botón sin recargar la página.

Se descartaron las otras dos opciones analizadas en el Sprint 1. **@angular/localize** no permite cambiar de idioma dentro de la aplicación y obliga a compilar y desplegar una versión por cada idioma, lo que complica el despliegue con Nginx. **Transloco** sí cumple lo que necesitamos, pero tiene una comunidad más pequeña y más funciones de las que EMI ocupa, por lo que ngx-translate resulta más fácil de aprender para el equipo.

### 2. Por qué ngx-translate
- Permite cambiar el idioma desde la tableta con un solo toque, sin recargar y sin perder la sesión.
- Solo hay una aplicación que compilar y desplegar.
- Es la librería más usada, así que hay mucha documentación y ejemplos.
- Su uso es sencillo: un pipe en el HTML y un servicio para cambiar el idioma.
- La versión 18 es compatible con Angular 22, que es la que usa el proyecto.

Como desventaja, traduce mientras la aplicación corre y la mantiene la comunidad y no el equipo de Angular, pero esto no afecta a un sistema del tamaño de EMI.

### 3. Idiomas soportados
La aplicación tendrá **español** (`es`), que será el idioma por defecto, e **inglés** (`en`). Se usan los códigos de dos letras (ISO 639-1) porque son los mismos que usan el navegador y el atributo `lang` de HTML. Para agregar otro idioma solo hay que crear su archivo JSON con las mismas claves y agregar su código a la lista de idiomas permitidos.

### 4. Estructura de archivos
Los archivos de traducción se guardan en la carpeta `public/i18n/` (en Angular 22 los archivos estáticos van en `public/` y ya no en `src/assets/`).

```
src/
└── app/
    ├── core/            (aquí va IdiomaService)
    └── features/
        ├── auth/
        ├── users/
        ├── supplies/
        ├── suppliers/
        └── movements/

public/
└── i18n/
    ├── es.json
    └── en.json
```

Ambos archivos deben tener **las mismas claves**. Cada módulo tiene su bloque y lo que se repite en varias pantallas (guardar, cancelar) va en `common`. Ejemplo:

```json
{
  "auth": { "login": "Iniciar sesión", "email": "Correo electrónico" },
  "users": { "title": "Usuarios", "create": "Crear usuario" },
  "common": { "save": "Guardar", "cancel": "Cancelar" },
  "errors": { "unauthorized": "Credenciales inválidas o sesión expirada." }
}
```

En las plantillas se usa así: `{{ 'auth.login' | translate }}`.

### 5. Cómo se implementa
La solución se divide en tres partes:

1. **Configuración global:** en `app.config.ts` se registra ngx-translate con el idioma de respaldo (`es`) y el cargador que lee los JSON. También hay que agregar `provideHttpClient()`, porque sin este el cargador no puede descargar los archivos.

```typescript
provideHttpClient(),
provideTranslateService({
  fallbackLang: 'es',
  loader: provideTranslateHttpLoader({ prefix: '/i18n/', suffix: '.json' })
})
```

2. **IdiomaService** (carpeta `core/`): es el único lugar donde se cambia el idioma. Decide el idioma inicial, hace el cambio, guarda la preferencia y actualiza el atributo `lang` de la página. Así los componentes no llaman a `TranslateService` por su cuenta.

3. **SelectorIdiomaComponent:** es un componente que solo dibuja el selector y le pide el cambio a `IdiomaService`. Son dos botones, **ES** y **EN**, ubicados en la barra superior y en el inicio de sesión. Como se usa en tabletas, cada botón mide al menos 44 × 44 píxeles, y tienen una etiqueta accesible con el nombre del idioma ("Español" y "English").

### 6. Cómo funciona el cambio de idioma
Cuando el usuario toca un idioma, el selector llama a `cambiarIdioma()` del servicio. El servicio revisa que el idioma sea permitido y usa `TranslateService.use()`. La primera vez que se pide un idioma, ngx-translate descarga su JSON (por ejemplo `/i18n/en.json`) y lo guarda en memoria, por lo que los siguientes cambios son inmediatos. Después, el pipe `translate` actualiza todos los textos de la pantalla sin recargar la página ni perder la sesión.

### 7. Idioma inicial y preferencia guardada
Al arrancar la aplicación se escoge el idioma en este orden:

1. La preferencia guardada en el `localStorage` (clave `emi.idioma`), si es válida.
2. El idioma del navegador (`navigator.language`, tomando solo las dos primeras letras: `es-CR` pasa a `es`).
3. El español, que es el idioma por defecto.

La preferencia se guarda con `localStorage`, dentro de un `try/catch`, porque algunos navegadores lo bloquean y la aplicación no debe fallar por eso. Hay que tomar en cuenta que la preferencia queda guardada en la **tableta** y no en el usuario, así que si varias personas comparten la tableta verán el último idioma elegido. Guardarla por usuario podría hacerse más adelante agregando una columna `idioma` a la tabla `usuarios`, pero no está en el alcance actual.

### 8. Traducciones faltantes
Si una clave está en `es.json` pero no en `en.json`, se muestra el texto en español (idioma de respaldo). Si no está en ninguno, se muestra la clave (por ejemplo `auth.login`), lo que ayuda a encontrar el error. Por eso la regla del equipo es que **todo texto nuevo se agrega en ambos archivos**, con la misma clave.

### 9. Fechas y números
Traducir los textos no cambia el formato de las fechas. Para esto se registran los datos del español con `registerLocaleData(localeEs)` y se le pasa el idioma activo al pipe `date`. Así el 9 de octubre de 2026 se ve como `9/10/26` en español y `10/9/26` en inglés.

### 10. Mensajes de error del backend
La traducción la hace solo el frontend. Actualmente la API en Go devuelve algunos mensajes fijos en español (por ejemplo "Contraseña incorrecta"), que no cambiarían al elegir inglés, por lo que no se muestran directamente. En su lugar, el frontend usa el **código de estado HTTP**:

| Código | Clave de traducción |
|---|---|
| 400 | `errors.badRequest` |
| 401 | `errors.unauthorized` |
| 404 | `errors.notFound` |
| 500 | `errors.server` |

Como mejora a futuro, se podría pedirle al equipo de backend que agregue un campo `code` en los errores (por ejemplo `CREDENCIALES_INVALIDAS`) para mostrar mensajes más específicos.

### 11. Conclusión
Con ngx-translate el sistema EMI podrá cambiar entre español e inglés desde la tableta de forma rápida y sencilla, con una sola aplicación desplegada. La solución queda organizada en tres partes (configuración, servicio y selector), lo que facilita su mantenimiento y permite agregar más idiomas en el futuro.
