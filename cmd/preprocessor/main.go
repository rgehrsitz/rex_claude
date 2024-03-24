package main

import (
	"fmt"
	"os"
	"rgehrsitz/rex/internal/preprocessor"
	"rgehrsitz/rex/internal/preprocessor/bytecode"
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

	// Compile rules to bytecode
	compiler := bytecode.NewCompiler()
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
