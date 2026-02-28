package analyzer_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/demidshumakher/loglinter/pkg/analyzer"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analyzer := analyzer.Analyzer(nil)

	analysistest.Run(t, testdata, analyzer, "example")
}
