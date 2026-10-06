// combos.go
//
// Genera las páginas de filtro combinado en content/filtros/: una por cada
// combinación de dos o tres dimensiones —década, país de origen y género—
// cuyo resultado no esté vacío. Lo que es una sola dimensión ya vive en las
// taxonomías (/decades/, /paises/, /generos/), así que aquí solo entran los
// cruces.
//
// Cada combinación es un page bundle con su index.md, cuyo nombre de carpeta
// es el slug canónico: la década, el país y el género unidos por guiones, en
// ese orden y saltando la dimensión que no participa:
//
//	1980-reino-unido           décadas × paises
//	1980-terror                décadas × generos
//	reino-unido-terror         paises × generos
//	1980-reino-unido-terror    las tres
//
// content/filtros/ no se escribe a mano: palette combos lo sincroniza entero,
// creando lo que falta y borrando lo que sobra, y palette validate comprueba
// que sigue al día. Las películas en borrador no entran en ningún combo.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// filtrosDir es la sección con las páginas combinadas, relativa a la raíz.
const filtrosDir = "content/filtros"

// contexto es una combinación de filtros activos. Una cadena vacía significa
// que esa dimensión no participa. El orden de los campos es el canónico.
type contexto struct {
	Decada string
	Pais   string
	Genero string
}

// numeroDimensiones cuenta las dimensiones con valor.
func (c contexto) numeroDimensiones() int {
	n := 0
	if c.Decada != "" {
		n++
	}
	if c.Pais != "" {
		n++
	}
	if c.Genero != "" {
		n++
	}
	return n
}

// slug es el nombre de carpeta y la URL de la página: las etiquetas de las
// dimensiones activas unidas por guiones, en orden canónico.
func (c contexto) slug() string {
	var partes []string
	if c.Decada != "" {
		partes = append(partes, slugify(c.Decada))
	}
	if c.Pais != "" {
		partes = append(partes, slugify(c.Pais))
	}
	if c.Genero != "" {
		partes = append(partes, slugify(c.Genero))
	}
	return strings.Join(partes, "-")
}

// titulo es el título humano de la página, para el frontmatter.
func (c contexto) titulo() string {
	var partes []string
	if c.Decada != "" {
		partes = append(partes, decadaEtiqueta(c.Decada))
	}
	if c.Pais != "" {
		partes = append(partes, c.Pais)
	}
	if c.Genero != "" {
		partes = append(partes, c.Genero)
	}
	return strings.Join(partes, " · ")
}

// filtroFM es el frontmatter de una página de filtro combinado.
//
// Los nombres de campo son deliberadamente distintos de los de las
// taxonomías (paises:, generos:, decades:): en Hugo, ese nombre plural es el
// que activa la taxonomía, y una página de /filtros/ no debe colarse en
// /paises/x/ ni en /generos/y/ como si fuera una película más.
type filtroFM struct {
	Title  string `yaml:"title"`
	Decada string `yaml:"decada,omitempty"`
	Pais   string `yaml:"pais,omitempty"`
	Genero string `yaml:"genero,omitempty"`
}

// slugify es la misma transliteración que Hugo aplica a los términos de las
// taxonomías: minúsculas, sin tildes, espacios por guiones.
//
// Los slugs de los enlaces no salen de aquí —las plantillas los leen de la
// URL del propio término— pero el nombre de carpeta tiene que coincidir para
// que /filtros/<slug>/ exista de verdad. Hugo hace esto mismo a través de su
// propia tabla de transliteración; aquí se cubre el español, que es de donde
// salen los términos de esta colección.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = acentos.Replace(s)
	s = strings.ReplaceAll(s, " ", "-")

	// Un nombre con espacios pegajosos o tildes abiertas no debe dejar
	// guiones dobles ni recortar el slug.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

// decadaEtiqueta es la década en español, igual que la plantilla
// decade-label.html: "Años 60" para 1968 y el año entero a partir de 2000.
func decadaEtiqueta(s string) string {
	if s == "" {
		return ""
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	if n >= 2000 {
		return fmt.Sprintf("Años %d", n)
	}
	return fmt.Sprintf("Años %d", n%100)
}

// cmdCombos mantiene content/filtros/ sincronizado con las películas:
// crea los combos que faltan, actualiza los que cambian y borra los que ya no
// tienen ninguna película detrás.
func cmdCombos(films []film, root string, dryRun bool) error {
	esperados, err := expectedCombos(films)
	if err != nil {
		return err
	}

	base := filepath.Join(root, filtrosDir)

	// En seco solo se informa; hay que saber el estado actual igualmente.
	actuales := map[string]bool{}
	entries, err := os.ReadDir(base)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				actuales[e.Name()] = true
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("no se pudo leer %s: %w", filtrosDir, err)
	}

	var crea, borra []string
	for slug := range esperados {
		if !actuales[slug] {
			crea = append(crea, slug)
		}
	}
	for dir := range actuales {
		if _, ok := esperados[dir]; !ok {
			borra = append(borra, dir)
		}
	}
	sort.Strings(crea)
	sort.Strings(borra)

	if dryRun {
		if len(crea) == 0 && len(borra) == 0 {
			fmt.Printf("%s está al día (%d combos).\n", filtrosDir, len(esperados))
			return nil
		}
		for _, slug := range crea {
			fmt.Printf("crearía /filtros/%s/\n", slug)
		}
		for _, slug := range borra {
			fmt.Printf("borraría /filtros/%s/\n", slug)
		}
		return nil
	}

	if err := os.MkdirAll(base, 0o755); err != nil {
		return fmt.Errorf("no se pudo crear %s: %w", filtrosDir, err)
	}

	// Se ordenan por slug para que el resultado no dependa del recorrido de
	// los mapas.
	slugs := make([]string, 0, len(esperados))
	for slug := range esperados {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	escritos := 0
	for _, slug := range slugs {
		ctx := esperados[slug]
		escrito, err := writeCombo(base, slug, ctx)
		if err != nil {
			return err
		}
		if escrito {
			escritos++
		}
	}

	// Los huérfanos se borran enteros: eran generados y no tienen nada que
	// conservar.
	for _, dir := range borra {
		path := filepath.Join(base, dir)
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("no se pudo borrar /filtros/%s/: %w", dir, err)
		}
	}

	fmt.Printf("%d combo(s) escritos, %d borrados -> %s\n", escritos, len(borra), base)
	return nil
}

// expectedCombos calcula el mapa slug → contexto de toda la colección.
//
// Una película con dos países y dos géneros participa en todos los cruces de
// cada país con cada género. Cada entrada del mapa viene de al menos una
// película, así que ningún combo generado está vacío. Si dos contextos
// distintos dieran el mismo slug, es un error: las URLs no pueden colisionar.
func expectedCombos(films []film) (map[string]contexto, error) {
	esperados := map[string]contexto{}

	for _, f := range films {
		if f.Draft {
			continue
		}

		decada := ""
		if len(f.Decades) > 0 {
			decada = f.Decades[0]
		}
		paises := f.Paises
		if len(paises) == 0 {
			paises = []string{""}
		}
		generos := f.Generos
		if len(generos) == 0 {
			generos = []string{""}
		}

		for _, pais := range paises {
			for _, genero := range generos {
				for _, ctx := range []contexto{
					{decada, pais, genero},
					{decada, pais, ""},
					{decada, "", genero},
					{"", pais, genero},
				} {
					if ctx.numeroDimensiones() < 2 {
						continue
					}
					if err := addContexto(esperados, ctx); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return esperados, nil
}

// addContexto añade un contexto al mapa, avisando si el slug ya estaba
// ocupado por otro contexto distinto.
func addContexto(esperados map[string]contexto, ctx contexto) error {
	slug := ctx.slug()
	if anterior, ok := esperados[slug]; ok && anterior != ctx {
		return fmt.Errorf("los contextos %q y %q comparten el slug %q; revísalo",
			anterior.titulo(), ctx.titulo(), slug)
	}
	esperados[slug] = ctx
	return nil
}

// writeCombo escribe el index.md de una página combinada. Si ya existe con el
// mismo contenido, no lo toca: así reejecutar `palette combos` sin cambios no
// ensucia el historial de git.
func writeCombo(base, slug string, ctx contexto) (bool, error) {
	fm := filtroFM{
		Title:  ctx.titulo(),
		Decada: ctx.Decada,
		Pais:   ctx.Pais,
		Genero: ctx.Genero,
	}

	yml, err := yaml.Marshal(fm)
	if err != nil {
		return false, fmt.Errorf("no se pudo escribir /filtros/%s/: %w", slug, err)
	}

	contenido := "---\n" + string(yml) + "---\n"
	path := filepath.Join(base, slug, "index.md")

	if actual, err := os.ReadFile(path); err == nil && string(actual) == contenido {
		return false, nil
	}

	dir := filepath.Join(base, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, fmt.Errorf("no se pudo crear /filtros/%s/: %w", slug, err)
	}

	if err := os.WriteFile(path, []byte(contenido), 0o644); err != nil {
		return false, fmt.Errorf("no se pudo escribir %s: %w", path, err)
	}
	return true, nil
}

// checkCombos comprueba que content/filtros/ está al día: todos los combos
// con películas existen, no hay combos huérfanos y el frontmatter de cada
// página coincide con su slug. Devuelve los fallos encontrados.
func checkCombos(root string, films []film) (failures, warnings []string) {
	esperados, err := expectedCombos(films)
	if err != nil {
		return []string{fmt.Sprintf("/filtros/: %v", err)}, nil
	}

	base := filepath.Join(root, filtrosDir)
	entries, err := os.ReadDir(base)
	if err != nil {
		return []string{fmt.Sprintf(
			"falta %s: ejecuta `go run ./tools/palette combos`", filtrosDir)}, nil
	}

	actuales := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			actuales[e.Name()] = true
		}
	}

	orden := func(m map[string]bool) []string {
		var out []string
		for k := range m {
			out = append(out, k)
		}
		sort.Strings(out)
		return out
	}

	for _, slug := range orden(actuales) {
		if _, ok := esperados[slug]; !ok {
			failures = append(failures, fmt.Sprintf(
				"/filtros/%s/ no tiene películas: ejecuta `go run ./tools/palette combos`", slug))
		}
	}

	faltan := map[string]bool{}
	for slug := range esperados {
		if !actuales[slug] {
			faltan[slug] = true
		}
	}
	for _, slug := range orden(faltan) {
		failures = append(failures, fmt.Sprintf(
			"falta /filtros/%s/: ejecuta `go run ./tools/palette combos`", slug))
	}

	// El frontmatter también es generado: si alguien lo toca a mano y años
	// después se regenera, el cambio se pierde sin que nadie lo sepa.
	for slug, esperado := range esperados {
		if !actuales[slug] {
			continue
		}
		path := filepath.Join(base, slug, "index.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("/filtros/%s/: %v", slug, err))
			continue
		}
		body, err := splitFrontMatter(string(raw))
		if err != nil {
			failures = append(failures, fmt.Sprintf("/filtros/%s/: %v", slug, err))
			continue
		}
		var fm filtroFM
		if err := yaml.Unmarshal([]byte(body), &fm); err != nil {
			failures = append(failures, fmt.Sprintf("/filtros/%s/: frontmatter inválido: %v", slug, err))
			continue
		}
		llegado := contexto{fm.Decada, fm.Pais, fm.Genero}
		if llegado != esperado || llegado.slug() != slug {
			failures = append(failures, fmt.Sprintf(
				"/filtros/%s/ no coincide con su frontmatter: ejecuta `go run ./tools/palette combos`", slug))
		}
	}

	return failures, warnings
}