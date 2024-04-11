package main

import (
	"os"
	"rgehrsitz/rex/internal/runtime"

	"github.com/rs/zerolog/log"
)

func main() {
	// Check if a file path is provided as an argument
	if len(os.Args) < 2 {
		log.Error().Msg("Usage: runtime <bytecode_file>")
		return
	}

	// Read the bytecode file
	bytecodeFilePath := os.Args[1]
	bytecodeBytes, err := os.ReadFile(bytecodeFilePath)
	if err != nil {
		log.Error().Err(err).Msg("Error reading bytecode file")
		return
	}

	// Create a new VM instance and run the bytecode
	vm := runtime.NewVM(bytecodeBytes)
	err = vm.Run()
	if err != nil {
		log.Error().Err(err).Msg("Error running bytecode")
		return
	}

	log.Info().Msg("Bytecode execution completed successfully.")

}
