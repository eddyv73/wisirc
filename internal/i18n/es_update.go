package i18n

func init() {
	add("es", map[string]string{
		"Update to %s":                                                           "Actualizar a %s",
		"⬆ %s available · /update":                                               "⬆ %s disponible · /update",
		"⬆ %s available · brew upgrade wisirc":                                   "⬆ %s disponible · brew upgrade wisirc",
		"✓ %s: reopen wisirc":                                                    "✓ %s: vuelve a abrir wisirc",
		"Checking for updates…":                                                  "Buscando actualizaciones…",
		"Downloading wisirc %s…":                                                 "Descargando wisirc %s…",
		"Could not check for updates: %s":                                        "No se pudo buscar actualizaciones: %s",
		"Could not update: %s":                                                   "No se pudo actualizar: %s",
		"wisirc %s is up to date (latest release: %s)":                           "wisirc %s está al día (última versión: %s)",
		"wisirc %s installed in %s: reopen wisirc to use it":                     "wisirc %s instalado en %s: vuelve a abrir wisirc para usarlo",
		"wisirc %s is installed: reopen wisirc to use it":                        "wisirc %s ya está instalado: vuelve a abrir wisirc para usarlo",
		"Update checks are off ([updates] check = false in config.toml)":         "La búsqueda de actualizaciones está apagada ([updates] check = false en config.toml)",
		"wisirc was installed with Homebrew: update it with brew upgrade wisirc": "wisirc se instaló con Homebrew: actualízalo con brew upgrade wisirc",
		"The download does not match SHA256SUMS; nothing was changed":            "La descarga no coincide con SHA256SUMS; no se cambió nada",
		"The new binary did not start; nothing was changed":                      "El binario nuevo no arrancó; no se cambió nada",
		"There is no release binary for this system; build it from source":       "No hay binario publicado para este sistema; compílalo desde el código",
		"Cannot write next to %s: %s":                                            "No se puede escribir junto a %s: %s",
	})
}
