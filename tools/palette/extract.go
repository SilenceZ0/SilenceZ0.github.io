// extract.go
//
// Medir una imagen: reducirla, contar sus colores y quedarse con cinco que se
// distinguen entre sí.

package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"sort"
)

const (
	// paletteSize es el número de colores de cada paleta.
	paletteSize = 5

	// sampleWidth y sampleHeight son la caja a la que se reduce cada imagen
	// antes de contar.
	//
	// Las capturas son mucho mayores, así que 160x160 deja del orden de
	// 25.000 píxeles: de sobra para que la paleta sea estable, y lo bastante
	// rápido para leer cien imágenes de un tirón.
	sampleWidth  = 160
	sampleHeight = 160
)

// extractPalette devuelve count colores de la imagen, todos distintos entre sí
// y ordenados de oscuro a claro para que la paleta se lea como un degradado.
func extractPalette(path string, count int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("no se pudo decodificar: %w", err)
	}

	picked := selectDistinct(mostCommon(resizeInside(src, sampleWidth, sampleHeight)), count)

	sort.SliceStable(picked, func(i, j int) bool {
		return luminance(picked[i]) < luminance(picked[j])
	})

	out := make([]string, len(picked))
	for i, c := range picked {
		out[i] = rgbToHex(c)
	}
	return out, nil
}

// resizeInside reduce la imagen para que quepa en maxW x maxH sin deformarla,
// promediando cada bloque de píxeles de origen.
//
// El promedio por bloques es lo que corresponde aquí, y no un kernel más
// elaborado como el Lanczos3 que usaba sharp. Dos razones:
//
//   - El kernel de Lanczos tiene lóbulos negativos, y al recortar a 0-255
//     produce halos en los bordes. Para contar píxeles eso significa inventar
//     colores que no están en la imagen, que es justo lo que una paleta de
//     color no debe hacer.
//   - Las capturas se reducen de 1200-1920 px a 160. A esa razón el promedio
//     por bloques y cualquier kernel de soporte amplio convergen al mismo
//     valor, así que la diferencia no se nota.
//
// Se promedia en espacio sRGB a propósito. Un promedio lineal de RGB es más
// correcto que uno en gamma, pero lo que interesa aquí es describir el color
// que se ve, no hacer física de color.
func resizeInside(src image.Image, maxW, maxH int) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	scale := math.Min(float64(maxW)/float64(w), float64(maxH)/float64(h))
	// No se agranda: una imagen más pequeña que la caja se queda como está.
	if scale > 1 {
		scale = 1
	}
	dstW := max(1, int(math.Round(float64(w)*scale)))
	dstH := max(1, int(math.Round(float64(h)*scale)))

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	for dy := 0; dy < dstH; dy++ {
		// El bloque de origen que cae en esta fila de destino. Como
		// (dy+1) <= dstH, y1 nunca pasa de b.Max.Y.
		y0 := b.Min.Y + dy*h/dstH
		y1 := b.Min.Y + (dy+1)*h/dstH
		if y1 <= y0 {
			y1 = y0 + 1
		}

		for dx := 0; dx < dstW; dx++ {
			x0 := b.Min.X + dx*w/dstW
			x1 := b.Min.X + (dx+1)*w/dstW
			if x1 <= x0 {
				x1 = x0 + 1
			}

			var sumR, sumG, sumB, n uint64
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					// RGBAModel quita la premultiplicacion del alfa, que es
					// lo mismo que hacia el removeAlpha() del extractor
					// original.
					c := color.RGBAModel.Convert(src.At(x, y)).(color.RGBA)
					sumR += uint64(c.R)
					sumG += uint64(c.G)
					sumB += uint64(c.B)
					n++
				}
			}

			o := dst.PixOffset(dx, dy)
			dst.Pix[o+0] = uint8(sumR / n)
			dst.Pix[o+1] = uint8(sumG / n)
			dst.Pix[o+2] = uint8(sumB / n)
			dst.Pix[o+3] = 0xff
		}
	}

	return dst
}

// mostCommon cuenta los píxeles por cubos de color y devuelve los cubos del
// más frecuente al menos frecuente.
func mostCommon(img *image.RGBA) []rgb {
	// Un struct de tres enteros como clave de mapa es mucho más rápido que una
	// cadena del tipo "168,168,192".
	type bucket struct{ R, G, B int }

	counts := make(map[bucket]int, 1024)

	// El orden de aparición importa: cuando dos cubos empatan en frecuencia
	// gana el que se vio primero, y el recorrido de un mapa de Go es
	// aleatorio. Se lleva aparte para poder reproducirlo siempre igual.
	var order []bucket

	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			o := img.PixOffset(x, y)
			// Cuantización gruesa: mantiene la paleta legible en vez de ruidosa.
			c := bucket{
				quantize(img.Pix[o+0]),
				quantize(img.Pix[o+1]),
				quantize(img.Pix[o+2]),
			}
			if _, seen := counts[c]; !seen {
				order = append(order, c)
			}
			counts[c]++
		}
	}

	ranked := make([]bucket, len(order))
	copy(ranked, order)

	// SliceStable para que los empates respeten el orden de aparición, igual
	// que un Array.sort de JavaScript.
	sort.SliceStable(ranked, func(i, j int) bool {
		return counts[ranked[i]] > counts[ranked[j]]
	})

	out := make([]rgb, len(ranked))
	for i, c := range ranked {
		out[i] = rgb{R: c.R, G: c.G, B: c.B}
	}
	return out
}

// quantize agrupa cada canal en cubos de 24.
//
// El recorte a 255 se hace al escribir el hexadecimal, no aquí: el cubo 264
// aparece de verdad entre los píxeles más claros y forma parte de la clave.
func quantize(v uint8) int {
	return int(math.Round(float64(v)/24)) * 24
}

// selectDistinct elige count colores distintos por frecuencia decreciente.
//
// Si con lo que hay no llega a count, deriva tonos aclarando los que ya ha
// cogido: pasa mucho con fotogramas muy planos o muy oscuros, donde no existen
// cinco colores distinguibles.
func selectDistinct(candidates []rgb, count int) []rgb {
	picked := make([]rgb, 0, count)

	for _, candidate := range candidates {
		if len(picked) >= count {
			break
		}
		if farEnough(picked, candidate) {
			picked = append(picked, candidate)
		}
	}

	if len(picked) > 0 {
		base := make([]rgb, len(picked))
		copy(base, picked)

		for shift := 34; len(picked) < count && shift < 160; shift += 34 {
			for _, seed := range base {
				if len(picked) >= count {
					break
				}
				derived := rgb{
					R: clampChannel(seed.R + shift),
					G: clampChannel(seed.G + shift),
					B: clampChannel(seed.B + shift),
				}
				if farEnough(picked, derived) {
					picked = append(picked, derived)
				}
			}
		}
	}

	return picked
}

// farEnough dice si c está al menos a minDistance de todos los ya cogidos.
func farEnough(picked []rgb, c rgb) bool {
	for _, p := range picked {
		if colorDistance(p, c) < minDistance {
			return false
		}
	}
	return true
}
