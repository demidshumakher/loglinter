package analyzer

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/demidshumakher/loglinter/pkg/rules"
)

const (
	slogPackage = "log/slog"
	zapPackage  = "go.uber.org/zap"
)

var logMethods = map[string][]string{
	slogPackage: {"Debug", "Info", "Warn", "Error", "Log", "LogAttrs"},
	zapPackage:  {"Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal"},
}

func Analyzer(cfg any) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "loglinter",
		Doc:      "Checks log messages for compliance with logging best practices",
		Run:      makeRunFunc(cfg),
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func makeRunFunc(cfg any) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		rules.Init()

		config := parseConfig(cfg)
		allRules := getRules(config)
		executor := newRuleExecutor(allRules, pass)
		analyzeCode(pass, executor)

		return nil, nil
	}
}

func parseConfig(cfg any) rulesConfig {
	result := rulesConfig{
		Rules: make(map[string]ruleConfig),
	}

	if cfg == nil {
		return result
	}

	cfgMap, ok := cfg.(map[string]any)
	if !ok {
		return result
	}

	if rulesCfg, ok := cfgMap["rules"].(map[string]any); ok {
		for ruleName, ruleCfg := range rulesCfg {
			if ruleData, ok := ruleCfg.(map[string]any); ok {
				rc := ruleConfig{Data: make(map[string]any)}
				if enabled, ok := ruleData["enabled"].(bool); ok {
					rc.Enabled = &enabled
				}
				for k, v := range ruleData {
					if k != "enabled" {
						rc.Data[k] = v
					}
				}
				result.Rules[ruleName] = rc
			}
		}
	}

	return result
}

type rulesConfig struct {
	Rules map[string]ruleConfig
}

type ruleConfig struct {
	Enabled *bool
	Data    map[string]any
}

func getRules(cfg rulesConfig) []rules.Rule {
	allRules, _ := rules.GetAllRules()
	enabledRules := make([]rules.Rule, 0, len(allRules))

	for _, rule := range allRules {
		enabled := true
		if rc, exists := cfg.Rules[rule.Name()]; exists {
			if rc.Enabled != nil {
				enabled = *rc.Enabled
			}
			if len(rc.Data) > 0 {
				rule.Configure(rc.Data)
			}
		}
		if enabled {
			enabledRules = append(enabledRules, rule)
		}
	}

	return enabledRules
}

func analyzeCode(pass *analysis.Pass, executor *ruleExecutor) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch node := n.(type) {
		case *ast.CallExpr:
			handleCallExpr(node, executor, pass)
		}
	})
}

func handleCallExpr(node *ast.CallExpr, executor *ruleExecutor, pass *analysis.Pass) {
	sel, ok := node.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	methodName := sel.Sel.Name

	if !isLogMethod(methodName) {
		return
	}

	if !isLogger(pass, sel.X) {
		return
	}

	executor.execute(node)
}

func isLogger(pass *analysis.Pass, expr ast.Expr) bool {
	ident := selToIdent(expr)
	if ident == nil {
		return false
	}

	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	typ := obj.Type()
	if typ == nil {
		return false
	}

	if ptr, ok := typ.(*types.Pointer); ok {
		if named, ok := ptr.Elem().(*types.Named); ok {
			pkg := named.Obj().Pkg()
			if pkg != nil && pkg.Path() == zapPackage {
				return named.Obj().Name() == "Logger"
			}
		}
	}

	if pkgName, ok := obj.(*types.PkgName); ok {
		return pkgName.Imported().Path() == slogPackage
	}

	return false
}

func selToIdent(expr ast.Expr) *ast.Ident {
	switch v := expr.(type) {
	case *ast.Ident:
		return v
	case *ast.SelectorExpr:
		if ident, ok := v.X.(*ast.Ident); ok {
			return ident
		}
	}
	return nil
}

func isLogMethod(methodName string) bool {
	allMethods := []string{}
	for _, methods := range logMethods {
		allMethods = append(allMethods, methods...)
	}
	return slices.Contains(allMethods, methodName)
}

type ruleExecutor struct {
	rules []rules.Rule
	pass  *analysis.Pass
}

func newRuleExecutor(allRules []rules.Rule, pass *analysis.Pass) *ruleExecutor {
	return &ruleExecutor{
		rules: allRules,
		pass:  pass,
	}
}

func (e *ruleExecutor) execute(call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}

	msgExpr := call.Args[0]
	msgValue := extractStringValue(msgExpr)

	ctx := &rules.CheckContext{
		MsgExpr: msgExpr,
		Msg:     msgValue,
	}

	for _, rule := range e.rules {
		if result := rule.Check(ctx); !result.Passed {
			e.reportViolation(msgExpr, result)
		}
	}
}

func isStringLiteral(expr ast.Expr) bool {
	_, ok := expr.(*ast.BasicLit)
	return ok
}

func (e *ruleExecutor) reportViolation(expr ast.Expr, result *rules.RuleResult) {
	diag := analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: result.Message,
	}

	if result.SuggestedFix != nil {
		newText := result.SuggestedFix.NewText
		if isStringLiteral(expr) {
			newText = `"` + newText + `"`
		}
		diag.SuggestedFixes = []analysis.SuggestedFix{
			{
				Message: result.SuggestedFix.Message,
				TextEdits: []analysis.TextEdit{
					{
						Pos:     expr.Pos(),
						End:     expr.End(),
						NewText: []byte(newText),
					},
				},
			},
		}
	}

	e.pass.Report(diag)
}

func getExprText(pass *analysis.Pass, expr ast.Expr) string {
	file := pass.Fset.File(expr.Pos())
	if file == nil {
		return ""
	}

	src, err := pass.ReadFile(file.Name())
	if err != nil {
		return ""
	}

	start := int(expr.Pos()) - file.Base()
	end := int(expr.End()) - file.Base()

	if start >= 0 && end <= len(src) {
		return string(src[start:end])
	}
	return ""
}

func extractStringValue(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch v := expr.(type) {
	case *ast.BasicLit:
		if v.Kind.String() == "STRING" && len(v.Value) >= 2 {
			return v.Value[1 : len(v.Value)-1]
		}
	case *ast.BinaryExpr:
		return extractConcatenatedValue(v)
	}
	return ""
}

func extractConcatenatedValue(expr *ast.BinaryExpr) string {
	var parts []string
	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch v := e.(type) {
		case *ast.BasicLit:
			if v.Kind.String() == "STRING" && len(v.Value) >= 2 {
				parts = append(parts, v.Value[1:len(v.Value)-1])
			}
		case *ast.BinaryExpr:
			walk(v.X)
			walk(v.Y)
		}
	}
	walk(expr)
	return strings.Join(parts, "")
}
