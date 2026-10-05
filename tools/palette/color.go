// color.go
//
// La matemática de color vive aquí y no en las plantillas. En una plantilla
// de Go, parsear un hexadecimal y calcular luminancias es ilegible; aquí es un
// archivo normal con sus pruebas. El extractor mide una vez, escribe el
// resultado en data/palettes.json y las plantillas solo pintan lo ya medido.
//
// Portado de src/utils/palette.ts de la versión de Astro, con los mismos
// coeficientes: cambiar un 0,65 por un 0,7 cambia el acento de todas las
// películas y las paletas deja de ser comparables con las de antes.

package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// minDistance es la distancia RGB mínima entre dos colores de una misma
// paleta. Por debajo de unos 46 dos muestras se leen como el mismo color y la
// paleta pierde su gracia.
const minDistance = 46

// rgb es un color de 8 bits por canal.
//
// Los canales son int y no uint8 porque durante la cuantización llegan valores
// mayores que 255 (ver quantize), y porque 264 es un cubo legítimo.
type rgb struct {
	R, G, B int
}

func clampChannel(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// rgbToHex escribe el color como #rrggbb, recortando a 255.
func rgbToHex(c rgb) string {
	return fmt.Sprintf("#%02x%02x%02x", clampChannel(c.R), clampChannel(c.G), clampChannel(c.B))
}

// parseHex acepta "#abc", "abc", "#aabbcc" y "aabbcc", en mayúsculas o no.
func parseHex(hex string) (rgb, bool) {
	body := strings.TrimPrefix(strings.TrimSpace(hex), "#")

	switch len(body) {
	case 3:
		// "#abc" es lo mismo que "#aabbcc".
		body = string([]byte{body[0], body[0], body[1], body[1], body[2], body[2]})
	case 6:
	default:
		return rgb{}, false
	}

	v, err := strconv.ParseUint(body, 16, 32)
	if err != nil {
		return rgb{}, false
	}
	return rgb{R: int(v >> 16 & 0xff), G: int(v >> 8 & 0xff), B: int(v & 0xff)}, true
}

// luminance es la luminancia perceived, de 0 (negro) a 1 (blanco).
func luminance(c rgb) float64 {
	return (0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)) / 255
}

// saturation es la saturación estilo HSV, de 0 (gris) a 1.
func saturation(c rgb) float64 {
	max, min := c.R, c.R
	if c.G > max {
		max = c.G
	}
	if c.B > max {
		max = c.B
	}
	if c.G < min {
		min = c.G
	}
	if c.B < min {
		min = c.B
	}
	if max == 0 {
		return 0
	}
	return float64(max-min) / float64(max)
}

// colorDistance es la distancia euclídea entre dos colores, de 0 a 441.
func colorDistance(a, b rgb) float64 {
	dr := float64(a.R - b.R)
	dg := float64(a.G - b.G)
	db := float64(a.B - b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// pickAccent elige el color con el que una película tiñe bordes, brillos y
// sombras.
//
// El extractor ordena cada paleta de oscuro a claro, así que palette[0] casi
// siempre es casi negro y no sirve como acento: un borde o un resplandor
// derivados de él son indistinguibles del estado en reposo. Aquí se busca un
// tono que sea a la vez saturado y visible sobre el fondo negro.
//
// Se calcula al medir la película y se guarda en palettes.json, en vez de
// recalcularse en cada build: es el mismo resultado, sin coste por render.
func pickAccent(palette []string) string {
	type scored struct {
		color rgb
		score float64
		lum   float64
	}

	all := make([]scored, 0, len(palette))
	for _, hex := range palette {
		c, ok := parseHex(hex)
		if !ok {
			continue
		}
		lum := luminance(c)
		// Se prefieren los tonos bien situados en el medio: demasiado oscuro
		// desaparece sobre negro, demasiado claro pierde el carácter de la
		// película.
		midRange := 1 - math.Min(math.Abs(lum-0.58)/0.58, 1)
		all = append(all, scored{
			color: c,
			score: saturation(c)*0.65 + midRange*0.35,
			lum:   lum,
		})
	}
	if len(all) == 0 {
		return "#ffffff"
	}

	// Se descartan los tonos inservibles antes de ordenar.
	pool := make([]scored, 0, len(all))
	for _, s := range all {
		if s.lum >= 0.25 && s.lum <= 0.9 {
			pool = append(pool, s)
		}
	}
	if len(pool) == 0 {
		pool = all
	}

	// SliceStable para que un empate conserve el orden de la paleta, igual que
	// un Array.sort de JavaScript.
	sort.SliceStable(pool, func(i, j int) bool { return pool[i].score > pool[j].score })

	return rgbToHex(pool[0].color)
}
