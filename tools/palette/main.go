// palette
//
// Mide las imágenes de cada película y guarda el resultado en
// data/palettes.json. Es el equivalente del extractor de Astro, en Go y sin
// dependencias más allá del parser de YAML.
//
// Garantiza, por imagen:
//   - exactamente 5 colores, todos distintos entre sí
//   - ordenados de oscuro a claro, para que la paleta se lea como degradado
//   - un color de acento, el que la película usa para bordes y brillos
//
// Uso:
//
//	go run ./tools/palette extract              todas las películas sin medir
//	go run ./tools/palette extract matrix-1999  solo esa
//	go run ./tools/palette extract --force      volver a medirlo todo
//	go run ./tools/palette list                 ver el estado de cada película
//	go run ./tools/palette validate             comprobar fichas, enlaces y medidas
//
// Se pueden pasar slugs o títulos, con o sin tildes, y vale que coincidan
// parcialmente. Las películas no indicadas conservan lo que tuvieran medido.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "\nerror:", err)
		os.Exit(1)
	}
}

const help = `palette: mide las imágenes de las películas y guarda sus paletas

Uso:
  palette extract [opciones] [slug...]   medir imágenes y escribir data/palettes.json
  palette list [opciones] [slug...]      ver el estado de cada película
  palette validate [opciones] [slug...]  comprobar frontmatter, enlaces y medidas

Opciones:
  --root RUTA   raíz del proyecto Hugo (por defecto ".")
  --force       volver a medir también las películas que ya están medidas
  --dry-run     no escribir nada, solo informar de lo que se haría
  -h, --help    esta ayuda

Se puede pasar un slug o un título, con coincidencia parcial y sin tildes:

  palette extract blade
  palette validate dracula matrix
`

func run(args []string) error {
	// Sin subcomando se miden, que es lo que se hace casi siempre.
	command := "extract"
	if len(args) > 0 && isCommand(args[0]) {
		command, args = args[0], args[1:]
	}

	var (
		root   string
		force  bool
		dryRun bool
	)

	fs := flag.NewFlagSet("palette", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&root, "root", ".", "raíz del proyecto Hugo")
	fs.BoolVar(&force, "force", false, "volver a medir también las películas ya medidas")
	fs.BoolVar(&dryRun, "dry-run", false, "no escribir nada")
	fs.Usage = func() { fmt.Fprint(os.Stderr, help) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	films, err := readFilms(root)
	if err != nil {
		return err
	}

	out := filepath.Join(root, palettesPath)
	measured, err := readPalettes(out)
	if err != nil {
		return err
	}

	if queries := fs.Args(); len(queries) > 0 {
		films, err = selectFilms(films, queries)
		if err != nil {
			return err
		}
	}

	switch command {
	case "extract":
		return cmdExtract(out, films, measured, force, dryRun)
	case "list":
		return cmdList(films, measured)
	case "validate":
		return cmdValidate(films, measured)
	default:
		return fmt.Errorf("orden desconocida %q", command)
	}
}

func isCommand(name string) bool {
	switch name {
	case "extract", "list", "validate":
		return true
	}
	return false
}

/* ------------------------------------------------------------- extract -- */

func cmdExtract(out string, films []film, measured paletteFile, force, dryRun bool) error {
	printHeader("SLUG", "AÑO", "IMÁGENES", "ESTADO", "PALETA")

	toMeasure := 0
	for _, f := range films {
		if existing, ok := measured[f.Slug]; ok && !force {
			printRow(f, len(f.images()), "medida", existing.Palette)
			continue
		}

		data, err := measure(f)
		if err != nil {
			return fmt.Errorf("%s: %w", f.Slug, err)
		}
		measured[f.Slug] = data
		toMeasure++

		state := "medida"
		if dryRun {
			state = "MEDIRÍA"
		}
		printRow(f, len(f.images()), state, data.Palette)
	}

	fmt.Println()

	if dryRun {
		fmt.Printf("--dry-run: no se ha escrito nada. %d película(s) por medir.\n", toMeasure)
		return nil
	}

	if toMeasure == 0 {
		fmt.Println("Nada que medir. Usa --force para volver a medirlo todo.")
		return nil
	}

	if err := writePalettes(out, measured); err != nil {
		return err
	}

	fmt.Printf("%d película(s) medida(s) -> %s\n", toMeasure, out)
	return nil
}

// measure mide todas las imágenes de una película.
//
// La paleta de la película sale de la portada, o del primer fotograma si no
// hay portada. El acento se calcula sobre esa misma paleta, para que la
// tarjeta y la cabecera de la ficha digan lo mismo.
func measure(f film) (filmData, error) {
	images := f.images()
	if len(images) == 0 {
		return filmData{}, errors.New("no hay imágenes: se esperaba " + coverPrefix + "* o " + frameDir + "/")
	}

	head, err := extractPalette(images[0], paletteSize)
	if err != nil {
		return filmData{}, fmt.Errorf("portada %s: %w", filepath.Base(images[0]), err)
	}

	data := filmData{
		Accent:  pickAccent(head),
		Palette: head,
	}

	for _, path := range f.Frames {
		palette, err := extractPalette(path, paletteSize)
		if err != nil {
			return filmData{}, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		data.Frames = append(data.Frames, palette)
	}

	return data, nil
}

/* ---------------------------------------------------------------- list -- */

func cmdList(films []film, measured paletteFile) error {
	printHeader("SLUG", "AÑO", "IMÁGENES", "DECADA", "PALETA")

	measuredCount := 0
	for _, f := range films {
		decade := "—"
		if len(f.Decades) > 0 {
			decade = f.Decades[0]
		}

		var palette []string
		if data, ok := measured[f.Slug]; ok {
			palette = data.Palette
			measuredCount++
		}

		printRow(f, len(f.images()), decade, palette)
	}

	fmt.Printf("\n%d película(s), %d medida(s)\n", len(films), measuredCount)
	return nil
}

/* ------------------------------------------------------------ validate -- */

// cmdValidate comprueba las fichas y las medidas.
//
// Sustituye al esquema de Zod de la versión de Astro. Hugo ya se encarga de
// leer el frontmatter; lo que Hugo no puede saber es si un enlace es de verdad
// un vídeo de YouTube, ni si lo que hay medido encaja con las imágenes que hay
// ahora mismo en el disco. Eso es lo que se comprueba aquí.
func cmdValidate(films []film, measured paletteFile) error {
	errorCount, warningCount := 0, 0

	for _, f := range films {
		data, ok := measured[f.Slug]
		failures, warnings := validateFilm(f, data, ok)

		for _, w := range warnings {
			fmt.Printf("  aviso  %s: %s\n", f.Slug, w)
			warningCount++
		}
		for _, e := range failures {
			fmt.Printf("  error  %s: %s\n", f.Slug, e)
			errorCount++
		}
	}

	// Una entrada de palettes.json sin carpeta detrás es casi siempre un slug
	// que se renombró. Se avisa, no se falla: la entrada se limpia sola con
	// `extract --force`.
	for _, slug := range staleSlugs(measured, films) {
		fmt.Printf("  aviso  palettes.json: %s ya no tiene carpeta en %s\n", slug, contentDir)
		warningCount++
	}

	fmt.Printf("\n%d película(s) comprobada(s), %d error(es), %d aviso(s)\n",
		len(films), errorCount, warningCount)

	if errorCount > 0 {
		return errors.New("la colección no está lista")
	}
	return nil
}

func validateFilm(f film, data filmData, hasData bool) (failures, warnings []string) {
	/* --- frontmatter --- */

	if strings.TrimSpace(f.Title) == "" {
		failures = append(failures, "falta title")
	}

	// 1888 es la primera proyección cinematográfica; 2100 evita un año
	// tecleado de más.
	if f.Year < 1888 || f.Year > 2100 {
		failures = append(failures, fmt.Sprintf("year %d está fuera del rango 1888-2100", f.Year))
	}

	want := decadeOf(f.Year)
	switch {
	case len(f.Decades) == 0:
		failures = append(failures, fmt.Sprintf("falta decades; para %d debería ser [%q]", f.Year, want))
	case !containsString(f.Decades, want):
		failures = append(failures,
			fmt.Sprintf("decades %v no incluye %q, que es la década de %d", f.Decades, want, f.Year))
	}

	// El género es obligatorio y sale de la lista cerrada. La lista es la
	// única fuente de verdad: Hugo parte el término del valor del frontmatter,
	// y sin exigir coincidencia exacta «Terror» y «terror» se partirían en
	// dos páginas de término distintas sin que nada avisara.
	if len(f.Generos) == 0 {
		failures = append(failures, fmt.Sprintf("falta generos; por ejemplo [%q]", "Terror"))
	}
	for _, g := range f.Generos {
		if !containsString(generosValidos, g) {
			failures = append(failures, fmt.Sprintf(
				"generos %q no está en la lista; válidos: %s", g, strings.Join(generosValidos, ", ")))
		}
	}

	// El orden es fijo para que los mensajes salgan siempre igual.
	for _, field := range []struct{ name, ref string }{
		{"trailerUrl", f.Trailer},
		{"soundtrackUrl", f.Soundtrack},
	} {
		if strings.TrimSpace(field.ref) == "" {
			continue
		}
		if youtubeID(field.ref) == "" {
			failures = append(failures, fmt.Sprintf(
				"%s no es un enlace de YouTube válido: %q (se espera un id de 11 caracteres)",
				field.name, field.ref))
		}
	}

	/* --- imágenes --- */

	if f.Cover == "" {
		warnings = append(warnings,
			"sin "+coverPrefix+"*, así que la portada será el primer fotograma")
	}
	if len(f.images()) == 0 {
		failures = append(failures, "no hay imágenes: se esperaba "+coverPrefix+"* o "+frameDir+"/")
	}

	/* --- medidas --- */

	if !hasData {
		failures = append(failures,
			"sin medir: ejecuta `go run ./tools/palette extract "+f.Slug+"`")
		return failures, warnings
	}

	failures = append(failures, validatePalette("palette", data.Palette, paletteSize)...)
	failures = append(failures, validatePalette("accent", []string{data.Accent}, 1)...)

	if len(data.Frames) != len(f.Frames) {
		failures = append(failures, fmt.Sprintf(
			"frames tiene %d paleta(s) y hay %d fotograma(s)", len(data.Frames), len(f.Frames)))
	} else {
		for i, palette := range data.Frames {
			label := fmt.Sprintf("frames[%d] (%s)", i, filepath.Base(f.Frames[i]))
			failures = append(failures, validatePalette(label, palette, paletteSize)...)
		}
	}

	return failures, warnings
}

func validatePalette(label string, palette []string, want int) []string {
	var failures []string

	if len(palette) != want {
		failures = append(failures, fmt.Sprintf("%s tiene %d color(es), se esperan %d",
			label, len(palette), want))
	}
	for i, hex := range palette {
		if _, ok := parseHex(hex); !ok {
			failures = append(failures, fmt.Sprintf("%s[%d] no es un hexadecimal: %q", label, i, hex))
		}
	}

	return failures
}

/* ------------------------------------------------------------- helpers -- */

// selectFilms filtra por slug o título.
//
// Se acepta coincidencia parcial, sin tildes y sin distinguir mayúsculas, para
// no tener que acordarse del slug exacto: `palette extract matrix` basta.
func selectFilms(films []film, queries []string) ([]film, error) {
	terms := make([]string, 0, len(queries))
	for _, q := range queries {
		if t := normalize(q); t != "" {
			terms = append(terms, t)
		}
	}

	var out []film
	for _, f := range films {
		haystack := normalize(f.Slug + " " + f.Title + " " + f.Director)
		if matchesAny(haystack, terms) {
			out = append(out, f)
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("ninguna película coincide con %s", strings.Join(queries, ", "))
	}
	return out, nil
}

// matchesAny dice si la película coincide con alguno de los términos.
//
// Con más de un término es un «o» y no un «y», que es lo que promete la ayuda:
// `palette extract blade alien` mide las dos. Un «y» solo serviría para dar
// con películas cuyo nombre, título y director cumplieran las tres condiciones a
// la vez, que no es lo que se está pidiendo, y hacía que la ayuda mentía.
func matchesAny(haystack string, terms []string) bool {
	// Sin términos no hay nada que filtrar, así que vale la película.
	if len(terms) == 0 {
		return true
	}

	for _, t := range terms {
		if strings.Contains(haystack, t) {
			return true
		}
	}
	return false
}

// acentos sustituye las vocales acentuadas y la eñe por su versión sin
// acento, para que "dracula" encuentre a "Drácula".
var acentos = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u",
	"Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ü", "U",
	"ñ", "n", "Ñ", "N",
)

func normalize(s string) string {
	return strings.ToLower(acentos.Replace(strings.ToLower(strings.TrimSpace(s))))
}

func staleSlugs(measured paletteFile, films []film) []string {
	var out []string
	for slug := range measured {
		found := false
		for _, f := range films {
			if f.Slug == slug {
				found = true
				break
			}
		}
		if !found {
			out = append(out, slug)
		}
	}
	// El recorrido de un mapa de Go es aleatorio, y aquí el orden importa.
	sort.Strings(out)
	return out
}

func containsString(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func printHeader(cols ...string) {
	fmt.Printf("%-34s %-5s %-9s %-8s %s\n", cols[0], cols[1], cols[2], cols[3], cols[4])
	fmt.Println(strings.Repeat("-", 84))
}

// printRow imprime una línea de la tabla. Una paleta vacía se marca como
// "sin medir": es lo único que se puede decir de una película a la que todavía
// no se le ha medido la portada.
func printRow(f film, images int, state string, palette []string) {
	if len(palette) == 0 {
		palette = []string{"sin medir"}
	}
	fmt.Printf("%-34s %-5d %-9d %-8s %s\n", f.Slug, f.Year, images, state, strings.Join(palette, " "))
}
