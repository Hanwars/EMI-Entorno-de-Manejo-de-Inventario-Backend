# HU 4 — Documentación: traducción y soporte multilingüe con ngx-translate

**Proyecto:** Sistema EMI (Entorno de Manejo de Inventario) — Coconutz BrewHouse
**Historia de usuario:** HU 4 — Implementación de traducción y soporte multilingüe
**Decisión:** el frontend usará **ngx-translate**, una librería de traducción en tiempo de ejecución, con archivos JSON por idioma.
**Documentos relacionados:** *Investigación HU 4*, *Alternativas para implementar traducciones en Angular (Sprint 1)* y *Estructura de archivos de idioma*.

---

## 1. Resumen de la decisión

| Aspecto | Definición |
|---|---|
| Librería | `@ngx-translate/core` y `@ngx-translate/http-loader` |
| Versión | 18 (estable desde junio de 2026; compatible con Angular 18 a 22) |
| Tipo de traducción | En tiempo de ejecución, con una sola aplicación compilada |
| Idiomas | Español (`es`, por defecto) e inglés (`en`) |
| Archivos de traducción | `public/i18n/es.json` y `public/i18n/en.json` |
| Idioma de respaldo | `es` |
| Idioma inicial | Preferencia guardada → idioma del navegador → español |
| Persistencia | `localStorage`, clave `emi.idioma` |
| Selector | Botón segmentado **ES / EN** en la barra superior y en el inicio de sesión |
| Traducción de errores del backend | Por código de estado HTTP, sin modificar la API |

---

## 2. Contexto y requisitos

El personal del restaurante usa el sistema desde **tabletas** y debe poder cambiar el idioma con un solo toque, sin recargar la aplicación ni perder la sesión. El sistema se despliega en una **red LAN** sobre una computadora de la empresa, por lo que conviene tener **una única aplicación** que compilar y publicar. El equipo está aprendiendo Angular, así que la solución debe ser sencilla de usar y de mantener.

---

## 3. Alternativas evaluadas

Angular ofrece dos enfoques: traducir al **compilar** (una versión de la aplicación por idioma) o traducir al **ejecutar** (una sola aplicación que carga los textos desde archivos JSON).

| Criterio | `@angular/localize` | **ngx-translate** | Transloco |
|---|---|---|---|
| Momento de la traducción | Compilación | Ejecución | Ejecución |
| Cambio de idioma dentro de la app | No (hay que ir a otra dirección) | Sí | Sí |
| Versiones a compilar y desplegar | Una por idioma | Una | Una |
| Origen | Equipo de Angular | Comunidad | Comunidad |
| Rendimiento | Muy bueno (textos ya compilados) | Bueno | Bueno |
| Plurales, fechas y números | Muy buen soporte | Menos estandarizado | Buen soporte |
| Comunidad y documentación | Oficial | La más amplia | Más pequeña |
| Curva de aprendizaje | Media | Baja | Media (más funciones de las necesarias) |
| Instalación | Manual | Manual | Incluye `ng add` |

### Motivos de descarte

- **`@angular/localize`:** no permite cambiar de idioma dentro de la aplicación y obliga a compilar y desplegar una versión por idioma, lo que complica la publicación con Nginx. Incumple el requisito principal de la HU.
- **Transloco:** cumple los requisitos y tiene mejor soporte de tipos y de *Signals*, pero cuenta con una comunidad menor y más funciones de las que EMI necesita. Entre las dos librerías de ejecución, la diferencia es en gran parte de preferencia. Sigue siendo una alternativa válida si el proyecto crece.

---

## 4. Justificación de la elección

1. **Cambio de idioma desde la tableta con un botón**, sin recargar y sin perder el estado de la aplicación.
2. **Un solo despliegue.** Los archivos `es.json` y `en.json` se publican junto con la aplicación; no hace falta ninguna dependencia externa, lo que encaja con un entorno LAN.
3. **Es la librería más usada**, con abundante documentación y respuestas a problemas comunes, algo valioso para un equipo que está aprendiendo.
4. **Uso sencillo:** un *pipe* en las plantillas (`{{ 'auth.login' | translate }}`) y un servicio para cambiar el idioma.
5. **Compatible con componentes standalone y con la versión de Angular del proyecto.** La versión 18 de la librería soporta Angular 18 a 22.

**Desventajas aceptadas:** traduce mientras la aplicación corre, por lo que es algo menos eficiente que la opción oficial; sus plurales son menos estandarizados; y la mantiene la comunidad, no el equipo de Angular. Ninguna afecta a un sistema de este tamaño.

---

## 5. Idiomas soportados

| Idioma | Código | Rol |
|---|---|---|
| Español | `es` | Por defecto y de respaldo |
| Inglés | `en` | Alternativo |

Se usan los códigos de dos letras de **ISO 639-1** porque coinciden con los que emplea el navegador (`navigator.language`), con el atributo `lang` de HTML y con los nombres de los archivos de traducción. Así se usa un único identificador en todas las capas.

La estrategia es extensible: para agregar un idioma se crea su archivo JSON con las mismas claves y se añade su código a la lista de idiomas permitidos de `IdiomaService`, sin tocar los componentes.

---

## 6. Arquitectura de la solución

Tres piezas con una responsabilidad cada una:

| Pieza | Responsabilidad |
|---|---|
| **Configuración global** (`app.config.ts`) | Registra ngx-translate: idioma de respaldo y cargador HTTP de los JSON. |
| **`IdiomaService`** (`core/`, *singleton*) | Único punto autorizado para cambiar el idioma: determina el idioma inicial, ejecuta el cambio, guarda la preferencia y actualiza el atributo `lang`. |
| **`SelectorIdiomaComponent`** | Componente de presentación. Dibuja el control y delega la acción en `IdiomaService`. Se reutiliza sin cambios en cualquier pantalla. |

Centralizar la lógica en el servicio evita que cada componente use `TranslateService` por su cuenta y garantiza que el cambio se comporte siempre igual.

```text
Usuario
   │ toca "EN"
   ▼
SelectorIdiomaComponent ──► IdiomaService.cambiarIdioma('en')
                                   │
                                   ├─► TranslateService.use('en') ──► GET /i18n/en.json (solo la primera vez)
                                   ├─► localStorage  emi.idioma = 'en'
                                   └─► <html lang="en">
                                   │
                                   ▼
              El pipe `translate` vuelve a resolver las claves
              → todos los textos cambian sin recargar la página
```

---

## 7. Estructura de archivos

```text
src/
└── app/
    ├── core/
    │   └── idioma.service.ts
    ├── shared/
    │   └── selector-idioma/
    │       └── selector-idioma.component.ts
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

> **Ubicación de los JSON:** en la versión de Angular del proyecto (v22), los archivos estáticos van en la carpeta `public/`, según la opción `assets` de `angular.json`. Esto sustituye la ubicación `src/assets/i18n/` de la definición inicial. El servidor los publica en `/i18n/`, la ruta que usa el cargador HTTP.

---

## 8. Estructura de los archivos JSON

### Convenciones

- Un archivo por idioma, nombrado con el código del idioma: `es.json`, `en.json`.
- **Ambos archivos tienen exactamente las mismas claves y la misma jerarquía.**
- Un bloque principal por módulo de `features/` (`auth`, `users`, `inventory`, etc.), más tres bloques transversales: `app`, `common` y `errors`.
- Los textos reutilizables (Guardar, Cancelar, Confirmar) van en `common`.
- Las claves se escriben en inglés y en `camelCase`; los valores, en el idioma del archivo.
- Los textos con datos variables usan parámetros: `"confirmExit": "¿Registrar salida de {{cantidad}} {{unidad}}?"`.

### `public/i18n/es.json`

```json
{
  "app": {
    "title": "EMI",
    "loading": "Cargando..."
  },
  "auth": {
    "login": "Iniciar sesión",
    "email": "Correo electrónico",
    "password": "Contraseña"
  },
  "users": {
    "title": "Usuarios",
    "create": "Crear usuario",
    "edit": "Editar usuario",
    "delete": "Eliminar usuario"
  },
  "inventory": {
    "title": "Inventario",
    "create": "Crear insumo",
    "search": "Buscar insumo"
  },
  "common": {
    "save": "Guardar",
    "cancel": "Cancelar",
    "confirm": "Confirmar",
    "language": "Idioma"
  },
  "errors": {
    "badRequest": "Los datos enviados no son válidos.",
    "unauthorized": "Credenciales inválidas o sesión expirada.",
    "notFound": "No se encontró el recurso solicitado.",
    "server": "Ocurrió un error en el servidor. Intente de nuevo.",
    "unknown": "Ocurrió un error inesperado."
  }
}
```

### `public/i18n/en.json`

```json
{
  "app": {
    "title": "EMI",
    "loading": "Loading..."
  },
  "auth": {
    "login": "Log in",
    "email": "Email",
    "password": "Password"
  },
  "users": {
    "title": "Users",
    "create": "Create user",
    "edit": "Edit user",
    "delete": "Delete user"
  },
  "inventory": {
    "title": "Inventory",
    "create": "Create supply",
    "search": "Search supply"
  },
  "common": {
    "save": "Save",
    "cancel": "Cancel",
    "confirm": "Confirm",
    "language": "Language"
  },
  "errors": {
    "badRequest": "The submitted data is not valid.",
    "unauthorized": "Invalid credentials or expired session.",
    "notFound": "The requested resource was not found.",
    "server": "A server error occurred. Please try again.",
    "unknown": "An unexpected error occurred."
  }
}
```

---

## 9. Instalación y configuración

### Instalación

```bash
npm install @ngx-translate/core @ngx-translate/http-loader
```

### `app.config.ts`

```typescript
import { ApplicationConfig, inject, provideAppInitializer } from '@angular/core';
import { provideHttpClient } from '@angular/common/http';
import { registerLocaleData } from '@angular/common';
import localeEs from '@angular/common/locales/es';
import { provideTranslateService } from '@ngx-translate/core';
import { provideTranslateHttpLoader } from '@ngx-translate/http-loader';
import { IdiomaService } from './core/idioma.service';

registerLocaleData(localeEs); // datos regionales del español para fechas y números

export const appConfig: ApplicationConfig = {
  providers: [
    provideHttpClient(), // el cargador HTTP lo necesita para descargar los JSON
    provideTranslateService({
      fallbackLang: 'es',
      loader: provideTranslateHttpLoader({ prefix: '/i18n/', suffix: '.json' }),
    }),
    // Resuelve el idioma inicial antes de mostrar la primera pantalla
    provideAppInitializer(() => inject(IdiomaService).inicializar()),
  ],
};
```

Puntos a tener en cuenta:

- **`provideHttpClient()` es indispensable.** Sin él, el cargador no puede descargar los archivos de traducción.
- En la versión 18 de la librería ya no existe `TranslateModule`: se usan funciones de proveedor (`provideTranslateService`). La opción de respaldo se llama `fallbackLang`; el antiguo nombre `defaultLang` fue reemplazado.
- **Nginx** debe servir la carpeta `i18n/` como archivos estáticos. Si al actualizar un texto la tableta sigue mostrando el anterior, es caché del navegador; se puede evitar agregando una versión al sufijo del cargador (por ejemplo `.json?v=1.0.0`).

---

## 10. Implementación

> Los fragmentos siguientes son una guía de implementación. Deben adaptarse a la estructura real del proyecto y verificarse al compilar.

### `core/idioma.service.ts`

```typescript
import { Injectable, inject, signal } from '@angular/core';
import { DOCUMENT } from '@angular/common';
import { TranslateService } from '@ngx-translate/core';
import { firstValueFrom } from 'rxjs';

export const IDIOMAS = ['es', 'en'] as const;
export type Idioma = (typeof IDIOMAS)[number];

const CLAVE_STORAGE = 'emi.idioma';
const IDIOMA_DEFECTO: Idioma = 'es';

@Injectable({ providedIn: 'root' })
export class IdiomaService {
  private readonly translate = inject(TranslateService);
  private readonly documento = inject(DOCUMENT);

  readonly idiomaActual = signal<Idioma>(IDIOMA_DEFECTO);

  /** Se ejecuta al arrancar la aplicación. */
  async inicializar(): Promise<void> {
    await this.aplicar(this.resolverIdiomaInicial());
  }

  /** Único punto de la aplicación autorizado para cambiar el idioma. */
  async cambiarIdioma(codigo: string): Promise<void> {
    if (!this.esPermitido(codigo)) return;
    await this.aplicar(codigo);
    this.guardarPreferencia(codigo);
  }

  private async aplicar(codigo: Idioma): Promise<void> {
    await firstValueFrom(this.translate.use(codigo));
    this.idiomaActual.set(codigo);
    this.documento.documentElement.lang = codigo;
  }

  private resolverIdiomaInicial(): Idioma {
    const guardado = this.leerPreferencia();
    if (guardado) return guardado;

    const navegador = navigator.language.slice(0, 2).toLowerCase();
    return this.esPermitido(navegador) ? navegador : IDIOMA_DEFECTO;
  }

  private esPermitido(codigo: string): codigo is Idioma {
    return (IDIOMAS as readonly string[]).includes(codigo);
  }

  private leerPreferencia(): Idioma | null {
    try {
      const valor = localStorage.getItem(CLAVE_STORAGE);
      return valor && this.esPermitido(valor) ? valor : null;
    } catch {
      return null; // almacenamiento bloqueado: se sigue sin preferencia
    }
  }

  private guardarPreferencia(codigo: Idioma): void {
    try {
      localStorage.setItem(CLAVE_STORAGE, codigo);
    } catch {
      /* si falla, la app sigue funcionando con el idioma actual */
    }
  }
}
```

### `shared/selector-idioma/selector-idioma.component.ts`

```typescript
import { Component, inject } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';
import { IdiomaService } from '../../core/idioma.service';

@Component({
  selector: 'app-selector-idioma',
  imports: [TranslatePipe],
  template: `
    <div class="selector-idioma" role="group" [attr.aria-label]="'common.language' | translate">
      @for (op of opciones; track op.codigo) {
        <button
          type="button"
          class="opcion"
          [class.activa]="idioma.idiomaActual() === op.codigo"
          [attr.aria-label]="op.nombre"
          [attr.aria-pressed]="idioma.idiomaActual() === op.codigo"
          [attr.lang]="op.codigo"
          (click)="idioma.cambiarIdioma(op.codigo)"
        >
          {{ op.etiqueta }}
        </button>
      }
    </div>
  `,
  styles: `
    .selector-idioma { display: inline-flex; }
    .opcion { min-width: 44px; min-height: 44px; } /* área táctil mínima */
    .opcion.activa { font-weight: 700; }
  `,
})
export class SelectorIdiomaComponent {
  protected readonly idioma = inject(IdiomaService);

  // El nombre accesible va en la propia lengua de cada opción
  protected readonly opciones = [
    { codigo: 'es', etiqueta: 'ES', nombre: 'Español' },
    { codigo: 'en', etiqueta: 'EN', nombre: 'English' },
  ];
}
```

### Uso en plantillas y en código

```html
<h1>{{ 'users.title' | translate }}</h1>
<button>{{ 'common.save' | translate }}</button>
<p>{{ 'movements.confirmExit' | translate: { cantidad: 5, unidad: 'L' } }}</p>
```

```typescript
// En TypeScript, cuando el texto no está en una plantilla
const mensaje = this.translate.instant('errors.server');
```

---

## 11. Flujo de cambio de idioma

1. El usuario toca una opción del selector.
2. El componente llama a `IdiomaService.cambiarIdioma()`.
3. El servicio valida que el código esté en la lista de idiomas permitidos y ejecuta `TranslateService.use()`.
4. **Primera vez en la sesión:** ngx-translate descarga `/i18n/en.json` (o `es.json`) con una petición `GET`. El contenido queda en memoria, de modo que los cambios posteriores entre idiomas ya cargados son inmediatos.
5. Los textos de pantalla se actualizan: el *pipe* `translate` vuelve a resolver sus claves, sin recargar la página y sin perder el estado ni la sesión.
6. El servicio guarda el código en `localStorage` y actualiza `<html lang>`. Ese atributo hace que los lectores de pantalla pronuncien el contenido con el idioma correcto y que el navegador no ofrezca traducciones automáticas innecesarias.

---

## 12. Idioma inicial y persistencia

### Resolución del idioma inicial

Al arrancar, `IdiomaService` aplica este orden de prioridad:

| Prioridad | Origen | Ejemplo |
|---|---|---|
| 1 | Preferencia guardada en `localStorage` (si es un idioma permitido) | `en` |
| 2 | Idioma del navegador (`navigator.language`, solo las dos primeras letras) | `es-CR` → `es`; `en-US` → `en` |
| 3 | Idioma por defecto | `es` |

Se ejecuta en el arranque (`provideAppInitializer`), así la primera pantalla aparece directamente en el idioma correcto y no hay parpadeo de textos en otro idioma.

### Persistencia

- **Dónde:** `localStorage`, clave `emi.idioma`, valor `es` o `en`.
- **Por qué:** conserva la preferencia entre sesiones, incluso al reiniciar la tableta, y no requiere cambios en la base de datos ni en la API.
- **Tolerancia a fallos:** el acceso se hace dentro de `try/catch`, porque algunos navegadores bloquean `localStorage` en modo privado. Si falla, la aplicación sigue funcionando con el idioma resuelto, aunque no se conserve la preferencia.
- **Alcance:** la preferencia queda asociada al **dispositivo**, no al usuario. Si varias personas comparten una tableta, todas verán el último idioma elegido.

---

## 13. Traducciones faltantes

| Situación | Comportamiento |
|---|---|
| La clave existe en `es.json` pero no en `en.json` | Se muestra el texto en español (idioma de respaldo). |
| La clave no existe en ningún archivo | Se muestra la propia clave (por ejemplo `auth.login`), lo que facilita detectar el error en desarrollo. |

### Regla del equipo

Todo texto nuevo se agrega en **ambos archivos**, con la misma clave y la misma jerarquía.

### Verificación automática (opcional)

El siguiente script compara las claves de `es.json` y `en.json` y avisa de las que falten o sobren. Se puede ejecutar antes de cada *commit* o en la integración continua.

`scripts/verificar-i18n.mjs`:

```javascript
// Verifica que es.json y en.json tengan exactamente las mismas claves.
// Uso: node scripts/verificar-i18n.mjs
import { readFileSync } from "node:fs";

const DIR = "public/i18n";
const BASE = "es";
const OTROS = ["en"];

function claves(obj, prefijo = "") {
  return Object.entries(obj).flatMap(([k, v]) =>
    v !== null && typeof v === "object"
      ? claves(v, `${prefijo}${k}.`)
      : [`${prefijo}${k}`]
  );
}

const leer = (idioma) =>
  new Set(claves(JSON.parse(readFileSync(`${DIR}/${idioma}.json`, "utf8"))));

const base = leer(BASE);
let errores = 0;

for (const idioma of OTROS) {
  const otro = leer(idioma);
  const faltantes = [...base].filter((k) => !otro.has(k));
  const sobrantes = [...otro].filter((k) => !base.has(k));
  faltantes.forEach((k) => console.error(`[${idioma}] falta la clave: ${k}`));
  sobrantes.forEach((k) => console.error(`[${idioma}] clave que no existe en ${BASE}: ${k}`));
  errores += faltantes.length + sobrantes.length;
}

if (errores > 0) {
  console.error(`\n${errores} diferencia(s) encontrada(s).`);
  process.exit(1);
}
console.log("OK: es.json y en.json tienen las mismas claves.");
```

Con un idioma nuevo, basta con agregar su código al arreglo `OTROS`.

---

## 14. Formato de fechas y números

Traducir los textos no cambia por sí solo el formato de fechas y números. Los *pipes* `date` y `number` de Angular usan el `LOCALE_ID`, que se fija al arrancar y no cambia en tiempo de ejecución. Por eso:

1. Se registran los datos regionales del español con `registerLocaleData(localeEs)` (ver `app.config.ts`).
2. Los *pipes* reciben el idioma activo como parámetro:

```html
{{ movimiento.fecha | date: 'shortDate' : undefined : idioma.idiomaActual() }}
```

Resultado para el 9 de octubre de 2026:

| Idioma | Formato |
|---|---|
| Español | `9/10/26` (día/mes) |
| Inglés | `10/9/26` (mes/día) |

---

## 15. Integración con el backend

La traducción es responsabilidad **exclusiva del frontend**; la API en Go no necesita conocer el idioma del usuario.

Actualmente la API devuelve los errores en dos formatos: `errorResponse()` envía el texto técnico en el campo `error`, y algunos manejadores envían mensajes fijos en español en `message` (por ejemplo, `"Contraseña incorrecta"`). Esos textos **no deben mostrarse directamente**, porque no cambiarían al elegir inglés.

### Estrategia inicial (sin modificar el backend)

El frontend elige el mensaje según el **código de estado HTTP**, que no depende del idioma:

| Estado HTTP | Clave de traducción |
|---|---|
| 400 | `errors.badRequest` |
| 401 | `errors.unauthorized` |
| 404 | `errors.notFound` |
| 500 | `errors.server` |
| Otro | `errors.unknown` |

```typescript
const CLAVES_ERROR: Record<number, string> = {
  400: 'errors.badRequest',
  401: 'errors.unauthorized',
  404: 'errors.notFound',
  500: 'errors.server',
};

export const claveDeError = (estado: number): string =>
  CLAVES_ERROR[estado] ?? 'errors.unknown';
```

### Mejora posterior (requiere acuerdo con backend)

Que la API agregue a sus errores un campo `code` con un identificador estable, por ejemplo `CREDENCIALES_INVALIDAS`, que el frontend use como clave de traducción para mostrar mensajes más específicos.

---

## 16. Criterios de verificación

| N.º | Prueba | Resultado esperado |
|---|---|---|
| 1 | Cambiar de ES a EN con el selector. | Todos los textos cambian sin recargar; la sesión se mantiene. |
| 2 | Recargar la página después de elegir EN. | Se conserva el inglés. |
| 3 | Primera apertura, sin preferencia, navegador en inglés. | La aplicación inicia en inglés. |
| 4 | Primera apertura, sin preferencia, navegador en otro idioma (por ejemplo, francés). | La aplicación inicia en español. |
| 5 | Clave presente solo en `es.json`. | Se muestra el texto en español. |
| 6 | Revisar `<html lang>` tras cambiar de idioma. | Coincide con el idioma activo. |
| 7 | Revisar el selector en una tableta y con lector de pantalla. | Opciones anunciadas como "Español" y "English"; área táctil de al menos 44 × 44 px. |
| 8 | Mostrar la fecha 9/10/2026 en ambos idiomas. | `9/10/26` en español y `10/9/26` en inglés. |
| 9 | Provocar un error 401 en cada idioma. | Mensaje de `errors.unauthorized` en el idioma activo. |
| 10 | Bloquear `localStorage` y usar la aplicación. | La aplicación funciona; la preferencia no se conserva. |
| 11 | Ejecutar `node scripts/verificar-i18n.mjs`. | Sin diferencias de claves entre los archivos. |

---

## 17. Alcance y decisiones pendientes

- **Preferencia por usuario:** queda fuera del alcance actual. Si se requiere, se agregaría una columna `idioma` a la tabla `usuarios` y se aplicaría al iniciar sesión.
- **Códigos de error del backend:** la mejora de la sección 15 debe acordarse con el equipo de backend.
- **Nombres de los bloques de claves:** el bloque `inventory` de los archivos JSON corresponde al módulo `supplies` de `features/`. Conviene que el equipo decida un único nombre para ambos y lo aplique en los dos idiomas.
- **Idiomas adicionales:** por ahora solo español e inglés.

---

## 18. Referencias

- Documentación de ngx-translate: <https://ngx-translate.org>
- Versiones de ngx-translate (v18.0.0): <https://github.com/ngx-translate/core/releases>
- ISO 639-1 (códigos de idioma): <https://www.loc.gov/standards/iso639-2/php/code_list.php>
- Documentos del equipo: *Investigación HU 4*, *Alternativas para implementar traducciones en Angular (Sprint 1)* y *Estructura de archivos de idioma*.
