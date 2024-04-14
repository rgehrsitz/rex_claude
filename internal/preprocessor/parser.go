package preprocessor

import (
	"encoding/json"
	"fmt"
	"rgehrsitz/rex/internal/rules"

	"github.com/rs/zerolog/log"
)

func ParseRules(rulesJSON []byte, context *rules.CompilationContext) ([]*rules.Rule, error) {
	log.Info().Msg("Started parsing rules...")
	var ruleDefs []json.RawMessage
	if err := json.Unmarshal(rulesJSON, &ruleDefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules JSON: %w", err)
	}

	var parsedRules []*rules.Rule
	for _, rJSON := range ruleDefs {
		rule, err := parseRule(rJSON)
		if err != nil {
			return nil, err
		}
		parsedRules = append(parsedRules, rule)
	}

	return parsedRules, nil
}

func parseRule(ruleJSON []byte) (*rules.Rule, error) {
	var rule rules.Rule
	err := json.Unmarshal(ruleJSON, &rule)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rule JSON: %w", err)
	}
	return &rule, nil
}

func ValidateRules(rules []*rules.Rule, context *rules.CompilationContext) error {
	log.Info().Msg("Started validating rules...")
	for _, rule := range rules {
		if err := validateRule(rule, context); err != nil {
			return err
		}
	}
	return nil
}

func validateRule(rule *rules.Rule, context *rules.CompilationContext) error {
	if len(rule.Conditions.All) == 0 && len(rule.Conditions.Any) == 0 {
		return fmt.Errorf("rule '%s' must have at least one condition", rule.Name)
	}

	if err := validateConditions(rule.Conditions.All, rule.Name); err != nil {
		return err
	}

	if err := validateConditions(rule.Conditions.Any, rule.Name); err != nil {
		return err
	}

	updateConsumedFacts(rule, context)

	return nil
}

func validateConditions(conditions []rules.Condition, ruleName string) error {
	log.Info().Msg("Started validating conditions...")
	for i, condition := range conditions {
		if err := validateCondition(condition, ruleName, i); err != nil {
			return err
		}
	}
	return nil
}

func validateCondition(condition rules.Condition, ruleName string, conditionIndex int) error {
	if condition.Fact == "" {
		return fmt.Errorf("missing 'fact' in condition %d of rule '%s'", conditionIndex, ruleName)
	}

	if !isValidOperator(condition.Operator) {
		return fmt.Errorf("invalid operator '%s' in condition %d of rule '%s'", condition.Operator, conditionIndex, ruleName)
	}

	if err := validateConditionValue(condition); err != nil {
		return fmt.Errorf("invalid value in condition %d of rule '%s': %w", conditionIndex, ruleName, err)
	}

	return nil
}

func validateConditionValue(condition rules.Condition) error {
	switch condition.Operator {
	case "equal", "notEqual", "lessThan", "lessThanOrEqual", "greaterThan", "greaterThanOrEqual":
		if _, ok := condition.Value.(float64); !ok {
			return fmt.Errorf("expected numeric value for operator '%s'", condition.Operator)
		}
	case "contains", "notContains":
		if _, ok := condition.Value.(string); !ok {
			return fmt.Errorf("expected string value for operator '%s'", condition.Operator)
		}
	default:
		return fmt.Errorf("unsupported operator '%s'", condition.Operator)
	}
	return nil
}

func isValidOperator(operator string) bool {
	validOperators := []string{"equal", "notEqual", "lessThan", "lessThanOrEqual", "greaterThan", "greaterThanOrEqual", "contains", "notContains"}
	for _, valid := range validOperators {
		if operator == valid {
			return true
		}
	}
	return false
}

func updateConsumedFacts(rule *rules.Rule, context *rules.CompilationContext) {
	traverseConditions(rule.Conditions.All, context)
	traverseConditions(rule.Conditions.Any, context)
}

func traverseConditions(conditions []rules.Condition, context *rules.CompilationContext) {
	for _, condition := range conditions {
		if condition.Fact != "" {
			context.ConsumedFacts[condition.Fact] = true
		}
		traverseConditions(condition.All, context)
		traverseConditions(condition.Any, context)
	}
}
