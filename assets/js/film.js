/*
 * Paleta por fotograma.
 *
 * Las paletas se extraen al medir las imágenes, así que pasar el ratón solo
 * cambia variables CSS: en el navegador no se calcula nada con píxeles.
 * El fotograma i lleva su paleta en data-frame-palette, escrita por
 * partials/frame-grid.html.
 */

(function () {
	"use strict";

	var article = document.querySelector(".film-detail");
	var band = document.querySelector("[data-palette-band]");
	var frames = document.querySelectorAll("[data-frame-palette]");
	var hint = document.querySelector("[data-palette-hint]");

	if (!article || !band || frames.length === 0) return;

	var swatches = band.querySelectorAll(".swatch");
	var filmPalette = (article.style.getPropertyValue("--accent") || "").trim();

	function apply(palette) {
		Array.prototype.forEach.call(swatches, function (swatch, index) {
			var color = palette[index];
			if (!color) return;

			swatch.style.setProperty("--swatch", color);
			// El botón es lo que guarda el valor que copia palette.js.
			swatch.setAttribute("data-hex", color);

			var label = swatch.querySelector(".swatch-hex");
			if (label) label.textContent = color;
		});

		// El acento también mueve bordes y brillos; se mantiene en sintonia.
		var accent = palette[Math.min(3, palette.length - 1)];
		if (accent) article.style.setProperty("--accent", accent);
	}

	function reset() {
		var original = article.getAttribute("data-film-palette");
		if (original) apply(original.split(",").filter(Boolean));
		if (filmPalette) article.style.setProperty("--accent", filmPalette);
	}

	// Se arranca con la paleta del primer fotograma para que la banda
	// coincida con lo que hay en pantalla.
	var initial = frames[0].getAttribute("data-frame-palette");
	if (initial) {
		article.setAttribute("data-film-palette", initial);
		apply(initial.split(",").filter(Boolean));
	}

	Array.prototype.forEach.call(frames, function (frame) {
		function show() {
			var palette = frame.getAttribute("data-frame-palette");
			if (palette) apply(palette.split(",").filter(Boolean));
		}

		frame.addEventListener("pointerenter", show);
		frame.addEventListener("focus", show);
		// En pantallas táctiles no hay hover, así que se reacciona al tocar sin
		// robar el clic que abre el visor.
		frame.addEventListener("pointerdown", function (event) {
			if (event.pointerType !== "mouse") show();
		});
		frame.addEventListener("pointerleave", reset);
		frame.addEventListener("blur", reset);
	});

	if (hint) hint.hidden = false;
})();
