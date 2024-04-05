package preprocessor

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"rgehrsitz/rex/internal/rules"
	"sort"
)

// OptimizeRules optimizes a slice of validated rules.
// OptimizeRules now also accepts a pointer to RuleEngineContext
func OptimizeRules(validatedRules []*rules.Rule, context *rules.RuleEngineContext) ([]*rules.Rule, error) {
	// Optimization logic remains mostly unchanged
	// You can now utilize 'context' for optimizations
	// For example, you might adjust optimizations based on the facts each rule consumes or produces

	optimizedRules := make([]*rules.Rule, len(validatedRules))
	copy(optimizedRules, validatedRules)

	// Apply various optimization strategies that might utilize 'context'
	optimizedRules, err := mergeRules(optimizedRules) // Assuming you adjust other functions similarly
	if err != nil {
		return nil, err
	}
	optimizedRules = prioritizeRules(optimizedRules)
	optimizedRules = simplifyConditions(optimizedRules)
	optimizedRules = precomputeExpressions(optimizedRules)
	optimizedRules = analyzeDependencies(optimizedRules)

	return optimizedRules, nil
}

// Placeholder functions for various optimization strategies:
func prioritizeRules(rulesToPrioritize []*rules.Rule) []*rules.Rule {
	// Create a copy of the rules slice to avoid modifying the original
	prioritizedRules := make([]*rules.Rule, len(rulesToPrioritize))
	copy(prioritizedRules, rulesToPrioritize)

	// Sort the rules based on their user-assigned priorities in descending order
	// Use sort.SliceStable for stable sorting.
	sort.SliceStable(prioritizedRules, func(i, j int) bool {
		// Handle cases where priority is not defined by treating them as lowest priority.
		priorityI := getRulePriority(prioritizedRules[i])
		priorityJ := getRulePriority(prioritizedRules[j])

		return priorityI > priorityJ
	})

	return prioritizedRules
}

// getRulePriority returns the priority of a rule, defaulting to 0 if not set.
func getRulePriority(r *rules.Rule) int {
	if r != nil {
		return r.Priority
	}
	return 0 // Default priority value if not set
}

func simplifyConditions(rulesToSimplify []*rules.Rule) []*rules.Rule {
	simplifiedRules := make([]*rules.Rule, 0, len(rulesToSimplify))
	for _, rule := range rulesToSimplify {
		simplifiedConditions := simplifyRuleConditions(rule.Conditions)
		if !equalConditions(simplifiedConditions, rule.Conditions) {
			simplifiedRule := &rules.Rule{
				Name:          rule.Name,
				Priority:      rule.Priority,
				Conditions:    simplifiedConditions,
				Event:         rule.Event,
				ProducedFacts: rule.ProducedFacts,
				ConsumedFacts: rule.ConsumedFacts,
			}
			simplifiedRules = append(simplifiedRules, simplifiedRule)
		} else {
			simplifiedRules = append(simplifiedRules, rule)
		}
	}
	return simplifiedRules
}

func simplifyRuleConditions(conditions rules.Conditions) rules.Conditions {
	simplified := rules.Conditions{
		All: simplifyAndDedupConditions(conditions.All),
		Any: simplifyAndDedupConditions(conditions.Any),
	}
	return simplified
}

func simplifyAndDedupConditions(conditions []rules.Condition) []rules.Condition {
	simplified := make([]rules.Condition, 0)
	for _, cond := range conditions {
		simplifiedCond := simplifyCondition(cond)
		if !containsCondition(simplified, simplifiedCond) {
			simplified = append(simplified, simplifiedCond)
		}
	}
	return simplified
}

func simplifyCondition(condition rules.Condition) rules.Condition {
	// First, recursively simplify any nested conditions.
	simplified := rules.Condition{
		Fact:      condition.Fact,
		Operator:  condition.Operator,
		Value:     condition.Value,
		ValueType: condition.ValueType,
		All:       simplifyAndDedupConditions(condition.All),
		Any:       simplifyAndDedupConditions(condition.Any),
	}

	// Example logical simplification: Identify redundant or overlapping conditions.
	// This is highly dependent on the logic of your conditions.
	// Below is a very basic placeholder logic.
	if canBeSimplified(simplified) {
		simplified = performLogicalSimplification(simplified)
	}

	return simplified
}
func canBeSimplified(condition rules.Condition) bool {
	// Example logic for a simple case where two "All" conditions might contradict or be redundant.
	if len(condition.All) >= 2 {
		// Placeholder logic: check for direct contradictions or redundancies
		// Real logic should be more comprehensive and based on actual operators and values.
		for i := 0; i < len(condition.All)-1; i++ {
			for j := i + 1; j < len(condition.All); j++ {
				if condition.All[i].Fact == condition.All[j].Fact {
					return true // Simplistic check; real logic should compare operators and values.
				}
			}
		}
	}
	// Check for other patterns that can be simplified.
	return false
}

func performLogicalSimplification(condition rules.Condition) rules.Condition {
	simplifiedCondition := condition // Start with the original condition

	// Simplify "All" conditions as an example.
	// This simplistic logic only considers direct redundancy based on the Fact.
	// A real implementation should consider operators and values.
	var newAll []rules.Condition
	seenFacts := make(map[string]bool)
	for _, cond := range condition.All {
		if _, seen := seenFacts[cond.Fact]; !seen {
			newAll = append(newAll, cond)
			seenFacts[cond.Fact] = true
		} // Else, it's a redundant condition and can be omitted.
	}
	simplifiedCondition.All = newAll

	// Similarly, apply simplification to "Any" conditions and nested conditions.

	return simplifiedCondition
}

func containsCondition(conditions []rules.Condition, condition rules.Condition) bool {
	for _, c := range conditions {
		if equalCondition(c, condition) {
			return true
		}
	}
	return false
}

func equalConditions(c1, c2 rules.Conditions) bool {
	if len(c1.All) != len(c2.All) || len(c1.Any) != len(c2.Any) {
		return false
	}
	for i := range c1.All {
		if !equalCondition(c1.All[i], c2.All[i]) {
			return false
		}
	}
	for i := range c1.Any {
		if !equalCondition(c1.Any[i], c2.Any[i]) {
			return false
		}
	}
	return true
}

func equalCondition(c1, c2 rules.Condition) bool {
	return c1.Fact == c2.Fact &&
		c1.Operator == c2.Operator &&
		c1.ValueType == c2.ValueType &&
		reflect.DeepEqual(c1.Value, c2.Value)
}

// func equalCondition(c1, c2 rules.Condition) bool {
// 	return c1.Fact == c2.Fact &&
// 		c1.Operator == c2.Operator &&
// 		c1.ValueType == c2.ValueType &&
// 		reflect.DeepEqual(c1.Value, c2.Value) &&
// 		equalConditions(rules.Conditions{All: c1.All, Any: c1.Any}, rules.Conditions{All: c2.All, Any: c2.Any})
// }

func precomputeExpressions(rules []*rules.Rule) []*rules.Rule {
	// Implement precomputation logic here.
	return rules
}

func analyzeDependencies(rules []*rules.Rule) []*rules.Rule {
	return rules
}

// mergeRules combines rules with identical conditions.
func mergeRules(rulesToMerge []*rules.Rule) ([]*rules.Rule, error) {
	// A map to identify and combine rules with identical conditions
	mergedRules := make(map[string]*rules.Rule)
	for _, rule := range rulesToMerge {
		key, _ := conditionsKey(rule.Conditions)
		if existingRule, found := mergedRules[key]; found {
			// Merge actions from the current rule into the existing rule
			existingRule.Event.Actions = append(existingRule.Event.Actions, rule.Event.Actions...)
			existingRule.ProducedFacts = append(existingRule.ProducedFacts, rule.ProducedFacts...)
			existingRule.ConsumedFacts = append(existingRule.ConsumedFacts, rule.ConsumedFacts...)
		} else {
			// If this set of conditions hasn't been seen before, add the rule to the map
			mergedRules[key] = rule
		}
	}

	// Convert the map back to a slice
	var optimizedRules []*rules.Rule
	for _, rule := range mergedRules {
		optimizedRules = append(optimizedRules, rule)
	}
	return optimizedRules, nil
}

// conditionsKey generates a unique key based on the conditions of a rule.
func conditionsKey(conds rules.Conditions) (string, error) {
	// Normalize conditions to ensure consistent ordering
	normalizedConditions, err := normalizeConditions(conds)
	if err != nil {
		return "", err
	}

	// Serialize the normalized conditions to JSON
	serializedConditions, err := json.Marshal(normalizedConditions)
	if err != nil {
		return "", fmt.Errorf("error marshaling conditions: %v", err)
	}

	// Use SHA256 hashing to create a unique key for the serialized conditions
	hash := sha256.Sum256(serializedConditions)
	return fmt.Sprintf("%x", hash), nil
}

// normalizeConditions ensures that conditions are in a consistent order.
func normalizeConditions(conds rules.Conditions) (rules.Conditions, error) {
	// Normalize 'All' conditions
	sortedAll, err := sortConditions(conds.All)
	if err != nil {
		return rules.Conditions{}, err
	}

	// Normalize 'Any' conditions
	sortedAny, err := sortConditions(conds.Any)
	if err != nil {
		return rules.Conditions{}, err
	}

	return rules.Conditions{All: sortedAll, Any: sortedAny}, nil
}

// sortConditions sorts conditions and their nested conditions.
func sortConditions(conditions []rules.Condition) ([]rules.Condition, error) {
	// Sort the slice of conditions by some consistent criteria
	sort.SliceStable(conditions, func(i, j int) bool {
		// Define a consistent sorting logic.
		if conditions[i].Fact != conditions[j].Fact {
			return conditions[i].Fact < conditions[j].Fact
		}
		if conditions[i].Operator != conditions[j].Operator {
			return conditions[i].Operator < conditions[j].Operator
		}
		if conditions[i].ValueType != conditions[j].ValueType {
			return conditions[i].ValueType < conditions[j].ValueType
		}

		// Custom comparison for Value based on ValueType
		return compareValues(conditions[i].Value, conditions[j].Value, conditions[i].ValueType)
	})

	// Recursively sort nested conditions
	for i, cond := range conditions {
		if len(cond.All) > 0 || len(cond.Any) > 0 {
			sortedNestedConds, err := normalizeConditions(rules.Conditions{All: cond.All, Any: cond.Any})
			if err != nil {
				return nil, fmt.Errorf("error sorting conditions: %v", err)
			}
			conditions[i].All = sortedNestedConds.All
			conditions[i].Any = sortedNestedConds.Any
		}
	}

	return conditions, nil
}

// compareValues compares two values based on their type.
func compareValues(v1, v2 interface{}, valueType string) bool {
	// Perform a type switch to determine how to compare the values
	switch valueType {
	case "int":
		val1, ok1 := v1.(int)
		val2, ok2 := v2.(int)
		if !ok1 || !ok2 {
			return false // Default to false if types do not match expectations
		}
		return val1 < val2
	case "float":
		val1, ok1 := v1.(float64)
		val2, ok2 := v2.(float64)
		if !ok1 || !ok2 {
			return false // Default to false if types do not match expectations
		}
		return val1 < val2
	case "string":
		val1, ok1 := v1.(string)
		val2, ok2 := v2.(string)
		if !ok1 || !ok2 {
			return false // Default to false if types do not match expectations
		}
		return val1 < val2
	case "bool":
		val1, ok1 := v1.(bool)
		val2, ok2 := v2.(bool)
		if !ok1 || !ok2 {
			return false // Default to false if types do not match expectations
		}
		// For booleans, false < true
		return !val1 && val2
	default:
		// If ValueType is unknown or not handled, default to false
		return false
	}
}
