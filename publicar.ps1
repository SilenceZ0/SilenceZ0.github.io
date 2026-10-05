<#
.SYNOPSIS
	Publica el sitio en el repositorio público de GitHub Pages.

.DESCRIPTION
	Compila el sitio y empuja el resultado a un repositorio que solo contiene
	la salida de `hugo`. El repositorio con el código, las fichas y las capturas
	se queda privado: a GitHub Pages solo va lo que hay en `public/`.

	Es un repositorio de usuario, `SilenceZ0.github.io`, así que el sitio sale
	en la raíz del dominio, sin subdirectorio.

.PARAMETER Repositorio
	Destino, como `propietario/repositorio`. Por defecto
	`SilenceZ0/SilenceZ0.github.io`.

.PARAMETER BaseURL
	URL pública del sitio. Es la que sale en los enlaces canónicos, el RSS y el
	sitemap. Por defecto la del repositorio de usuario.

.PARAMETER SaltarValidacion
	No ejecutar `palette validate` antes de compilar. Está para el bucle de
	pruebas; para publicar, déjalo quieto.

.EXAMPLE
	.\publicar.ps1
	Publica en https://silencez0.github.io/

.EXAMPLE
	.\publicar.ps1 -Repositorio SilenceZ0/prueba -BaseURL https://silencez0.github.io/prueba/
	Publica primero en un repositorio de prueba, para no tocar el bueno.
#>

[CmdletBinding()]
param(
	[string] $Repositorio = 'SilenceZ0/SilenceZ0.github.io',
	[string] $BaseURL = 'https://silencez0.github.io/',
	[switch] $SaltarValidacion
)

$ErrorActionPreference = 'Stop'

# `winget` añade Hugo y Go al PATH del sistema, pero el shell de esta sesión se
# abrió antes. Sin esta línea, `hugo` no se encuentra y el error parece del
# proyecto cuando en realidad es del entorno.
$env:Path = [Environment]::GetEnvironmentVariable('Path', 'Machine') + ';' +
	[Environment]::GetEnvironmentVariable('Path', 'User')

$raiz = $PSScriptRoot
Set-Location $raiz

function Die($mensaje) {
	Write-Host ''
	Write-Host "  $mensaje" -ForegroundColor Red
	Write-Host ''
	exit 1
}

# El repositorio va como `propietario/repositorio`. Sin el propietario, `git push`
# responde "remote: Not Found", que parece un problema de permisos o de red y
# en realidad es que la dirección estaba incompleta. Se comprueba aquí para que
# el error diga la verdad.
if ($Repositorio -notmatch '^[^/]+/[^/]+$') {
	Die "El repositorio tiene que ir como propietario/repositorio, y es '$Repositorio'."
}

Write-Host ''
Write-Host '  Publicando PelículasData' -ForegroundColor Cyan
Write-Host "  origen:  $raiz"
Write-Host "  destino: https://github.com/$Repositorio"
Write-Host "  url:     $BaseURL"
Write-Host ''

# 1. Las dos validaciones del AGENTS.md. Una compilación que avisa está rota
#    aunque genere páginas: los avisos de Hugo suelen ser funciones que se van
#    a dejar de existir.
if (-not $SaltarValidacion) {
	Write-Host '  1/4  Comprobando fichas, enlaces y medidas' -ForegroundColor Cyan
	& go run ./tools/palette validate
	if ($LASTEXITCODE -ne 0) { Die 'Las fichas no pasan la validación. No se publica nada.' }
}

Write-Host '  2/4  Compilando el sitio' -ForegroundColor Cyan

# `--cleanDestinationDir` no es opcional. Sin él Hugo deja en `public/` los
# ficheros de las compilaciones anteriores: en el estado en que se encontró esta
# carpeta había trece versiones del CSS y cinco de intro.js, y a GitHub Pages
# habrían subido todas. Nadie las pide, pero ocupan sitio y confunden.
#
# `--baseURL` va en la línea de comandos y no en hugo.toml a propósito. Así el
# archivo sigue diciendo localhost y las URLs canónicas apuntan a localhost
# cuando se trabaja en local, que es lo que el AGENTS.md quiere.
$salida = & hugo --minify --gc --cleanDestinationDir --baseURL $BaseURL 2>&1
$codigo = $LASTEXITCODE

# Hugo escribe los avisos en la salida estándar, no en la de error, y sale con
# 0 aunque tenga WARN. Una compilación que avisa está rota, así que se mira el
# texto.
$avisos = $salida | Where-Object { $_ -match 'WARN|ERROR' }
if ($codigo -ne 0) {
	$salida | ForEach-Object { Write-Host "      $_" }
	Die "Hugo ha fallado con el código $codigo."
}
if ($avisos) {
	$avisos | ForEach-Object { Write-Host "      $_" -ForegroundColor Yellow }
	Die 'La compilación ha dado avisos. No se publica nada.'
}

# Se busca la línea del recuento y de ella solo se saca el número de páginas.
#
# El patrón tiene que saltar la barra vertical del tableau de Hugo: entre
# "Pages" y el número hay espacios y ese carácter, así que un `\s+` a secas no
# llega al dígito y el recuento salía como desconocido. Esa barra tampoco se
# copia a ningún otro sitio: rompe el análisis de un .ps1 aunque vaya dentro de
# una cadena.
$lineaPaginas = $salida | Where-Object { $_ -match 'Pages.*?\d' } | Select-Object -First 1
$paginas = if ($lineaPaginas -match 'Pages.*?(\d+)') { $matches[1] } else { '?' }
Write-Host "      $paginas páginas"

if (-not (Test-Path 'public\404.html')) {
	Die 'Falta public\404.html. Sin él se vería la página de error del servidor.'
}

# 3. Se copia a una carpeta aparte en vez de usar `public/` directamente como
#    repositorio. El motivo es `public/` está en el .gitignore del proyecto: si
#    se metiera un .git dentro, `git status` del repositorio de código empezaría
#    a pensar que hay un submódulo, y `git clean` podría llevárselo por delante.
$destino = Join-Path ([System.IO.Path]::GetTempPath()) 'publicar-peliculasdata'
if (Test-Path $destino) { Remove-Item $destino -Recurse -Force }
Write-Host '  3/4  Preparando el repositorio de salida' -ForegroundColor Cyan
Copy-Item 'public' $destino -Recurse

# `404.html` en la raíz de un repositorio de GitHub Pages no lo sirve Jekyll,
# que busca archivos que empiezan por guion bajo y se come los directorios. Con
# .nojekyll se publica el sitio tal cual sale de Hugo.
Set-Content -Path (Join-Path $destino '.nojekyll') -Value '' -NoNewline

Push-Location $destino
try {
	git init --quiet --initial-branch=main
	git config user.name 'SilenceZ0'
	git config user.email 'silencez0@users.noreply.github.com'

	# Lo que hay aquí es HTML ya compilado: no se vuelve a compilar ni se
	# vuelve a interpretar. Convertirle los finales de línea solo añade ruido,
	# así que se apaga la conversión en este repositorio.
	git config core.autocrlf false

	git add --all

	# El mensaje lleva el número de páginas, que es lo único que resume esta
	# publicación: como la carpeta se inicializa de cero en cada vuelta, el
	# repositorio público se queda con un único commit, no con un historial. El
	# código está en el repositorio privado, que sí tiene historial.
	#
	# El mensaje va en un archivo y no en el `-m` a propósito. Al pasar un texto
	# con acentos como argumento, Windows PowerShell lo entrega a git en la
	# codificación de la consola y la palabra "páginas" se guarda partida en dos
	# bytes. Y el archivo se escribe con WriteAllText y no con Set-Content porque
	# este último, con -Encoding UTF8, pone una marca de orden de bytes al
	# principio y el mensaje se guardaría con un carácter invisible delante.
	$mensaje = Join-Path $destino 'COMMIT_EDITMSG.txt'
	[System.IO.File]::WriteAllText($mensaje, "Publicar $paginas páginas`n",
		[System.Text.UTF8Encoding]::new($false))
	$ErrorActionPreference = 'Continue'
	git commit --quiet --file $mensaje
	Remove-Item $mensaje -Force

	# --force porque el repositorio es una carpeta de salida, no un proyecto con
	# historial que valga la pena. También hace falta en la primera publicación,
	# si el repositorio se creó en GitHub con un README.
	#
	# Git escribe sus errores en la salida de error, y con
	# `$ErrorActionPreference = 'Stop'` PowerShell los convierte en errores
	# nativos que llenan la consola de texto en rojo y sin explicación. La
	# preferencia ya está bajada arriba; aquí se mira el código de salida, que es
	# lo que dice si la operación ha ido bien.
	git push --force --quiet "https://github.com/$Repositorio.git" main
	if ($LASTEXITCODE -ne 0) {
		Die 'El push ha fallado. Suele ser que el repositorio no existe, o que
falta el permiso, o que hay otro push pendiente por alguien más.'
	}
}
finally {
	Pop-Location
	Remove-Item $destino -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host '  4/4  Publicado' -ForegroundColor Cyan
Write-Host ''
Write-Host "  $BaseURL" -ForegroundColor Green
Write-Host ''
Write-Host '  GitHub Pages tarda hasta un par de minutos en servirlo. Si sale 404,'
Write-Host '  comprueba en Settings -> Pages que la fuente de publicación sea'
Write-Host '  Deploy from a branch, con la rama main.'
Write-Host ''
