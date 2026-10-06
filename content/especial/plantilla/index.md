---
title: "Plantilla de artículo"
date: 2026-10-06
draft: true
---

Este es un artículo de ejemplo. **Está en borrador**, así que no se publica:
sirve para ver cómo queda la prosa de la sección y para copiar de él.

## Qué tipo de contenido cabe aquí

Recomendaciones, listas, comparaciones, lo que sea. El cuerpo es Markdown
normal:

- títulos con `##` y `###`,
- **negritas** y *cursivas*,
- enlaces [a cualquier parte](https://gohugo.io/),
- citas:

> La ciencia ficción no predice el futuro: nos obliga a mirar el presente de
> costado.

- y tablas:

| Película | Año | Década |
| :------- | --: | :----- |
| Stalker  | 1979 | 1970 |

## Imágenes y vídeo

- **La imagen de la tarjeta** del índice es `cover.jpg`, dentro de la carpeta
  del artículo, junto al `index.md`. Da igual el tamaño o la proporción del
  original: el CSS la recorta siempre a 16/9, así que todas las tarjetas
  quedan igual.
- **Las imágenes del cuerpo** van también junto al `index.md` y se citan en
  el Markdown: `![Descripción](foto.jpg)`.
- **Los vídeos de YouTube** se incrustan con el shortcode `video`, para
  escucharlos sin salir de la web:

  ```text
  {{< video url="https://www.youtube.com/watch?v=..." label="Banda sonora" >}}
  ```

  Acepta el enlace completo o el id suelto. Si el artículo lleva dos vídeos,
  dales `anchor` distinto (`anchor="trailer"`, `anchor="banda"`).

Para crear uno propio:

```sh
hugo new especial/mi-articulo/index.md
```

Quita `draft: true` cuando el artículo esté listo para publicarse. Si le
pones `description`, esa frase es la que sale en los buscadores; si no, se
usa el resumen del principio.
