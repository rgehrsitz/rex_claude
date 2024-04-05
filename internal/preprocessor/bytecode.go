package preprocessor

import (
	"rgehrsitz/rex/internal/preprocessor/bytecode"
	"rgehrsitz/rex/internal/rules"
)

func ConvertRulesToBytecode(optimizedRules []*rules.Rule) ([]byte, error) {
	// Initialize a new RuleEngineContext
	context := rules.NewRuleEngineContext()
	// Create a new instance of the compiler
	compiler := bytecode.NewCompiler(context)

	// Compile the optimized rules using the compiler
	compiledBytecode, err := compiler.Compile(optimizedRules)
	if err != nil {
		return nil, err
	}

	return compiledBytecode, nil
}
