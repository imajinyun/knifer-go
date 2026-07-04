package namingcheck

import (
	"go/ast"
	"strings"
)

func checkWriterCloseContract(fn funcDecl) []Violation {
	if fn.Decl.Body == nil || !writerCloseCheckApplies(fn.Path) {
		return nil
	}
	writerVars := map[string]bool{}
	var violations []Violation
	ast.Inspect(fn.Decl.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.AssignStmt:
			recordWriterAssignments(n, writerVars)
		case *ast.DeferStmt:
			if discardsWriterClose(n.Call, writerVars) {
				violations = append(violations, Violation{
					Path: fn.Path,
					Line: fn.Line,
					Msg:  fn.Name + ": write-path Close errors must be handled",
				})
			}
		}
		return true
	})
	return violations
}

func writerCloseCheckApplies(path string) bool {
	allowed := []string{
		"internal/file/",
		"internal/httpx/http/",
		"internal/httpx/resty/",
		"internal/imgx/",
		"internal/net/",
		"internal/poi/",
		"internal/xml/",
		"internal/zip/",
	}
	for _, prefix := range allowed {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func recordWriterAssignments(assign *ast.AssignStmt, writerVars map[string]bool) {
	for i, rhs := range assign.Rhs {
		call, ok := rhs.(*ast.CallExpr)
		if !ok || !isWriterOpenCall(call) || i >= len(assign.Lhs) {
			continue
		}
		if ident, ok := assign.Lhs[i].(*ast.Ident); ok && ident.Name != "_" {
			writerVars[ident.Name] = true
		}
	}
}

func isWriterOpenCall(call *ast.CallExpr) bool {
	name := callName(call.Fun)
	switch {
	case name == "os.Create", name == "os.OpenFile", name == "os.CreateTemp":
		return true
	case strings.HasSuffix(name, ".openFile"):
		return len(call.Args) >= 3
	case strings.HasSuffix(name, ".createTemp"):
		return true
	default:
		return false
	}
}

func discardsWriterClose(call *ast.CallExpr, writerVars map[string]bool) bool {
	if closeReceiverIsWriter(call, writerVars) {
		return true
	}
	lit, ok := call.Fun.(*ast.FuncLit)
	if !ok || lit.Body == nil {
		return false
	}
	discards := false
	ast.Inspect(lit.Body, func(node ast.Node) bool {
		if discards || node == nil {
			return !discards
		}
		assign, ok := node.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}
		lhs, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhs.Name != "_" {
			return true
		}
		rhs, ok := assign.Rhs[0].(*ast.CallExpr)
		if ok && closeReceiverIsWriter(rhs, writerVars) {
			discards = true
			return false
		}
		return true
	})
	return discards
}

func closeReceiverIsWriter(call *ast.CallExpr, writerVars map[string]bool) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Close" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && writerVars[ident.Name]
}
