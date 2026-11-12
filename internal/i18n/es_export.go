package i18n

func init() {
	add("es", map[string]string{
		"/export halloy [file]": "/exportar halloy [archivo]",
		"  /export halloy [file]   your servers as a Halloy config":                                               "  /exportar halloy [archivo]   tus servidores como config de Halloy",
		"there are no servers to export":                                                                          "no hay servidores para exportar",
		"Exported %d server(s) to %s":                                                                             "Exportados %d servidor(es) a %s",
		"%s already exists; pick another file: /export halloy <file>":                                             "%s ya existe; elige otro archivo: /exportar halloy <archivo>",
		"It holds your passwords in plain text: keep it private. Copy its [servers.*] into Halloy's config.toml.": "Lleva tus contraseñas en texto plano: guárdalo en privado. Copia sus [servers.*] al config.toml de Halloy.",
	})
}
