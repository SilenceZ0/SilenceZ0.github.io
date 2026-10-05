// palettes.go
//
// data/palettes.json: lo único que el extractor escribe y lo único que las
// plantillas leen. Las fichas nunca se tocan, así que volver a medir no
// ensucia el contenido.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// palettesPath es el archivo de datos medidos, relativo a la raíz.
const palettesPath = "data/palettes.json"

// filmData es lo medido de una película.
//
// No incluye las dimensiones de las imágenes: Hugo ya las sabe, con .Width y
// .Height del propio recurso, así que guardarlas sería duplicar un dato que
// puede quedarse viejo en cuanto se recorta una imagen.
type filmData struct {
	// Accent es el color con el que la película tiñe bordes y brillos.
	Accent string `json:"accent"`

	// Palette sale de la portada: es la que ven las tarjetas y la cabecera.
	Palette []string `json:"palette"`

	// Frames tiene una paleta por fotograma, en el mismo orden que los
	// archivos de frames/.
	Frames [][]string `json:"frames,omitempty"`
}

// paletteFile es el mapa slug -> medidas.
type paletteFile map[string]filmData

// readPalettes lee data/palettes.json. Si no existe todavía, devuelve un mapa
// vacío: es el estado normal de un proyecto recién clonado.
func readPalettes(path string) (paletteFile, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return paletteFile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	var out paletteFile
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%s: JSON inválido: %w", path, err)
	}
	if out == nil {
		out = paletteFile{}
	}
	return out, nil
}

// writePalettes escribe el archivo de datos medidos.
//
// encoding/json ordena las claves de un mapa, así que la salida es siempre la
// misma y los diffs no son ruido por reordenar.
func writePalettes(path string, data paletteFile) error {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
