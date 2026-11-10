package bird_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// types.go lists facade enum constants by hand, so a value the spec adds reaches
// only the wire type; a facade enum that names any value must name them all.
func TestFacadeEnumConstantsCoverTheWireEnum(t *testing.T) {
	t.Parallel()

	wire := wireEnumValues(t)
	aliases, facade := facadeEnums(t, wire)

	checked := 0
	for name, wireType := range aliases {
		have, named := facade[name]
		if !named {
			continue
		}
		checked++
		var missing []string
		for v := range wire[wireType] {
			if !have[v] {
				missing = append(missing, strconv.Quote(v))
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			t.Errorf("bird.%s has no constant for %s; add it beside the others in types.go", name, strings.Join(missing, ", "))
		}
	}
	if checked == 0 {
		t.Fatal("found no facade enum constants; the source scan is broken")
	}
}

// wireEnumValues maps each generated wire enum type to its values, leaving out
// the "<nil>" member of a nullable enum, which no caller compares against.
func wireEnumValues(t *testing.T) map[string]map[string]bool {
	t.Helper()
	file := parseFile(t, filepath.Join("internal", "oapi", "oapi.gen.go"))
	values := map[string]map[string]bool{}
	eachConst(file, func(name string, typ ast.Expr, value ast.Expr) {
		ident, ok := typ.(*ast.Ident)
		if !ok {
			return
		}
		if v, ok := stringLit(value); ok && v != "<nil>" {
			if values[ident.Name] == nil {
				values[ident.Name] = map[string]bool{}
			}
			values[ident.Name][v] = true
		}
	})
	return values
}

// Constants are either typed literals (types.go) or re-exports (the generated facades).
func facadeEnums(t *testing.T, wire map[string]map[string]bool) (map[string]string, map[string]map[string]bool) {
	t.Helper()
	wireConst := map[string][2]string{}
	eachConst(parseFile(t, filepath.Join("internal", "oapi", "oapi.gen.go")), func(name string, typ ast.Expr, value ast.Expr) {
		if ident, ok := typ.(*ast.Ident); ok {
			if v, ok := stringLit(value); ok {
				wireConst[name] = [2]string{ident.Name, v}
			}
		}
	})

	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	aliases := map[string]string{}
	facade := map[string]map[string]bool{}
	add := func(typ, value string) {
		if facade[typ] == nil {
			facade[typ] = map[string]bool{}
		}
		facade[typ][value] = true
	}
	for _, src := range sources {
		if strings.HasSuffix(src, "_test.go") {
			continue
		}
		file := parseFile(t, src)
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				if sel, ok := wireSelector(ts.Type); ok && ts.Assign.IsValid() && wire[sel] != nil {
					aliases[ts.Name.Name] = sel
				}
			}
		}
		eachConst(file, func(name string, typ ast.Expr, value ast.Expr) {
			if ident, ok := typ.(*ast.Ident); ok {
				if v, ok := stringLit(value); ok {
					add(ident.Name, v)
				}
				return
			}
			if sel, ok := wireSelector(value); ok {
				if c, ok := wireConst[sel]; ok {
					add(c[0], c[1])
				}
			}
		})
	}
	// A re-exported constant is recorded under its wire type; credit it to
	// every facade alias of that type.
	for name, wireType := range aliases {
		for v := range facade[wireType] {
			add(name, v)
		}
	}
	return aliases, facade
}

func parseFile(t *testing.T, path string) *ast.File {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s (run make -C clients/sdk-go prepare to generate it): %v", path, err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

// eachConst visits each single-value constant; both packages spell type and
// value on every line, so there is no implicit repetition to carry forward.
func eachConst(file *ast.File, visit func(name string, typ, value ast.Expr)) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs := spec.(*ast.ValueSpec)
			if len(vs.Names) != 1 || len(vs.Values) != 1 {
				continue
			}
			visit(vs.Names[0].Name, vs.Type, vs.Values[0])
		}
	}
}

func stringLit(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	v, err := strconv.Unquote(lit.Value)
	return v, err == nil
}

func wireSelector(expr ast.Expr) (string, bool) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "oapi" {
		return "", false
	}
	return sel.Sel.Name, true
}
