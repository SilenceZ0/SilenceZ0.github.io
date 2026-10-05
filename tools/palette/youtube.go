// youtube.go
//
// Reconocer un enlace de YouTube. Es la misma comprobación que hacía el
// extractor de Astro y sirve para dos cosas: avisar de un enlace roto antes de
// publicar, y dejar constancia del id normalizado que luego incrusta la
// plantilla.

package main

import (
	"net/url"
	"regexp"
	"strings"
)

// idPattern es el formato de un id de YouTube: siempre 11 caracteres de
// [A-Za-z0-9_-].
//
// El largo exacto es lo que evita que un texto cualquiera, como
// "no-es-un-link", se tome por un id válido y produzca un vídeo roto.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// hostsYouTube son los dominios aceptados. Se admite también cualquier
// subdominio, como music.youtube.com, igual que en la versión de Astro.
var hostsYouTube = []string{"youtube.com", "youtube-nocookie.com", "youtu.be"}

// schemePattern detecta si la referencia ya trae protocolo.
var schemePattern = regexp.MustCompile(`(?i)^https?://`)

// youtubeID devuelve el id de 11 caracteres de una referencia de YouTube, o
// "" si no lo es. Acepta el id suelto, watch?v=, youtu.be/, /embed/,
// /shorts/ y /live/, con o sin protocolo.
func youtubeID(ref string) string {
	value := strings.TrimSpace(ref)
	if value == "" {
		return ""
	}

	// Id suelto, del tipo "dQw4w9WgXcQ".
	if idPattern.MatchString(value) {
		return value
	}

	// Se toleran referencias sin protocolo, del tipo "youtu.be/abc123".
	candidate := value
	if !schemePattern.MatchString(value) {
		candidate = "https://" + value
	}

	parsed, err := url.Parse(candidate)
	if err != nil {
		return ""
	}

	host := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
	if !isYouTubeHost(host) {
		return ""
	}

	segments := segmentsOf(parsed.Path)

	if host == "youtu.be" {
		if len(segments) == 0 {
			return ""
		}
		return validID(segments[0])
	}

	// /watch?v=<id>
	if parsed.Path == "/watch" {
		return validID(parsed.Query().Get("v"))
	}

	// /embed/<id>, /shorts/<id>, /live/<id>
	if len(segments) >= 2 {
		switch segments[0] {
		case "embed", "shorts", "live":
			return validID(segments[1])
		}
	}

	return ""
}

func isYouTubeHost(host string) bool {
	for _, supported := range hostsYouTube {
		if host == supported || strings.HasSuffix(host, "."+supported) {
			return true
		}
	}
	return false
}

// segmentsOf parte una ruta en sus segmentos no vacíos, como el filter(Boolean)
// de JavaScript.
func segmentsOf(path string) []string {
	var out []string
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func validID(candidate string) string {
	if idPattern.MatchString(candidate) {
		return candidate
	}
	return ""
}
