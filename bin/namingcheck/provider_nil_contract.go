package namingcheck

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"
)

func checkProviderNilOptionContract(fn funcDecl) []Violation {
	if fn.Decl.Body == nil || !strings.HasPrefix(fn.Name, "With") || !returnsOptionType(fn.Decl.Type.Results) {
		return nil
	}
	params := providerLikeOptionParams(fn.Decl.Type.Params)
	if len(params) == 0 || !returnedOptionFuncAssignsUnguardedParam(fn.Decl.Body, params) {
		return nil
	}
	return []Violation{violation(fn, "nil provider/function option parameters must not overwrite existing providers")}
}

func providerLikeOptionParams(params *ast.FieldList) map[string]bool {
	out := map[string]bool{}
	if params == nil {
		return out
	}
	for _, param := range params.List {
		for _, name := range param.Names {
			if name == nil || name.Name == "_" {
				continue
			}
			if providerLikeParam(name.Name, param.Type) {
				out[name.Name] = true
			}
		}
	}
	return out
}

func providerLikeParam(name string, typ ast.Expr) bool {
	if !nilableProviderOptionType(typ) {
		return false
	}
	return providerLikeName(name) || providerLikeName(exprString(typ))
}

func nilableProviderOptionType(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.FuncType, *ast.InterfaceType, *ast.ChanType:
		return true
	case *ast.StarExpr:
		return true
	case *ast.Ident:
		return nilableNamedProviderType(e.Name)
	case *ast.SelectorExpr:
		return nilableNamedProviderType(e.Sel.Name)
	default:
		return false
	}
}

func nilableNamedProviderType(name string) bool {
	name = strings.ToLower(name)
	if strings.HasSuffix(name, "func") {
		return true
	}
	keywords := []string{
		"provider",
		"factory",
		"reader",
		"writer",
		"formatter",
		"parser",
		"resolver",
		"matcher",
		"replacer",
		"generator",
		"filter",
		"hook",
		"runner",
		"listener",
		"transport",
	}
	return slices.ContainsFunc(keywords, func(keyword string) bool {
		return strings.Contains(name, keyword)
	})
}

func providerLikeName(name string) bool {
	name = strings.ToLower(name)
	if name == "fn" {
		return true
	}
	keywords := []string{
		"provider",
		"factory",
		"func",
		"reader",
		"writer",
		"clock",
		"runner",
		"lookup",
		"dial",
		"marshal",
		"unmarshal",
		"random",
		"source",
		"transport",
		"stat",
		"mkdir",
		"rename",
		"remove",
		"parse",
		"parser",
		"format",
		"resolver",
		"matcher",
		"replacer",
		"generator",
		"filter",
		"hook",
		"valid",
		"sprint",
		"compile",
		"request",
		"decoder",
		"encoder",
		"exec",
		"finalizer",
		"listener",
		"client",
		"server",
	}
	return slices.ContainsFunc(keywords, func(keyword string) bool {
		return strings.Contains(name, keyword)
	})
}

func returnedOptionFuncAssignsUnguardedParam(body *ast.BlockStmt, params map[string]bool) bool {
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		if found || node == nil {
			return !found
		}
		ret, ok := node.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, result := range ret.Results {
			lit, ok := result.(*ast.FuncLit)
			if ok && lit.Body != nil && blockAssignsUnguardedParam(lit.Body, params, nil) {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func blockAssignsUnguardedParam(body *ast.BlockStmt, params, guarded map[string]bool) bool {
	if body == nil {
		return false
	}
	blockGuarded := cloneStringBoolMap(guarded)
	for _, stmt := range body.List {
		if stmtAssignsUnguardedParam(stmt, params, blockGuarded) {
			return true
		}
		if guardedParam, ok := nilReturnGuard(stmt, params); ok {
			blockGuarded[guardedParam] = true
		}
	}
	return false
}

func stmtAssignsUnguardedParam(stmt ast.Stmt, params, guarded map[string]bool) bool {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		return assignWritesUnguardedProviderParam(s, params, guarded)
	case *ast.ExprStmt:
		return exprUsesUnguardedProviderParam(s.X, params, guarded)
	case *ast.BlockStmt:
		return blockAssignsUnguardedParam(s, params, guarded)
	case *ast.IfStmt:
		thenGuarded := cloneStringBoolMap(guarded)
		for param := range nonNilGuardParams(s.Cond, params) {
			thenGuarded[param] = true
		}
		if blockAssignsUnguardedParam(s.Body, params, thenGuarded) {
			return true
		}
		return elseAssignsUnguardedParam(s.Else, params, guarded)
	case *ast.ForStmt:
		return blockAssignsUnguardedParam(s.Body, params, guarded)
	case *ast.RangeStmt:
		return blockAssignsUnguardedParam(s.Body, params, guarded)
	case *ast.SwitchStmt:
		return caseClausesAssignUnguardedParam(s.Body, params, guarded)
	case *ast.TypeSwitchStmt:
		return caseClausesAssignUnguardedParam(s.Body, params, guarded)
	case *ast.SelectStmt:
		return caseClausesAssignUnguardedParam(s.Body, params, guarded)
	default:
		return false
	}
}

func exprUsesUnguardedProviderParam(expr ast.Expr, params, guarded map[string]bool) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || !selectorCallTarget(call.Fun) {
		return false
	}
	for _, arg := range call.Args {
		ident, ok := arg.(*ast.Ident)
		if ok && params[ident.Name] && !guarded[ident.Name] {
			return true
		}
	}
	return false
}

func elseAssignsUnguardedParam(stmt ast.Stmt, params, guarded map[string]bool) bool {
	switch s := stmt.(type) {
	case nil:
		return false
	case *ast.BlockStmt:
		return blockAssignsUnguardedParam(s, params, guarded)
	case *ast.IfStmt:
		return stmtAssignsUnguardedParam(s, params, guarded)
	default:
		return false
	}
}

func caseClausesAssignUnguardedParam(body *ast.BlockStmt, params, guarded map[string]bool) bool {
	if body == nil {
		return false
	}
	for _, stmt := range body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		if blockAssignsUnguardedParam(&ast.BlockStmt{List: clause.Body}, params, guarded) {
			return true
		}
	}
	return false
}

func assignWritesUnguardedProviderParam(assign *ast.AssignStmt, params, guarded map[string]bool) bool {
	if assign.Tok != token.ASSIGN {
		return false
	}
	for i, rhs := range assign.Rhs {
		ident, ok := rhs.(*ast.Ident)
		if !ok || !params[ident.Name] || guarded[ident.Name] || i >= len(assign.Lhs) {
			continue
		}
		if selectorAssignTarget(assign.Lhs[i]) {
			return true
		}
	}
	return false
}

func selectorAssignTarget(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil {
		return false
	}
	_, ok = sel.X.(*ast.Ident)
	return ok
}

func selectorCallTarget(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil {
		return false
	}
	_, ok = sel.X.(*ast.Ident)
	return ok
}

func nonNilGuardParams(expr ast.Expr, params map[string]bool) map[string]bool {
	out := map[string]bool{}
	e, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return out
	}
	if e.Op == token.LAND {
		for param := range nonNilGuardParams(e.X, params) {
			out[param] = true
		}
		for param := range nonNilGuardParams(e.Y, params) {
			out[param] = true
		}
		return out
	}
	if e.Op != token.NEQ {
		return out
	}
	if param, ok := nilComparisonParam(e.X, e.Y, params); ok {
		out[param] = true
	}
	return out
}

func nilReturnGuard(stmt ast.Stmt, params map[string]bool) (string, bool) {
	ifStmt, ok := stmt.(*ast.IfStmt)
	if !ok || ifStmt.Else != nil || !blockTerminates(ifStmt.Body) {
		return "", false
	}
	binary, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok || binary.Op != token.EQL {
		return "", false
	}
	return nilComparisonParam(binary.X, binary.Y, params)
}

func nilComparisonParam(left, right ast.Expr, params map[string]bool) (string, bool) {
	if param, ok := identParam(left, params); ok && isNilIdent(right) {
		return param, true
	}
	if param, ok := identParam(right, params); ok && isNilIdent(left) {
		return param, true
	}
	return "", false
}

func identParam(expr ast.Expr, params map[string]bool) (string, bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok || !params[ident.Name] {
		return "", false
	}
	return ident.Name, true
}

func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func blockTerminates(body *ast.BlockStmt) bool {
	if body == nil || len(body.List) == 0 {
		return false
	}
	switch body.List[len(body.List)-1].(type) {
	case *ast.ReturnStmt:
		return true
	default:
		return false
	}
}

func cloneStringBoolMap(in map[string]bool) map[string]bool {
	out := map[string]bool{}
	for key, value := range in {
		out[key] = value
	}
	return out
}
