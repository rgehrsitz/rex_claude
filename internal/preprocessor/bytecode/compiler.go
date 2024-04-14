// File: compiler.go
package bytecode

import (
	"fmt"
	"math"
	"rgehrsitz/rex/internal/rules"

	"github.com/rs/zerolog/log"
)

// Compile compiles a set of optimized rules into bytecode, recording fact usage.
func Compile(rules []*rules.Rule, context *rules.CompilationContext) ([]byte, error) {
	var (
		bytecodeBuffer []byte
		factIndex      = make(map[string]int) // Index facts for quick access
	)

	for _, rule := range rules {
		ruleBytecode, err := compileRule(*rule, &factIndex)
		if err != nil {
			log.Error().Err(err).Str("rule", rule.Name).Msg("Failed to compile rule")
			return nil, err
		}
		bytecodeBuffer = append(bytecodeBuffer, ruleBytecode...)
	}

	return bytecodeBuffer, nil
}

// compileRule compiles an individual rule into bytecode, updating the fact index.
func compileRule(rule rules.Rule, factIndex *map[string]int) ([]byte, error) {
	var code []byte
	code = append(code, byte(RULE_START))
	logBytecodeStep("After appending RULE_START", code)

	// Initialize fact index positions before compiling conditions or actions
	initializeFactIndex(rule, factIndex)

	conditionsBytecode, err := compileConditions(rule.Conditions, factIndex)
	if err != nil {
		return nil, err
	}
	code = append(code, conditionsBytecode...)
	logBytecodeStep("After compiling conditions", code)

	eventBytecode, err := compileEvent(rule.Event, factIndex)
	if err != nil {
		return nil, err
	}
	code = append(code, eventBytecode...)
	logBytecodeStep("After compiling event", code)
	code = append(code, byte(RULE_END))
	logBytecodeStep("After appending RULE_END", code)

	return code, nil
}

// Initializes fact index mapping each fact to a unique position.
func initializeFactIndex(rule rules.Rule, factIndex *map[string]int) {
	index := 0
	for _, fact := range rule.ConsumedFacts {
		if _, exists := (*factIndex)[fact]; !exists {
			(*factIndex)[fact] = index
			index++
		}
	}
	for _, fact := range rule.ProducedFacts {
		if _, exists := (*factIndex)[fact]; !exists {
			(*factIndex)[fact] = index
			index++
		}
	}
}

// compileConditions handles both single and nested conditions.
func compileConditions(conditions rules.Conditions, factIndex *map[string]int) ([]byte, error) {
	var code []byte
	code = append(code, byte(COND_START))
	logBytecodeStep("After appending COND_START", code)

	// Compile 'all' conditions
	for i, cond := range conditions.All {
		compiledCond, err := compileSingleCondition(cond, factIndex)
		if err != nil {
			return nil, err
		}
		code = append(code, compiledCond...)
		logBytecodeStep(fmt.Sprintf("After compiling 'all' condition %d", i), code)

		if i < len(conditions.All)-1 {
			code = append(code, byte(JUMP_IF_FALSE))
			// This is where we should set JumpPos
			jumpPos := len(code)                  // This will point to the next byte where offset will be written
			code = append(code, byte(0), byte(0)) // Placeholder for jump position
			conditions.All[i].JumpPos = jumpPos
			logBytecodeStep(fmt.Sprintf("After appending JUMP_IF_FALSE for 'all' condition %d", i), code)
		}
	}

	// Compile 'any' conditions
	for i, cond := range conditions.Any {
		compiledCond, err := compileSingleCondition(cond, factIndex)
		if err != nil {
			return nil, err
		}
		code = append(code, compiledCond...)
		logBytecodeStep(fmt.Sprintf("After compiling 'any' condition %d", i), code)

		if i < len(conditions.Any)-1 {
			code = append(code, byte(JUMP_IF_TRUE))
			// This is where we should set JumpPos
			jumpPos := len(code)                  // This will point to the next byte where offset will be written
			code = append(code, byte(0), byte(0)) // Placeholder for jump position
			conditions.Any[i].JumpPos = jumpPos
			logBytecodeStep(fmt.Sprintf("After appending JUMP_IF_TRUE for 'any' condition %d", i), code)
		}
	}

	code = append(code, byte(COND_END))
	logBytecodeStep("After appending COND_END", code)

	// Patch jump positions
	patchJumpPositions(code, conditions)

	return code, nil
}

// compileSingleCondition compiles a single condition, potentially recursive for nested conditions.
func compileSingleCondition(cond rules.Condition, factIndex *map[string]int) ([]byte, error) {
	var code []byte

	// Generate appropriate bytecode based on condition operator
	switch cond.Operator {
	case "equal":
		switch cond.ValueType {
		case "int":
			code = append(code, byte(EQ_INT))
		case "float":
			code = append(code, byte(EQ_FLOAT))
		case "string":
			code = append(code, byte(EQ_STRING))
		default:
			return nil, fmt.Errorf("unsupported value type for equal operator: %s", cond.ValueType)
		}
	case "notEqual":
		switch cond.ValueType {
		case "int":
			code = append(code, byte(NEQ_INT))
		case "float":
			code = append(code, byte(NEQ_FLOAT))
		case "string":
			code = append(code, byte(NEQ_STRING))
		default:
			return nil, fmt.Errorf("unsupported value type for notEqual operator: %s", cond.ValueType)
		}
	case "greaterThan":
		switch cond.ValueType {
		case "int":
			code = append(code, byte(GT_INT))
		case "float":
			code = append(code, byte(GT_FLOAT))
		default:
			return nil, fmt.Errorf("unsupported value type for greaterThan operator: %s", cond.ValueType)
		}
	case "greaterThanOrEqual":
		switch cond.ValueType {
		case "int":
			code = append(code, byte(GTE_INT))
		case "float":
			code = append(code, byte(GTE_FLOAT))
		default:
			return nil, fmt.Errorf("unsupported value type for greaterThanOrEqual operator: %s", cond.ValueType)
		}
	case "lessThan":
		switch cond.ValueType {
		case "int":
			code = append(code, byte(LT_INT))
		case "float":
			code = append(code, byte(LT_FLOAT))
		default:
			return nil, fmt.Errorf("unsupported value type for lessThan operator: %s", cond.ValueType)
		}
	case "lessThanOrEqual":
		switch cond.ValueType {
		case "int":
			code = append(code, byte(LTE_INT))
		case "float":
			code = append(code, byte(LTE_FLOAT))
		default:
			return nil, fmt.Errorf("unsupported value type for lessThanOrEqual operator: %s", cond.ValueType)
		}
	default:
		return nil, fmt.Errorf("unsupported operator: %s", cond.Operator)
	}

	// Load fact value
	factIdx, ok := (*factIndex)[cond.Fact]
	if !ok {
		return nil, fmt.Errorf("fact not found: %s", cond.Fact)
	}
	code = append(code, byte(LOAD_FACT))
	code = append(code, byte(factIdx))

	// Load comparison value
	switch cond.ValueType {
	case "int":
		code = append(code, byte(LOAD_CONST_INT))
		value := cond.Value.(int)
		code = append(code, byte(value>>24), byte(value>>16), byte(value>>8), byte(value))
	case "float":
		code = append(code, byte(LOAD_CONST_FLOAT))
		value := cond.Value.(float64)
		bits := math.Float64bits(value)
		code = append(code, byte(bits>>56), byte(bits>>48), byte(bits>>40), byte(bits>>32),
			byte(bits>>24), byte(bits>>16), byte(bits>>8), byte(bits))
	case "string":
		code = append(code, byte(LOAD_CONST_STRING))
		value := cond.Value.(string)
		code = append(code, byte(len(value)))
		code = append(code, []byte(value)...)
	case "bool":
		code = append(code, byte(LOAD_CONST_BOOL))
		value := cond.Value.(bool)
		if value {
			code = append(code, byte(1))
		} else {
			code = append(code, byte(0))
		}
	default:
		return nil, fmt.Errorf("unsupported value type: %s", cond.ValueType)
	}

	return code, nil
}

// compileEvent processes actions associated with a rule's event.
func compileEvent(event rules.Event, factIndex *map[string]int) ([]byte, error) {
	var code []byte

	code = append(code, byte(ACTION_START))

	for _, action := range event.Actions {
		switch action.Type {
		case "updateFact":
			code = append(code, byte(UPDATE_FACT))
			factIndex := (*factIndex)[action.Target]
			code = append(code, byte(factIndex))
			// Append the new value based on its type
			switch value := action.Value.(type) {
			case int:
				code = append(code, byte(LOAD_CONST_INT))
				code = append(code, byte(value>>24), byte(value>>16), byte(value>>8), byte(value))
			case float64:
				code = append(code, byte(LOAD_CONST_FLOAT))
				bits := math.Float64bits(value)
				code = append(code, byte(bits>>56), byte(bits>>48), byte(bits>>40), byte(bits>>32),
					byte(bits>>24), byte(bits>>16), byte(bits>>8), byte(bits))
			case string:
				code = append(code, byte(LOAD_CONST_STRING))
				code = append(code, byte(len(value)))
				code = append(code, []byte(value)...)
			case bool:
				code = append(code, byte(LOAD_CONST_BOOL))
				if value {
					code = append(code, byte(1))
				} else {
					code = append(code, byte(0))
				}
			default:
				return nil, fmt.Errorf("unsupported value type for updateStore action: %T", value)
			}
		case "sendMessage":
			code = append(code, byte(SEND_MESSAGE))
			// Append the message target and content
			target := action.Target
			code = append(code, byte(len(target)))
			code = append(code, []byte(target)...)
			content := action.Value.(string)
			code = append(code, byte(len(content)))
			code = append(code, []byte(content)...)
		default:
			return nil, fmt.Errorf("unsupported action type: %s", action.Type)
		}
	}

	code = append(code, byte(ACTION_END))

	return code, nil
}

const InstructionLength = 3 // Adjust according to actual length

func patchJumpPositions(code []byte, conditions rules.Conditions) {
	logBytecodeStep("Before patching jumps", code)
	// Patch jump positions for all types of conditions
	patchJumps(code, conditions.All)
	patchJumps(code, conditions.Any)
	logBytecodeStep("After patching jumps", code)
}

func patchJumps(code []byte, jumps []rules.Condition) {
	for _, cond := range jumps {
		if cond.JumpPos < InstructionLength {
			log.Error().Msg("Invalid jump position, less than instruction length")
			continue
		}
		if cond.JumpPos > 0 {
			log.Trace().Msgf("Preparing to patch jump for condition: %s at position: %d", cond.Fact, cond.JumpPos)

			if cond.JumpPos < InstructionLength {
				log.Error().Msg("Invalid jump position, less than instruction length")
				continue
			}

			if cond.JumpPos >= len(code)-InstructionLength {
				log.Error().Msg("Invalid jump position, exceeds bytecode length")
				continue
			}

			jumpPos := cond.JumpPos
			jumpOffset := len(code) - (jumpPos + InstructionLength)
			if jumpOffset < 0 || jumpOffset > 65535 {
				log.Error().Str("condition", cond.Fact).Msgf("Jump offset out of bounds: %d", jumpOffset)
				continue
			}

			code[jumpPos+1] = byte(jumpOffset >> 8)
			code[jumpPos+2] = byte(jumpOffset & 0xFF)
			log.Trace().Msgf("Patched jump for condition: %s at position: %d with offset: %d", cond.Fact, jumpPos, jumpOffset)
		}
	}
}

func logBytecodeStep(description string, code []byte) {
	log.Trace().Msgf("%s: current bytecode length=%d, last instruction=%d", description, len(code), code[len(code)-1])

}
