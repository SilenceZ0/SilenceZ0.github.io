# Fotogramas de Películas

Colección de fotogramas de películas con su **paleta de color extraída
automáticamente** de la propia imagen. La web es estática: no hay servidor, ni
base de datos, ni llamadas a APIs externas.

El acento de color de cada película (bordes, brillos, marcas de la línea de
tiempo) se elige automáticamente como el tono más saturado y luminoso de su
paleta, no como el primer color de la lista.

Esta es la versión **Hugo** del sitio. La anterior, con Astro, está archivada en
[`SilenceZ0/fotogramas-de-peliculas`](https://github.com/SilenceZ0/fotogramas-de-peliculas).

## Requisitos

- **Hugo Extended 0.167 o superior.** La versión *Extended* es obligatoria: sin
  ella no hay soporte de Sass/Webpack y `assets/` no se procesa.
- **Go 1.23 o superior**, solo para el extractor de paletas. El sitio se
  construye sin él.
- Las imágenes son capturas propias: por eso **no** se usa TMDB ni ninguna otra
  API de metadatos. Las fichas se escriben a mano y solo se pega el enlace de
  YouTube.

## Comandos

| Comando                            | Qué hace                                          |
| :--------------------------------- | :------------------------------------------------ |
| `hugo server`                      | Servidor de desarrollo en `localhost:1313`         |
| `hugo --minify`                    | Compila el sitio estático en `./public/`           |
| `hugo --gc --minify`               | Igual, y limpia la caché de recursos              |
| `go run ./tools/palette extract`   | Mide las imágenes y genera las paletas             |
| `go run ./tools/palette list`      | Ver qué hay y qué falta por medir                  |
| `go run ./tools/palette combos`    | Genera los cruces de filtros en `content/filtros/`  |
| `go run ./tools/palette validate`  | Comprueba fichas, enlaces, cruces y medidas         |
| `.\publicar.ps1`                   | Publica en GitHub Pages                            |

No hay `package.json` ni paso de `npm install`: Hugo no necesita nada instalado
para construir el sitio.

## Añadir una película

1. **Crea la carpeta** `content/peliculas/<slug>/`. El nombre de la carpeta es el
   slug y también la URL (`/peliculas/<slug>/`), así que no hace falta ningún
   campo `id` ni `slug` dentro del Markdown.

2. **Coloca las imágenes** dentro, junto al `index.md`:

   ```text
   content/peliculas/alien-el-octavo-pasajero-1979/
   ├── index.md
   ├── cover.jpg
   └── frames/
       ├── frame-01.jpg
       ├── frame-02.jpg
       └── …
   ```

   Esto es un *page bundle*: un directorio con su `index.md` y sus recursos. Es
   la diferencia más grande con la versión de Astro. Allí cada ficha llevaba una
   lista de rutas (`cover:` y `frames:`) que había que mantener a mano y que se
   desincronizaba en cuanto añadías una captura. Aquí **la ficha solo dice qué
   película es**: soltar el archivo en `frames/` es toda la operación.

   Los nombres llevan cero a la izquierda (`frame-01.jpg`), y por eso el orden
   alfabético coincide con el orden de visionado y con el orden en que Hugo
   ordena los recursos. `cover.*` en la raíz es la portada; si no hay portada,
   la paleta de la película sale del primer fotograma.

3. **Crea el `index.md`.** El archetype lo escribe por ti:

   ```sh
   hugo new peliculas/alien-el-octavo-pasajero-1979/index.md
   ```

   ```markdown
   ---
   title: "Alien: The Eighth Passenger"
   year: 1979
   paises: ["Reino Unido", "Estados Unidos"]
   director: "Ridley Scott"
   decades: ["1970"]
   trailerUrl: "https://www.youtube.com/watch?v=OL5dD-EwcaU"
   soundtrackUrl: "https://www.youtube.com/watch?v=EsmHr45H204"
   ---

   El párrafo va aquí, en el cuerpo del archivo, sin comillas.
   ```

   Dos detalles que hacen fallar la validación si no se respetan:

   - **Los títulos con dos puntos van entrecomillados**: `"2001: A Space Odyssey"`,
     sin comillas YAML interpreta el `:` como separador de clave.
   - **Los enlaces de YouTube se validan.** Admite `watch?v=`, `youtu.be/`,
     `/embed/`, `/shorts/` y `/live/`, con o sin protocolo. El id debe tener 11
     caracteres, así que un texto como `no-es-un-link` se rechaza en lugar de
     generar un vídeo roto. El sitio incrusta con `youtube-nocookie.com` y
     carga diferida.

   `decades` y `generos` son los dos campos que se pueden *"olvidar"* sin que
   Hugo se rompa, pero `palette validate` no los perdona: la década tiene que
   cuadrar con `year` y el género tiene que salir de la lista cerrada de
   `tools/palette`. Se dejan explícitos en la ficha para que sea visible de un
   vistazo que la taxonomía está bien puesta.

4. **Mide las imágenes**:

   ```sh
   go run ./tools/palette extract
   ```

   El extractor escribe `data/palettes.json` con la paleta de cada imagen y el
   color de acento de cada película. Ese archivo **se genera solo**: está
   separado del contenido a propósito, para que reejecutar el extractor no toque
   los archivos de las películas ni ensucie el historial de git.

5. `go run ./tools/palette combos` (solo si cambió la década, el país o el
   género de una ficha), `go run ./tools/palette validate` y `hugo --minify`.

## Medir solo algunas películas

Medir todas lleva unos segundos con diez películas y es una molestia con ochenta.
Se puede elegir cuáles:

```sh
go run ./tools/palette list              # qué hay y qué falta por medir
go run ./tools/palette extract matrix-1999
go run ./tools/palette extract blade alien life  # varias a la vez
go run ./tools/palette extract --force   # volver a medirlo todo
go run ./tools/palette extract --dry-run # ver qué se mediría, sin escribir
```

Se acepta el slug o el título, con o sin tildes y sin importar las mayúsculas,
y vale con coincidencia parcial: `"the thing"`, `batalla` o `1982` funcionan.

**Las películas no indicadas conservan lo que ya tenían medido.** Medir una sola no
toca el resto del archivo: es lo que hace falta cuando has cambiado cuatro
capturas de una película y no quieres arrastrar el resto.

## Añadir una serie

Las series son fichas de texto: portada, año, plataforma y notas. Sin
fotogramas, sin paleta y sin tráiler, así que la herramienta de paletas **no
toca esta carpeta**: sigue leyendo solo `content/peliculas/`.

1. **Crea la carpeta** `content/series/<slug>/`, con el `index.md` y el
   `cover.jpg` dentro:

   ```sh
   hugo new series/<slug>/index.md
   ```

   ```markdown
   ---
   title: "Título de la serie"
   year: 2021
   platform: "Plataforma o cadena"
   draft: true
   ---

   Las notas de la ficha van aquí, en el cuerpo del archivo.
   ```

2. **Quita `draft: true`** cuando la tengas lista. El ejemplo de arranque,
   `content/series/plantilla/`, se queda en borrador a propósito: no se
   publica y sirve para copiar de él.

3. `hugo --minify`.

`year` admite también un intervalo entre comillas (`year: "2019–2022"`). Las
series **no llevan `decades`**: esa taxonomía es de películas y sus páginas
dicen «X películas», algo que una serie no puede prometer.

## Añadir un artículo

Los artículos viven en `content/especial/`, uno por carpeta, con sus imágenes
junto al `index.md` si el artículo las lleva.

```sh
hugo new especial/<slug>/index.md
```

```markdown
---
title: "Título del artículo"
date: 2026-10-06
draft: true
---

El cuerpo va aquí, en Markdown: títulos, listas, citas, tablas y enlaces.
```

- **Se publica al quitar `draft: true`.** El índice `/especial/` los ordena de
  más reciente a más antiguo según `date`.
- **La tarjeta del índice lleva imagen**: pon `cover.jpg` en la carpeta del
  artículo (junto al `index.md`) y saldrá recortada a 16/9, siempre con el
  mismo tamaño sea cual sea el original. Sin `cover.jpg`, sale el marcador de
  posición.
- **Las imágenes del cuerpo** van también junto al `index.md` y se citan con
  `![Descripción](foto.jpg)`.
- **Los vídeos de YouTube se incrustan** con el shortcode `video`, sin salir
  de la web:

  ```text
  {{< video url="https://www.youtube.com/watch?v=..." label="Banda sonora" >}}
  ```

  Acepta el enlace completo o el id suelto, y usa youtube-nocookie. Si el
  artículo lleva dos vídeos, dales `anchor` distinto a cada uno.
- **El resumen del índice** sale del principio del cuerpo; para cortarlo en
  otro sitio, separa con un `<!--more-->`.
- **`description` es opcional**: si no lo pones, la meta descripción de la
  página sale del resumen.
- El ejemplo `content/especial/plantilla/` está en borrador y sirve de
  arranque.

## Cómo se eligen los colores

`palette extract` reduce cada imagen a 160×160, cuenta píxeles por color
cuantizado en cubos de 24 y se queda con los 5 tonos más frecuentes que estén
separados entre sí (distancia mínima de 46, o los dos primeros se leerían como
el mismo color). Si una imagen es demasiado plana para llenar 5 colores
distintos —un fotograma muy oscuro, por ejemplo— deriva variantes aclarando los
que ya tiene. El resultado siempre son 5 colores únicos, ordenados de oscuro a
claro, para que la paleta se lea como un degradado.

El **acento** de la película sale de esa misma paleta: se descartan los tonos con
luminancia fuera de 0,25–0,9 (demasiado oscuro desaparece sobre el fondo negro,
demasiado claro pierde el carácter) y gana el que más combine saturación con
estar en la zona media. No es el primero de la lista, que casi siempre es casi
negro.

Las dimensiones intrínsecas **no** se guardan: Hugo ya las sabe, con `.Width` y
`.Height` del propio recurso, y se escriben como `width` y `height` en el HTML.
Sin ellas el navegador no puede reservar el hueco de la imagen y la página da
saltos al cargar.

En la ficha, pasar el ratón por un fotograma cambia la paleta del encabezado.
Solo se intercambian variables CSS: el trabajo de píxeles ya está hecho en el
build, así que en el navegador no se calcula nada.

### Una nota sobre las paletas medidas

El extractor de Go **reimplementa** el de la versión de Astro (que usaba
`sharp`), no lo llama. Consecuencia práctica: las paletas no son idénticas a las
de la versión de Astro. De 250 colores comparados, coinciden 211; los 39
restantes difieren en un cubo de cuantización o en el puesto dentro de la
paleta.

La causa no es el filtro de reducción —probar con Lanczos3, el kernel que usa
`sharp`, solo mejoraba 39→37— sino el **decodificador JPEG**: el `image/jpeg` de
Go y el `libjpeg` de libvips no hacen la transformada discreta del coseno igual.
Un píxel que cae en un cubo distinto cambia el recuento de ese cubo, y con la
paleta ordenada por frecuencia eso acaba moviendo un color de sitio.

Las garantías sí se cumplen: 5 colores distintos, de oscuro a claro, hexadecimal
válido y acento coherente con la paleta. Lo que cambia es la paleta concreta que
sale de cada imagen, que es una decisión de implementación, no de contenido.

## Décadas, origen y género: por qué son una taxonomía

En la versión de Astro, filtrar por década eran botones con JavaScript que
ocultaban tarjetas de la portada. Esos filtros no se podían compartir, ni
indexar, ni funcionar sin JavaScript.

En Hugo una taxonomía es contenido de verdad. Hay tres: `decades`, que es el
año de la película redondeado a década, `pais`, que es su país de origen, y
`genero`, que es su género. Las tres se declaran en `hugo.toml` y Hugo genera
las páginas solo:

| URL                     | Qué es                        | Plantilla                |
| :---------------------- | :---------------------------- | :----------------------- |
| `/`                     | Mural de portadas             | `layouts/index.html`     |
| `/peliculas/`           | Índice de todas las películas    | `_default/list.html`     |
| `/peliculas/<slug>/`    | Ficha de una película         | `_default/single.html`   |
| `/decades/`             | Lista de décadas              | `taxonomy/taxonomy.html` |
| `/decades/1980/`        | Las películas de una década      | `taxonomy/term.html`     |
| `/paises/`              | Lista de países de origen        | `taxonomy/taxonomy.html` |
| `/paises/<país>/`       | Las películas de un país         | `taxonomy/term.html`     |
| `/generos/`             | Lista de géneros                 | `taxonomy/taxonomy.html` |
| `/generos/<género>/`    | Las películas de un género       | `taxonomy/term.html`     |
| `/filtros/`             | Entrada a los cruces             | `filtros/list.html`      |
| `/filtros/<slug>/`      | Un cruce de filtros              | `filtros/single.html`    |
| `/series/`              | Índice de series              | `series/list.html`       |
| `/series/<slug>/`       | Ficha de una serie            | `series/single.html`     |
| `/especial/`            | Índice de artículos           | `especial/list.html`     |
| `/especial/<slug>/`     | Un artículo                   | `especial/single.html`   |

Las tres filas de pestañas —«Décadas», «Origen» y «Género»— son enlaces a esas
páginas; el partial que las dibuja, con su etiqueta, es `filters.html`.
Funcionan sin JavaScript, se pueden compartir y se pueden indexar. El recuento
sale de la propia taxonomía, así que añadir una película de 1965 hace aparecer
«Años 60» sola, sin tocar una sola plantilla.

Los tres filtros se **combinan entre sí**, porque cada cruce con películas es
una página de verdad en `/filtros/`. Las filas son contextuales: en
`/decades/1980/`, la fila «Origen» enlaza a `/filtros/1980-reino-unido/` y la
«Género» a `/filtros/1980-terror/`, así que se elige un país o un género sin
perder la década. La regla es «**Todas** quita solo esa dimensión»: en un cruce
de dos o tres filtros, «Todas» de la fila de la década deja el cruce de país y
género intacto, y si con lo que queda solo hay una dimensión, la URL se queda
en la página de término de esa taxonomía.

**Esas páginas no se escriben a mano**: las genera `go run ./tools/palette
combos`, una por cada cruce de dos o tres dimensiones con al menos una
película, y `palette validate` comprueba que la sección está al día. El
frontmatter lleva `decada:`, `pais:` y `genero:` en singular, a propósito: con
los nombres plurales de las taxonomías (`paises:`, `generos:`), Hugo colaría la
página en `/paises/x/` como si fuera una película más. El slug de cada cruce es
la década, el país y el género unidos por guiones, en ese orden:

```
1980-reino-unido           décadas × orígenes
1980-terror                décadas × géneros
reino-unido-terror         orígenes × géneros
1980-reino-unido-terror    las tres dimensiones
```

Los slugs los construye la misma transliteración que Hugo aplica a los
términos —de ahí que `1970-union-sovietica-ciencia-ficcion` no lleve tilde—,
y solo se generan los cruces con películas: uno vacío no tiene página y no
aparece en las filas. Las películas en borrador (`draft: true`) tampoco
participan. Si un cruce deja de tener películas, `palette combos` lo borra, y
si una ficha cambia de país, regenera el que haga falta.

**El campo de cada taxonomía se escribe con el nombre plural**
(`paises: ["Reino Unido"]`, `generos: ["Terror"]`), que es como lo pide Hugo.
Con el singular (`pais:`, `genero:`) **no hay ningún aviso**: Hugo genera la
página de la taxonomía sin términos, sin páginas de término y con el estado
vacío «Aún no hay orígenes», como si no hubiera ninguna película con ese dato.

**Los géneros son una lista cerrada** que vive en `tools/palette` (`films.go`,
`generosValidos`) y que exige `palette validate`: cada `generos` de cada ficha
tiene que salir exactamente de esa lista. Es lo que evita que «Terror» y
«terror» se partan en dos páginas de término distintas sin avisar. Los géneros
que ninguna película use no crean página: Hugo solo genera los términos vivos.

**Los países también son una lista cerrada**, en el mismo archivo
(`paisesValidos`), y por la misma razón: sin ella, «Canadá» y «Canada» o
«Estados Unidos» y «EEUU» partirían `/paises/` en dos páginas de término sin
que Hugo ni el compilador dijeran nada. La lista **no es un catálogo de los
países del mundo**, sino los nombres que se pueden usar: para dar de alta uno
nuevo basta con añadir una línea a `paisesValidos` y escribirlo en la ficha.
Si `palette validate` se queja de un país, es porque aún no está en la lista.

Nombres canónicos, tal y como van en el frontmatter (copia y pega; el orden es
el alfabético sin tener en cuenta las tildes):

```text
Alemania, Alemania del Oeste, Argentina, Australia, Austria, Bélgica, Brasil,
Canadá, Checoslovaquia, Chile, China, Colombia, Corea del Sur, Croacia, Cuba,
Dinamarca, Egipto, Ecuador, Eslovaquia, Eslovenia, España, Estados Unidos,
Estonia, Filipinas, Finlandia, Francia, Grecia, Hong Kong, Hungría, India,
Indonesia, Irán, Irlanda, Islandia, Israel, Italia, Japón, Kenia, Letonia,
Líbano, Lituania, Luxemburgo, Marruecos, México, Nigeria, Noruega,
Nueva Zelanda, Países Bajos, Pakistán, Paraguay, Perú, Polonia, Portugal,
Reino Unido, República Checa, Rumanía, Rusia, Senegal, Serbia, Singapur,
Sudáfrica, Suecia, Suiza, Taiwán, Tailandia, Túnez, Turquía, Ucrania,
Unión Soviética, Uruguay, Venezuela, Vietnam, Yugoslavia
```

Los nombres de la lista **no tienen por qué tener página**: solo aparece en la
fila «Origen» el país que esté en uso, es decir, el que tenga al menos una
película. Al añadir la primera, la pestaña sale sola y `palette combos` crea
sus cruces en `/filtros/`, sin tocar ninguna plantilla.

**Un detalle de Hugo que muerde:** desde la v0.146 las plantillas de
taxonomía **no** se buscan en `layouts/_default/`, sino en `layouts/taxonomy/`.
Un archivo en el sitio equivocado no da ningún aviso: si `term.html` está en
`_default/`, Hugo renderiza `/decades/1980/` con la plantilla de la taxonomía,
donde `.Pages` son los términos en lugar de las películas, y el resultado es una
página que se ve bien pero está vacía. Por eso las dos plantillas viven en
`layouts/taxonomy/`, con el nombre explícito.

## La navegación

La barra superior con las pestañas (Inicio, Películas, Series, Especial) está
en **todas las** páginas: la incluye `baseof.html`, así que sale en la portada, los
índices, las fichas, las décadas y los artículos. Como las pestañas de década,
son enlaces a páginas de verdad: sin JavaScript, compartibles y indexables.

- **El menú vive en `hugo.toml`**, en bloques `[[menus.main]]` con `pageRef`.
  Cambiar el orden es mover el `weight`; añadir una pestaña es añadir su bloque
  y crear las plantillas de la sección. Ninguna URL está escrita a mano.
- **La pestaña activa** la calcula `layouts/partials/site-nav.html`: en la
  portada se marca «Inicio»; en un índice o una ficha, la de su sección. Las
  páginas de década marcan «Películas», porque una década no deja de ser un
  listado de películas: son taxonomía y no sección, y eso se atiende a mano.
- **Los recuentos salen de cada sección** (`.RegularPages`), igual que los de
  década, y una sección vacía no enseña el cero.
- **La barra no es fija**, a propósito: al no estar posicionada no crea
  contexto de apilamiento y no compite en z-index con el visor de fotogramas.

## Estructura

```text
hugo-fotogramas/
├── hugo.toml                  Configuración del sitio y el menú de pestañas
├── go.mod, go.sum             Solo para el extractor de paletas
├── archetypes/
│   ├── default.md
│   ├── peliculas.md           Plantilla de una película nueva
│   ├── series.md              Plantilla de una serie nueva
│   └── especial.md            Plantilla de un artículo nuevo
├── content/peliculas/         Lo que escribes tú
│   └── <slug>/
│       ├── index.md           La ficha
│       ├── cover.jpg          La portada
│       └── frames/            Los fotogramas
├── content/series/            Series: fichas de texto con portada
│   ├── _index.md
│   └── <slug>/index.md        (+ cover.jpg)
├── content/especial/          Artículos largos
│   ├── _index.md
│   └── <slug>/index.md
├── data/
│   └── palettes.json          Paletas y acentos: lo que genera el extractor
├── layouts/
│   ├── _default/
│   │   ├── baseof.html        Esqueleto: <html>, head, navegación, main
│   │   ├── single.html        Ficha de una película
│   │   └── list.html          Rejilla de /peliculas/
│   ├── taxonomy/
│   │   ├── taxonomy.html      /decades/, /paises/ y /generos/
│   │   └── term.html          /decades/1980/, /paises/<país>/ y /generos/<género>/
│   ├── series/
│   │   ├── list.html          /series/
│   │   └── single.html        /series/<slug>/
│   ├── especial/
│   │   ├── list.html          /especial/
│   │   └── single.html        /especial/<slug>/
│   ├── index.html             /, el mural de portadas
│   └── partials/
│       ├── head.html, scripts.html
│       ├── site-nav.html      Las pestañas de sección, en todas las páginas
│       ├── film-card.html, frame-grid.html, cover.html
│       ├── poster.html         La película del mural de la portada
│       ├── series-card.html   La tarjeta de series, sin paleta
│       ├── palette.html, video-box.html
│       ├── filters.html       Las dos filas de pestañas: décadas y origen
│       ├── decade-tabs.html, pais-tabs.html
│       ├── lightbox.html      El diálogo del visor de fotogramas
│       └── func/              Funciones: accent, fecha, film-data, youtube-id, …
├── assets/
│   ├── css/main.css           Todo el CSS está aquí
│   └── js/                    lightbox, palette, film
├── static/                    favicon
└── tools/palette/             El extractor, en Go
    ├── main.go                La línea de órdenes
    ├── films.go               Descubre los bundles y lee el frontmatter
    ├── extract.go             Imagen → paleta
    ├── color.go               La matemática de color y el acento
    ├── youtube.go             Reconoce un enlace de YouTube
    └── palettes.go            Lee y escribe data/palettes.json
```

Todo el CSS vive en `assets/css/main.css` y no dentro de las plantillas, para
que el HTML generado no lleve CSS incrustado.

## De Astro a Hugo: qué cambió y por qué

| En Astro | En Hugo | Por qué |
| :--- | :--- | :--- |
| `src/content/films/<slug>.md` | `content/peliculas/<slug>/index.md` | Un *page bundle* deja las imágenes junto a la ficha. El directorio se llama `peliculas` para que la URL salga sola. |
| `cover:` y `frames:` en el frontmatter | Los archivos se descubren solos | Una lista de rutas se desincroniza en cuanto añades una captura. |
| `image-data.json` con `frameSizes` | `data/palettes.json` sin tamaños | Hugo ya sabe las dimensiones del recurso. |
| `npm run extract` (Node + sharp) | `go run ./tools/palette` (Go) | Sin dependencias binarias, y el código es legible. |
| Esquema de Zod | `palette validate` | Hugo lee el frontmatter, pero no sabe si un enlace es de verdad un vídeo de YouTube. |
| Botones de década con JS | Taxonomías `decades`, `pais`, `genero` | Filtros de verdad: se comparten, se indexan, funcionan sin JS. |
| `getFilms()` en `utils/films.ts` | `.RegularPages.ByParam "year"` | Lo hace Hugo. |
| Componentes `.astro` | `layouts/` y `layouts/partials/` | La forma de Hugo. |

El acento también cambió de sitio: en Astro se calculaba en cada render; aquí
lo calcula el extractor y se guarda en `palettes.json`. Es el mismo resultado y
no se recalcula en cada build.

## Detalles de Hugo que costarían un rato

Anotados porque son los que más sorprenden:

- **Las plantillas de taxonomía van en `layouts/taxonomy/`**, no en `_default/`
  (ver arriba).
- **El campo de una taxonomía se escribe con el nombre plural** en el
  frontmatter (`paises:`, `generos:`, no `pais:` ni `genero:`). Con el
  singular Hugo no avisa: genera la página de la taxonomía vacía y ninguna
  página de término.
- **Los términos llegan en minúsculas** (`.Term`) y la página de término titula
  con mayúscula inicial en cada palabra («Corea Del Sur», «Ciencia Ficción»).
  El nombre tal y como está escrito en las fichas solo existe en el
  frontmatter, y lo recuperan `func/pais-nombre.html` y
  `func/genero-nombre.html` recorriendo las películas del término.
- **Las URLs no llevan tildes** (`removePathAccents` en `hugo.toml`): el
  término «Ciencia ficción» es `/generos/ciencia-ficcion/`, no
  `/generos/ciencia-ficción/`. Solo afecta a los slugs con acento, que hoy son
  los géneros.
- **`site.Language.LanguageCode` y `site.Data` están obsoletos** desde la
  v0.146: son `.Language.Locale` y `hugo.Data`.
- **`page.Pages` no tiene `.First` ni `.Last`.** Se usa `index .Pages 0`.
- **Un partial que escribe texto devuelve HTML sin escapar**, así que los `&` de
  una query-string llegan tal cual al atributo. `func/youtube-embed.html`
  devuelve con `return` a propósito, para que se escapen exactamente una vez y
  el HTML sea válido.
- **No existen `regexMatch` ni `strings.ContainsRE`.** `strings.FindRE` devuelve
  las coincidencias enteras; para capturar grupos está
  `strings.FindRESubmatch`.
- **`--printUnusedTemplates` da falsos positivos** con `{{ define "main" }}`: no
  detecta los bloques de las plantillas base. No es una fuente fiable.
- **`hugo.ServerPort` no existe** en la 0.167. La plantilla con `hugo --minify`
  compila sin quejarse y el servidor se cae con un error 500 en cada página. El
  puerto sale de `baseURL`.
- **No hay que inyectar `livereload.js`.** Hugo lo pone solo, al principio de
  `<head>`, con `data-no-instant`. Añadirlo a mano deja dos y se abren dos
  conexiones. La recarga en vivo ya funciona sin tocar nada.

## La pantalla de bienvenida

Antes del sitio hay una capa con una nube de puntos, dibujada en un `<canvas>`
por `assets/js/intro.js`. Debajo se cargan las portadas mientras aparece, así que
no hay salto de carga al entrar.

Decisiones que no son obvias y conviene no deshacer:

- **Es una capa encima, no una página.** La portada sigue siendo `/`, así que no
  hay trampa de "atrás" en el historial.
- **Sale una vez por sesión**, con `sessionStorage`. En cuanto se muestra deja la
  marca, así que recargar no la trae de vuelta. Para verla otra vez hay que abrir
  una ventana privada o borrar los datos del sitio.
- **Se puede forzar desde la barra de direcciones**, que es lo único que hace
  falta para poder trabajar en el lienzo:

  | URL          | Qué hace                                    |
  | :----------- | :------------------------------------------ |
  | `/?intro=1`  | La enseña siempre, y **sin** guardar la marca |
  | `/?intro=0`  | La esconde siempre                          |
  | `/`          | Lo de normal: una vez por sesión             |

  Con `?intro=1` la sesión sigue como si la intro no se hubiera visto, para que
  quitando el parámetro vuelva a salir sola.

- **La decisión está en un solo sitio.** Todo pasa por el JavaScript en línea de
  `partials/intro-gate.html`, que es lo único que corre antes del primer pintado.
  El CSS **no** vuelve a mirar nada. Si el CSS decidiera por su cuenta, podría
  pasar que la intro se quedara invisible con el contenido marcado `inert`: una
  web que no se puede usar.
- **`prefers-reduced-motion` gana siempre.** Se comprueba *después* del parámetro
  de la URL a propósito, así que ni `?intro=1` la enseña con esa opción puesta.
  Es lo único que el visitante no debería poder saltarse desde la barra.
- **El contenido se marca `inert` con JavaScript**, sobre los hijos de `<body>`.
  No se envuelve en un `div` nuevo, porque el ancho máximo y el centrado viven en
  `body` y un envoltorio los partiría en dos.
- **No lleva `role="dialog"`.** Es decorativa; un diálogo modal
  (`aria-modal`) dejaría a un lector de pantalla encerrado sin poder leer lo que
  hay debajo.

## Publicar

La web está en **[silencez0.github.io](https://silencez0.github.io/)**, servida
por GitHub Pages desde el repositorio público
[`SilenceZ0/SilenceZ0.github.io`](https://github.com/SilenceZ0/SilenceZ0.github.io).

```powershell
.\publicar.ps1
```

Ese repositorio es **una carpeta de salida**: dentro solo hay lo que produce
`hugo`, nada más. **Este repositorio sigue siendo privado**, con el código, las
fichas y las capturas. Es deliberado: a GitHub Pages va el sitio, no el material
con el que está hecho.

El script valida las fichas, compila, y empuja. Sale con `--baseURL` en la línea
de comandos, no tocando `hugo.toml`, así que en local los enlaces canónicos y el
RSS siguen apuntando a `localhost`. Para publicar antes de tiempo en otro sitio:

```powershell
.\publicar.ps1 -Repositorio 'propietario/repositorio' -BaseURL 'https://ejemplo/'
```

### Lo que hay que tener en cuenta

- **Un repositorio de usuario, no uno de proyecto.** Por eso el sitio sale en la
  raíz del dominio y no bajo un subdirectorio. Un repositorio de proyecto
  (`silencez0.github.io/peliculasdata/`) obligaría a que todas las rutas
  llevaran ese prefijo.
- **`--cleanDestinationDir` es obligatorio.** Sin él, Hugo deja en `public/` los
  ficheros de builds anteriores. Pasó: había trece versiones del CSS y cinco de
  `intro.js` acumuladas, y se habrían publicado todas.
- **El `baseURL` no se escribe en el repositorio de salida.** Vive en el comando
  de publicación, y por eso el `--force` del `push` es correcto: el destino es
  una carpeta generada, no un proyecto con historial que valga la pena.
- **El repositorio público tiene un solo commit.** La carpeta se inicializa de
  cero en cada publicación y se empuja por la fuerza, así que en GitHub no hay
  historial: solo el estado actual. El historial de verdad está en el
  repositorio privado, que es el del código.
- **La tipografía se sirve desde aquí**, en `static/fonts/`. Antes venía de
  Google Fonts. EB Garamond es una fuente variable: un solo archivo de 111 KB
  cubre los pesos 400 a 600. La licencia es SIL OFL 1.1 y su texto está junto al
  `.woff2`, en `static/fonts/OFL.txt`.
- **La ruta de la fuente va desde la raíz**, con la barra delante, en
  `assets/css/main.css` y en el `preload` de `layouts/partials/head.html`. El CSS
  minificado vive en `/css/main.min.<huella>.css`, así que una ruta relativa se
  resolvería contra `/css/` y no encontraría el archivo. Hugo no avisa de esto:
  compila limpio y lo único que se ve es la tipografía de reserva. Si algún día
  el sitio se sirviera bajo un subdirectorio, hay que cambiar las dos barras.

### Dominio propio

Ahora se sirve en `silencez0.github.io`. Si algún día quieres un dominio propio,
los `.es` rondan los 10 €/año, se pueden añadir en Cloudflare y apuntar con un
archivo `CNAME` dentro del repositorio de salida. Sigue funcionando por HTTPS sin
configurar nada más.