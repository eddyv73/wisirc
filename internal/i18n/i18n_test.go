package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestTranslateAndFallback(t *testing.T) {
	defer SetLang("en")
	add("es", map[string]string{"Hello %s": "Hola %s"})
	if got := T("Hello %s", "ana"); got != "Hello ana" {
		t.Fatalf("en: %q", got)
	}
	if !SetLang("es") || Lang() != "es" {
		t.Fatal("SetLang es")
	}
	if got := T("Hello %s", "ana"); got != "Hola ana" {
		t.Fatalf("es: %q", got)
	}
	if got := T("Untranslated"); got != "Untranslated" {
		t.Fatalf("fallback: %q", got)
	}
	if SetLang("xx") || Lang() != "es" {
		t.Fatal("unknown language accepted")
	}
	delete(catalogs["es"], "Hello %s")
}

func TestDetect(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	cases := []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{"LANG": "es_MX.UTF-8"}, "es"},
		{map[string]string{"LANG": "en_US.UTF-8"}, "en"},
		{map[string]string{"LANG": "fr_FR.UTF-8"}, "en"},
		{map[string]string{"LC_ALL": "es_ES.UTF-8", "LANG": "en_US"}, "es"},
		{map[string]string{"LC_MESSAGES": "es", "LANG": "en_US"}, "es"},
		{map[string]string{}, "en"},
	}
	for _, c := range cases {
		if got := Detect(env(c.env)); got != c.want {
			t.Errorf("%v: got %s", c.env, got)
		}
	}
}

func TestLanguages(t *testing.T) {
	ls := Languages()
	if len(ls) < 2 || ls[0].Code != "en" || ls[1].Code != "es" || ls[1].Name != "Español" {
		t.Fatalf("%+v", ls)
	}
}

// TestCatalogsComplete scans the module for i18n.T("literal", ...) calls and
// checks every message has a translation with the same format verbs, and
// that no catalog entry is left unused.
func TestCatalogsComplete(t *testing.T) {
	root, _ := filepath.Abs("../..")
	used := map[string]string{}
	fset := token.NewFileSet()
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		// skip hidden folders (VCS, editor and tool state) and build output
		if d.IsDir() && p != root && (strings.HasPrefix(d.Name(), ".") || d.Name() == "dist") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "T" {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); !ok || x.Name != "i18n" {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok {
				t.Errorf("%s: i18n.T needs a string literal", fset.Position(call.Pos()))
				return true
			}
			s, _ := strconv.Unquote(lit.Value)
			used[s] = fset.Position(call.Pos()).String()
			return true
		})
		return nil
	})
	for _, l := range Languages()[1:] {
		cat := catalogs[l.Code]
		for msg, pos := range used {
			tr, ok := cat[msg]
			if !ok {
				t.Errorf("%s: missing %s translation for %q", pos, l.Code, msg)
				continue
			}
			if verbs(msg) != verbs(tr) {
				t.Errorf("%s: %q → %q changes the format verbs", l.Code, msg, tr)
			}
		}
		for msg := range cat {
			if _, ok := used[msg]; !ok {
				t.Errorf("%s: unused catalog entry %q", l.Code, msg)
			}
		}
	}
}

func verbs(s string) string {
	var out []string
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '%' {
			out = append(out, s[i:i+2])
			i++
		}
	}
	return strings.Join(out, " ")
}
