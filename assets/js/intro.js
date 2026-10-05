/*
 * intro.js: la nube de puntos de la pantalla de bienvenida.
 *
 * Es el sketch de ../constelacion/porting a pelo: los mismos puntos que se
 * conectan con los que tienen cerca y que el ratón va arrastrando. La
 * diferencia de fondo es que aquí no hay p5, y no por pereza sino porque el
 * sitio entero lleva 222 líneas de JavaScript y cero dependencias. Meter
 * p5 desde un CDN serían 250 KB y la primera dependencia externa del
 * proyecto, y en una web que se montó sin servidor ni llamadas a nada
 * externo eso no encaja.
 *
 * Lo que se pierde al quitar p5 es noise(), el ruido de Perlin que le da a la
 * nube el aire orgánico. Aquí se sustituye por tres senos cruzados. No es lo
 * mismo, pero a esta escala, con las líneas tan tenues, la diferencia no se
 * aprecia.
 *
 * Tres cosas que el sketch hacía y aquí no se pueden permitir:
 *
 *   - El bucle no paraba nunca. Al cerrar la intro se suelta el
 *     requestAnimationFrame, o la nube se sigue calculando detrás del mural
 *     mientras la persona lee la web.
 *   - Cada clic soltaba veinte partículas sin techo, y como el bucle compara
 *     todos contra todos, a los veinte clics ya son 150.000 comparaciones por
 *     fotograma. Hay un límite, y aquí las partículas no nacen de los clics.
 *   - El lienzo se dibujaba a la resolución real, que en un móvil con
 *     pantalla 3x multiplica el trabajo por nueve. Se limita a 2.
 *
 * Los puntos ya no son blancos: cada uno nace con el acento de una película,
 * que Hugo pasa en data-accentos desde data/palettes.json. Las líneas sí
 * quedan en un gris neutro, y no por austeridad: colorearlas obligaría a
 * poner strokeStyle una vez por pareja, casi diez mil veces por fotograma,
 * y el color ya está en los puntos.
 */

(function () {
	"use strict";

	var raiz = document.documentElement;

	// El script se carga en todas las páginas, pero solo hace nada cuando el
	// script en línea de <head> ha decidido que la intro se muestra. Si este
	// archivo se Adelanta a la decisión, el lienzo se pondría a dibujar
	// detrás de un mural que ya se ve.
	if (raiz.getAttribute("data-intro") !== "abierta") return;

	var intro = document.querySelector(".intro");
	var lienzo = intro.querySelector(".intro-lienzo");
	var boton = intro.querySelector(".intro-boton");
	var contexto = lienzo.getContext("2d");

	var FONDO = "#0f0f15";
	var MAX_PUNTOS = 240;
	var MIN_PUNTOS = 40;
	var POR_PIXEL = 13000;
	var DISTANCIA = 104;
	var ALFA_LINEA = 0.34;
	var RADIO_RATON = 190;
	var ENTRADA = 2.4;
	var ESCALA_MAXIMA = 2;

	var menosMovimiento = window.matchMedia("(prefers-reduced-motion: reduce)");

	var acentos = (lienzo.getAttribute("data-accentos") || "")
		.split(/\s+/)
		.filter(function (hex) {
			return /^#[0-9a-f]{6}$/i.test(hex);
		});

	// data/palettes.json siempre trae los acentos. Este reserva solo cubre el
	// caso de que el JSON se quedara sin la clave, y con un gris medio la
	// nube se vería igual de viva.
	if (!acentos.length) acentos = ["#9a9aa8", "#c0c0c0", "#7890a8"];

	var ancho = 0;
	var alto = 0;
	var escala = 1;
	var puntos = [];
	var ratonX = -9999;
	var ratonY = -9999;
	var costumbre = 0;
	var cuadro = 0;
	var tapados = [];
	var cerrado = false;

	function azar(minimo, maximo) {
		return minimo + Math.random() * (maximo - minimo);
	}

	function crear() {
		return {
			x: azar(0, ancho),
			y: azar(0, alto),
			vx: azar(-0.2, 0.2),
			vy: azar(-0.2, 0.2),
			fase: azar(0, 6.283),
			tono: acentos[puntos.length % acentos.length]
		};
	}

	// Los puntos van con el área de la pantalla y no con un número fijo. El
	// radio de conexión está en píxeles, así que en un monitor grande los mismos
	// 140 puntos quedarían separados y la constelación se vería como motas
	// sueltas en vez de como una red. Con 13000 píxeles por punto la densidad
	// se queda parecida en cualquier pantalla.
	function cuantos() {
		var objetivo = Math.round((ancho * alto) / POR_PIXEL);
		if (objetivo < MIN_PUNTOS) objetivo = MIN_PUNTOS;
		if (objetivo > MAX_PUNTOS) objetivo = MAX_PUNTOS;
		return objetivo;
	}

	function sembrar() {
		var objetivo = cuantos();
		while (puntos.length > objetivo) puntos.pop();
		while (puntos.length < objetivo) puntos.push(crear());
	}

	function medir() {
		escala = Math.min(window.devicePixelRatio || 1, ESCALA_MAXIMA);
		ancho = window.innerWidth;
		alto = window.innerHeight;
		lienzo.width = Math.round(ancho * escala);
		lienzo.height = Math.round(alto * escala);
		contexto.setTransform(escala, 0, 0, escala, 0, 0);
		sembrar();
	}

	function avanzar() {
		for (var i = 0; i < puntos.length; i++) {
			var p = puntos[i];

			// La deriva, con tres senos cruzados en lugar de ruido de Perlin.
			p.fase += 0.004;
			p.vx += Math.cos(p.fase * 1.7 + p.y / 260) * 0.006;
			p.vy += Math.sin(p.fase * 1.3 + p.x / 260) * 0.006;

			// El ratón atrae lo que tiene cerca.
			var dx = ratonX - p.x;
			var dy = ratonY - p.y;
			var lejos = Math.sqrt(dx * dx + dy * dy);
			if (lejos < RADIO_RATON && lejos > 0.01) {
				var fuerza = (1 - lejos / RADIO_RATON) * 0.05;
				p.vx += (dx / lejos) * fuerza;
				p.vy += (dy / lejos) * fuerza;
			}

			p.vx *= 0.985;
			p.vy *= 0.985;

			var vel = Math.sqrt(p.vx * p.vx + p.vy * p.vy);
			if (vel > 1.3) {
				p.vx = (p.vx / vel) * 1.3;
				p.vy = (p.vy / vel) * 1.3;
			}

			p.x += p.vx;
			p.y += p.vy;

			// Wrap-around: la nube nunca se va por un borde.
			if (p.x < 0) p.x += ancho;
			else if (p.x > ancho) p.x -= ancho;
			if (p.y < 0) p.y += alto;
			else if (p.y > alto) p.y -= alto;
		}
	}

	function pintar() {
		contexto.globalAlpha = 1;
		contexto.fillStyle = FONDO;
		contexto.fillRect(0, 0, ancho, alto);

		var entrada = Math.min(1, costumbre / ENTRADA);
		var limite = DISTANCIA * DISTANCIA;

		// Las líneas. Se compara el cuadrado de la distancia porque para el
		// alfa basta con la proporción, y así se ahorra una raíz cuadrada por
		// pareja, que son casi diez mil por fotograma.
		contexto.strokeStyle = "#ffffff";
		contexto.lineWidth = 1;
		for (var i = 0; i < puntos.length; i++) {
			for (var j = i + 1; j < puntos.length; j++) {
				var dx = puntos[i].x - puntos[j].x;
				var dy = puntos[i].y - puntos[j].y;
				var cuadrado = dx * dx + dy * dy;
				if (cuadrado > limite) continue;
				contexto.globalAlpha = ALFA_LINEA * (1 - cuadrado / limite) * entrada;
				contexto.beginPath();
				contexto.moveTo(puntos[i].x, puntos[i].y);
				contexto.lineTo(puntos[j].x, puntos[j].y);
				contexto.stroke();
			}
		}

		// Los puntos, que sí llevan el acento de su película.
		for (var k = 0; k < puntos.length; k++) {
			contexto.globalAlpha = 0.9 * entrada;
			contexto.fillStyle = puntos[k].tono;
			contexto.beginPath();
			contexto.arc(puntos[k].x, puntos[k].y, 1.7, 0, 6.283);
			contexto.fill();
		}

		contexto.globalAlpha = 1;
	}

	function bucle() {
		if (cerrado) return;
		avanzar();
		pintar();
		costumbre += 1 / 60;
		cuadro = window.requestAnimationFrame(bucle);
	}

	function tapar() {
		var hijos = document.body.children;
		for (var i = 0; i < hijos.length; i++) {
			var hijo = hijos[i];
			if (hijo === intro || hijo.contains(intro)) continue;
			hijo.setAttribute("inert", "");
			tapados.push(hijo);
		}
	}

	function destapar() {
		for (var i = 0; i < tapados.length; i++) {
			tapados[i].removeAttribute("inert");
		}
		tapados = [];
	}

	function retirar() {
		intro.remove();
		raiz.removeAttribute("data-intro");
	}

	function cerrar() {
		if (cerrado) return;
		cerrado = true;

		// Lo primero: el bucle. Si la nube siguiera corriendo durante el
		// desvanecido, se notaría el tirón al pararse a mitad.
		window.cancelAnimationFrame(cuadro);

		document.removeEventListener("keydown", alPulsarTecla);
		window.removeEventListener("pointermove", alMover);
		window.removeEventListener("resize", medir);
		menosMovimiento.removeEventListener("change", alCambiaElMovimiento);

		// El contenido se devuelve antes de que termine el desvanecido, para
		// que quien está con el teclado no espere cuatrocientos milisegundos
		// a que la capa se vaya.
		destapar();
		enfocarElContenido();

		intro.classList.add("intro--saliendo");
		window.setTimeout(retirar, 420);
	}

	function enfocarElContenido() {
		var zona = document.querySelector("main");
		if (!zona) return;
		zona.setAttribute("tabindex", "-1");
		zona.focus();
	}

	function alPulsarTecla(evento) {
		if (evento.key !== "Escape") return;
		evento.preventDefault();
		cerrar();
	}

	function alMover(evento) {
		ratonX = evento.clientX;
		ratonY = evento.clientY;
	}

	function alCambiaElMovimiento() {
		if (menosMovimiento.matches) cerrar();
	}

	medir();

	intro.addEventListener("click", cerrar);
	document.addEventListener("keydown", alPulsarTecla);
	window.addEventListener("pointermove", alMover);
	window.addEventListener("resize", medir);
	menosMovimiento.addEventListener("change", alCambiaElMovimiento);

	tapar();
	boton.focus();
	bucle();
})();