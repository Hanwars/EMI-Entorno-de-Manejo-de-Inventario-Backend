# HU 4	Investigación - Implementación de traducción y soporte multilingüe

## Definición de estructura de archivos de traducción
src/  
├── app/  
│   ├── core/  
│   ├── features/  
│   │   ├── auth/  
│   │   ├── users/  
│   │   ├── supplies/  
│   │   ├── suppliers/  
│   │   └── movements/  
│   └── ...  
│  
└── assets/  
    └── i18n/  
        ├── es.json  
        └── en.json  

## Estructura de los archivos .json
- es.json: es para los textos en español.
- en.json: es para los textos en ingles.

### Ejemplo de archivo JSON
{  
  &nbsp;&nbsp;&nbsp;&nbsp;"app": {  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"title": "EMI",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"loading": "Cargando..."  
  &nbsp;&nbsp;&nbsp;&nbsp;},  
  &nbsp;&nbsp;&nbsp;&nbsp;"auth": {  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"login": "Iniciar sesión",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"email": "Correo electrónico",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"password": "Contraseña"  
  &nbsp;&nbsp;&nbsp;&nbsp;},  
  &nbsp;&nbsp;&nbsp;&nbsp;"users": {  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"title": "Usuarios",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"create": "Crear usuario",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"edit": "Editar usuario",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"delete": "Eliminar usuario"  
  &nbsp;&nbsp;&nbsp;&nbsp;},  
  &nbsp;&nbsp;&nbsp;&nbsp;"inventory": {  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"title": "Inventario",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"create": "Crear insumo",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"search": "Buscar insumo"  
  &nbsp;&nbsp;&nbsp;&nbsp;},  
  &nbsp;&nbsp;&nbsp;&nbsp;"common": {  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"save": "Guardar",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"cancel": "Cancelar",  
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"confirm": "Confirmar"  
  &nbsp;&nbsp;&nbsp;&nbsp;}  
}  

