package preprocessor

import (
	"rgehrsitz/rex/internal/rules"
)

// OptimizeRules optimizes a slice of validated rules.
func OptimizeRules(validatedRules []*rules.Rule) ([]*rules.Rule, error) {
	// Placeholder for optimized rules - initially just a copy of the validated rules.
	optimizedRules := make([]*rules.Rule, len(validatedRules))
	copy(optimizedRules, validatedRules)

	// Apply various optimization strategies
	optimizedRules, err := mergeRules(optimizedRules)
	if err != nil {
		return nil, err
	}
	optimizedRules = prioritizeRules(optimizedRules)
	optimizedRules = simplifyConditions(optimizedRules)
	optimizedRules = precomputeExpressions(optimizedRules)
	optimizedRules = analyzeDependencies(optimizedRules)

	// Further optimization steps can be added here.

	return optimizedRules, nil
}

// Placeholder functions for various optimization strategies:
func prioritizeRules(rules []*rules.Rule) []*rules.Rule {
	// Implement prioritization logic here.
	return rules
}

func simplifyConditions(rules []*rules.Rule) []*rules.Rule {
	// Implement condition simplification logic here.
	return rules
}

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
		key := conditionsKey(rule.Conditions)
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
// This example implementation may need to be adapted to ensure that the key
// is truly unique based on the rule's logic.
func conditionsKey(conds rules.Conditions) string {
	// This function should return a string that uniquely represents the conditions.
	// For now, it returns a placeholder key.
	if conds.All == nil {
		return "placeholder_key"
	}
	return "placeholder_key"
}
