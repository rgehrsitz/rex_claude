package preprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRule_ValidRule(t *testing.T) {
	validRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "age",
                    "value": 30,
                    "operator": "="
                },
                {
                    "fact": "name",
                    "value": "John",
                    "operator": "="
                }
            ]
        },
        "action": {
            "type": "updateStore",
            "target": "name",
            "value": "Hello, John!"
        }
    }`
	rule, err := ParseRule([]byte(validRuleJSON))
	require.NoError(t, err, "Unexpected error")
	assert.NotNil(t, rule, "Expected a rule, got nil")
}

func TestParseRule_InvalidRuleWithMismatchedValueType(t *testing.T) {
	invalidRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "age",
                    "value": "30",
                    "valueType": "int",
                    "operator": "="
                }
            ]
        },
        "action": {
            "type": "notify",
            "target": "name",
            "value": "Hello, John!"
        }
    }`
	_, err := ParseRule([]byte(invalidRuleJSON))
	assert.Error(t, err, "Expected an error, got nil")
}

func TestParseRule_InvalidRuleWithUnsupportedOperation(t *testing.T) {
	invalidRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "name",
                    "value": "John",
                    "operator": "<"
                }
            ]
        },
        "action": {
            "type": "updateStore",
            "target": "name",
            "value": "Hello, John!"
        }
    }`
	_, err := ParseRule([]byte(invalidRuleJSON))
	assert.Error(t, err, "Expected an error, got nil")
}

func TestParseRule_ValidRuleWithNestedConditions(t *testing.T) {
	validNestedRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "age",
                    "value": 30,
                    "operator": "="
                },
                {
                    "any": [
                        {
                            "fact": "city",
                            "value": "New York",
                            "operator": "="
                        },
                        {
                            "fact": "city",
                            "value": "Los Angeles",
                            "operator": "="
                        }
                    ]
                }
            ]
        },
        "action": {
            "type": "updateStore",
            "target": "name",
            "value": "Hello, user from New York or Los Angeles!"
        }
    }`
	rule, err := ParseRule([]byte(validNestedRuleJSON))
	require.NoError(t, err, "Unexpected error")
	assert.NotNil(t, rule, "Expected a rule, got nil")
}

func TestParseRule_ValidRuleWithSupportedOperators(t *testing.T) {
	validOperatorsRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "age",
                    "value": 30,
                    "operator": "greaterThanOrEqual"
                },
                {
                    "fact": "name",
                    "value": "John",
                    "operator": "notEqual"
                },
                {
                    "fact": "isStudent",
                    "value": true,
                    "operator": "equal"
                }
            ]
        },
        "action": {
            "type": "updateStore",
            "target": "name",
            "value": "Hello, adult non-student!"
        }
    }`
	rule, err := ParseRule([]byte(validOperatorsRuleJSON))
	require.NoError(t, err, "Unexpected error")
	assert.NotNil(t, rule, "Expected a rule, got nil")
}

func TestParseRule_InvalidRuleWithMissingRequiredFields(t *testing.T) {
	invalidMissingFieldsRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "age"
                }
            ]
        }
    }`
	_, err := ParseRule([]byte(invalidMissingFieldsRuleJSON))
	assert.Error(t, err, "Expected an error, got nil")
}

func TestParseRule_ValidRuleWithDeeplyNestedConditions(t *testing.T) {
	nestedRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "temperature",
                    "value": 22,
                    "operator": "greaterThanOrEqual"
                },
                {
                    "any": [
                        {
                            "fact": "weather",
                            "value": "rainy",
                            "operator": "equal"
                        },
                        {
                            "all": [
                                {
                                    "fact": "weather",
                                    "value": "cloudy",
                                    "operator": "equal"
                                },
                                {
                                    "fact": "humidity",
                                    "value": 80,
                                    "operator": "greaterThan"
                                }
                            ]
                        }
                    ]
                }
            ]
        },
        "action": {
            "type": "sendAlert",
            "target": "user",
            "value": "Bring an umbrella!"
        }
    }`

	rule, err := ParseRule([]byte(nestedRuleJSON))
	require.NoError(t, err, "Unexpected error parsing rule with deeply nested conditions")
	assert.NotNil(t, rule, "Expected a non-nil rule")
}

func TestParseRule_InvalidRuleWithUnsupportedOperator(t *testing.T) {
	unsupportedOperatorRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "age",
                    "value": 25,
                    "operator": "modulo"
                }
            ]
        },
        "action": {
            "type": "notify",
            "target": "user",
            "value": "Unsupported operator test"
        }
    }`

	_, err := ParseRule([]byte(unsupportedOperatorRuleJSON))
	assert.Error(t, err, "Expected an error due to unsupported operator")
}

func TestParseRule_InvalidRuleMissingFact(t *testing.T) {
	missingFactRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "value": 30,
                    "operator": "equal"
                }
            ]
        },
        "action": {
            "type": "updateStore",
            "target": "userStatus",
            "value": "Active"
        }
    }`

	_, err := ParseRule([]byte(missingFactRuleJSON))
	assert.Error(t, err, "Expected an error due to missing 'fact' in a condition")
}

func TestParseRule_InvalidRuleWithTypeMismatch(t *testing.T) {
	typeMismatchRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "age",
                    "value": "twenty-five",
                    "valueType": "int",
                    "operator": "equal"
                }
            ]
        },
        "action": {
            "type": "adjustStatus",
            "target": "userAge",
            "value": "Invalid age"
        }
    }`

	_, err := ParseRule([]byte(typeMismatchRuleJSON))
	assert.Error(t, err, "Expected an error due to type mismatch between 'valueType' and actual 'value'")
}

func TestParseRule_NumericTypeHandling(t *testing.T) {
	numericTypeRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "temperature",
                    "value": 20.5,
                    "operator": "="
                },
                {
                    "fact": "age",
                    "value": 30,
                    "operator": "="
                }
            ]
        },
        "action": {
            "type": "notify",
            "target": "climateControl",
            "value": "Adjusting temperature for optimal comfort."
        }
    }`

	rule, err := ParseRule([]byte(numericTypeRuleJSON))
	require.NoError(t, err, "Unexpected error parsing rule with numeric values")
	assert.NotNil(t, rule, "Expected a non-nil rule")
	// Additional checks can be performed here to ensure that numeric types are correctly interpreted.
}
func TestParseRule_ComplexNestedConditions(t *testing.T) {
	complexNestedRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "day",
                    "value": "Monday",
                    "operator": "equal"
                },
                {
                    "any": [
                        {
                            "all": [
                                {
                                    "fact": "weather",
                                    "value": "sunny",
                                    "operator": "equal"
                                },
                                {
                                    "fact": "temperature",
                                    "value": 75,
                                    "operator": "greaterThan"
                                }
                            ]
                        },
                        {
                            "fact": "holiday",
                            "value": true,
                            "operator": "equal"
                        }
                    ]
                }
            ]
        },
        "action": {
            "type": "activate",
            "target": "outdoorActivities",
            "value": "Scheduled activities for the day."
        }
    }`

	rule, err := ParseRule([]byte(complexNestedRuleJSON))
	require.NoError(t, err, "Unexpected error parsing rule with complex nested conditions")
	assert.NotNil(t, rule, "Expected a non-nil rule")
}

func TestParseRule_UnsupportedValueType(t *testing.T) {
	unsupportedValueTypeRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "mood",
                    "value": "happy",
                    "valueType": "emoji",
                    "operator": "equal"
                }
            ]
        },
        "action": {
            "type": "adjustLighting",
            "target": "room",
            "value": "Bright and colorful"
        }
    }`

	_, err := ParseRule([]byte(unsupportedValueTypeRuleJSON))
	assert.Error(t, err, "Expected an error due to unsupported ValueType")
}

func TestParseRule_NoConditions(t *testing.T) {
	noConditionsRuleJSON := `{
        "conditions": {
        },
        "action": {
            "type": "logEvent",
            "target": "system",
            "value": "This rule has no conditions."
        }
    }`

	_, err := ParseRule([]byte(noConditionsRuleJSON))
	// Depending on your application's logic, adjust the assertion accordingly.
	assert.Error(t, err, "Expected an error due to no conditions in rule")
	// OR
	// require.NoError(t, err, "Unexpected error parsing rule with no conditions")
}

func TestParseRule_RedundantConditionsInAllBlock(t *testing.T) {
	redundantConditionsRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "temperature",
                    "value": 30,
                    "operator": "greaterThan"
                },
                {
                    "fact": "temperature",
                    "value": 30,
                    "operator": "greaterThan"
                }
            ]
        },
        "action": {
            "type": "adjustThermostat",
            "target": "indoor",
            "value": "decrease"
        }
    }`

	_, err := ParseRule([]byte(redundantConditionsRuleJSON))
	assert.Error(t, err, "Expected an error due to redundant conditions in 'All' block")
}

func TestParseRule_RedundantConditionsInAnyBlock(t *testing.T) {
	redundantConditionsRuleJSON := `{
        "conditions": {
            "any": [
                {
                    "fact": "dayOfWeek",
                    "value": "Saturday",
                    "operator": "equal"
                },
                {
                    "fact": "dayOfWeek",
                    "value": "Saturday",
                    "operator": "equal"
                }
            ]
        },
        "action": {
            "type": "triggerNotification",
            "target": "user",
            "value": "It's the weekend!"
        }
    }`

	_, err := ParseRule([]byte(redundantConditionsRuleJSON))
	assert.Error(t, err, "Expected an error due to redundant conditions in 'Any' block")
}

func TestParseRule_ContradictoryConditionsInAllBlock(t *testing.T) {
	contradictoryConditionsRuleJSON := `{
        "conditions": {
            "all": [
                {
                    "fact": "temperature",
                    "value": 30,
                    "operator": "lessThan"
                },
                {
                    "fact": "temperature",
                    "value": 30,
                    "operator": "greaterThanOrEqual"
                }
            ]
        },
        "action": {
            "type": "adjustThermostat",
            "target": "indoor",
            "value": "increase"
        }
    }`

	_, err := ParseRule([]byte(contradictoryConditionsRuleJSON))
	assert.Error(t, err, "Expected an error due to contradictory conditions in 'All' block")
}

func TestParseRule_ContradictoryConditionsInAnyBlock(t *testing.T) {
	contradictoryConditionsRuleJSON := `{
        "conditions": {
            "any": [
                {
                    "fact": "lightLevel",
                    "value": 50,
                    "operator": "lessThan"
                },
                {
                    "fact": "lightLevel",
                    "value": 50,
                    "operator": "greaterThanOrEqual"
                }
            ]
        },
        "action": {
            "type": "adjustLighting",
            "target": "indoor",
            "value": "increase"
        }
    }`

	_, err := ParseRule([]byte(contradictoryConditionsRuleJSON))
	assert.Error(t, err, "Expected an error due to contradictory conditions in 'Any' block")
}

func TestParseRule_AmbiguousConditionsInAnyBlock(t *testing.T) {
	ambiguousConditionsRuleJSON := `{
        "conditions": {
            "any": [
                {
                    "fact": "temperature",
                    "value": 30,
                    "operator": "greaterThan"
                },
                {
                    "fact": "temperature",
                    "value": 35,
                    "operator": "greaterThan"
                }
            ]
        },
        "action": {
            "type": "adjustThermostat",
            "target": "indoor",
            "value": "decrease"
        }
    }`

	_, err := ParseRule([]byte(ambiguousConditionsRuleJSON))
	assert.Error(t, err, "Expected an error due to ambiguous conditions in 'Any' block")
}
