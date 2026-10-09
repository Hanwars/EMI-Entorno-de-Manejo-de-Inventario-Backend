ALTERNATIVAS PARA IMPLEMENTAR TRADUCCIONES EN ANGULAR
Proyecto EMI (Entorno de Manejo de Inventario) - Coconutz BrewHouse
Sprint 1
--


Los dos enfoques posibles ya que en Angular hay dos formas de manejar las traducciones:

La primera es la traducción en tiempo de compilación. En este enfoque, los textos se traducen cuando se construye la aplicación, y el resultado es una versión distinta de la aplicación para cada idioma. Es la forma oficial que ofrece Angular.

La segunda es la traducción en tiempo de ejecución. Aquí la aplicación es una sola, y los textos se cargan desde archivos de traducción (normalmente en formato JSON) mientras la aplicación está funcionando. Esto permite cambiar de idioma con un botón, sin recargar ni tener versiones separadas.

 Alternativa 1: @angular/localize (oficial de Angular)

Es la herramienta creada por el propio equipo de Angular. Se marcan los textos en las plantillas HTML con el atributo i18n, se extraen a un archivo de traducción y, al compilar, Angular genera una aplicación completa por cada idioma.

Ventajas:
- Es la solución oficial, por lo que tiene soporte asegurado mientras exista Angular.
- Tiene muy buen rendimiento, porque los textos ya vienen traducidos en el código compilado.
- Maneja bien temas más avanzados de internacionalización, como plurales, fechas y números según el idioma.

Desventajas:
- No permite cambiar de idioma dentro de la aplicación. Para pasar de español a inglés hay que ir a otra dirección donde está la otra versión compilada.
- Hay que compilar y desplegar una versión por cada idioma, lo que complica el despliegue en el servidor con Nginx.
- Usar traducciones dentro del código TypeScript es menos cómodo.

4. Alternativa 2: ngx-translate

Es la librería de traducción más usada en la comunidad de Angular. Funciona con archivos JSON, uno por idioma (por ejemplo es.json y en.json), que se cargan mientras la aplicación corre. En las plantillas se usa un pipe, por ejemplo {{ 'menu.inventario' | translate }}, y para cambiar de idioma basta con llamar al servicio TranslateService con el idioma deseado.

Ventajas:
- Permite cambiar el idioma en cualquier momento desde la aplicación.
- Es la librería más popular, por lo que hay mucha documentación, tutoriales y respuestas a problemas comunes.
- Su forma de uso es sencilla y fácil de aprender.
- Su versión 17 funciona tanto con módulos como con componentes standalone, que es lo que usa EMI.
- Solo hay una aplicación que compilar y desplegar.

Desventajas:
- Es un poco menos eficiente que la opción oficial, porque traduce mientras la aplicación corre.
- El manejo de plurales es menos estandarizado que en @angular/localize.
- Es mantenida por la comunidad y no por el equipo de Angular.

5. Alternativa 3: Transloco

Es una librería más reciente que también traduce en tiempo de ejecución con archivos JSON y fue pensada como una evolución de ngx-translate. Actualmente se publica con el nombre @jsverse/transloco (antes se llamaba @ngneat/transloco). Para cambiar de idioma se usa TranslocoService.

Ventajas:
- También permite cambiar de idioma dentro de la aplicación.
- Permite cargar traducciones por secciones solo cuando se necesitan, lo que es útil en aplicaciones grandes.
- Tiene mejor soporte de tipos en TypeScript y se integra bien con los Signals de las versiones modernas de Angular.
- Incluye un comando de instalación (ng add) que deja la configuración básica lista.

Desventajas:
- Tiene una comunidad más pequeña que ngx-translate, por lo que hay menos ejemplos y tutoriales.
- Tiene más funciones de las que EMI necesita, lo que puede hacerla un poco más difícil de aprender al inicio.

6. Comparación general

Si comparamos las tres opciones, la principal diferencia está en si se puede o no cambiar de idioma dentro de la aplicación. @angular/localize no lo permite y requiere una versión por idioma, mientras que ngx-translate y Transloco sí lo permiten con una sola aplicación. Entre estas dos últimas, la decisión es en gran parte de preferencia: ngx-translate destaca por su popularidad y sencillez, y Transloco por ser más moderna.

7. Recomendación para EMI

Para EMI se recomienda una librería de traducción en tiempo de ejecución, porque el personal debe poder cambiar el idioma desde la tableta con un botón, sin recargar la aplicación y sin que el equipo tenga que desplegar dos versiones en el servidor. Por esta razón lo mejor seria descartar @angular/localize para este proyecto.

Entre las dos opciones restantes, lo recomendable seria ngx-translate, ya que es la más usada, tiene más documentación disponible y su forma de uso es sencilla, lo que facilita el trabajo de un equipo que está aprendiendo Angular.