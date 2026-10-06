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

Para crear uno propio:

```sh
hugo new especial/mi-articulo/index.md
```

Quita `draft: true` cuando el artículo esté listo para publicarse. Si le
pones `description`, esa frase es la que sale en los buscadores; si no, se
usa el resumen del principio.
