package i18n

func init() {
	add("es", map[string]string{
		"start with a copy of this wisirc or Halloy config (file or directory)":                               "arranca con una copia de este config de wisirc o de Halloy (archivo o carpeta)",
		"%s already exists; wisirc never replaces a config: move it away or choose another one with --config": "%s ya existe; wisirc nunca reemplaza un config: muévelo o elige otro con --config",
		"Copied %d server(s) from the Halloy config %s":                                                       "Copié %d servidor(es) del config de Halloy %s",
		"Copied the wisirc config from %s (%d server(s))":                                                     "Copié el config de wisirc de %s (%d servidor(es))",
	})
}
