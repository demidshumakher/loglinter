package rules

import (
	"fmt"
)

type RuleRegistry struct {
	builders map[string]RuleBuilder
}

func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		builders: make(map[string]RuleBuilder),
	}
}

var globalRegistry = NewRuleRegistry()

func (r *RuleRegistry) Register(name string, builder RuleBuilder) error {
	if name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}
	if _, exists := r.builders[name]; exists {
		return fmt.Errorf("rule %q is already registered", name)
	}

	r.builders[name] = builder
	return nil
}

func (r *RuleRegistry) Get(name string) (Rule, error) {
	builder, exists := r.builders[name]
	if !exists {
		return nil, fmt.Errorf("rule %q not found", name)
	}
	return builder(), nil
}

func (r *RuleRegistry) GetAll() ([]Rule, error) {
	rules := make([]Rule, 0, len(r.builders))
	for _, builder := range r.builders {
		rules = append(rules, builder())
	}
	return rules, nil
}

func RegisterRule(name string, builder RuleBuilder) error {
	return globalRegistry.Register(name, builder)
}

func GetAllRules() ([]Rule, error) {
	return globalRegistry.GetAll()
}
