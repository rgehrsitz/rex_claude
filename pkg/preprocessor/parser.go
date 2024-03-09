// pkg/preprocessor/parser.go

package preprocessor

import (
	"encoding/json"
	"errors"
	"fmt"
	"rgehrsitz/rex/pkg/rules"
)

func ParseRule(ruleJSON []byte) (*rules.Rule, error) {
	fmt.Printf("Original JSON: %s\n", string(ruleJSON))

	var rule rules.Rule
	err := json.Unmarshal(ruleJSON, &rule)
	if err != nil {
		return nil, err
	}

	// Debug: Print the parsed rule before validation
	fmt.Printf("Parsed Rule: %+v\n", rule)

	// Validate the rule
	err = ValidateRule(&rule)
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func ValidateRule(rule *rules.Rule) error {
	fmt.Printf("Unmarshaled Rule: %+v\n", *rule)

	// Validate rule name
	if rule.Name == "" {
		return errors.New("rule name cannot be empty")
	}

	// Validate rule conditions
	if len(rule.Conditions.All) == 0 && len(rule.Conditions.Any) == 0 {
		return errors.New("rule must have at least one condition")
	}

	// Validate condition operators
	for _, condition := range rule.Conditions.All {
		if !isValidOperator(condition.Operator) {
			return errors.New("invalid operator in condition")
		}
	}
	for _, condition := range rule.Conditions.Any {
		if !isValidOperator(condition.Operator) {
			return errors.New("invalid operator in condition")
		}
	}

	// Validate event type
	if rule.Event.EventType == "" {
		return errors.New("event type cannot be empty")
	}

	// Validate produced and consumed facts
	for _, fact := range rule.ProducedFacts {
		if fact == "" {
			return errors.New("produced fact cannot be empty")
		}
	}
	for _, fact := range rule.ConsumedFacts {
		if fact == "" {
			return errors.New("consumed fact cannot be empty")
		}
	}

	// Validate event actions
	for _, action := range rule.Event.Actions {
		if action.Type == "" {
			return errors.New("action type cannot be empty")
		}
		if action.Target == "" {
			return errors.New("action target cannot be empty")
		}
	}

	// Validate top-level conditions
	err := validateConditions(rule.Conditions.All)
	if err != nil {
		return err
	}
	err = validateConditions(rule.Conditions.Any)
	if err != nil {
		return err
	}

	// Validate nested conditions
	for _, condition := range rule.Conditions.All {
		err := validateCondition(&condition)
		if err != nil {
			return err
		}
	}
	for _, condition := range rule.Conditions.Any {
		err := validateCondition(&condition)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateConditions(conditions []rules.Condition) error {
	for _, condition := range conditions {
		fmt.Printf("Validating Condition: %+v\n", condition)

		if condition.Fact == "" && len(condition.All) == 0 && len(condition.Any) == 0 {
			return errors.New("condition must have either fact or nested conditions")
		}
		if condition.Fact != "" && !isValidOperator(condition.Operator) {
			return errors.New("invalid operator in condition")
		}
		if len(condition.Any) > 0 {
			err := validateConditions(condition.Any)
			if err != nil {
				return err
			}
		}
		err := validateConditions(condition.All)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateCondition(condition *rules.Condition) error {
	if condition.Fact == "" && len(condition.All) == 0 && len(condition.Any) == 0 {
		return errors.New("condition must have either fact or nested conditions")
	}
	if condition.Fact != "" && !isValidOperator(condition.Operator) {
		return errors.New("invalid operator in condition")
	}
	return nil
}

func isValidOperator(operator string) bool {
	for _, supportedOperator := range rules.SupportedOperators {
		if operator == supportedOperator {
			return true
		}
	}
	return false
}
