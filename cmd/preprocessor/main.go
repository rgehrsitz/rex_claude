package main

import (
	"flag"
	"fmt"
	"os"
	"rgehrsitz/rex/internal/preprocessor"
	"rgehrsitz/rex/internal/preprocessor/bytecode"
	"rgehrsitz/rex/internal/rules" // Make sure to import the package where RuleEngineContext is defined

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func main() {
	// Setup command-line flags
	logLevel := flag.String("loglevel", "info", "Set log level: panic, fatal, error, warn, info, debug, trace")
	logOutput := flag.String("logoutput", "console", "Set log output: console or file")
	inputFile := flag.String("input", "", "Path to the input JSON file")
	flag.Parse()

	// Configure zerolog based on the flags
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	level, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid log level")
	}
	zerolog.SetGlobalLevel(level)

	switch *logOutput {
	case "console":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "3:04PM"})
	case "file":
		file, err := os.Create("logs.txt")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create log file")
		}
		defer file.Close()
		log.Logger = log.Output(file)
	default:
		log.Fatal().Msg("Invalid log output option")
	}

	// Check for input file argument
	if *inputFile == "" {
		log.Fatal().Msg("No input file specified")
	}

	// Process the input file
	ruleJSON, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read input file")
	}

	compilationContext := createCompilationContext()
	compiledBytecode, err := compileRules(ruleJSON, compilationContext)
	if err != nil {
		log.Error().Err(err).Msg("Failed to compile rules")
		return
	}

	err = os.WriteFile("bytecode.bin", compiledBytecode, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Error writing bytecode to file")
		return
	}
}

func createCompilationContext() *rules.CompilationContext {
	return &rules.CompilationContext{
		FactIndex:     make(map[string]int),
		ConsumedFacts: make(map[string]bool),
		ProducedFacts: make(map[string]bool),
	}
}

func compileRules(ruleJSON []byte, context *rules.CompilationContext) ([]byte, error) {
	parsedRules, err := preprocessor.ParseRules(ruleJSON, context)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}

	err = preprocessor.ValidateRules(parsedRules, context)
	if err != nil {
		return nil, fmt.Errorf("failed to validate rules: %w", err)
	}

	updateContextWithRuleFacts(parsedRules, context)

	optimizedRules, err := preprocessor.OptimizeRules(parsedRules, context)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize rules: %w", err)
	}

	bytecodeBytes, err := bytecode.Compile(optimizedRules, context)
	if err != nil {
		return nil, fmt.Errorf("error compiling rules to bytecode: %w", err)
	}

	return bytecodeBytes, nil
}

func updateContextWithRuleFacts(rules []*rules.Rule, context *rules.CompilationContext) {
	for _, rule := range rules {
		for _, fact := range rule.ConsumedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
		for _, fact := range rule.ProducedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				context.FactIndex[fact] = len(context.FactIndex)
			}
		}
	}
}
