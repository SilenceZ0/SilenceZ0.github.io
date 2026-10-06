---
title: "{{ replace (path.Base (path.Dir .File.Path)) "-" " " | title }}"
year: {{ now.Year }}
platform: ""
draft: true
---

{{- /*
  Plantilla de una serie nueva. Se crea con:

    hugo new series/<slug>/index.md

  Después hay que poner cover.jpg en esa misma carpeta, junto al index.md.
  Las series son fichas de texto: no llevan fotogramas, ni paleta, ni tráiler,
  y por eso la herramienta de paletas no las toca (sigue leyendo solo
  content/peliculas/).

  `year` es el año de estreno, y `platform` la plataforma o cadena. Si la serie
  se emitió en varios años, el intervalo se puede escribir en `year` como
  texto: `year: "2019–2022"`.
*/ -}}
