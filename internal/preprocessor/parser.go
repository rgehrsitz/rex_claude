// pkg/preprocessor/parser.go

package preprocessor

import (
	"encoding/json"
	"errors"
	"fmt"
	"rgehrsitz/rex/internal/rules"
)

// ParseRule parses a JSON byte array into a Rule struct and validates it.
func ParseRule(ruleJSON []byte) (*rules.Rule, error) {
	var rule rules.Rule
	err := json.Unmarshal(ruleJSON, &rule)
	if err != nil {
		return nil, err
	}

	// Debug statement to print out the parsed JSON
	//fmt.Printf("Parsed JSON: %+v\n", rule)

	// Validate that the rule has conditions
	if len(rule.Conditions.All) == 0 && len(rule.Conditions.Any) == 0 {
		return nil, errors.New("a rule must have at least one condition")
	}

	// Validate the conditions of the rule
	if err = validateConditions(rule.Conditions); err != nil {
		return nil, err
	}

	return &rule, nil
}

// validateConditions recursively validates all conditions in a Conditions struct.
func validateConditions(conditions rules.Conditions) error {
	for _, cond := range conditions.All {
		if err := validateCondition(cond); err != nil {
			return err
		}
	}
	for _, cond := range conditions.Any {
		if err := validateCondition(cond); err != nil {
			return err
		}
	}
	return nil
}

// validateCondition validates a single Condition struct.
func validateCondition(condition rules.Condition) error {
	// Skip direct type and operator validation if this condition is just for nesting other conditions
	if condition.Fact == "" && (len(condition.All) > 0 || len(condition.Any) > 0) {
		// Validate nested 'All' conditions
		if err := validateNestedConditions(condition.All); err != nil {
			return err
		}
		// Validate nested 'Any' conditions
		if err := validateNestedConditions(condition.Any); err != nil {
			return err
		}
		// If there are only nested conditions and they are valid, no further checks are needed
		return nil
	}
	// Validate based on the explicit ValueType
	if condition.ValueType != "" {
		expectedType := getTypeString(condition.Value)
		if condition.ValueType != expectedType {
			return fmt.Errorf("ValueType does not match the type of Value: expected %s, got %s", condition.ValueType, expectedType)
		}
	} else {
		// Infer and validate the type if ValueType is not specified
		condition.ValueType = getTypeString(condition.Value)
		if condition.ValueType == "" {
			return errors.New("unsupported or missing type of Value")
		}
	}

	// Check for required 'fact' field
	if condition.Fact == "" && len(condition.All) == 0 && len(condition.Any) == 0 {
		return errors.New("missing 'fact' in condition")
	}

	// Normalize the operator to its canonical form.
	canonicalOperator := NormalizeOperator(condition.Operator)

	// Validate the operation based on the ValueType
	if !isOperatorValidForType(canonicalOperator, condition.ValueType) {
		return fmt.Errorf("unsupported operation '%s' for type '%s'", canonicalOperator, condition.ValueType)
	}

	// Recursively validate nested conditions
	if err := validateNestedConditions(condition.All); err != nil {
		return err
	}
	if err := validateNestedConditions(condition.Any); err != nil {
		return err
	}

	return nil
}

// getTypeString returns the type of the value as a string.
func getTypeString(value interface{}) string {
	switch value.(type) {
	case float64:
		return "float" // Assumes all numbers are parsed as float64 by default; an int can be represented as a float without loss of precision
	case string:
		return "string"
	case bool:
		return "bool"
	default:
		return ""
	}
}

// isOperatorValidForType checks if the operator is valid for the given ValueType.
func isOperatorValidForType(operator, valueType string) bool {
	validOperators := map[string][]string{
		"int":    {"equal", "notEqual", "lessThan", "lessThanOrEqual", "greaterThan", "greaterThanOrEqual"},
		"float":  {"equal", "notEqual", "lessThan", "lessThanOrEqual", "greaterThan", "greaterThanOrEqual"},
		"string": {"equal", "notEqual", "contains", "notContains"},
		"bool":   {"equal", "notEqual"},
	}

	for _, validOp := range validOperators[valueType] {
		if operator == validOp {
			return true
		}
	}
	return false
}

// validateNestedConditions recursively validates a slice of nested conditions.
func validateNestedConditions(conditions []rules.Condition) error {
	for _, cond := range conditions {
		if err := validateCondition(cond); err != nil {
			return err
		}
	}
	return nil
}

// Map aliases to canonical operator names.
var operatorAliases = map[string]string{
	"=":  "equal",
	"!=": "notEqual",
	"<":  "lessThan",
	"<=": "lessThanOrEqual",
	">":  "greaterThan",
	">=": "greaterThanOrEqual",
	// Add other aliases as necessary.
}

// NormalizeOperator converts an operator alias to its canonical form.
func NormalizeOperator(operator string) string {
	if canonical, ok := operatorAliases[operator]; ok {
		return canonical
	}
	return operator
}
