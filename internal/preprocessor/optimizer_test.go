package preprocessor

import (
	"rgehrsitz/rex/internal/rules"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCompareValues will test the compareValues function for various data types.
func TestCompareValues(t *testing.T) {
	// Test for integer comparison
	assert.True(t, compareValues(5, 10, "int"), "Expected 5 to be less than 10 for 'int' type")
	assert.False(t, compareValues(10, 5, "int"), "Expected 10 not to be less than 5 for 'int' type")

	// Test for float comparison
	assert.True(t, compareValues(5.0, 10.0, "float"), "Expected 5.0 to be less than 10.0 for 'float' type")
	assert.False(t, compareValues(10.0, 5.0, "float"), "Expected 10.0 not to be less than 5.0 for 'float' type")

	// Test for string comparison
	assert.True(t, compareValues("apple", "banana", "string"), "Expected 'apple' to be less than 'banana' for 'string' type")
	assert.False(t, compareValues("banana", "apple", "string"), "Expected 'banana' not to be less than 'apple' for 'string' type")

	// Test for boolean comparison
	assert.True(t, compareValues(false, true, "bool"), "Expected false to be less than true for 'bool' type")
	assert.False(t, compareValues(true, false, "bool"), "Expected true not to be less than false for 'bool' type")

	// Test for type mismatch handling (should return false)
	assert.False(t, compareValues("5", 5, "int"), "Expected type mismatch to return false")
}

// TestSortConditions will test the sortConditions function for correct ordering of conditions.
func TestSortConditions(t *testing.T) {
	// Unsorted conditions
	conditions := []rules.Condition{
		{Fact: "temperature", Operator: "greaterThan", Value: 30, ValueType: "float"},
		{Fact: "humidity", Operator: "lessThan", Value: 80, ValueType: "int"},
		{Fact: "status", Operator: "equal", Value: "open", ValueType: "string"},
	}

	// Expected sorted order based on the Fact field
	expectedOrder := []string{"humidity", "status", "temperature"}

	sortedConditions, err := sortConditions(conditions)
	assert.NoError(t, err, "sortConditions should not produce an error")

	for i, cond := range sortedConditions {
		assert.Equal(t, expectedOrder[i], cond.Fact, "Expected condition to be in the correct order")
	}

	// Add more tests as needed to cover different scenarios, including sorting with nested conditions.
}

// TestConditionsKey will test the conditionsKey function to ensure it generates a unique key for each distinct set of conditions.
func TestConditionsKey(t *testing.T) {
	conditions1 := rules.Conditions{
		All: []rules.Condition{{Fact: "temperature", Operator: "greaterThan", Value: 30}},
	}
	conditions2 := rules.Conditions{
		All: []rules.Condition{{Fact: "temperature", Operator: "greaterThan", Value: 30}},
	}
	conditions3 := rules.Conditions{
		All: []rules.Condition{{Fact: "humidity", Operator: "lessThan", Value: 80}},
	}

	key1, err := conditionsKey(conditions1)
	assert.NoError(t, err, "conditionsKey should not produce an error")
	key2, err := conditionsKey(conditions2)
	assert.NoError(t, err, "conditionsKey should not produce an error")
	key3, err := conditionsKey(conditions3)
	assert.NoError(t, err, "conditionsKey should not produce an error")

	// Two identical sets of conditions should produce the same key
	assert.Equal(t, key1, key2, "Identical conditions should produce the same key")

	// Different conditions should produce different keys
	assert.NotEqual(t, key1, key3, "Different conditions should produce different keys")

	// Add more tests to cover various complexities, including nested conditions.
}

// TestCompareValues_Equality tests comparing equal values.
func TestCompareValues_Equality(t *testing.T) {
	assert.False(t, compareValues(5, 5, "int"), "Expected equal values to return false")
	assert.False(t, compareValues(3.14, 3.14, "float"), "Expected equal values to return false")
	assert.False(t, compareValues("hello", "hello", "string"), "Expected equal values to return false")
	assert.False(t, compareValues(true, true, "bool"), "Expected equal values to return false")
}

// TestSortConditions_NestedConditions tests sorting with nested conditions.

func TestSortConditions_NestedConditions(t *testing.T) {
	conditions := []rules.Condition{
		{
			Fact:     "temperature",
			Operator: "greaterThan",
			Value:    30,
			All: []rules.Condition{
				{Fact: "humidity", Operator: "lessThan", Value: 80},
				{Fact: "status", Operator: "equal", Value: "open"},
			},
		},
		{
			Fact:     "temperature",
			Operator: "lessThan",
			Value:    20,
			Any: []rules.Condition{
				{Fact: "day", Operator: "equal", Value: "Monday"},
				{Fact: "time", Operator: "greaterThan", Value: "09:00"},
			},
		},
	}

	sortedConditions, err := sortConditions(conditions)
	assert.NoError(t, err, "sortConditions should not produce an error")

	// Assert the correct sorting order of the top-level conditions
	assert.Equal(t, "temperature", sortedConditions[0].Fact, "Expected 'temperature' to be the first condition")
	assert.Equal(t, "temperature", sortedConditions[1].Fact, "Expected 'temperature' to be the second condition")

	// Assert the correct sorting order of the nested 'All' conditions
	assert.Equal(t, "humidity", sortedConditions[0].All[0].Fact, "Expected 'humidity' to be the first nested 'All' condition")
	assert.Equal(t, "status", sortedConditions[0].All[1].Fact, "Expected 'status' to be the second nested 'All' condition")

	// Assert the correct sorting order of the nested 'Any' conditions
	assert.Equal(t, "day", sortedConditions[1].Any[0].Fact, "Expected 'day' to be the first nested 'Any' condition")
	assert.Equal(t, "time", sortedConditions[1].Any[1].Fact, "Expected 'time' to be the second nested 'Any' condition")
}

// TestConditionsKey_ComplexConditions tests key generation with complex nested conditions.
func TestConditionsKey_ComplexConditions(t *testing.T) {
	conditions1 := rules.Conditions{
		All: []rules.Condition{
			{
				Fact:     "temperature",
				Operator: "greaterThan",
				Value:    30,
				Any: []rules.Condition{
					{Fact: "humidity", Operator: "lessThan", Value: 80},
					{Fact: "status", Operator: "equal", Value: "open"},
				},
			},
		},
	}

	conditions2 := rules.Conditions{
		All: []rules.Condition{
			{
				Fact:     "temperature",
				Operator: "greaterThan",
				Value:    30,
				Any: []rules.Condition{
					{Fact: "status", Operator: "equal", Value: "open"},
					{Fact: "humidity", Operator: "lessThan", Value: 80},
				},
			},
		},
	}

	key1, err := conditionsKey(conditions1)
	assert.NoError(t, err, "conditionsKey should not produce an error")

	key2, err := conditionsKey(conditions2)
	assert.NoError(t, err, "conditionsKey should not produce an error")

	assert.Equal(t, key1, key2, "Equivalent complex conditions should produce the same key")
}

func TestPrioritizeRules(t *testing.T) {
	// Define a set of mock rules with varying priorities
	mockRules := []*rules.Rule{
		{
			Name:     "Rule1",
			Priority: 1, // Lowest priority
		},
		{
			Name:     "Rule2",
			Priority: 3, // Highest priority
		},
		{
			Name:     "Rule3",
			Priority: 2,
		},
		{
			Name: "Rule4", // No priority defined, should be treated as lowest priority
		},
	}

	// Execute the prioritizeRules function
	prioritizedRules := prioritizeRules(mockRules)

	// Verify the order of the rules after prioritization
	assert.Equal(t, "Rule2", prioritizedRules[0].Name, "Rule2 should be first due to highest priority")
	assert.Equal(t, "Rule3", prioritizedRules[1].Name, "Rule3 should be second")
	assert.Equal(t, "Rule1", prioritizedRules[2].Name, "Rule1 should be third")
	assert.Equal(t, "Rule4", prioritizedRules[3].Name, "Rule4 should be last (default priority)")

	// Test with rules having the same priority
	mockRulesSamePriority := []*rules.Rule{
		{
			Name:     "RuleA",
			Priority: 1,
		},
		{
			Name:     "RuleB",
			Priority: 1, // Same priority as RuleA
		},
	}

	prioritizedRulesSamePriority := prioritizeRules(mockRulesSamePriority)

	// Since RuleA and RuleB have the same priority, their relative order should be maintained
	assert.Equal(t, "RuleA", prioritizedRulesSamePriority[0].Name, "RuleA should come first due to original order")
	assert.Equal(t, "RuleB", prioritizedRulesSamePriority[1].Name, "RuleB should come second due to original order")

	// Edge cases such as an empty rule set can also be tested
	emptyRules := []*rules.Rule{}
	prioritizedEmptyRules := prioritizeRules(emptyRules)
	assert.Empty(t, prioritizedEmptyRules, "Prioritizing an empty rule set should result in an empty slice")
}

func TestPrioritizeRules2(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		inputRules     []*rules.Rule
		expectedOutput []*rules.Rule
	}{
		{
			name: "Rules with different priorities",
			inputRules: []*rules.Rule{
				{Name: "Rule1", Priority: 2},
				{Name: "Rule2", Priority: 1},
				{Name: "Rule3", Priority: 3},
			},
			expectedOutput: []*rules.Rule{
				{Name: "Rule3", Priority: 3},
				{Name: "Rule1", Priority: 2},
				{Name: "Rule2", Priority: 1},
			},
		},
		{
			name: "Rules with same priorities",
			inputRules: []*rules.Rule{
				{Name: "Rule1", Priority: 2},
				{Name: "Rule2", Priority: 2},
				{Name: "Rule3", Priority: 2},
			},
			expectedOutput: []*rules.Rule{
				{Name: "Rule1", Priority: 2},
				{Name: "Rule2", Priority: 2},
				{Name: "Rule3", Priority: 2},
			},
		},
		{
			name: "Rules with missing priorities",
			inputRules: []*rules.Rule{
				{Name: "Rule1", Priority: 2},
				{Name: "Rule2"},
				{Name: "Rule3", Priority: 1},
			},
			expectedOutput: []*rules.Rule{
				{Name: "Rule1", Priority: 2},
				{Name: "Rule3", Priority: 1},
				{Name: "Rule2"},
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := prioritizeRules(tc.inputRules)
			assert.Equal(t, tc.expectedOutput, output)
		})
	}
}

func TestGetRulePriority(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		rule     *rules.Rule
		expected int
	}{
		{
			name:     "Rule with defined priority",
			rule:     &rules.Rule{Priority: 3},
			expected: 3,
		},
		{
			name:     "Rule with missing priority",
			rule:     &rules.Rule{},
			expected: 0,
		},
		{
			name:     "Nil rule",
			rule:     nil,
			expected: 0,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			priority := getRulePriority(tc.rule)
			assert.Equal(t, tc.expected, priority)
		})
	}
}
