// test/preprocessor/preprocessor_test.go

package preprocessor_test

import (
	"rgehrsitz/rex/pkg/preprocessor"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRule_ValidRule(t *testing.T) {
	// Test case 1: Valid rule
	validRuleJSON := `{
        "name": "ExampleRule",
        "priority": 1,
        "conditions": {
            "all": [
                {
                    "fact": "temperature",
                    "operator": "greaterThan",
                    "value": 30
                }
            ]
        },
        "event": {
            "eventType": "TemperatureExceeded",
            "actions": [
                {
                    "type": "updateStore",
                    "target": "alerts",
                    "value": "Temperature exceeded threshold"
                }
            ]
        },
        "producedFacts": ["temperatureAlert"],
        "consumedFacts": ["temperature"]
    }`

	rule, err := preprocessor.ParseRule([]byte(validRuleJSON))
	require.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, "ExampleRule", rule.Name)
	assert.Equal(t, 1, rule.Priority)
	assert.Len(t, rule.Conditions.All, 1)
	assert.Equal(t, "temperature", rule.Conditions.All[0].Fact)
	assert.Equal(t, "greaterThan", rule.Conditions.All[0].Operator)
	assert.Equal(t, 30, rule.Conditions.All[0].Value)
	assert.Equal(t, "TemperatureExceeded", rule.Event.EventType)
	assert.Len(t, rule.Event.Actions, 1)
	assert.Equal(t, "updateStore", rule.Event.Actions[0].Type)
	assert.Equal(t, "alerts", rule.Event.Actions[0].Target)
	assert.Equal(t, "Temperature exceeded threshold", rule.Event.Actions[0].Value)
	assert.Equal(t, []string{"temperatureAlert"}, rule.ProducedFacts)
	assert.Equal(t, []string{"temperature"}, rule.ConsumedFacts)
}

func TestParseRule_InvalidJSON(t *testing.T) {
	// Test case 2: Invalid JSON
	invalidRuleJSON := `{
        "name": "InvalidRule",
        "priority": "high",
        "conditions": {
            "all": [
                {
                    "fact": "temperature",
                    "operator": "greaterThan",
                    "value": 30
                }
            ]
        }
    }`

	_, err := preprocessor.ParseRule([]byte(invalidRuleJSON))
	assert.Error(t, err)
}

func TestParseRule_InvalidRule(t *testing.T) {
	// Test case 3: Invalid rule
	invalidRuleJSON := `{
        "name": "",
        "priority": 1,
        "conditions": {
            "all": [
                {
                    "fact": "temperature",
                    "operator": "invalid",
                    "value": 30
                }
            ]
        },
        "event": {
            "eventType": "",
            "actions": [
                {
                    "type": "",
                    "target": "",
                    "value": ""
                }
            ]
        }
    }`

	_, err := preprocessor.ParseRule([]byte(invalidRuleJSON))
	assert.Error(t, err)
}

func TestParseRule_RuleWithIntAndFloat(t *testing.T) {
	// Test case 4: Rule with integer and float values
	ruleWithIntAndFloatJSON := `{
        "name": "RuleWithIntAndFloat",
        "priority": 2,
        "conditions": {
            "all": [
                {
                    "fact": "age",
                    "operator": "greaterThan",
                    "value": 30
                },
                {
                    "fact": "temperature",
                    "operator": "lessThan",
                    "value": 98.6
                }
            ]
        },
        "event": {
            "eventType": "HighTemperature"
        }
    }`

	rule, err := preprocessor.ParseRule([]byte(ruleWithIntAndFloatJSON))
	require.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, "RuleWithIntAndFloat", rule.Name)
	assert.Equal(t, 2, rule.Priority)
	assert.Len(t, rule.Conditions.All, 2)
	assert.Equal(t, "age", rule.Conditions.All[0].Fact)
	assert.Equal(t, "greaterThan", rule.Conditions.All[0].Operator)
	assert.Equal(t, 30, rule.Conditions.All[0].Value)
	assert.Equal(t, "temperature", rule.Conditions.All[1].Fact)
	assert.Equal(t, "lessThan", rule.Conditions.All[1].Operator)
	assert.Equal(t, 98.6, rule.Conditions.All[1].Value)
	assert.Equal(t, "HighTemperature", rule.Event.EventType)
}

func TestParseRule_RuleWithNestedConditions(t *testing.T) {
	// Test case 5: Rule with nested conditions
	ruleWithNestedConditionsJSON := `{
    "name": "RuleWithNestedConditions",
    "priority": 3,
    "conditions": {
        "all": [
            {
                "fact": "age",
                "operator": "greaterThan",
                "value": 30
            },
            {
                "any": [
                    {
                        "fact": "temperature",
                        "operator": "greaterThan",
                        "value": 98.6
                    },
                    {
                        "fact": "heartRate",
                        "operator": "greaterThan",
                        "value": 100
                    }
                ]
            }
        ]
    },
    "event": {
        "eventType": "HighRisk"
    }
}`

	rule, err := preprocessor.ParseRule([]byte(ruleWithNestedConditionsJSON))
	require.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, "RuleWithNestedConditions", rule.Name)
	assert.Equal(t, 3, rule.Priority)
	assert.Len(t, rule.Conditions.All, 2)
	assert.Equal(t, "age", rule.Conditions.All[0].Fact)
	assert.Equal(t, "greaterThan", rule.Conditions.All[0].Operator)
	assert.Equal(t, 30, rule.Conditions.All[0].Value)
	assert.Len(t, rule.Conditions.All[1].Any, 2)
	assert.Equal(t, "temperature", rule.Conditions.All[1].Any[0].Fact)
	assert.Equal(t, "greaterThan", rule.Conditions.All[1].Any[0].Operator)
	assert.Equal(t, 98.6, rule.Conditions.All[1].Any[0].Value)
	assert.Equal(t, "heartRate", rule.Conditions.All[1].Any[1].Fact)
	assert.Equal(t, "greaterThan", rule.Conditions.All[1].Any[1].Operator)
	assert.Equal(t, 100, rule.Conditions.All[1].Any[1].Value)
	assert.Equal(t, "HighRisk", rule.Event.EventType)
}
