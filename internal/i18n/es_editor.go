package i18n

func init() {
	add("es", map[string]string{
		"No text editor found: set $EDITOR (for example EDITOR=nano)": "No encontré un editor de texto: define $EDITOR (por ejemplo EDITOR=nano)",
		"Editor %s not found: check $VISUAL and $EDITOR":              "No encontré el editor %s: revisa $VISUAL y $EDITOR",
	})
}
