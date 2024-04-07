package main

import (
	"fmt"
	"os"
	"rgehrsitz/rex/internal/preprocessor"
	"rgehrsitz/rex/internal/preprocessor/bytecode"
	"rgehrsitz/rex/internal/rules" // Make sure to import the package where RuleEngineContext is defined
)

func main() {
	// Assume the first argument is the path to the JSON file
	if len(os.Args) < 2 {
		panic("No input file specified")
	}
	inputFilePath := os.Args[1]

	// Read the input file
	ruleJSON, err := os.ReadFile(inputFilePath)
	if err != nil {
		panic(err)
	}

	// Initialize a new RuleEngineContext
	context := rules.NewRuleEngineContext()

	// Parse and validate the rules with the context
	validatedRules, err := preprocessor.ParseAndValidateRules(ruleJSON, context)
	if err != nil {
		panic(err)
	}

	// **Insert the new code here to update the context with all facts from validatedRules**
	for _, rule := range validatedRules {
		for _, fact := range rule.ConsumedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				index := len(context.FactIndex) // Assign a new index
				context.FactIndex[fact] = index
			}
		}
		for _, fact := range rule.ProducedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				index := len(context.FactIndex) // Assign a new index
				context.FactIndex[fact] = index
			}
		}
	}

	// OptimzeRules is assumed to be updated to accept a context as well
	optimizedRules, err := preprocessor.OptimizeRules(validatedRules, context) // Assuming OptimizeRules is updated
	if err != nil {
		panic(err)
	}

	// Compile rules to bytecode, assuming the compiler can accept or use context if necessary
	compiler := bytecode.NewCompiler(context) // Assuming NewCompiler is updated to accept a context
	bytecodeBytes, err := compiler.Compile(optimizedRules)
	if err != nil {
		fmt.Println("Error compiling rules to bytecode:", err)
		return
	}

	// Write bytecode to a file
	err = os.WriteFile("bytecode.bin", bytecodeBytes, 0644)
	if err != nil {
		fmt.Println("Error writing bytecode to file:", err)
		return
	}
}
