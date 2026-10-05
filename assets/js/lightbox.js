/*
 * Visor de fotogramas a pantalla completa.
 *
 * Los botones del markup llevan data-frame-src y data-frame-index, así que el
 * script no necesita saber nada de la película: solo recoge lo que hay.
 */

(function () {
	"use strict";

	var lightbox = document.querySelector("[data-lightbox]");
	var lightboxImg = document.querySelector("[data-lightbox-img]");
	var closeButton = document.querySelector(".lightbox-close");
	var prevButton = document.querySelector("[data-lightbox-prev]");
	var nextButton = document.querySelector("[data-lightbox-next]");
	var currentEl = document.querySelector("[data-lightbox-current]");
	var totalEl = document.querySelector("[data-lightbox-total]");

	if (
		!lightbox ||
		!lightboxImg ||
		!closeButton ||
		!prevButton ||
		!nextButton ||
		!currentEl ||
		!totalEl
	) {
		return;
	}

	var closeButtons = document.querySelectorAll("[data-lightbox-close]");
	var galleryTriggers = document.querySelectorAll(
		"[data-lightbox-gallery] [data-frame-src]"
	);

	var frames = [];
	var currentIndex = 0;
	var isOpen = false;
	var lastTrigger = null;

	function collectFrames() {
		frames = Array.prototype.map
			.call(galleryTriggers, function (el) {
				return el.getAttribute("data-frame-src") || "";
			})
			.filter(Boolean);
	}

	function show(index) {
		if (frames.length === 0) return;
		// El módulo mantiene el índice dentro del rango, también hacia atrás.
		currentIndex = ((index % frames.length) + frames.length) % frames.length;
		lightboxImg.src = frames[currentIndex];
		lightboxImg.alt = "Fotograma " + (currentIndex + 1) + " de " + frames.length;
		currentEl.textContent = String(currentIndex + 1);
	}

	function close() {
		if (!isOpen) return;
		isOpen = false;
		lightbox.setAttribute("aria-hidden", "true");
		// Mantiene el panel fuera del orden de tabulación mientras está
		// cerrado; si no, quien navega con teclado aterriza en un diálogo
		// invisible.
		lightbox.setAttribute("inert", "");
		document.body.style.overflow = "";
		lightboxImg.src = "";
		if (lastTrigger) lastTrigger.focus();
		lastTrigger = null;
	}

	function open(index, trigger) {
		collectFrames();
		totalEl.textContent = String(frames.length);
		show(index);
		isOpen = true;
		lastTrigger = trigger;
		lightbox.removeAttribute("inert");
		lightbox.setAttribute("aria-hidden", "false");
		document.body.style.overflow = "hidden";
		closeButton.focus();
	}

	Array.prototype.forEach.call(galleryTriggers, function (trigger) {
		trigger.addEventListener("click", function () {
			var indexAttr = trigger.getAttribute("data-frame-index");
			open(indexAttr ? parseInt(indexAttr, 10) : 0, trigger);
		});
	});

	Array.prototype.forEach.call(closeButtons, function (btn) {
		btn.addEventListener("click", close);
	});

	prevButton.addEventListener("click", function () {
		show(currentIndex - 1);
	});
	nextButton.addEventListener("click", function () {
		show(currentIndex + 1);
	});

	document.addEventListener("keydown", function (event) {
		if (!isOpen) return;
		if (event.key === "Escape") {
			event.preventDefault();
			close();
		}
		if (event.key === "ArrowRight") {
			event.preventDefault();
			show(currentIndex + 1);
		}
		if (event.key === "ArrowLeft") {
			event.preventDefault();
			show(currentIndex - 1);
		}
	});

	lightbox.addEventListener("click", function (event) {
		if (event.target === lightbox) close();
	});
})();
