package rules

import (
	"slices"
	"unicode"
)

const RuleNoSpecialCharsName = "no_special_chars"

var allowedPunctuation = []rune{'.', ',', ':', ';', '-', '_', '/', '\\', '(', ')', '[', ']', '{', '}', '=', '+', '"', '\'', '<', '>', '?'}

type NoSpecialCharsRule struct {
	BaseRule
}

func NewNoSpecialCharsRule() Rule {
	return &NoSpecialCharsRule{
		BaseRule: NewBaseRule(RuleNoSpecialCharsName, "Checks that log messages don't contain special characters or emojis"),
	}
}

func (r *NoSpecialCharsRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() {
		return ResultPass()
	}

	valid, _ := CheckNoSpecialChars(ctx.Msg)
	if valid {
		return ResultPass()
	}

	cleanedMsg := cleanSpecialChars(ctx.Msg)
	return ResultFailWithSuggestion(
		"log message should not contain special characters or emojis",
		"Remove special characters",
		cleanedMsg,
	)
}

func CheckNoSpecialChars(msg string) (bool, rune) {
	runes := []rune(msg)
	dotCount := 0

	for i, ch := range runes {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || unicode.IsSpace(ch) {
			continue
		}

		if ch == '.' {
			dotCount++
		}

		if isRepeatedChar(runes, i, ch) {
			return false, ch
		}

		if slices.Contains(allowedPunctuation, ch) {
			continue
		}

		if isEmoji(ch) {
			return false, ch
		}

		if isProblematicSpecialChar(ch) {
			return false, ch
		}
	}

	if dotCount >= 3 {
		return false, '.'
	}

	return true, 0
}

func isRepeatedChar(runes []rune, pos int, ch rune) bool {
	if pos > 0 && runes[pos-1] == ch {
		return true
	}
	if pos < len(runes)-1 && runes[pos+1] == ch {
		return true
	}
	return false
}

func isEmoji(ch rune) bool {
	return (ch >= 0x1F600 && ch <= 0x1F64F) || // Emoticon
		(ch >= 0x1F300 && ch <= 0x1F5FF) || // Misc Symbols and Pictographs
		(ch >= 0x1F680 && ch <= 0x1F6FF) || // Transport and Map
		(ch >= 0x1F1E0 && ch <= 0x1F1FF) || // Regional (flags)
		(ch >= 0x2600 && ch <= 0x26FF) || // Misc symbols
		(ch >= 0x2700 && ch <= 0x27BF) || // Dingbats
		(ch >= 0xFE00 && ch <= 0xFE0F) || // Variation Selectors
		(ch >= 0x1F900 && ch <= 0x1F9FF) // Supplemental Symbols and Pictographs
}

func isProblematicSpecialChar(ch rune) bool {
	return slices.Contains([]rune{'@', '#', '$', '%', '^', '&', '*', '|', '`', '~', '!'}, ch)
}

func cleanSpecialChars(msg string) string {
	var result []rune
	for _, ch := range msg {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || unicode.IsSpace(ch) {
			result = append(result, ch)
			continue
		}
		if slices.Contains(allowedPunctuation, ch) {
			result = append(result, ch)
		}
	}
	return string(result)
}
