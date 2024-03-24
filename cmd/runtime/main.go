package main

import (
	"fmt"
	"os"
	"rgehrsitz/rex/internal/runtime"
)

func main() {
	// Check if a file path is provided as an argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: runtime <bytecode_file>")
		return
	}

	// Read the bytecode file
	bytecodeFilePath := os.Args[1]
	bytecodeBytes, err := os.ReadFile(bytecodeFilePath)
	if err != nil {
		fmt.Println("Error reading bytecode file:", err)
		return
	}

	// Create a new VM instance and run the bytecode
	vm := runtime.NewVM(bytecodeBytes)
	err = vm.Run()
	if err != nil {
		fmt.Println("Error running bytecode:", err)
		return
	}

	fmt.Println("Bytecode execution completed successfully.")
}
