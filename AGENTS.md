# AGENTS.md

Instrucciones para trabajar en este repositorio. Está en español, y el código
también: los comentarios se escriben en español y sin palabras en inglés que se
colen sin querer.

Para el mapa del workspace (qué carpetas hay y cuál está activa), lee
`../AGENTS.md`.

## Qué es esto

**Fotogramas de Películas**, la versión **Hugo** del sitio. Está en marcha. La
versión con Astro está en `../cine-art/` y está **congelada**: es un archivo
histórico, no se le hace commit ni se le sube nada.

Sitio estático de una galería de fotogramas de películas, con la paleta de color
extraída automáticamente de cada imagen. 12 películas, 48 fotogramas, 39 páginas.

Está publicada en **[silencez0.github.io](https://silencez0.github.io/)**, pero
**este repositorio sigue siendo privado**: lo publicado es una carpeta de salida
con lo que produce Hugo, en otro repositorio. Ver «Cómo se publica».
La portada es un **mural de portadas** (`layouts/partials/poster.html`), con
el título sobre la imagen al pasar el ratón. El acento de cada película se
calcula con la paleta de su portada, y la paleta de cada fotograma sale de
contar píxeles por cubos de color.

**La migración de Astro a Hugo está terminada y verificada.** No hay nada
pendiente que retomar. Si te piden el estado, la respuesta es: hecha,
verificada y subida.

## Development

El servidor de desarrollo es `hugo server`, en `http://localhost:1313/`.
Arranca **en segundo plano**, porque es un proceso que no termina nunca:

```
hugo server --port 1313 --bind 127.0.0.1
```

Para pararlo, termina el proceso `hugo` (`Stop-Process -Name hugo -Force` en
PowerShell). No hay equivalente a `astro dev stop`: Hugo no gestiona su propio
demonio.

Antes de dar cualquier cosa por terminada, tienen que pasar las dos cosas:

```
go run ./tools/palette validate
hugo --minify
```

El primero tiene que salir con `0 error(es)` y el segundo sin una sola línea
`WARN` ni `ERROR`. Un build que avisa está roto aunque genere páginas: los
avisos de Hugo suelen ser funciones que se van a dejar de existir.

## Comandos

| Comando                           | Para qué                                        |
| :-------------------------------- | :---------------------------------------------- |
| `hugo server`                     | Desarrollo, con recarga en vivo                 |
| `hugo --minify`                   | Compilar a `./public/`                          |
| `hugo list all`                   | Ver las URLs que genera cada página             |
| `go run ./tools/palette list`     | Estado de medición de cada película             |
| `go run ./tools/palette extract`  | Medir imágenes → `data/palettes.json`           |
| `go run ./tools/palette validate` | Validar fichas, enlaces y medidas               |

## Convenciones del proyecto

- **Una película es un page bundle**: `content/peliculas/<slug>/` con su
  `index.md`, su `cover.*` y su `frames/`. El nombre de la carpeta es el slug y
  es también la URL.
- **Las imágenes no se declaran en el frontmatter.** Si aparece un `cover:` o un
  `frames:` en un `index.md`, está mal: el bundle las recoge solas.
- **Los nombres de fotograma llevan cero a la izquierda** (`frame-01.jpg`). El
  orden alfabético tiene que coincidir con el orden de visionado, porque es lo
  que empareja cada imagen con su paleta en `data/palettes.json`.
- **`data/palettes.json` se genera, no se escribe a mano.** Si una paleta cambia,
  se regenera con `palette extract`.
- **Todo el CSS en `assets/css/main.css`.** Ninguna llave en las plantillas.
- **La matemática de color va en Go, no en las plantillas.** Si hace falta
  parsear un hexadecimal o calcular una luminancia, va en `tools/palette/`.
- **El país de origen va en `paises: [...]`**, con el nombre plural, aunque la
  lista tenga un solo país. Es la taxonomía `pais → paises` del `hugo.toml` y
  alimenta las pestañas «Origen».

## Antes de tocar las plantillas de taxonomía

Desde la v0.146 las plantillas de taxonomía se buscan en `layouts/taxonomy/`, no
en `layouts/_default/`. Si están en el sitio equivocado, Hugo **no avisa**:
renderiza `/decades/1980/` con la plantilla de la taxonomía, donde `.Pages` son
los términos y no las películas. Cae una página vacía y parece un problema de
contenido.

La regla: `taxonomy/taxonomy.html` es `/decades/` **y** `/paises/`,
`taxonomy/term.html` es `/decades/<década>/` **y** `/paises/<país>/`. Las dos
taxonomías comparten plantilla, así que hay que ramificar por
`.Data.Plural == "paises"`; sin esa pregunta el título de un país lo decide
`func/decade-label.html`, que hace `int` del nombre y devuelve «Años 0».
Ramifica con `if/else`, no con `cond`: `cond` evalúa los dos argumentos y la
etiqueta de década se ejecutaría igualmente.

Y otro que también muerde: **el campo del frontmatter lleva el nombre plural**
(`paises:`, no `pais:`). Con el singular Hugo **no da ningún aviso**: genera la
página de la taxonomía sin términos, sin páginas de término y con el estado
vacío «Aún no hay orígenes», como si no hubiera ninguna película con ese dato.

## Cómo se publica

El sitio está en <https://silencez0.github.io/>. **Este repositorio sigue
privado**, y eso no va a cambiar: lo publicado es una carpeta de salida con lo
que produce Hugo, en el repositorio público
[`SilenceZ0/SilenceZ0.github.io`](https://github.com/SilenceZ0/SilenceZ0.github.io).

Se publica con `.\publicar.ps1`, que valida las fichas, compila y empuja. **No
añadas un segundo camino de publicación**: si se acrescenta otro, se acaba
publicando una versión desde el sitio equivocado.

Tres cosas que hay que respetar:

- **`baseURL` se pasa en el comando, no se escribe.** Sigue siendo
  `http://localhost:1313/` en `hugo.toml`, y en local los enlaces canónicos y el
  RSS apuntan a `localhost`. El script lo cambia solo al publicar. Si lo pones
  en `hugo.toml`, en local salen mal todas las URL absolutas.
- **`--cleanDestinationDir` no se quita.** Sin él, Hugo acumula en `public/` los
  ficheros de los builds anteriores. Pasó: llegaron a verse trece versiones del
  CSS y cinco de `intro.js` en la misma carpeta.
- **El `push` va con `--force`, y es lo correcto.** El destino es una carpeta
  generada. No tiene historial que conservar.

La tipografía se sirve desde `static/fonts/`, no desde Google Fonts. La ruta del
`.woff2` va **desde la raíz**, con la barra delante, tanto en el `@font-face` de
`assets/css/main.css` como en el `preload` de `layouts/partials/head.html`, y en
los dos sitios tiene que seguir siendo así. El CSS minificado vive en
`/css/main.min.<huella>.css`, así que una ruta relativa se resolvería contra
`/css/` y no encontraría el archivo. **Hugo no avisa de eso**: compila limpio y lo
único que se ve es la tipografía de reserva en el navegador.

## Nada de APIs externas

Las capturas son del usuario. **Prohibido** usar TMDB o cualquier otra API de
metadatos, carteles o imágenes. Las fichas se escriben a mano y lo único que se
pega desde fuera es el enlace de YouTube.

## Documentation

Documentación de Hugo: https://gohugo.io/documentation/

Consulta estas guías antes de tocar lo que corresponde:

- [Añadir páginas, enrutado y layouts](https://gohugo.io/en-gb/getting-started/quick-start/)
- [Page bundles y recursos](https://gohugo.io/en-gb/content-management/page-bundles/)
- [Organizar recursos](https://gohugo.io/en-gb/content-management/organize-resources/)
- [Taxonomías](https://gohugo.io/content-management/taxonomies/)
- [Buscar en las plantillas](https://gohugo.io/en-gb/templates/introduction/) y
  [funciones disponibles](https://gohugo.io/en-gb/functions/)
- [Asset pipelines: `assets/` y el CSS](https://gohugo.io/en-gb/hugo-pipeline/asset-pipeline/)

## Atajos de Hugo y de Go que ya se han pagado

Lista de lo que costó tiempo descubrir. Antes de añadir una función, mírala aquí:
si ya está, úsala.

- `site.Language.LanguageCode` y `site.Data` están obsoletos desde la v0.146:
  son `.Language.Locale` y `hugo.Data`.
- `page.Pages` **no** tiene `.First` ni `.Last`: se usa `index .Pages 0`.
- `regexMatch` y `strings.ContainsRE` **no existen**. `strings.FindRE` devuelve las
  coincidencias enteras; para capturar grupos, `strings.FindRESubmatch`.
- Un **partial que escribe texto devuelve HTML sin escapar**, así que los `&` de
  una query-string llegan tal cual al atributo. Para escaparlos exactamente una
  vez, el partial devuelve con `return`.
- `{{ define "main" }}` + `--printUnusedTemplates` da falsos positivos: no
  detecta los bloques de las plantillas base. No es una fuente fiable.
- `sliceStable`/`sort.SliceStable` importan: en Go el orden de `sort.Slice` no es
  estable, y el extractor depende del orden estable para que los empates de
  frecuencia den el mismo color en cada ejecución.
- **`hugo.ServerPort` no existe** en la 0.167: `hugo --minify` compila sin
  quejarse y el servidor devuelve un error 500 en cada página. El puerto sale de
  `baseURL`.
- **No inyectes `livereload.js`.** Hugo lo pone solo al principio de `<head>`,
  con `data-no-instant`. Si se añade a mano quedan dos y se abren dos conexiones.

## La pantalla de bienvenida

La capa con la nube de puntos vive en `layouts/partials/intro-gate.html` (la
decisión), `intro.html` (el marcado) y `assets/js/intro.js` (el lienzo).

- **Un solo punto de decisión.** Todo, incluido `prefers-reduced-motion`, se
  decide en el JavaScript en línea de `intro-gate.html`, que es lo único que
  corre antes del primer pintado. **El CSS no vuelve a mirar nada**, a propósito:
  si el CSS también decidiera, podría quedar la intro invisible con el contenido
  marcado `inert`, y eso es una web que no se puede usar.
- **Una vez por sesión**, con `sessionStorage`. La marca se escribe en cuanto se
  muestra, así que recargar no la devuelve. Para verla otra vez hay que usar
  `/?intro=1`, que la enseña **sin** guardar la marca, o abrir una ventana
  privada.
- **`prefers-reduced-motion` se comprueba después del parámetro de la URL**, así
  que gana a `?intro=1` sin excepción. No lo muevas por encima: es lo único que
  el visitante no debería poder saltarse desde la barra de direcciones.
- **El contenido se marca `inert` con JavaScript**, sobre los hijos de `<body>`.
  No lo envuelvas en un `div` nuevo: el ancho máximo y el centrado están en
  `body` y un envoltorio los parte en dos.
- **No lleva `role="dialog"`.** Es decorativa; con `aria-modal` un lector de
  pantalla quedaría encerrado sin poder leer lo que hay debajo.

Está razonado entero en el `README.md`, en la sección «La pantalla de
bienvenida».

## Verificar contra la versión anterior

Este sitio es una migración desde Astro (`SilenceZ0/fotogramas-de-peliculas`),
que queda archivada y congelada. Si tocas el extractor de paletas, conviene
comparar el resultado contra el JSON de referencia de aquella versión:

- `../cine-art/src/data/image-data.json` — lo que produjo el extractor de Astro,
  con `sharp`.
- `data/palettes.json` — lo que produce el de Go.

Se compara el campo `palette` (la portada) con `palette`, y el array `frames` con
`framePalettes`.

**No persigas la coincidencia exacta, y no pierdas el tiempo intentando.** El
extractor de Go *reimplementa* el de Astro, no lo llama, y los decodificadores
JPEG no son los mismos: el `image/jpeg` de Go y el `libjpeg` de libvips no hacen
la transformada discreta del coseno igual, así que algunos píxeles caen en cubos
de cuantización distintos. De 250 colores comparados, coinciden 211; los 39
restantes difieren en un cubo o en el puesto dentro de la paleta. Está explicado
en el README.

Lo que sí tiene que cumplirse siempre son las garantías, y las comprueba
`palette validate`: 5 colores distintos, de oscuro a claro, hexadecimal válido y
acento coherente con la paleta.