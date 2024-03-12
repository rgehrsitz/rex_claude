package main

import (
	"fmt"
	"os"
	"rgehrsitz/rex/internal/preprocessor"
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

	// Parse and validate the rules
	validatedRules, err := preprocessor.ParseAndValidateRules(ruleJSON)
	if err != nil {
		panic(err)
	}

	// Optimze the rules
	optimizedRules, err := preprocessor.OptimizeRules(validatedRules)
	if err != nil {
		panic(err)
	}

	// Parse and validate the rules
	bytecode, err := preprocessor.ConvertRulesToBytecode(optimizedRules)
	if err != nil {
		panic(err)
	}

	// Save the bytecode to a same name as inputFilePath, but with .bc extension
	outputFilePath := fmt.Sprintf("%s.bc", inputFilePath)
	err = os.WriteFile(outputFilePath, bytecode, 0644)
	if err != nil {
		panic(err)
	}
}
