package bytecode

import (
	"encoding/json"
	"rgehrsitz/rex/internal/preprocessor"
	"rgehrsitz/rex/internal/rules"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileSimpleRule(t *testing.T) {
	// Define the JSON for the simple rule
	ruleJSON := `[
        {
            "name": "SimpleRule",
            "conditions": {
                "all": [
                    {
                        "fact": "temperature",
                        "operator": "greaterThan",
                        "value": 30,
                        "valueType": "int"
                    }
                ],
                "any": []
            },
            "event": {
                "actions": [
                    {
                        "type": "updateFact",
                        "target": "ac_status",
                        "value": true
                    }
                ]
            },
            "producedFacts": ["ac_status"],
            "consumedFacts": ["temperature"]
        }
    ]`

	// // Parse the rule JSON
	// var ruleset []*rules.Rule
	// err := json.Unmarshal([]byte(ruleJSON), &ruleset)
	// require.NoError(t, err, "Failed to parse rule JSON")

	// Initialize the RuleEngineContext
	context := rules.NewRuleEngineContext()

	// Create the compiler instance
	//compiler := NewCompiler(context)

	ruleset, err := preprocessor.ParseRules([]byte(ruleJSON), nil)
	if err != nil {
		t.Fatal(err)
	}

	//After parsing the rules into validatedRules and before compiling:
	for _, rule := range ruleset {
		for _, fact := range rule.ConsumedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex) // Assign a new index
			}
		}
		for _, fact := range rule.ProducedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex) // Assign a new index
			}
		}
	}

	// Compile the ruleset
	bytecode, err := Compile(ruleset, context)
	require.NoError(t, err, "Compilation failed")

	// Assert that the bytecode is not nil or empty
	assert.NotEmpty(t, bytecode, "Compiled bytecode should not be empty")

	// Detailed bytecode assertion
	expectedBytecode := []byte{
		17, 0, // LOAD_FACT "temperature"
		19, 30, 0, 0, 0, // LOAD_CONST_INT 30
		4,        // GT_INT
		26, 5, 0, // JUMP_IF_FALSE 5 bytes ahead (corrected offset)
		28, 1, // UPDATE_FACT "ac_status"
		22, 1, // LOAD_CONST_BOOL true
	}

	assert.Equal(t, expectedBytecode, bytecode, "The generated bytecode does not match the expected sequence")

}

func TestCompileMultipleConditionsRule(t *testing.T) {
	// Define JSON for a rule with multiple conditions
	ruleJSON := `[
		{
			"name": "ComplexRule",
			"conditions": {
				"all": [
					{
						"fact": "temperature",
						"operator": "greaterThan",
						"value": 25,
						"valueType": "int"
					},
					{
						"fact": "humidity",
						"operator": "lessThan",
						"value": 50,
						"valueType": "int"
					}
				],
				"any": []
			},
			"event": {
				"actions": [
					{
						"type": "updateFact",
						"target": "ac_status",
						"value": true
					}
				]
			},
			"producedFacts": ["ac_status"],
			"consumedFacts": ["temperature", "humidity"]
		}
	]`

	// Parse the rule JSON into a ruleset
	var ruleset []*rules.Rule
	err := json.Unmarshal([]byte(ruleJSON), &ruleset)
	require.NoError(t, err, "Failed to parse rule JSON")

	// Initialize the RuleEngineContext
	context := rules.NewRuleEngineContext()

	// Create the compiler instance
	//compiler := NewCompiler(context)

	// Index the facts involved in the rules
	for _, rule := range ruleset {
		for _, fact := range rule.ConsumedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
		for _, fact := range rule.ProducedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
	}

	// Compile the ruleset
	bytecode, err := Compile(ruleset, context)
	require.NoError(t, err, "Compilation failed")

	// Expected bytecode for multiple conditions
	expectedBytecode := []byte{
		17, 0, // LOAD_FACT "temperature"
		19, 25, 0, 0, 0, // LOAD_CONST_INT 25
		4,         // GT_INT
		26, 16, 0, // JUMP_IF_FALSE 16 bytes ahead
		17, 1, // LOAD_FACT "humidity"
		19, 50, 0, 0, 0, // LOAD_CONST_INT 50
		2,        // LT_INT
		26, 5, 0, // JUMP_IF_FALSE 5 bytes ahead
		28, 2, // UPDATE_FACT "ac_status"
		22, 1, // LOAD_CONST_BOOL true
	}

	assert.Equal(t, expectedBytecode, bytecode, "Compiled bytecode does not match the expected sequence")
}

func TestCompileAnyConditionsRule(t *testing.T) {
	// Define JSON for a rule with "any" conditions
	ruleJSON := `[
		{
			"name": "VentilationRule",
			"conditions": {
				"all": [],
				"any": [
					{
						"fact": "temperature",
						"operator": "greaterThan",
						"value": 28,
						"valueType": "int"
					},
					{
						"fact": "humidity",
						"operator": "lessThan",
						"value": 40,
						"valueType": "int"
					}
				]
			},
			"event": {
				"actions": [
					{
						"type": "updateFact",
						"target": "fan_status",
						"value": true
					}
				]
			},
			"producedFacts": ["fan_status"],
			"consumedFacts": ["temperature", "humidity"]
		}
	]`

	// Parse the rule JSON into a ruleset
	var ruleset []*rules.Rule
	err := json.Unmarshal([]byte(ruleJSON), &ruleset)
	require.NoError(t, err, "Failed to parse rule JSON")

	// Initialize the RuleEngineContext
	context := rules.NewRuleEngineContext()

	// Create the compiler instance
	//compiler := NewCompiler(context)

	// Index the facts involved in the rules
	for _, rule := range ruleset {
		for _, fact := range rule.ConsumedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
		for _, fact := range rule.ProducedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
	}

	// Compile the ruleset
	bytecode, err := Compile(ruleset, context)
	require.NoError(t, err, "Compilation failed")

	// Expected bytecode for "any" conditions
	expectedBytecode := []byte{
		17, 0, // LOAD_FACT "temperature"
		19, 28, 0, 0, 0, // LOAD_CONST_INT 28
		4,         // GT_INT
		25, 12, 0, // JUMP_IF_TRUE 12 bytes ahead to action label
		17, 1, // LOAD_FACT "humidity"
		19, 40, 0, 0, 0, // LOAD_CONST_INT 40
		2,        // LT_INT
		26, 5, 0, // JUMP_IF_FALSE 2 bytes ahead to action label
		28, 2, // UPDATE_FACT "fan_status"
		22, 1, // LOAD_CONST_BOOL true
	}

	assert.Equal(t, expectedBytecode, bytecode, "Compiled bytecode does not match the expected sequence")
}

func TestCompileNestedConditionsRule(t *testing.T) {
	// Define JSON for a rule with nested conditions
	ruleJSON := `[
		{
			"name": "NestedConditionsRule",
			"conditions": {
				"all": [
					{
						"fact": "temperature",
						"operator": "greaterThan",
						"value": 25,
						"valueType": "int"
					},
					{
						"any": [
							{
								"fact": "humidity",
								"operator": "lessThan",
								"value": 40,
								"valueType": "int"
							},
							{
								"fact": "room_occupied",
								"operator": "equal",
								"value": true,
								"valueType": "bool"
							}
						]
					}
				]
			},
			"event": {
				"actions": [
					{
						"type": "updateFact",
						"target": "ac_status",
						"value": true
					}
				]
			},
			"producedFacts": ["ac_status"],
			"consumedFacts": ["temperature", "humidity", "room_occupied"]
		}
	]`

	// Parse the rule JSON into a ruleset
	var ruleset []*rules.Rule
	err := json.Unmarshal([]byte(ruleJSON), &ruleset)
	require.NoError(t, err, "Failed to parse rule JSON")

	// Initialize the RuleEngineContext
	context := rules.NewRuleEngineContext()

	// Create the compiler instance
	//compiler := NewCompiler(context)

	// Index the facts involved in the rules
	for _, rule := range ruleset {
		for _, fact := range rule.ConsumedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
		for _, fact := range rule.ProducedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
	}

	// Compile the ruleset
	bytecode, err := Compile(ruleset, context)
	require.NoError(t, err, "Compilation failed")

	// Expected bytecode for nested conditions
	expectedBytecode := []byte{
		17, 0, // LOAD_FACT "temperature"
		19, 25, 0, 0, 0, // LOAD_CONST_INT 25
		4,         // GT_INT
		26, 24, 0, // JUMP_IF_FALSE 24 bytes ahead to end label
		17, 1, // LOAD_FACT "humidity"
		19, 40, 0, 0, 0, // LOAD_CONST_INT 40
		2,        // LT_INT
		25, 9, 0, // JUMP_IF_TRUE 9 bytes ahead to action
		17, 2, // LOAD_FACT "room_occupied"
		22, 1, // LOAD_CONST_BOOL true
		0,        // EQ_BOOL
		26, 5, 0, // JUMP_IF_FALSE 5 bytes ahead to end
		28, 3, // UPDATE_FACT "ac_status"
		22, 1, // LOAD_CONST_BOOL true
	}

	assert.Equal(t, expectedBytecode, bytecode, "Compiled bytecode does not match the expected sequence")
}

func TestCompileMultipleRulesWithMixedConditions(t *testing.T) {
	// Define JSON for multiple rules with mixed conditions
	ruleJSON := `[
		{
			"name": "TemperatureRule",
			"conditions": {
				"all": [
					{
						"fact": "temperature",
						"operator": "greaterThan",
						"value": 30,
						"valueType": "int"
					}
				]
			},
			"event": {
				"actions": [
					{
						"type": "updateFact",
						"target": "ac_status",
						"value": true
					}
				]
			},
			"producedFacts": ["ac_status"],
			"consumedFacts": ["temperature"]
		},
		{
			"name": "HumidityRule",
			"conditions": {
				"any": [
					{
						"fact": "humidity",
						"operator": "lessThan",
						"value": 40,
						"valueType": "int"
					},
					{
						"fact": "room_occupied",
						"operator": "equal",
						"value": true,
						"valueType": "bool"
					}
				]
			},
			"event": {
				"actions": [
					{
						"type": "updateFact",
						"target": "dehumidifier_status",
						"value": true
					}
				]
			},
			"producedFacts": ["dehumidifier_status"],
			"consumedFacts": ["humidity", "room_occupied"]
		}
	]`

	// Parse the rule JSON into a ruleset
	var ruleset []*rules.Rule
	err := json.Unmarshal([]byte(ruleJSON), &ruleset)
	require.NoError(t, err, "Failed to parse rule JSON")

	// Initialize the RuleEngineContext
	context := rules.NewRuleEngineContext()

	// Create the compiler instance
	//compiler := NewCompiler(context)

	// Index the facts involved in the rules
	for _, rule := range ruleset {
		for _, fact := range rule.ConsumedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
		for _, fact := range rule.ProducedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
	}

	// Compile the ruleset
	bytecode, err := Compile(ruleset, context)
	require.NoError(t, err, "Compilation failed")

	// Expected bytecode for multiple rules with mixed conditions
	expectedBytecode := []byte{
		// TemperatureRule
		17, 0, // LOAD_FACT "temperature"
		19, 30, 0, 0, 0, // LOAD_CONST_INT 30
		4,        // GT_INT
		26, 5, 0, // JUMP_IF_FALSE 5 bytes ahead to end
		28, 1, // UPDATE_FACT "ac_status"
		22, 1, // LOAD_CONST_BOOL true
		// HumidityRule
		17, 2, // LOAD_FACT "humidity"
		19, 40, 0, 0, 0, // LOAD_CONST_INT 40
		2,        // LT_INT
		25, 9, 0, // JUMP_IF_TRUE 9 bytes ahead to action
		17, 3, // LOAD_FACT "room_occupied"
		22, 1, // LOAD_CONST_BOOL true
		0,        // EQ_BOOL
		26, 1, 0, // JUMP_IF_FALSE 5 bytes ahead to end
		28, 4, // UPDATE_FACT "dehumidifier_status"
		22, 1, // LOAD_CONST_BOOL true
	}

	assert.Equal(t, expectedBytecode, bytecode, "Compiled bytecode does not match the expected sequence")
}
