// films.go
//
// Una película es un page bundle: una carpeta con index.md y las imágenes al
// lado. El nombre de la carpeta es el slug, que es también la clave en
// data/palettes.json, y las imágenes se descubren por nombre en vez de
// declararse en el frontmatter.
//
// Esa es la diferencia grande con la versión de Astro: allí cada ficha llevaba
// una lista de rutas (cover: y frames:) que había que mantener a mano y que se
// desincronizaba en cuanto añadías una captura. Aquí la ficha solo dice qué
// película es.

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// contentDir es la carpeta con los page bundles, relativa a la raíz.
	contentDir = "content/peliculas"

	// frameDir es la subcarpeta con los fotogramas, dentro de cada bundle.
	frameDir = "frames"

	// coverPrefix es el nombre que debe tener la portada.
	coverPrefix = "cover."
)

// imageExt son las extensiones que se consideran imágenes.
//
// webP está en la lista porque Hugo sí lo muestra, aunque la biblioteca
// estándar de Go no lo sabe decodificar: si evera un .webp, el extractor lo
// dice en vez de medirlo a medias.
var imageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// film es una película leída del disco.
type film struct {
	Slug       string // nombre de la carpeta; también la clave en palettes.json
	Dir        string // ruta de la carpeta
	IndexPath  string // ruta del index.md
	Title      string
	Year       int
	Director   string
	Paises     []string
	Decades    []string
	Generos    []string
	Draft      bool
	Trailer    string
	Soundtrack string

	Cover  string   // ruta absoluta; "" si no hay
	Frames []string // rutas absolutas, ordenadas por nombre
}

// images devuelve la portada y los fotogramas, en el orden en que se miden.
func (f film) images() []string {
	if f.Cover != "" {
		return append([]string{f.Cover}, f.Frames...)
	}
	return f.Frames
}

// frontMatter es el frontmatter de index.md.
//
// Los campos que solo existían en la versión de Astro (cover, frames) no
// aparecen: en un page bundle las imágenes se descubren solas.
type frontMatter struct {
	Title         string   `yaml:"title"`
	Year          int      `yaml:"year"`
	Director      string   `yaml:"director"`
	Paises        []string `yaml:"paises"`
	Decades       []string `yaml:"decades"`
	Generos       []string `yaml:"generos"`
	Draft         bool     `yaml:"draft"`
	TrailerURL    string   `yaml:"trailerUrl"`
	SoundtrackURL string   `yaml:"soundtrackUrl"`
}

// generosValidos es la lista cerrada de géneros de la colección.
//
// Es la única lista que existe: Hugo crea los términos de la taxonomía desde
// lo que haya en cada frontmatter, y validate exige que cada `generos` salga
// de aquí. Sin esa lista, «Terror» y «terror» se partirían en dos páginas de
// término distintas sin que nada avisara, igual que pasa con un `pais:` en
// singular. Los géneros que ninguna película use no crean página: Hugo solo
// genera los términos que están vivos.
var generosValidos = []string{
	"Acción",
	"Animación",
	"Aventura",
	"Bélico",
	"Ciencia ficción",
	"Comedia",
	"Crimen",
	"Documental",
	"Drama",
	"Expresionismo",
	"Extraterrestres",
	"Fantasía",
	"Musical",
	"Misterio",
	"Romance",
	"Superhéroes",
	"Terror",
	"Thriller",
	"Western",
}

// paisesValidos es la lista cerrada de países de origen de la colección.
//
// Es la gemela de generosValidos y existe por la misma razón: Hugo crea el
// término de la taxonomía desde lo que encuentre en el frontmatter, y sin
// exigir coincidencia exacta «Canadá» y «Canada» (o «Estados Unidos» y
// «EEUU») partirían /paises/ en dos páginas sin que nada avisara.
//
// No es un catálogo de los países del mundo, sino la lista de los que se
// pueden usar: para dar de alta uno nuevo basta con añadirlo aquí y usarlo en
// una ficha. Los que no use ninguna película no crean página, igual que con
// los géneros.
//
// El orden es el alfabético ignorando acentos, para que al mirar la lista se
// vea de un vistazo si un nombre ya está o no.
var paisesValidos = []string{
	"Alemania",
	"Alemania del Oeste",
	"Argentina",
	"Australia",
	"Austria",
	"Bélgica",
	"Brasil",
	"Canadá",
	"Checoslovaquia",
	"Chile",
	"China",
	"Colombia",
	"Corea del Sur",
	"Croacia",
	"Cuba",
	"Dinamarca",
	"Egipto",
	"Ecuador",
	"Eslovaquia",
	"Eslovenia",
	"España",
	"Estados Unidos",
	"Estonia",
	"Filipinas",
	"Finlandia",
	"Francia",
	"Grecia",
	"Hong Kong",
	"Hungría",
	"India",
	"Indonesia",
	"Irán",
	"Irlanda",
	"Islandia",
	"Israel",
	"Italia",
	"Japón",
	"Kenia",
	"Letonia",
	"Líbano",
	"Lituania",
	"Luxemburgo",
	"Marruecos",
	"México",
	"Nigeria",
	"Noruega",
	"Nueva Zelanda",
	"Países Bajos",
	"Pakistán",
	"Paraguay",
	"Perú",
	"Polonia",
	"Portugal",
	"Reino Unido",
	"República Checa",
	"Rumanía",
	"Rusia",
	"Senegal",
	"Serbia",
	"Singapur",
	"Sudáfrica",
	"Suecia",
	"Suiza",
	"Taiwán",
	"Tailandia",
	"Túnez",
	"Turquía",
	"Ucrania",
	"Unión Soviética",
	"Uruguay",
	"Venezuela",
	"Vietnam",
	"Yugoslavia",
}

// readFilms devuelve las películas de content/peliculas, ordenadas por slug.
func readFilms(root string) ([]film, error) {
	base := filepath.Join(root, contentDir)

	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer %s: %w", contentDir, err)
	}

	var films []film
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		slug := entry.Name()
		dir := filepath.Join(base, slug)
		indexPath := filepath.Join(dir, "index.md")

		if _, err := os.Stat(indexPath); err != nil {
			return nil, fmt.Errorf("%s: falta index.md, no es un page bundle", slug)
		}

		f, err := readFilm(slug, dir, indexPath)
		if err != nil {
			return nil, err
		}
		films = append(films, f)
	}

	sort.Slice(films, func(i, j int) bool { return films[i].Slug < films[j].Slug })
	return films, nil
}

func readFilm(slug, dir, indexPath string) (film, error) {
	raw, err := os.ReadFile(indexPath)
	if err != nil {
		return film{}, fmt.Errorf("%s: %w", slug, err)
	}

	body, err := splitFrontMatter(string(raw))
	if err != nil {
		return film{}, fmt.Errorf("%s: %w", slug, err)
	}

	var fm frontMatter
	if err := yaml.Unmarshal([]byte(body), &fm); err != nil {
		return film{}, fmt.Errorf("%s: frontmatter inválido: %w", slug, err)
	}

	cover, frames := discoverImages(dir)

	return film{
		Slug:       slug,
		Dir:        dir,
		IndexPath:  indexPath,
		Title:      fm.Title,
		Year:       fm.Year,
		Director:   fm.Director,
		Paises:     fm.Paises,
		Decades:    fm.Decades,
		Generos:    fm.Generos,
		Draft:      fm.Draft,
		Trailer:    fm.TrailerURL,
		Soundtrack: fm.SoundtrackURL,
		Cover:      cover,
		Frames:     frames,
	}, nil
}

// splitFrontMatter separa el bloque --- ... --- del cuerpo Markdown.
func splitFrontMatter(text string) (string, error) {
	// El BOM rompe la comparación con "---", y los CRLF ensucian el cierre.
	text = strings.TrimPrefix(text, "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")

	rest, ok := strings.CutPrefix(text, "---")
	if !ok {
		return "", errors.New("no empieza por ---")
	}

	rest = strings.TrimLeft(rest, "\n")

	end := strings.Index(rest, "\n---")
	if end == -1 {
		return "", errors.New("no hay cierre de frontmatter")
	}

	return rest[:end], nil
}

// discoverImages encuentra la portada y los fotogramas del bundle.
//
// Se descubren por nombre, no por una lista en el frontmatter: añadir una
// captura es soltar el archivo en frames/ y nada más. Los nombres llevan cero
// a la izquierda (frame-01.jpg), así que el orden alfabético es el orden de
// visionado, y es el mismo orden que usa la plantilla al ordenarlos.
func discoverImages(dir string) (string, []string) {
	cover := findCover(dir)

	entries, err := os.ReadDir(filepath.Join(dir, frameDir))
	if err != nil {
		return cover, nil
	}

	frames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !imageExt[strings.ToLower(filepath.Ext(entry.Name()))] {
			continue
		}
		frames = append(frames, filepath.Join(dir, frameDir, entry.Name()))
	}

	// os.ReadDir ya viene ordenado por nombre, pero se dice explícitamente
	// porque de este orden depende el emparejamiento con las paletas.
	sort.Strings(frames)

	return cover, frames
}

// findCover busca la portada en la raíz del bundle.
func findCover(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	best := ""
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !imageExt[strings.ToLower(filepath.Ext(name))] {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(name), coverPrefix) {
			continue
		}

		// cover.jpg gana a cover-portada.jpg: el nombre más corto es el que
		// Hugo encuentra con Resources.GetMatch "cover.*".
		if best == "" || len(name) < len(best) {
			best = name
		}
	}

	if best == "" {
		return ""
	}
	return filepath.Join(dir, best)
}

// decadeOf es la década de un año, del tipo 1982 -> "1980".
func decadeOf(year int) string {
	return strconv.Itoa(year / 10 * 10)
}
