package bird_test

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// A caller outside this module cannot import internal/oapi, so every wire type
// an exported bird type, field, method or signature hands them needs a bird
// name, and every such enum needs its constants.
func TestFacadeReachesNoUnaliasedWireType(t *testing.T) {
	t.Parallel()

	wire := map[string]ast.Expr{}
	methods := map[string][]ast.Expr{}
	for _, decl := range parseFile(t, filepath.Join("internal", "oapi", "oapi.gen.go")).Decls {
		eachTypeSpec(decl, func(ts *ast.TypeSpec) {
			// An alias to a predeclared type (an ID = string) reaches callers as
			// that type and needs no bird name.
			if ident, ok := ts.Type.(*ast.Ident); ok && ts.Assign.IsValid() && types.Universe.Lookup(ident.Name) != nil {
				return
			}
			wire[ts.Name.Name] = ts.Type
		})
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv != nil && fn.Name.IsExported() {
			recv := fn.Recv.List[0].Type
			if star, ok := recv.(*ast.StarExpr); ok {
				recv = star.X
			}
			if ident, ok := recv.(*ast.Ident); ok {
				methods[ident.Name] = append(methods[ident.Name], ast.Expr(fn.Type))
			}
		}
	}
	enums := wireEnumValues(t)

	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	aliased := map[string]bool{}
	// Event.AsAny returns a webhook payload as any, so the union is a root.
	queue := []string{"WebhookEvent"}
	for _, src := range sources {
		if strings.HasSuffix(src, "_test.go") {
			continue
		}
		for _, decl := range parseFile(t, src).Decls {
			eachTypeSpec(decl, func(ts *ast.TypeSpec) {
				if !ts.Name.IsExported() {
					return
				}
				if sel, ok := wireSelector(ts.Type); ok && ts.Assign.IsValid() {
					aliased[sel] = true
					queue = append(queue, sel)
					return
				}
				queue = append(queue, wireRefs(ts.Type, facadeIdent)...)
			})
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.IsExported() {
				queue = append(queue, wireRefs(fn.Type, facadeIdent)...)
			}
		}
	}

	seen := map[string]bool{}
	leaks := map[string]bool{}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		typ, ok := wire[name]
		if !ok || seen[name] {
			continue
		}
		seen[name] = true
		if !aliased[name] {
			leaks[name] = true
		}
		local := func(n string) bool { return wire[n] != nil }
		queue = append(queue, wireRefs(typ, local)...)
		for _, method := range methods[name] {
			queue = append(queue, wireRefs(method, local)...)
		}
	}

	var missing []string
	for name := range leaks {
		missing = append(missing, name)
	}
	sort.Strings(missing)
	for _, name := range missing {
		t.Errorf("an exported bird type reaches oapi.%s, which has no bird alias; regenerate (make -C clients/sdk-go prepare)", name)
	}
	if len(seen) == 0 {
		t.Fatal("reached no wire types; the source scan is broken")
	}

	aliasesOf, facade := facadeEnums(t, enums)
	named := map[string]bool{}
	for alias, wireType := range aliasesOf {
		if len(facade[alias]) > 0 {
			named[wireType] = true
		}
	}
	for name := range seen {
		if len(enums[name]) > 0 && len(facade[name]) == 0 && !named[name] {
			t.Errorf("an exported bird type reaches enum oapi.%s, which has no bird constants; regenerate (make -C clients/sdk-go prepare)", name)
		}
	}
}

// A bare identifier in the facade names a bird type, never a wire one.
func facadeIdent(string) bool { return false }

func eachTypeSpec(decl ast.Decl, visit func(*ast.TypeSpec)) {
	gen, ok := decl.(*ast.GenDecl)
	if !ok {
		return
	}
	for _, spec := range gen.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok {
			visit(ts)
		}
	}
}

// wireRefs lists the oapi types a type expression names: struct fields and
// interface methods only when exported, every parameter and result of a func.
// Inside internal/oapi a reference is a bare identifier that local reports as
// a declared wire type; in the facade it is an oapi selector.
func wireRefs(expr ast.Expr, local func(string) bool) []string {
	var out []string
	var walk func(ast.Expr)
	fields := func(list *ast.FieldList, exportedOnly bool) {
		if list == nil {
			return
		}
		for _, f := range list.List {
			if exportedOnly && len(f.Names) > 0 && !f.Names[0].IsExported() {
				continue
			}
			walk(f.Type)
		}
	}
	walk = func(e ast.Expr) {
		switch e := e.(type) {
		case *ast.Ident:
			if local(e.Name) {
				out = append(out, e.Name)
			}
		case *ast.SelectorExpr:
			if sel, ok := wireSelector(e); ok {
				out = append(out, sel)
			}
		case *ast.StarExpr:
			walk(e.X)
		case *ast.ParenExpr:
			walk(e.X)
		case *ast.ArrayType:
			walk(e.Elt)
		case *ast.Ellipsis:
			walk(e.Elt)
		case *ast.MapType:
			walk(e.Key)
			walk(e.Value)
		case *ast.ChanType:
			walk(e.Value)
		case *ast.IndexExpr:
			walk(e.X)
			walk(e.Index)
		case *ast.IndexListExpr:
			walk(e.X)
			for _, i := range e.Indices {
				walk(i)
			}
		case *ast.StructType:
			fields(e.Fields, true)
		case *ast.InterfaceType:
			fields(e.Methods, true)
		case *ast.FuncType:
			fields(e.TypeParams, false)
			fields(e.Params, false)
			fields(e.Results, false)
		}
	}
	walk(expr)
	return out
}
