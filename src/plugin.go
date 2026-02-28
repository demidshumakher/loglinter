package loglinter

import (
	"golang.org/x/tools/go/analysis"

	"github.com/golangci/plugin-module-register/register"

	"github.com/demidshumakher/loglinter/pkg/analyzer"
)

func init() {
	register.Plugin("loglinter", func(conf any) (register.LinterPlugin, error) {
		return &plugin{config: conf}, nil
	})
}

type plugin struct {
	config any
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		analyzer.Analyzer(p.config),
	}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
