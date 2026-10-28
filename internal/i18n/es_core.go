package i18n

// Spanish for the core packages: cmd/wisirc, app, irc, state, commands,
// config and halloy.
func init() {
	add("es", map[string]string{
		// cmd/wisirc
		"path to the config file": "ruta del archivo de configuración",
		"print the version":       "muestra la versión",
		"Created an example config at %s. Edit it and run wisirc again.":                                                    "Creé un config de ejemplo en %s. Edítalo y vuelve a ejecutar wisirc.",
		"Found %d new server(s) in your Halloy config: ≡ → Servers → [Import from Halloy] to bring them in with one click.": "Encontré %d servidor(es) nuevos en tu config de Halloy: ≡ → Servidores → [Importar de Halloy] para traerlos con un clic.",

		// irc
		"not connected":              "no conectado",
		"Disconnected":               "Desconectado",
		"Disconnected: %s":           "Desconectado: %s",
		"Disconnected: ping timeout": "Desconectado: ping timeout",
		"Connecting to %s…":          "Conectando a %s…",
		"Connected to %s":            "Conectado a %s",
		"Could not connect: %s":      "No se pudo conectar: %s",
		"SASL failed: %s":            "SASL falló: %s",

		// state
		"%s joined %s":             "%s se unió a %s",
		"%s left %s":               "%s salió de %s",
		"%s kicked %s from %s":     "%s expulsó a %s de %s",
		"%s quit":                  "%s salió",
		"You are now %s":           "Ahora eres %s",
		"%s is now %s":             "%s ahora es %s",
		"%s set mode: %s":          "%s cambió modo: %s",
		"%s changed the topic: %s": "%s cambió el topic: %s",
		"Topic: %s":                "Topic: %s",

		// commands
		"Usage: %s":                     "Uso: %s",
		"You are not in a channel":      "No estás en un canal",
		"Unknown command: /%s":          "Comando desconocido: /%s",
		"/me <action>":                  "/me <acción>",
		"/join #channel[,#other] [key]": "/join #canal[,#otro] [clave]",
		"/part [#channel] [message]":    "/part [#canal] [mensaje]",
		"/msg <nick|#channel> <text>":   "/msg <nick|#canal> <texto>",
		"/query <nick> [text]":          "/query <nick> [texto]",
		"/nick <new>":                   "/nick <nuevo>",
		"/topic [#channel] [text]":      "/topic [#canal] [texto]",
		"/kick <nick> [reason]":         "/kick <nick> [razón]",
		"/mode [target] <modes> [args]": "/mode [objetivo] <modos> [args]",
		"/quote <IRC line>":             "/quote <línea IRC>",

		// app
		"Usage: /connect <name|host> [port] [options]":         "Uso: /connect <nombre|host> [puerto] [opciones]",
		"Usage: -sasl user:password":                           "Uso: -sasl usuario:clave",
		"%s is not a new connection":                           "%s no es una conexión nueva",
		"The server needs a name":                              "El servidor necesita un nombre",
		"%s is in config.toml: edit it with ≡ → Edit config":   "%s está en config.toml: edítalo con ≡ → Editar config",
		"%s is in config.toml: delete it with ≡ → Edit config": "%s está en config.toml: bórralo con ≡ → Editar config",
		"A server named %s already exists":                     "Ya existe un servidor llamado %s",
		"There is no server %s":                                "No existe el servidor %s",
		"I don't know where the config is":                     "no sé dónde está el config",
		"Could not save servers.toml: %s":                      "No se pudo guardar servers.toml: %s",
		"Extra argument: %s":                                   "Argumento de más: %s",
		"Invalid port: %s":                                     "Puerto inválido: %s",
		"Unknown option: %s":                                   "Opción desconocida: %s",
		"Missing value for %s":                                 "Falta el valor de %s",

		// config
		"unknown key ignored: %s":                         "clave desconocida ignorada: %s",
		"%s is also in config.toml (config.toml is used)": "%s también está en config.toml (se usa config.toml)",

		// halloy
		"no server, not imported": "sin server, no se importa",
		"password_file could not be read; set the password in ≡ → Servers":                     "password_file no se pudo leer; pon la clave en ≡ → Servidores",
		"password_command is not run for security; set the password in ≡ → Servers":            "password_command no se ejecuta por seguridad; pon la clave en ≡ → Servidores",
		"nick_password is not supported; use SASL or /msg NickServ identify":                   "nick_password no está soportado; usa SASL o /msg NickServ identify",
		"sasl.external (certificate) is not supported":                                         "sasl.external (certificado) no está soportado",
		"sasl.plain.password_file could not be read; set the password in ≡ → Servers":          "sasl.plain.password_file no se pudo leer; pon la clave en ≡ → Servidores",
		"sasl.plain.password_command is not run for security; set the password in ≡ → Servers": "sasl.plain.password_command no se ejecuta por seguridad; pon la clave en ≡ → Servidores",
		"the Halloy config has no servers":                                                     "la config de Halloy no tiene servidores",
	})
}
