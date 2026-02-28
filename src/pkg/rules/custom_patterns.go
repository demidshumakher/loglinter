package rules

import (
	"fmt"
	"log"
	"regexp"
)

const RuleCustomPatternsName = "custom_patterns"

type CustomPatternsRule struct {
	BaseRule
	patterns      []string
	compiledRegex []*regexp.Regexp
}

func NewCustomPatternsRule() Rule {
	return &CustomPatternsRule{
		BaseRule:      NewBaseRule(RuleCustomPatternsName, "Checks log messages against custom regex patterns"),
		patterns:      []string{},
		compiledRegex: []*regexp.Regexp{},
	}
}

func (r *CustomPatternsRule) Configure(config map[string]any) error {
	if err := r.BaseRule.Configure(config); err != nil {
		return err
	}

	if patterns, ok := config["patterns"].([]any); ok {
		r.patterns = make([]string, len(patterns))
		r.compiledRegex = make([]*regexp.Regexp, 0, len(patterns))

		for i, p := range patterns {
			s, ok := p.(string)
			if !ok {
				log.Fatalf("custom_patterns: pattern at index %d is not a string", i)
			}
			re, err := regexp.Compile(s)
			if err != nil {
				log.Fatalf("custom_patterns: failed to compile pattern %q: %v", s, err)
			}
			r.patterns[i] = s
			r.compiledRegex = append(r.compiledRegex, re)
		}
	}

	return nil
}

func (r *CustomPatternsRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() {
		return ResultPass()
	}

	for _, re := range r.compiledRegex {
		if re.MatchString(ctx.Msg) {
			return ResultFail(fmt.Sprintf("log message matches pattern: %s", re.String()))
		}
	}

	return ResultPass()
}
