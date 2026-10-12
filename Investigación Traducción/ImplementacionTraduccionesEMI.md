# Implementación de traducciones y soporte multilingüe — EMI

## Descripción

Esta implementación establece una estructura centralizada para administrar las traducciones del frontend de **EMI (Entorno de Manejo de Inventario)** mediante **ngx-translate**. Permite mostrar la interfaz en español e inglés, cambiar el idioma sin recargar la aplicación y conservar la preferencia durante la navegación y entre sesiones.

La solución está planteada para una aplicación Angular con componentes standalone.

## Objetivos

- Centralizar las traducciones en archivos JSON.
- Sustituir los textos fijos de las plantillas por claves de traducción.
- Permitir seleccionar el idioma desde la interfaz.
- Mantener el idioma elegido durante la navegación y las siguientes visitas.
- Facilitar que nuevos módulos incorporen traducciones sin implementar su propia lógica de idioma.

## Estructura de archivos

```text
src/
└── app/
    ├── core/
    │   ├── services/
    │   │   └── idioma.service.ts
    │   └── components/
    │       └── selector-idioma/
    │           ├── selector-idioma.component.ts
    │           ├── selector-idioma.component.html
    │           └── selector-idioma.component.css
    ├── features/
    │   ├── auth/
    │   ├── users/
    │   ├── supplies/
    │   ├── suppliers/
    │   └── movements/
    ├── app.config.ts
    └── ...

public/
└── i18n/
    ├── es.json
    └── en.json
```

> Los archivos JSON se ubican en `public/i18n/` según la estructura documentada para el proyecto. El cargador los solicita mediante las rutas `/i18n/es.json` y `/i18n/en.json`.

## 1. Instalación

Desde la raíz del proyecto Angular, instalar las dependencias si todavía no están instaladas:

```bash
npm install @ngx-translate/core @ngx-translate/http-loader
```

## 2. Archivos de traducción

Los dos archivos deben conservar la misma jerarquía y las mismas claves. Los textos compartidos por varios módulos deben colocarse en `common`.

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
    "password": "Contraseña",
    "logout": "Cerrar sesión"
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
  "suppliers": {
    "title": "Proveedores",
    "create": "Crear proveedor"
  },
  "movements": {
    "title": "Movimientos",
    "create": "Registrar movimiento"
  },
  "common": {
    "save": "Guardar",
    "cancel": "Cancelar",
    "confirm": "Confirmar",
    "search": "Buscar",
    "edit": "Editar",
    "delete": "Eliminar",
    "language": "Idioma"
  },
  "errors": {
    "badRequest": "La solicitud no es válida.",
    "unauthorized": "Credenciales inválidas o sesión expirada.",
    "notFound": "No se encontró el recurso solicitado.",
    "server": "Ocurrió un error en el servidor."
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
    "email": "Email address",
    "password": "Password",
    "logout": "Log out"
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
  "suppliers": {
    "title": "Suppliers",
    "create": "Create supplier"
  },
  "movements": {
    "title": "Movements",
    "create": "Register movement"
  },
  "common": {
    "save": "Save",
    "cancel": "Cancel",
    "confirm": "Confirm",
    "search": "Search",
    "edit": "Edit",
    "delete": "Delete",
    "language": "Language"
  },
  "errors": {
    "badRequest": "The request is invalid.",
    "unauthorized": "Invalid credentials or expired session.",
    "notFound": "The requested resource was not found.",
    "server": "A server error occurred."
  }
}
```

Los textos anteriores son una base inicial. Deben añadirse las claves correspondientes a todas las pantallas y acciones reales del sistema.

## 3. Configuración global

En `src/app/app.config.ts`, integrar los proveedores de ngx-translate en el arreglo `providers` existente. No se deben eliminar los proveedores que ya utiliza la aplicación.

```typescript
import { provideHttpClient } from '@angular/common/http';
import {
  provideTranslateService,
  provideTranslateHttpLoader
} from '@ngx-translate/core';
import { provideAppInitializer, inject } from '@angular/core';
import { IdiomaService } from './core/services/idioma.service';

// Dentro de providers:
provideHttpClient(),

provideTranslateService({
  fallbackLang: 'es',
  loader: provideTranslateHttpLoader({
    prefix: '/i18n/',
    suffix: '.json'
  })
}),

provideAppInitializer(() =>
  inject(IdiomaService).inicializar()
)
```

El inicializador establece el idioma inicial antes de mostrar la interfaz. Este ejemplo corresponde al enfoque standalone descrito para Angular 22 y ngx-translate 18; debe adaptarse si la configuración real del proyecto difiere.

## 4. Servicio centralizado de idioma

Crear `src/app/core/services/idioma.service.ts`:

```typescript
import { Injectable } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { firstValueFrom } from 'rxjs';

type Idioma = 'es' | 'en';

@Injectable({
  providedIn: 'root'
})
export class IdiomaService {
  private idiomasDisponibles: Idioma[] = ['es', 'en'];

  idiomaActual: Idioma = 'es';

  constructor(private translate: TranslateService) {}

  async inicializar(): Promise<void> {
    let idiomaGuardado: string | null = null;

    try {
      idiomaGuardado = localStorage.getItem('emi.idioma');
    } catch {
      // La aplicación puede continuar sin almacenamiento local.
    }

    const idiomaNavegador =
      typeof navigator !== 'undefined'
        ? navigator.language.slice(0, 2).toLowerCase()
        : 'es';

    let idiomaInicial: Idioma = 'es';

    if (this.esIdiomaValido(idiomaGuardado)) {
      idiomaInicial = idiomaGuardado;
    } else if (this.esIdiomaValido(idiomaNavegador)) {
      idiomaInicial = idiomaNavegador;
    }

    await this.cambiarIdioma(idiomaInicial);
  }

  async cambiarIdioma(idioma: string): Promise<void> {
    if (!this.esIdiomaValido(idioma)) {
      return;
    }

    try {
      await firstValueFrom(this.translate.use(idioma));

      this.idiomaActual = idioma;
      document.documentElement.lang = idioma;

      try {
        localStorage.setItem('emi.idioma', idioma);
      } catch {
        // El cambio sigue activo aunque no se pueda guardar.
      }
    } catch (error) {
      console.error('No se pudo cargar el idioma:', error);
    }
  }

  private esIdiomaValido(
    idioma: string | null
  ): idioma is Idioma {
    return idioma !== null &&
      this.idiomasDisponibles.includes(idioma as Idioma);
  }
}
```

El servicio es singleton mediante `providedIn: 'root'`. La preferencia se guarda en el navegador bajo la clave `emi.idioma`, por lo que queda asociada al dispositivo y no a una cuenta específica.

## 5. Selector de idioma

Crear `src/app/core/components/selector-idioma/selector-idioma.component.ts`:

```typescript
import { Component } from '@angular/core';
import { IdiomaService } from '../../services/idioma.service';

@Component({
  selector: 'app-selector-idioma',
  standalone: true,
  templateUrl: './selector-idioma.component.html',
  styleUrl: './selector-idioma.component.css'
})
export class SelectorIdiomaComponent {
  constructor(public idiomaService: IdiomaService) {}

  cambiarIdioma(idioma: string): void {
    this.idiomaService.cambiarIdioma(idioma);
  }
}
```

Crear `selector-idioma.component.html`:

```html
<div class="selector-idioma" aria-label="Seleccionar idioma">
  <button
    type="button"
    [class.activo]="idiomaService.idiomaActual === 'es'"
    [attr.aria-pressed]="idiomaService.idiomaActual === 'es'"
    aria-label="Español"
    (click)="cambiarIdioma('es')">
    ES
  </button>

  <button
    type="button"
    [class.activo]="idiomaService.idiomaActual === 'en'"
    [attr.aria-pressed]="idiomaService.idiomaActual === 'en'"
    aria-label="English"
    (click)="cambiarIdioma('en')">
    EN
  </button>
</div>
```

Crear `selector-idioma.component.css`:

```css
.selector-idioma {
  display: flex;
  gap: 4px;
}

.selector-idioma button {
  min-width: 44px;
  min-height: 44px;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 6px;
  background: white;
  color: #333;
  cursor: pointer;
}

.selector-idioma button.activo {
  background: #245c46;
  color: white;
  border-color: #245c46;
}
```

Importar `SelectorIdiomaComponent` en el componente de layout o barra superior, y también en la pantalla de inicio de sesión si utiliza una plantilla independiente. Conservar los imports existentes.

## 6. Usar traducciones en las plantillas

En cada componente standalone que utilice el pipe, importar `TranslatePipe` desde `@ngx-translate/core`:

```typescript
import { TranslatePipe } from '@ngx-translate/core';

// Agregar TranslatePipe a imports del componente.
```

Ejemplo de plantilla:

```html
<h1>{{ 'inventory.title' | translate }}</h1>

<button type="button">
  {{ 'inventory.create' | translate }}
</button>

<input
  type="text"
  [placeholder]="'inventory.search' | translate"
/>

<button type="button">
  {{ 'common.save' | translate }}
</button>

<button type="button">
  {{ 'common.cancel' | translate }}
</button>
```

Las claves reemplazan los textos fijos de la plantilla. Al cambiar el idioma, ngx-translate vuelve a resolver las claves sin recargar la página.

## 7. Agregar traducciones a nuevos módulos

Para incorporar un módulo nuevo, agregar sus claves a ambos archivos JSON. Por ejemplo, para un módulo de reportes:

En `es.json`:

```json
"reports": {
  "title": "Reportes",
  "generate": "Generar reporte"
}
```

En `en.json`:

```json
"reports": {
  "title": "Reports",
  "generate": "Generate report"
}
```

Estos objetos se agregan dentro del objeto raíz existente; no se debe reemplazar todo el archivo. El componente nuevo solo necesita utilizar las claves con el pipe `translate`.

## 8. Pruebas de aceptación

- [ ] Existen `public/i18n/es.json` y `public/i18n/en.json`.
- [ ] Ambos archivos son JSON válidos y contienen las mismas claves.
- [ ] Los textos de prueba se muestran correctamente en español e inglés.
- [ ] El selector cambia el idioma sin recargar la página.
- [ ] El idioma se mantiene al navegar entre módulos.
- [ ] La preferencia se conserva después de recargar la página.
- [ ] La interfaz mantiene el idioma inicial antes de mostrar los textos.
- [ ] Un nuevo módulo puede agregar claves sin modificar el servicio.
- [ ] El selector es accesible mediante teclado y sus opciones tienen etiquetas comprensibles.

## Consideraciones

- El idioma se guarda en el navegador/dispositivo, no en la base de datos ni asociado a una cuenta de usuario.
- Las traducciones de texto no cambian automáticamente el formato regional de fechas y números.
- Los mensajes fijos que devuelve el backend deben mapearse a claves de traducción en el frontend si también deben cambiar de idioma.
- Si una clave no existe en el idioma activo, se debe revisar que exista en el idioma de respaldo (`es`).
- La implementación debe integrarse con la configuración y los componentes existentes; los fragmentos son una base de referencia y no una confirmación de que el código ya compiló en el repositorio.

## Resultado esperado

EMI contará con una estructura centralizada de traducciones, un servicio único para administrar el idioma, un selector reutilizable y una convención para que los módulos incorporen nuevos textos en español e inglés.
