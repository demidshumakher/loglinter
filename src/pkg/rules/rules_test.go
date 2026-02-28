package rules

import (
	"go/ast"
	"testing"
)

func TestCheckLowercase(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantValid bool
		wantFix   string
	}{
		{"valid lowercase", "starting server", true, ""},
		{"invalid uppercase", "Starting server", false, "starting server"},
		{"invalid uppercase long", "Database connection failed", false, "database connection failed"},
		{"empty", "", true, ""},
		{"invalid starts with number", "123 items processed", false, ""},
		{"invalid starts with bracket", "(Starting) server", false, ""},
		{"invalid starts with quote", "\"Starting\" server", false, ""},
		{"valid lowercase after space", " server starting", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, fix := CheckLowercase(tt.msg)
			if valid != tt.wantValid {
				t.Errorf("CheckLowercase(%q) valid = %v, want %v", tt.msg, valid, tt.wantValid)
			}
			if fix != tt.wantFix {
				t.Errorf("CheckLowercase(%q) fix = %q, want %q", tt.msg, fix, tt.wantFix)
			}
		})
	}
}

func TestCheckEnglishOnly(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantValid bool
	}{
		{"valid english", "starting server", true},
		{"valid with numbers", "user 123 logged in", true},
		{"valid with punctuation", "server started on port 8080", true},
		{"valid with emoji", "server started 😀", true},
		{"invalid cyrillic", "запуск сервера", false},
		{"invalid chinese", "启动服务器", false},
		{"invalid arabic", "بدء التشغيل", false},
		{"mixed english and cyrillic", "starting запуск server", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := CheckEnglishOnly(tt.msg)
			if valid != tt.wantValid {
				t.Errorf("CheckEnglishOnly(%q) valid = %v, want %v", tt.msg, valid, tt.wantValid)
			}
		})
	}
}

func TestCheckNoSpecialChars(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantValid bool
	}{
		{"valid simple", "server started", true},
		{"valid with period", "server started.", true},
		{"valid with comma", "server started, listening", true},
		{"valid with colon", "port: 8080", true},
		{"invalid exclamation", "server started!", false},
		{"invalid double exclamation", "connection failed!!", false},
		{"invalid ellipsis", "something went wrong...", false},
		{"invalid at symbol", "user @localhost", false},
		{"invalid hash", "error #404", false},
		{"invalid emoji", "server started 😀", false},
		{"invalid multiple special", "error!!!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := CheckNoSpecialChars(tt.msg)
			if valid != tt.wantValid {
				t.Errorf("CheckNoSpecialChars(%q) valid = %v, want %v", tt.msg, valid, tt.wantValid)
			}
		})
	}
}

func TestSensitiveWordsRule(t *testing.T) {
	rule := NewSensitiveWordsRule().(*SensitiveWordsRule)

	t.Run("default words loaded", func(t *testing.T) {
		if len(rule.words) == 0 {
			t.Error("Expected default words to be loaded")
		}
	})

	t.Run("check with sensitive variable - Ident", func(t *testing.T) {
		rule.words = []string{"password"}
		ctx := &CheckContext{
			MsgExpr: &ast.Ident{Name: "password"},
		}
		result := rule.Check(ctx)
		if result.Passed {
			t.Error("Check() should fail for sensitive variable")
		}
	})

	t.Run("check with sensitive variable in binary op", func(t *testing.T) {
		rule.words = []string{"password"}

		// Create a binary expression: "user password: " + password
		binExpr := &ast.BinaryExpr{
			X:  &ast.BasicLit{Kind: 1, Value: `"user password: "`},
			Op: 8, // ADD
			Y:  &ast.Ident{Name: "password"},
		}
		ctx := &CheckContext{
			MsgExpr: binExpr,
		}
		result := rule.Check(ctx)
		if result.Passed {
			t.Error("Check() should fail for sensitive variable in binary op")
		}
	})

	t.Run("check with SelectorExpr", func(t *testing.T) {
		rule.words = []string{"password"}
		selExpr := &ast.SelectorExpr{
			X:   &ast.Ident{Name: "config"},
			Sel: &ast.Ident{Name: "password"},
		}
		ctx := &CheckContext{
			MsgExpr: selExpr,
		}
		result := rule.Check(ctx)
		if result.Passed {
			t.Error("Check() should fail for sensitive field")
		}
	})

	t.Run("configure with custom words replaces defaults", func(t *testing.T) {
		rule2 := NewSensitiveWordsRule()
		rule2.Configure(map[string]any{
			"words": []any{"custom1", "custom2"},
		})
		sw := rule2.(*SensitiveWordsRule)
		if len(sw.words) != 2 {
			t.Errorf("Expected 2 custom words, got %d", len(sw.words))
		}
	})

	t.Run("non-sensitive variable passes", func(t *testing.T) {
		rule.words = []string{"password"}
		ctx := &CheckContext{
			MsgExpr: &ast.Ident{Name: "username"},
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should pass for non-sensitive variable")
		}
	})
}

func TestBaseRule(t *testing.T) {
	rule := NewBaseRule("test_rule", "Test description")

	t.Run("Name", func(t *testing.T) {
		if rule.Name() != "test_rule" {
			t.Errorf("Name() = %q, want %q", rule.Name(), "test_rule")
		}
	})

	t.Run("Description", func(t *testing.T) {
		if rule.Description() != "Test description" {
			t.Errorf("Description() = %q, want %q", rule.Description(), "Test description")
		}
	})

	t.Run("Enabled", func(t *testing.T) {
		if !rule.Enabled() {
			t.Error("Enabled() returned false")
		}
	})

	t.Run("SetEnabled", func(t *testing.T) {
		rule.SetEnabled(false)
		if rule.Enabled() {
			t.Error("SetEnabled() did not disable rule")
		}
	})
}

func TestResultHelpers(t *testing.T) {
	t.Run("ResultPass", func(t *testing.T) {
		result := ResultPass()
		if !result.Passed {
			t.Error("ResultPass() returned failing result")
		}
	})

	t.Run("ResultFail", func(t *testing.T) {
		result := ResultFail("test message")
		if result.Passed {
			t.Error("ResultFail() returned passing result")
		}
	})

	t.Run("ResultFailWithSuggestion", func(t *testing.T) {
		result := ResultFailWithSuggestion("msg", "fix msg", "new text")
		if result.Passed {
			t.Error("ResultFailWithSuggestion() returned passing result")
		}
		if result.SuggestedFix == nil {
			t.Error("ResultFailWithSuggestion() returned nil SuggestedFix")
		}
	})
}
