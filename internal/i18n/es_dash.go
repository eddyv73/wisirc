package i18n

func init() {
	add("es", map[string]string{
		"Dashboard":                   "Panel",
		"connected":                   "conectado",
		"connecting…":                 "conectando…",
		"disconnected":                "desconectado",
		"no lag samples yet":          "sin mediciones de latencia aún",
		"measuring…":                  "midiendo…",
		"no channels":                 "sin canales",
		"%d new":                      "%d sin leer",
		"1 mention":                   "1 mención",
		"%d mentions":                 "%d menciones",
		"all read":                    "todo leído",
		"servers: %d · connected: %d": "servidores: %d · conectados: %d",
		"average %dms":                "promedio %dms",
		"No servers yet: add one in ≡ → Servers": "Aún no hay servidores: agrega uno en ≡ → Servidores",
	})
}
