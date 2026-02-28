package rules

import (
	"fmt"
	"go/ast"
	"strings"
)

const RuleSensitiveWordsName = "sensitive_words"

var DefaultSensitiveWords = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"api_key",
	"apikey",
	"auth",
	"credential",
	"private_key",
	"access_token",
	"refresh_token",
	"bearer",
	"secret_key",
	"encryption_key",
}

type SensitiveWordsRule struct {
	BaseRule
	words []string
}

func NewSensitiveWordsRule() Rule {
	return &SensitiveWordsRule{
		BaseRule: NewBaseRule(RuleSensitiveWordsName, "Checks that log messages don't contain sensitive variables"),
		words:    DefaultSensitiveWords,
	}
}

func (r *SensitiveWordsRule) Configure(config map[string]any) error {
	if err := r.BaseRule.Configure(config); err != nil {
		return err
	}

	if words, ok := config["words"].([]any); ok && len(words) > 0 {
		r.words = make([]string, len(words))
		for i, w := range words {
			if s, ok := w.(string); ok {
				r.words[i] = s
			}
		}
	}

	return nil
}

func (r *SensitiveWordsRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() {
		return ResultPass()
	}

	if sensitiveVar := r.findSensitiveVar(ctx.MsgExpr); sensitiveVar != "" {
		return ResultFail(fmt.Sprintf("log message contains sensitive variable: %s", sensitiveVar))
	}

	return ResultPass()
}

func (r *SensitiveWordsRule) findSensitiveVar(expr ast.Expr) string {
	var walk func(ast.Expr) string
	walk = func(e ast.Expr) string {
		switch v := e.(type) {
		case *ast.Ident:
			if r.isSensitiveWord(v.Name) {
				return v.Name
			}
		case *ast.SelectorExpr:
			if r.isSensitiveWord(v.Sel.Name) {
				return v.Sel.Name
			}
		case *ast.BinaryExpr:
			if result := walk(v.X); result != "" {
				return result
			}
			return walk(v.Y)
		case *ast.StarExpr:
			return walk(v.X)
		case *ast.UnaryExpr:
			return walk(v.X)
		case *ast.CallExpr:
			for _, arg := range v.Args {
				if result := walk(arg); result != "" {
					return result
				}
			}
			if sel, ok := v.Fun.(*ast.SelectorExpr); ok {
				if r.isSensitiveWord(sel.Sel.Name) {
					return sel.Sel.Name
				}
			}
			if ident, ok := v.Fun.(*ast.Ident); ok {
				if r.isSensitiveWord(ident.Name) {
					return ident.Name
				}
			}
		}
		return ""
	}
	return walk(expr)
}

func (r *SensitiveWordsRule) isSensitiveWord(word string) bool {
	wordLower := strings.ToLower(word)
	for _, w := range r.words {
		if strings.ToLower(w) == wordLower {
			return true
		}
	}
	return false
}
