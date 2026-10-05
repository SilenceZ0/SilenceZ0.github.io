{{- /*
  Plantilla para una película nueva. Se crea con:

    hugo new content peliculas/the-matrix-1999/index.md

  Después hay que poner cover.jpg y frames/frame-01.jpg… en esa misma carpeta.
  No hay que escribir ninguna lista de imágenes: el bundle las recoge solas y
  `go run ./tools/palette extract -- <slug>` mide sus paletas.

  Ojo: `decades` tiene que coincidir con el año. `palette validate` lo avisa.
*/ -}}
---
title: "{{ replace (path.Base (path.Dir .File.Path)) "-" " " | title }}"
year: 2024
director: ""
decades: ["2020"]
trailerUrl: ""
soundtrackUrl: ""
---
