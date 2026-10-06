---
title: "{{ replace (path.Base (path.Dir .File.Path)) "-" " " | title }}"
date: {{ .Date }}
draft: true
---

{{- /*
  Plantilla de artículo para la sección Especial. Se crea con:

    hugo new especial/mi-articulo/index.md

  El artículo se queda en borrador hasta que se quita `draft: true`, así que
  no se publica a medias. El cuerpo es Markdown normal: títulos, listas,
  citas, imágenes (suelta los archivos junto al index.md) y enlaces.
*/ -}}
