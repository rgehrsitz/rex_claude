package main

import (
	"flag"
	"os"
	"rgehrsitz/rex/internal/preprocessor"
	"rgehrsitz/rex/internal/preprocessor/bytecode"
	"rgehrsitz/rex/internal/rules" // Make sure to import the package where RuleEngineContext is defined

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func main() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Use pkgerrors for detailed error stack trace
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Define and parse command line flags for log level and output type
	logLevel := flag.String("loglevel", "info", "Set log level: panic, fatal, error, warn, info, debug, trace")
	logOutput := flag.String("logoutput", "console", "Set log output: console or file")
	flag.Parse()

	// Set the global log level based on the flag
	level, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid log level")
	}
	zerolog.SetGlobalLevel(level)

	// Set log output based on the flag
	switch *logOutput {
	case "console":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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

	log.Info().Msg("Starting the preprocessor")

	// Assume the first argument is the path to the JSON file
	if len(os.Args) < 2 {
		log.Fatal().Msg("No input file specified")
	}
	inputFilePath := os.Args[1]

	// Read the input file
	ruleJSON, err := os.ReadFile(inputFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read input file")
	}

	// Initialize a new RuleEngineContext
	context := rules.NewRuleEngineContext()

	// Parse and validate the rules with the context
	validatedRules, err := preprocessor.ParseAndValidateRules(ruleJSON, context)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse and validate rules")
		return // Using return instead of panic to allow for a graceful shutdown
	}

	// **Insert the new code here to update the context with all facts from validatedRules**
	for _, rule := range validatedRules {
		for _, fact := range rule.ConsumedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				index := len(context.FactIndex) // Assign a new index
				context.FactIndex[fact] = index
				log.Debug().Msg("Context updated with facts from validated rules")
			}
		}
		for _, fact := range rule.ProducedFacts {
			if _, exists := context.FactIndex[fact]; !exists {
				index := len(context.FactIndex) // Assign a new index
				context.FactIndex[fact] = index
				log.Debug().Msg("Context updated with facts from validated rules")
			}
		}
	}

	// OptimzeRules is assumed to be updated to accept a context as well
	optimizedRules, err := preprocessor.OptimizeRules(validatedRules, context)
	if err != nil {
		log.Error().Err(err).Msg("Failed to optimize rules")
		return
	}

	// Compile rules to bytecode, assuming the compiler can accept or use context if necessary
	compiler := bytecode.NewCompiler(context) // Assuming NewCompiler is updated to accept a context
	bytecodeBytes, err := compiler.Compile(optimizedRules)
	if err != nil {
		log.Error().Err(err).Msg("Error compiling rules to bytecode")
		return
	}

	// Write bytecode to a file
	err = os.WriteFile("bytecode.bin", bytecodeBytes, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Error writing bytecode to file")
		return
	}
}
