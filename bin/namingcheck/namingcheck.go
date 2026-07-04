// Package namingcheck enforces AST-based naming contracts for knifer-go APIs.
package namingcheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Violation struct {
	Path string
	Line int
	Msg  string
}

func (v Violation) String() string {
	if v.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", v.Path, v.Line, v.Msg)
	}
	return fmt.Sprintf("%s: %s", v.Path, v.Msg)
}

type CheckConfig struct {
	Root string
}

type funcDecl struct {
	Name    string
	Path    string
	Line    int
	Package string
	Decl    *ast.FuncDecl
}

func Check(config CheckConfig) ([]Violation, error) {
	root := config.Root
	if root == "" {
		root = "."
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	var funcs []funcDecl
	testRefs := map[string]map[string]bool{}

	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == ".aiflow" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		rel := slashRel(root, path)
		if strings.HasSuffix(path, "_test.go") {
			collectTestRefs(testRefs, file)
			return nil
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil {
				continue
			}
			pos := fset.Position(fn.Name.Pos())
			funcs = append(funcs, funcDecl{
				Name:    fn.Name.Name,
				Path:    rel,
				Line:    pos.Line,
				Package: file.Name.Name,
				Decl:    fn,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var violations []Violation
	for _, fn := range funcs {
		violations = append(violations, checkEContract(fn)...)
		violations = append(violations, checkMustContract(fn, testRefs)...)
		violations = append(violations, checkWithContract(fn)...)
		violations = append(violations, checkTrustBoundaryNameContract(fn)...)
		violations = append(violations, checkProviderNilOptionContract(fn)...)
		violations = append(violations, checkWriterCloseContract(fn)...)
	}
	slices.SortFunc(violations, func(a, b Violation) int {
		if a.Path != b.Path {
			return strings.Compare(a.Path, b.Path)
		}
		if a.Line != b.Line {
			return a.Line - b.Line
		}
		return strings.Compare(a.Msg, b.Msg)
	})
	return violations, nil
}

func checkEContract(fn funcDecl) []Violation {
	if !strings.HasSuffix(fn.Name, "E") || isErrorObjectEName(fn.Name) {
		return nil
	}
	if returnsError(fn.Decl.Type.Results) {
		return nil
	}
	return []Violation{violation(fn, "functions ending in E must return error")}
}

func checkMustContract(fn funcDecl, testRefs map[string]map[string]bool) []Violation {
	if !strings.HasPrefix(fn.Name, "Must") {
		return nil
	}
	var violations []Violation
	if !containsPanicPath(fn.Decl.Body) {
		violations = append(violations, violation(fn, "Must functions must panic directly or delegate to another Must function"))
	}
	if !hasTestReference(testRefs, fn.Name) {
		violations = append(violations, violation(fn, "Must functions must have a test or example referencing the function"))
	}
	return violations
}

func checkWithContract(fn funcDecl) []Violation {
	if !strings.HasPrefix(fn.Name, "With") || isBuilderWithMethod(fn) || isScopedWithFunction(fn) || isFreeBuilderWithFunction(fn) {
		return nil
	}
	if returnsOptionType(fn.Decl.Type.Results) {
		return nil
	}
	return []Violation{violation(fn, "With functions must return an option type")}
}

func checkTrustBoundaryNameContract(fn funcDecl) []Violation {
	if !containsWord(fn.Name, "Safe") || allowedTrustBoundaryName(fn) {
		return nil
	}
	return []Violation{violation(fn, "Safe names are reserved for trust-boundary APIs")}
}

func returnsError(results *ast.FieldList) bool {
	if results == nil {
		return false
	}
	for _, result := range results.List {
		if exprString(result.Type) == "error" {
			return true
		}
	}
	return false
}

func returnsOptionType(results *ast.FieldList) bool {
	if results == nil || len(results.List) == 0 {
		return false
	}
	for _, result := range results.List {
		if !isOptionTypeName(exprString(result.Type)) {
			return false
		}
	}
	return true
}

func isOptionTypeName(name string) bool {
	name = strings.TrimPrefix(name, "*")
	name = strings.TrimPrefix(name, "[]")
	return strings.HasSuffix(name, "Option")
}

func containsPanicPath(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		if found || node == nil {
			return !found
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name := callName(call.Fun)
		lowerName := strings.ToLower(name)
		if name == "panic" || strings.Contains(lowerName, "panic") || strings.HasPrefix(name, "Must") || strings.Contains(name, ".Must") {
			found = true
			return false
		}
		return true
	})
	return found
}

func collectTestRefs(refs map[string]map[string]bool, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		switch n := node.(type) {
		case *ast.Ident:
			addRef(refs, n.Name, file.Name.Name)
		case *ast.SelectorExpr:
			addRef(refs, n.Sel.Name, file.Name.Name)
		}
		return true
	})
}

func addRef(refs map[string]map[string]bool, name, pkg string) {
	if name == "" {
		return
	}
	if _, ok := refs[name]; !ok {
		refs[name] = map[string]bool{}
	}
	refs[name][pkg] = true
}

func hasTestReference(refs map[string]map[string]bool, name string) bool {
	return len(refs[name]) > 0
}

func isBuilderWithMethod(fn funcDecl) bool {
	return fn.Decl.Recv != nil
}

func isScopedWithFunction(fn funcDecl) bool {
	return fn.Name == "WithScopedGlobalConfig"
}

func isFreeBuilderWithFunction(fn funcDecl) bool {
	params := fn.Decl.Type.Params
	results := fn.Decl.Type.Results
	if params == nil || results == nil || len(params.List) == 0 || len(results.List) != 1 {
		return false
	}
	return exprString(params.List[0].Type) == exprString(results.List[0].Type)
}

func isErrorObjectEName(name string) bool {
	switch name {
	case "LogE", "LogAtE", "LogAtEWithOptions":
		return true
	default:
		return false
	}
}

func allowedTrustBoundaryName(fn funcDecl) bool {
	if strings.Contains(fn.Path, "/testdata/") {
		return true
	}
	exactNames := map[string]bool{
		"ContentLengthSafe":            true,
		"ContentLengthSafeWithOptions": true,
		"DownloadBytesSafeE":           true,
		"DownloadFileSafe":             true,
		"DownloadFileSafeWithOptions":  true,
		"DownloadSafe":                 true,
		"DownloadStringSafeE":          true,
		"GetSafe":                      true,
		"GetStringSafeE":               true,
		"HeadSafe":                     true,
		"IsSafeIdentifier":             true,
		"LoadRemoteSafe":               true,
		"LoadRemoteSafeWithOptions":    true,
		"NewSafeRequest":               true,
		"OpenSafe":                     true,
		"OpenSafeWithOptions":          true,
		"OptionsSafe":                  true,
		"PatchSafe":                    true,
		"PostFormSafeE":                true,
		"PostJSONSafeE":                true,
		"PostSafe":                     true,
		"PostStringSafeE":              true,
		"PutSafe":                      true,
		"SafeDownloadedFilename":       true,
		"SafeJoin":                     true,
		"SafeJoinDownloadPath":         true,
		"CheckedConvert":               true,
	}
	if exactNames[fn.Name] {
		return true
	}
	return strings.HasPrefix(fn.Name, "DeleteSafe")
}

func containsWord(name, word string) bool {
	return strings.Contains(name, word)
}

func exprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprString(e.X)
	case *ast.ArrayType:
		return "[]" + exprString(e.Elt)
	case *ast.SelectorExpr:
		return exprString(e.X) + "." + e.Sel.Name
	case *ast.IndexExpr:
		return exprString(e.X)
	case *ast.IndexListExpr:
		return exprString(e.X)
	default:
		return ""
	}
}

func callName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return callName(e.X) + "." + e.Sel.Name
	case *ast.IndexExpr:
		return callName(e.X)
	case *ast.IndexListExpr:
		return callName(e.X)
	default:
		return ""
	}
}

func violation(fn funcDecl, msg string) Violation {
	return Violation{Path: fn.Path, Line: fn.Line, Msg: fn.Name + ": " + msg}
}

func slashRel(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
