// pkg/preprocessor/parser.go

package preprocessor

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"rgehrsitz/rex/internal/rules"
)

// parseAndValidateRules parses a JSON array of rules and validates each rule.
func ParseAndValidateRules(rulesJSON []byte) ([]*rules.Rule, error) {
	var ruleDefs []json.RawMessage
	if err := json.Unmarshal(rulesJSON, &ruleDefs); err != nil {
		return nil, err
	}

	var validatedRules []*rules.Rule
	for _, rJSON := range ruleDefs {
		rule, err := ParseRule(rJSON)
		if err != nil {
			return nil, err
		}
		validatedRules = append(validatedRules, rule)
	}

	return validatedRules, nil
}

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
	if err = validateConditions(&rule.Conditions); err != nil {
		return nil, err
	}

	return &rule, nil
}

// validateConditions recursively validates all conditions in a Conditions struct.
func validateConditions(conditions *rules.Conditions) error {
	for _, cond := range conditions.All {
		if err := validateCondition(&cond); err != nil {
			return err
		}
	}
	for _, cond := range conditions.Any {
		if err := validateCondition(&cond); err != nil {
			return err
		}
	}

	// Check for redundant conditions
	if hasRedundantConditions(conditions.All) {
		return errors.New("redundant conditions found in 'All' block")
	}
	if hasRedundantConditions(conditions.Any) {
		return errors.New("redundant conditions found in 'Any' block")
	}

	// Check for contradictory conditions
	if hasContradictoryConditions(conditions.All) {
		return errors.New("contradictory conditions found in 'All' block")
	}
	if hasContradictoryConditions(conditions.Any) {
		return errors.New("contradictory conditions found in 'Any' block")
	}

	if hasAmbiguousConditions(conditions.Any) {
		return errors.New("ambiguous conditions found in 'Any' block")
	}

	return nil
}

// validateCondition validates a single Condition struct.
func validateCondition(condition *rules.Condition) error {

	// Infer and assign ValueType if not explicitly provided
	if condition.ValueType == "" {
		inferredType := getTypeString(condition.Value)
		condition.ValueType = inferredType

		// Typecast the value based on the inferred type
		switch inferredType {
		case "int":
			floatValue, ok := condition.Value.(float64)
			if !ok {
				return fmt.Errorf("invalid value for int type: %v", condition.Value)
			}
			condition.Value = int64(floatValue)
		case "float":
			// No need to typecast, JSON numbers are already unmarshalled as float64
		case "string", "bool":
			// No need to typecast, JSON strings and bools are already unmarshalled correctly
		default:
			return fmt.Errorf("unsupported value type: %s", inferredType)
		}
	}

	// Skip direct type and operator validation if this condition is just for nesting other conditions
	// fmt.Printf("condition: %+v\n", condition)
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

	// // Recursively validate nested conditions
	// if err := validateNestedConditions(condition.All); err != nil {
	// 	return err
	// }
	// if err := validateNestedConditions(condition.Any); err != nil {
	// 	return err
	// }

	// Check for contradictory conditions
	// if hasContradictoryConditions(condition.All) {
	// 	return errors.New("contradictory conditions found in 'All' block")
	// }
	// if hasContradictoryConditions(condition.Any) {
	// 	return errors.New("contradictory conditions found in 'Any' block")
	// }

	// Check for ambiguous conditions
	// if hasAmbiguousConditions(condition.Any) {
	// 	return errors.New("ambiguous conditions found in 'Any' block")
	// }

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
	switch v := value.(type) {
	case int, int32, int64:
		return "int"
	case float64:
		// Check if the float64 value is an integer
		if float64(int64(v)) == v {
			return "int"
		}
		return "float"
	case string:
		return "string"
	case bool:
		return "bool"
	default:
		// Log or handle the unexpected type accordingly
		return "unknown"
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
		if err := validateCondition(&cond); err != nil {
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

func hasRedundantConditions(conditions []rules.Condition) bool {
	// Check for redundant conditions within the same level of nesting
	for i := 0; i < len(conditions); i++ {
		//print out the conditions
		fmt.Printf("condition: %+v\n", conditions[i])
		for j := i + 1; j < len(conditions); j++ {
			if equalCondition(conditions[i], conditions[j]) {
				return true
			}
		}
	}
	return false
}
func hasContradictoryConditions(conditions []rules.Condition) bool {
	// Check for contradictory conditions within the same level of nesting
	for i := 0; i < len(conditions); i++ {
		for j := i + 1; j < len(conditions); j++ {
			if isContradictory(conditions[i], conditions[j]) {
				return true
			}
		}
	}
	return false
}

func isContradictory(cond1, cond2 rules.Condition) bool {
	// Check if the two conditions have the same fact
	if cond1.Fact != cond2.Fact {
		return false
	}

	// Infer the valueType if not explicitly set
	valueType := cond1.ValueType
	if valueType == "" {
		valueType = getTypeString(cond1.Value)
	}

	// Check if the two conditions have contradictory operators or values
	switch cond1.Operator {
	case "equal":
		if cond2.Operator == "notEqual" && reflect.DeepEqual(cond1.Value, cond2.Value) {
			return true
		}
	case "notEqual":
		if cond2.Operator == "equal" && reflect.DeepEqual(cond1.Value, cond2.Value) {
			return true
		}
	case "lessThan":
		if cond2.Operator == "greaterThanOrEqual" && compareValuesForEquality(cond1.Value, cond2.Value, valueType) {
			return true
		}
	case "lessThanOrEqual":
		if cond2.Operator == "greaterThan" && compareValuesForEquality(cond1.Value, cond2.Value, valueType) {
			return true
		}
	case "greaterThan":
		if cond2.Operator == "lessThanOrEqual" && compareValuesForEquality(cond2.Value, cond1.Value, valueType) {
			return true
		}
	case "greaterThanOrEqual":
		if cond2.Operator == "lessThan" && compareValuesForEquality(cond2.Value, cond1.Value, valueType) {
			return true
		}
	}

	return false
}

func hasAmbiguousConditions(conditions []rules.Condition) bool {
	// Create a map to store the facts and their corresponding conditions
	factConditionsMap := make(map[string][]rules.Condition)

	// Iterate over the conditions and group them by fact
	for _, cond := range conditions {
		factConditionsMap[cond.Fact] = append(factConditionsMap[cond.Fact], cond)
	}

	// Check for ambiguous conditions within each group of conditions with the same fact
	for _, conditionsWithSameFact := range factConditionsMap {
		if len(conditionsWithSameFact) > 1 {
			for i := 0; i < len(conditionsWithSameFact); i++ {
				for j := i + 1; j < len(conditionsWithSameFact); j++ {
					if isAmbiguous(conditionsWithSameFact[i], conditionsWithSameFact[j]) {
						return true
					}
				}
			}
		}
	}

	return false
}

func isAmbiguous(cond1, cond2 rules.Condition) bool {
	// Check if the two conditions have the same fact, operator, and value type
	if cond1.Fact == cond2.Fact && cond1.Operator == cond2.Operator && cond1.ValueType == cond2.ValueType {
		// Check if the two conditions have different values
		if !reflect.DeepEqual(cond1.Value, cond2.Value) {
			return true
		}
	}

	return false
}

func compareValuesForEquality(v1, v2 interface{}, valueType string) bool {
	switch valueType {
	case "int":
		// Assuming all numbers are treated as float64 due to JSON unmarshalling.
		// Convert both to float64 for comparison to handle JSON's default behavior.
		val1, ok1 := v1.(float64)
		val2, ok2 := v2.(float64)
		if !ok1 || !ok2 {
			return false
		}
		return val1 == val2
	case "float":
		val1, ok1 := v1.(float64)
		val2, ok2 := v2.(float64)
		if !ok1 || !ok2 {
			return false
		}
		return val1 == val2
	case "string":
		val1, ok1 := v1.(string)
		val2, ok2 := v2.(string)
		if !ok1 || !ok2 {
			return false
		}
		return val1 == val2
	case "bool":
		val1, ok1 := v1.(bool)
		val2, ok2 := v2.(bool)
		if !ok1 || !ok2 {
			return false
		}
		return val1 == val2
	default:
		return false
	}
}
