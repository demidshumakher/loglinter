package rules

import "sync"

var initOnce sync.Once

func Init() {
	initOnce.Do(func() {
		RegisterRule(RuleLowercaseName, NewLowercaseRule)
		RegisterRule(RuleNoSpecialCharsName, NewNoSpecialCharsRule)
		RegisterRule(RuleEnglishOnlyName, NewEnglishOnlyRule)
		RegisterRule(RuleSensitiveWordsName, NewSensitiveWordsRule)
		RegisterRule(RuleCustomPatternsName, NewCustomPatternsRule)
	})
}
