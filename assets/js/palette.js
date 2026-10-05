/*
 * Copiar el color de una muestra al portapapeles.
 *
 * Los botones los escribe partials/palette.html y llevan el valor en
 * data-hex, que es también lo que actualiza film.js al pasar el ratón por un
 * fotograma: sin actualizar ese atributo se vería el color nuevo pero se
 * copiaría el anterior.
 */

(function () {
	"use strict";

	var buttons = document.querySelectorAll("[data-hex]");

	Array.prototype.forEach.call(buttons, function (button) {
		button.addEventListener("click", function () {
			var hex = button.getAttribute("data-hex");
			if (!hex) return;

			var label = button.querySelector(".swatch-hex");
			var previous = label ? label.textContent : "";

			function done(text) {
				if (!label) return;
				label.textContent = text;
				window.setTimeout(function () {
					label.textContent = previous;
				}, 1400);
			}

			// navigator.clipboard necesita un contexto seguro, así que se
			// recurre a un textarea oculto cuando no lo hay (un http normal
			// por la red local, por ejemplo).
			if (navigator.clipboard && window.isSecureContext) {
				navigator.clipboard.writeText(hex).then(
					function () {
						done("Copiado");
					},
					function () {
						done("No se pudo copiar");
					}
				);
				return;
			}

			try {
				var area = document.createElement("textarea");
				area.value = hex;
				area.setAttribute("readonly", "");
				area.style.position = "fixed";
				area.style.opacity = "0";
				document.body.appendChild(area);
				area.select();
				document.execCommand("copy");
				area.remove();
				done("Copiado");
			} catch (error) {
				done("No se pudo copiar");
			}
		});
	});
})();
