package i18n

func init() {
	add("es", map[string]string{
		"Edit…":                   "Editar…",
		"Enable":                  "Activar",
		"Disable":                 "Desactivar",
		"disabled":                "desactivado",
		"Undo edit":               "Deshacer edición",
		"%s enabled: connecting…": "%s activado: conectando…",
		"%s disabled: it won't connect at startup":          "%s desactivado: no se conectará al iniciar",
		"%s is not saved: save it first":                    "%s no está guardado: guárdalo primero",
		"Could not save the disabled servers: %s":           "No se pudieron guardar los servidores desactivados: %s",
		"%s is in config.toml: rename it there":             "%s está en config.toml: cámbiale el nombre ahí",
		"%s uses config.toml again":                         "%s vuelve a usar config.toml",
		"Undo the changes to %s and use config.toml again?": "¿Deshacer los cambios de %s y volver a usar config.toml?",
		"Yes, undo  (Y)":                                    "Sí, deshacer  (S)",
		" %s is in config.toml: your changes are saved to servers.toml, config.toml stays as it is":           " %s está en config.toml: tus cambios se guardan en servers.toml y config.toml no se toca",
		" ● connected  ◌ connecting  ○ disconnected  ⊘ disabled":                                              " ● conectado  ◌ conectando  ○ desconectado  ⊘ desactivado",
		" servers.toml is written by wisirc · config.toml is edited by hand · ✎ edited here":                  " servers.toml lo guarda wisirc · config.toml se edita a mano · ✎ editado aquí",
		" ↑/↓ choose · Enter edit · n new · c connect · x enable/disable · d delete · Esc close":              " ↑/↓ elegir · Enter editar · n nuevo · c conectar · x activar/desactivar · d borrar · Esc cerrar",
		"Enter/click go · c connect/disconnect · e edit · x enable/disable · right click menu · F2/Esc close": "Enter/clic ir · c conectar/desconectar · e editar · x activar/desactivar · clic derecho menú · F2/Esc cerrar",
	})
}
