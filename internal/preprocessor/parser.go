package preprocessor

import (
	"encoding/json"
	"fmt"
	"rgehrsitz/rex/internal/rules"

	"github.com/rs/zerolog/log"
)

// ParseRules parses a JSON array of rules.
func ParseRules(rulesJSON []byte, context *rules.CompilationContext) ([]*rules.Rule, error) {
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

// parseRule decodes a single JSON rule into a rules.Rule object.
func parseRule(ruleJSON []byte) (*rules.Rule, error) {
	type tempRule struct {
		Name       string           `json:"name"`
		Priority   int              `json:"priority"`
		Conditions rules.Conditions `json:"conditions"`
		Event      rules.Event      `json:"event"`
	}

	var temp tempRule
	err := json.Unmarshal(ruleJSON, &temp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rule JSON: %w", err)
	}

	consumedFacts := extractConsumedFacts(temp.Conditions)
	producedFacts := extractProducedFacts(temp.Event)

	rule := &rules.Rule{
		Name:          temp.Name,
		Priority:      temp.Priority,
		Conditions:    convertConditions(temp.Conditions),
		Event:         temp.Event,
		ProducedFacts: producedFacts,
		ConsumedFacts: consumedFacts,
	}

	return rule, nil
}

func extractConsumedFacts(conds rules.Conditions) []string {
	factSet := make(map[string]bool)
	var collectFacts func(conditions []rules.Condition)
	collectFacts = func(conditions []rules.Condition) {
		for _, cond := range conditions {
			if cond.Fact != "" {
				factSet[cond.Fact] = true
			}
			collectFacts(cond.All)
			collectFacts(cond.Any)
		}
	}
	collectFacts(conds.All)
	collectFacts(conds.Any)
	facts := make([]string, 0, len(factSet))
	for fact := range factSet {
		facts = append(facts, fact)
	}
	return facts
}

func extractProducedFacts(event rules.Event) []string {
	factSet := make(map[string]bool)
	for _, action := range event.Actions {
		if action.Type == "updateFact" || action.Type == "sendMessage" {
			factSet[action.Target] = true
		}
	}
	facts := make([]string, 0, len(factSet))
	for fact := range factSet {
		facts = append(facts, fact)
	}
	return facts
}

// convertConditions processes conditions and determines the type for each value
func convertConditions(conds rules.Conditions) rules.Conditions {
	for i, cond := range conds.All {
		conds.All[i].Value, conds.All[i].ValueType = determineValueType(cond.Value)
	}
	for i, cond := range conds.Any {
		conds.Any[i].Value, conds.Any[i].ValueType = determineValueType(cond.Value)
	}
	return conds
}

// determineValueType determines the type of the value and returns the value with its type
func determineValueType(v interface{}) (interface{}, string) {
	switch val := v.(type) {
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return int(i), "int"
		}
		if f, err := val.Float64(); err == nil {
			return f, "float"
		}
	case float64:
		if float64(int64(val)) == val {
			return int(val), "int"
		}
		return val, "float"
	case string, bool:
		return val, fmt.Sprintf("%T", val)
	}
	return nil, "unknown"
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

	updateFacts(rule, context)

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

	if !isValidOperator(condition.Operator, condition.ValueType) {
		return fmt.Errorf("invalid operator '%s' in condition %d of rule '%s'", condition.Operator, conditionIndex, ruleName)
	}

	if err := validateConditionValue(condition); err != nil {
		return fmt.Errorf("invalid value in condition %d of rule '%s': %w", conditionIndex, ruleName, err)
	}

	return nil
}

func validateConditionValue(condition rules.Condition) error {
	switch condition.ValueType {
	case "int":
		if _, ok := condition.Value.(int); !ok {
			return fmt.Errorf("expected integer value for operator '%s', got %T", condition.Operator, condition.Value)
		}
	case "float":
		if _, ok := condition.Value.(float64); !ok {
			return fmt.Errorf("expected float value for operator '%s', got %T", condition.Operator, condition.Value)
		}
	case "string":
		if _, ok := condition.Value.(string); !ok {
			return fmt.Errorf("expected string value for operator '%s', got %T", condition.Operator, condition.Value)
		}
	default:
		return fmt.Errorf("unsupported or invalid value type '%s'", condition.ValueType)
	}

	// Additionally, check if the operation is valid for the given value type
	if !isValidOperator(condition.Operator, condition.ValueType) {
		return fmt.Errorf("invalid operator '%s' for value type '%s'", condition.Operator, condition.ValueType)
	}

	return nil
}

func isValidOperator(operator string, valueType string) bool {
	validOperators := map[string][]string{
		"int":    {"equal", "notEqual", "lessThan", "lessThanOrEqual", "greaterThan", "greaterThanOrEqual"},
		"float":  {"equal", "notEqual", "lessThan", "lessThanOrEqual", "greaterThan", "greaterThanOrEqual"},
		"string": {"equal", "notEqual", "contains", "notContains"},
	}

	ops, ok := validOperators[valueType]
	if !ok {
		return false
	}

	for _, op := range ops {
		if operator == op {
			return true
		}
	}
	return false
}

func updateFacts(rule *rules.Rule, context *rules.CompilationContext) {
	log.Debug().Msg("Updating facts for rule: " + rule.Name)
	for _, fact := range rule.ConsumedFacts {
		context.ConsumedFacts[fact] = true
		updateFactIndex(fact, context)
	}
	for _, fact := range rule.ProducedFacts {
		context.ProducedFacts[fact] = true
		updateFactIndex(fact, context)
	}
}

func updateFactIndex(fact string, context *rules.CompilationContext) {
	if _, exists := context.FactIndex[fact]; !exists {
		newIndex := len(context.FactIndex)
		context.FactIndex[fact] = newIndex
		log.Debug().Msgf("Fact '%s' added to index at position %d", fact, newIndex)
	}
}
