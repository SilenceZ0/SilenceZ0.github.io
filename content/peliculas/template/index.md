---
# ═══════════════════════════════════════════════════════════════════════════
#  PLANTILLA DE FICHA DE PELÍCULA
# ═══════════════════════════════════════════════════════════════════════════
#  Esta carpeta es una ficha de ejemplo con TODAS las posibilidades del
#  sitio. Está en borrador, así que no se publica: es material de consulta y
#  para copiar y pegar.
#
#  Para verla en local hay que arrancar el servidor CON los borradores:
#
#      hugo server -D
#
#  (sin la -D, ni esta ficha ni las plantillas de series/especial se generan).
#
#  Pasos para una película nueva:
#
#    1. Copia esta carpeta entera y renómbrala al slug de la película. El
#       nombre de la carpeta es el slug Y la URL: content/peliculas/<slug>/
#    2. Sustituye cover.* y los archivos de frames/ por los de la película.
#    3. Edita los campos de abajo (los comentarios explican cada uno) y
#       escribe la sinopsis en el cuerpo, después del cierre `---`.
#    4. go run ./tools/palette extract <slug>
#    5. go run ./tools/palette combos   (si cambió la década, el país o el género)
#    6. go run ./tools/palette validate  (tiene que dar 0 errores)
#    7. hugo --minify                    (sin una sola línea WARN ni ERROR)
#    8. Quita `draft: true` para publicar.
#
#  No existe ningún campo `id`, `slug`, `cover` ni `frames`: el slug es la
#  carpeta y las imágenes se descubren por su nombre, solas.
# ═══════════════════════════════════════════════════════════════════════════

# OBLIGATORIO · El título de la ficha. Si lleva dos puntos, entre comillas
# (sin ellas, YAML cree que el `:` separa una clave): "2001: A Space Odyssey".
title: "Plantilla de película"

# OBLIGATORIO · Año de estreno, entero, entre 1888 y 2100. De aquí sale la
# década de la taxonomía.
year: 1979

# OPCIONAL · Se muestra como «Dirigida por …». Si no lo pones, no aparece.
director: "Ridley Scott"

# OBLIGATORIO · Países de origen, EN PLURAL y en lista, aunque sea uno solo.
# Tienen que estar en la lista cerrada `paisesValidos` (tools/palette/films.go).
# Con el nombre en singular (`pais:`) Hugo no genera nada y NO avisa.
paises: ["Reino Unido", "Estados Unidos"]

# OBLIGATORIO · Décadas, EN PLURAL y en lista. Tiene que incluir la de `year`:
# 1979 -> "1970", 2005 -> "2000". `palette validate` lo comprueba.
decades: ["1970"]

# OBLIGATORIO · Géneros, EN PLURAL y en lista, de la lista cerrada
# `generosValidos` (tools/palette/films.go). La lista entera está al final de
# este archivo. Un género fuera de la lista lo rechaza `palette validate`.
generos: ["Ciencia ficción", "Terror"]

# OPCIONALES · Enlaces de YouTube. Se admite el enlace completo (watch?v=,
# youtu.be/, /embed/, /shorts/, /live/, con o sin protocolo) o el id de 11
# caracteres suelto. Un enlace inválido falla en `palette validate` y no
# incrusta nada. Si faltan, su caja de vídeo simplemente no se dibuja.
trailerUrl: "https://www.youtube.com/watch?v=Eu9ZFTXXEiw"
soundtrackUrl: "https://youtu.be/6pnev1fa2bY?list=PLqnnuEVGcRQxuqYpNjdR32_dVRzDmVs1O"

# OPCIONAL · Meta descripción propia, solo si quieres controlarla a mano. Si
# no la pones, se genera sola con el título y el año.
# description: "…"

# BORRADOR · Mientras sea `true`, la ficha no se publica. La plantilla se
# queda así a propósito; quita la línea (o pon `false`) para publicar.
draft: true
---

**Esta es la plantilla de ficha de película.** Los datos de ejemplo son de
*Alien* (1979) y las imágenes son de reserva. Está en borrador, así que no se
publica; para verla en local hay que arrancar `hugo server -D`.

## Qué se ve en una ficha

Una ficha reúne, de arriba abajo:

- **La portada** (`cover.*`) como imagen de cabecera, con un degradado encima.
- **El título**, los **países**, el **año** y los **géneros**, que son enlaces
  a sus páginas de taxonomía (`/paises/…`, `/decades/…`, `/generos/…`).
- **La dirección**: la línea «Dirigida por …».
- **Las notas**: este mismo cuerpo, en Markdown.
- **La paleta**: 5 colores sacados de la portada, más el color de acento.
- **El recuento de fotogramas**.
- **El tráiler** y **la banda sonora**, cada uno en su caja de vídeo.
- **La rejilla de fotogramas**: al pasar el ratón por uno, cambia la paleta.

Todo lo que no esté en el frontmatter o en `frames/` no se pinta: no hay que
rellenar huecos. La paleta y el acento **no se escriben**, los calcula el
extractor a partir de las imágenes.

## El frontmatter, campo a campo

| Campo | ¿Obligatorio? | Qué es y qué pinta |
| :---- | :-----------: | :----------------- |
| `title` | Sí | Título de la película. Con `:` dentro, entre comillas. |
| `year` | Sí | Año de estreno (1888–2100). Determina la década. |
| `paises` | Sí | Lista de países, en plural, de la lista cerrada. |
| `decades` | Sí | Lista de décadas, en plural; debe incluir la de `year`. |
| `generos` | Sí | Lista de géneros, en plural, de la lista cerrada. |
| `director` | No | Se muestra como «Dirigida por …». |
| `trailerUrl` | No | Vídeo de YouTube del tráiler. |
| `soundtrackUrl` | No | Vídeo de YouTube de la banda sonora. |
| `description` | No | Meta descripción propia (si no, se genera sola). |
| `draft` | No | Si es `true`, la ficha no se publica. |

Los tres campos de taxonomía van **en plural** (`paises`, `decades`,
`generos`), aunque la lista tenga un solo valor. Son los que crean las páginas
de origen, década y género y los cruces de `/filtros/`; con el singular Hugo
no genera nada y no avisa de nada.

## Las imágenes no se declaran

Van dentro de la carpeta, junto a este `index.md`:

```text
content/peliculas/<slug>/
├── index.md
├── cover.jpg          <- la portada (cover.*)
└── frames/
    ├── frame-01.jpg   <- los fotogramas, con cero a la izquierda
    ├── frame-02.jpg
    └── …
```

- La portada es cualquier `cover.*` en la raíz. Extensiones que el sitio
  reconoce: `.jpg`, `.jpeg`, `.png`, `.gif` y `.webp`. Si hay varias, gana la
  de nombre más corto (`cover.jpg` antes que `cover-portada.jpg`).
- Los fotogramas van en `frames/`. **El cero a la izquierda importa**: el
  orden alfabético es el orden de visionado, y es el mismo orden que empareja
  cada imagen con su paleta medida.
- **Sin portada**: el primer fotograma hace de portada y `palette validate`
  lo avisa. **Sin fotogramas**: la ficha sale con «Aún no hay fotogramas» y
  la paleta se saca de la portada.
- Cada vez que añadas o quites imágenes hay que volver a medir:
  `go run ./tools/palette extract <slug>` (o `--force` para todas).

## El cuerpo

Es la sinopsis o las notas de la ficha, en Markdown normal: párrafos,
**negrita**, *cursiva*, listas, enlaces y citas. Lo que escribas aquí aparece
como las notas de la película.

En una ficha **no** se incrustan vídeos a mano: el tráiler y la banda sonora
salen de `trailerUrl` y `soundtrackUrl`. Los shortcodes `lead` y `video` son
de los artículos de `/especial/`, no de las fichas de película.

## Géneros y países válidos

Los géneros salen de la lista cerrada `generosValidos` de
`tools/palette/films.go`. Estos son todos:

```text
Acción, Animación, Aventura, Bélico, Ciencia ficción, Comedia, Crimen,
Documental, Drama, Extraterrestres, Fantasía, Musical, Misterio, Romance,
Superhéroes, Terror, Thriller, Western
```

Los países salen de `paisesValidos`, en el mismo archivo. **No es un catálogo
del mundo, sino los nombres que se pueden usar**: para dar de alta uno nuevo,
se añade a esa lista y se escribe igual en la ficha. Estos son todos:

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

Los que no use ninguna película no crean página: Hugo solo genera los términos
vivos. La lista canónica también está en el `README.md`.

## Antes de publicar

```sh
go run ./tools/palette extract <slug>   # mide las imágenes
go run ./tools/palette combos           # solo si cambió década, país o género
go run ./tools/palette validate         # tiene que dar 0 error(es)
hugo --minify                           # sin WARN ni ERROR
```

Y, cuando la ficha esté lista, quita `draft: true`. La URL será
`/peliculas/<slug>/`, donde `<slug>` es el nombre de esta carpeta.
