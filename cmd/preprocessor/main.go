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

	// Parse and preprocess the rules
	optimizedRules, err := preprocessor.OptimizeRules(ruleJSON)
	if err != nil {
		panic(err)
	}

	// Output the optimized rules
	// This could be to a file or stdout, depending on your needs
	// For simplicity, just printing to stdout here
	fmt.Println(string(optimizedRules))
}
