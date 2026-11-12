package i18n

func init() {
	add("es", map[string]string{
		// first-run setup
		"Setup":                       "Configuración inicial",
		"Welcome to wisirc":           "Te damos la bienvenida a wisirc",
		"  ·  step %d of %d":          "  ·  paso %d de %d",
		"Language":                    "Idioma",
		"Who are you on IRC?":         "¿Quién eres en IRC?",
		"Pick a look":                 "Elige un estilo",
		"Where do you want to start?": "¿Dónde quieres empezar?",
		"↑/↓ to choose; the screen changes as you move.":                         "↑/↓ para elegir; la pantalla cambia mientras te mueves.",
		"Your nick is how people see you. The real name is optional and public.": "Tu nick es como te ven los demás. El nombre real es opcional y público.",
		"↑/↓ to preview each theme.":                                             "↑/↓ para ver cada tema.",
		"You can add more servers later from ≡ → Servers.":                       "Puedes agregar más servidores después desde ≡ → Servidores.",
		"Nick":      "Nick",
		"Real name": "Nombre real",
		"Import from Halloy (found on this computer)": "Importar de Halloy (está en este equipo)",
		"Libera.Chat, joining #wisirc":                "Libera.Chat, entrando a #wisirc",
		"My own server (address, ZNC, password…)":     "Mi propio servidor (dirección, ZNC, contraseña…)",
		"None for now": "Ninguno por ahora",
		"   ↑ %d more": "   ↑ %d más",
		"   ↓ %d more": "   ↓ %d más",
		"Next →":       "Siguiente →",
		"← Back":       "← Atrás",
		"Finish":       "Terminar",
		"Skip setup":   "Saltar",
		" Enter next · Shift+Tab back · Esc skip":                                              " Enter siguiente · Shift+Tab atrás · Esc saltar",
		"A nick is one word: letters, digits and - _ [ ] { } | ^ `, not starting with a digit": "Un nick es una sola palabra: letras, números y - _ [ ] { } | ^ `, sin empezar por un número",
		"Could not save the %s: %s":                                                            "No se pudo guardar %s: %s",
		"language":                                                                             "el idioma",
		"theme":                                                                                "el tema",
		"nick":                                                                                 "el nick",
		"Welcome, %s! Change language, theme and servers any time from ≡ (F1).":          "¡Hola, %s! Cambia el idioma, el tema y los servidores cuando quieras desde ≡ (F1).",
		"Setup skipped. Add a server any time: ≡ → Servers, or /connect irc.libera.chat": "Configuración saltada. Agrega un servidor cuando quieras: ≡ → Servidores, o /connect irc.libera.chat",
		"Add a server any time: ≡ → Servers, or /connect irc.libera.chat":                "Agrega un servidor cuando quieras: ≡ → Servidores, o /connect irc.libera.chat",
	})
}
