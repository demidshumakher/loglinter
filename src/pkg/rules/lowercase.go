package rules

import (
	"unicode"
)

const RuleLowercaseName = "lowercase"

type LowercaseRule struct {
	BaseRule
}

func NewLowercaseRule() Rule {
	return &LowercaseRule{
		BaseRule: NewBaseRule(RuleLowercaseName, "Checks that log messages start with a lowercase letter"),
	}
}

func (r *LowercaseRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() {
		return ResultPass()
	}

	valid, suggestion := CheckLowercase(ctx.Msg)
	if valid {
		return ResultPass()
	}

	return ResultFailWithSuggestion(
		"log message should start with a lowercase letter",
		"Change to lowercase",
		suggestion,
	)
}

func CheckLowercase(msg string) (bool, string) {
	if msg == "" {
		return true, ""
	}

	runes := []rune(msg)
	firstCh := runes[0]

	if !unicode.IsLetter(firstCh) {
		return false, ""
	}

	if unicode.IsUpper(firstCh) {
		runes[0] = unicode.ToLower(firstCh)
		return false, string(runes)
	}

	return true, ""
}
