// File: compiler.go

package bytecode

import (
	"encoding/binary"
	"fmt"
	"math"
	"rgehrsitz/rex/internal/rules"

	"github.com/rs/zerolog/log"
)

// Compiler compiles optimized rules into bytecode.
type Compiler struct {
	instructions       []Instruction
	bytecode           []byte
	labelOffsets       map[string]int
	labelCounter       int
	context            *rules.RuleEngineContext
	jumpsNeedingLabels []jumpLabelPair
}

type jumpLabelPair struct {
	instructionIndex int    // Position in the instructions slice
	label            string // The label this jump is associated with
}

// NewCompiler creates a new instance of the bytecode compiler.
func NewCompiler(context *rules.RuleEngineContext) *Compiler {
	return &Compiler{
		instructions:       []Instruction{},
		bytecode:           []byte{},
		labelOffsets:       make(map[string]int),
		labelCounter:       0,
		context:            context,
		jumpsNeedingLabels: make([]jumpLabelPair, 0),
	}
}

// Compile compiles a set of rules into bytecode.
func (c *Compiler) Compile(rules []*rules.Rule) ([]byte, error) {
	for _, rule := range rules {
		if err := c.compileRule(rule); err != nil {
			return nil, err
		}
	}

	// After compiling all rules, resolve label offsets to finalize the bytecode
	if err := c.resolveLabelOffsets(); err != nil {
		return nil, err
	}

	return c.bytecode, nil
}

// generateUniqueLabel generates a unique label for use in the bytecode.
func (c *Compiler) generateUniqueLabel(base string) string {
	label := fmt.Sprintf("%s_%d", base, c.labelCounter)
	log.Debug().Str("Label", label).Msg("Generated unique label")
	c.labelCounter++
	return label
}

// emitInstruction appends an instruction to the compiler's list of instructions and updates its bytecode position.
func (c *Compiler) emitInstruction(opcode Opcode, operands ...byte) {
	// Calculate the current bytecode position based on the actual bytecode size.
	currentBytecodePosition := len(c.bytecode)

	// Append the new instruction to the bytecode.
	c.bytecode = append(c.bytecode, byte(opcode))
	c.bytecode = append(c.bytecode, operands...)

	// Append the new instruction to the instructions slice for reference.
	c.instructions = append(c.instructions, Instruction{
		Opcode:           opcode,
		Operands:         operands,
		BytecodePosition: currentBytecodePosition,
	})

	log.Debug().
		Int("Opcode", int(opcode)).
		Str("Operation", opcode.String()).
		Interface("Operands", operands).
		Int("BytecodePosition", currentBytecodePosition).
		Msg("Emitted instruction")

}

// emitLabel emits a label instruction and records its offset.
func (c *Compiler) emitLabel(label string) {
	// The label offset should be the current length of the bytecode slice,
	// which represents the position in the bytecode where the label is defined.
	labelOffset := len(c.bytecode)

	c.labelOffsets[label] = labelOffset

	log.Debug().
		Str("Label", label).
		Int("BytecodePosition", labelOffset).
		Msg("Emitted label")
}

// compileRule compiles a single rule into bytecode.
func (c *Compiler) compileRule(rule *rules.Rule) error {
	log.Debug().
		Str("RuleID", rule.Name).
		Msg("Starting compilation of rule")

	startLabel := c.generateUniqueLabel("rule_start")
	endLabel := c.generateUniqueLabel("rule_end")
	c.emitLabel(startLabel)

	if err := c.compileConditions(rule.Conditions, endLabel); err != nil {
		return err
	}

	// Compile the actions
	for _, action := range rule.Event.Actions {
		switch action.Type {
		case "updateFact":
			factIndex, err := c.getFactIndex(action.Target)
			if err != nil {
				return err
			}
			c.emitInstruction(UPDATE_FACT, byte(factIndex))
			c.emitLoadConstantInstruction(action.Value, "bool")
		// Add cases for other action types as needed
		default:
			log.Error().
				Str("ActionType", action.Type).
				Msg("Unsupported action type encountered")

			return fmt.Errorf("unsupported action type: %s", action.Type)
		}
	}

	c.emitLabel(endLabel)
	log.Info().
		Int("BytecodeSize", len(c.bytecode)).
		Msg("Compilation completed successfully")

	return nil
}

// compileConditions compiles conditions (including nested conditions) into bytecode.
func (c *Compiler) compileConditions(conditions rules.Conditions, endLabel string) error {
	for i := range conditions.All {
		// Use the index to obtain a pointer to each condition
		if err := c.compileCondition(&conditions.All[i], endLabel, false); err != nil {
			return err
		}
	}

	for i := range conditions.Any {
		// Generate a unique label for the action part of the "any" conditions
		actionLabel := c.generateUniqueLabel("action")
		// Use the index to obtain a pointer to each condition
		if err := c.compileCondition(&conditions.Any[i], actionLabel, true); err != nil {
			return err
		}
		// Emit a jump instruction to skip to the end if the condition is true,
		// since for "any" conditions, we want to perform the action if any condition is true.
		c.emitInstruction(JUMP, []byte(endLabel)...)
		// When emitting a jump instruction related to a label
		c.jumpsNeedingLabels = append(c.jumpsNeedingLabels, jumpLabelPair{
			instructionIndex: len(c.instructions) - 1, // Index of the jump instruction just added
			label:            endLabel,                // The label the jump is associated with
		})
	}

	return nil
}

// compileCondition compiles a single condition or nested block into bytecode.
func (c *Compiler) compileCondition(condition *rules.Condition, jumpLabel string, jumpIfTrue bool) error {
	placeholder := []byte{0x00, 0x00} // Using 2 bytes for the placeholder

	// Handle nested `all` conditions
	if len(condition.All) > 0 {
		for _, nestedCond := range condition.All {
			if err := c.compileCondition(&nestedCond, jumpLabel, false); err != nil {
				return err
			}
		}
		return nil // All `all` conditions processed
	}

	// Handle nested `any` conditions
	if len(condition.Any) > 0 {
		anyEndLabel := c.generateUniqueLabel("any_end")
		for _, nestedCond := range condition.Any {
			if err := c.compileCondition(&nestedCond, jumpLabel, true); err != nil {
				return err
			}
		}
		c.emitInstruction(JUMP, placeholder...)
		c.jumpsNeedingLabels = append(c.jumpsNeedingLabels, jumpLabelPair{
			instructionIndex: len(c.instructions) - 1, // Index of the jump instruction just added
			label:            jumpLabel,               // The label the jump is associated with
		})
		c.emitLabel(anyEndLabel)
		return nil // All `any` conditions processed
	}

	// Compile simple condition based on `Fact`, `Operator`, `Value`
	factIndex, err := c.getFactIndex(condition.Fact) // Check for an error from getFactIndex
	if err != nil {
		return err // Return the error if the fact is not found
	}

	log.Debug().
		Str("Fact", condition.Fact).
		Int("FactIndex", factIndex).
		Msg("Compiling condition for fact")

	c.emitInstruction(LOAD_FACT, byte(factIndex))
	c.emitLoadConstantInstruction(condition.Value, condition.ValueType) // Adjust for value type

	// Emit the comparison instruction based on `Operator`
	comparisonOpcode := c.getComparisonOpcode(condition.Operator)
	c.emitInstruction(comparisonOpcode)

	// Conditional jump based on the result
	if jumpIfTrue {
		c.emitInstruction(JUMP_IF_TRUE, placeholder...)
	} else {
		c.emitInstruction(JUMP_IF_FALSE, placeholder...)
	}

	// After emitting JUMP_IF_FALSE or JUMP_IF_TRUE
	jumpType := "JUMP_IF_FALSE"
	if jumpIfTrue {
		jumpType = "JUMP_IF_TRUE"
	}

	log.Debug().
		Str("JumpType", jumpType).
		Int("PlaceholderBytecodePosition", len(c.bytecode)-2).
		Msg("Emitted conditional jump with placeholder")

	// Append jump needing label resolution
	c.jumpsNeedingLabels = append(c.jumpsNeedingLabels, jumpLabelPair{
		instructionIndex: len(c.instructions) - 1, // Index of the jump instruction just added
		label:            jumpLabel,               // The label the jump is associated with
	})

	return nil
}

// resolveLabelOffsets replaces label placeholders with actual instruction offsets.
func (c *Compiler) resolveLabelOffsets() error {
	log.Info().Msg("Starting to resolve labels to offsets")

	log.Debug().
		Interface("FinalInstructions", c.instructions).
		Msg("Final Instructions before resolving labels")

	log.Debug().
		Interface("LabelOffsets", c.labelOffsets).
		Msg("Label Offsets")

	// Resolve jumps to label offsets based on the BytecodePosition
	for _, jump := range c.jumpsNeedingLabels {
		labelOffset, exists := c.labelOffsets[jump.label]
		if !exists {
			log.Error().
				Str("Label", jump.label).
				Msg("Error: label not defined")

			return fmt.Errorf("label %s not defined", jump.label)
		}

		placeholderPosition := c.instructions[jump.instructionIndex].BytecodePosition
		log.Debug().
			Str("Label", jump.label).
			Int("LabelOffset", labelOffset).
			Int("PlaceholderBytecodePosition", placeholderPosition).
			Msg("Resolving label to bytecode position")

		// Replace placeholder at placeholderPosition with actual labelOffset
		// binary.LittleEndian.PutUint16(c.bytecode[placeholderPosition:], uint16(labelOffset))
		binary.LittleEndian.PutUint16(c.bytecode[placeholderPosition:], uint16(labelOffset-placeholderPosition-2))

	}

	return nil
}

// getFactIndex retrieves the index of a fact in the fact table.
func (c *Compiler) getFactIndex(factName string) (int, error) {
	index, exists := c.context.FactIndex[factName]
	if !exists {
		return -1, fmt.Errorf("fact '%s' not defined in the context", factName)
	}
	return index, nil
}

// emitLoadConstantInstruction emits instructions to load a constant value of various types.
func (c *Compiler) emitLoadConstantInstruction(value interface{}, valueType string) {
	switch valueType {
	case "int":
		var intValue int
		switch v := value.(type) {
		case float64:
			// Force convert float64 to int if valueType is 'int'
			intValue = int(v)
		case int:
			intValue = v
		default:
			log.Fatal().
				Str("ExpectedType", "int").
				Interface("ActualType", value).
				Msg("Unsupported conversion")

		}
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(intValue))
		c.emitInstruction(LOAD_CONST_INT, buf...)

	case "float":
		var floatValue float64
		switch v := value.(type) {
		case int:
			// Force convert int to float64 if valueType is 'float'
			floatValue = float64(v)
		case float64:
			floatValue = v
		default:
			log.Fatal().
				Str("Type", fmt.Sprintf("%T", value)).
				Msg("Unsupported conversion: value type not expected for float")

		}
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, math.Float64bits(floatValue))
		c.emitInstruction(LOAD_CONST_FLOAT, buf...)

	case "string":
		strValue, ok := value.(string)
		if !ok {
			log.Fatal().
				Str("ValueType", fmt.Sprintf("%T", value)).
				Msg("Unsupported conversion: value is not a string as expected")
		}

		strBytes := []byte(strValue)
		// Assuming a single byte to denote length for simplicity, adjust as necessary.
		if len(strBytes) > 255 {
			panic("String value too long for LOAD_CONST_STRING instruction")
		}
		// Emit length followed by string bytes
		c.emitInstruction(LOAD_CONST_STRING, append([]byte{byte(len(strBytes))}, strBytes...)...)

	case "bool":
		boolValue, ok := value.(bool)
		if !ok {
			log.Fatal().
				Str("ValueType", fmt.Sprintf("%T", value)).
				Msg("Unsupported conversion: value is not a bool as expected")
		}
		var buf byte = 0x00
		if boolValue {
			buf = 0x01
		}
		c.emitInstruction(LOAD_CONST_BOOL, buf)

	default:
		panic(fmt.Sprintf("Unsupported valueType: '%s'", valueType))
	}
}

// Adjust getComparisonOpcode to match your operators
func (c *Compiler) getComparisonOpcode(operator string) Opcode {
	switch operator {
	case "greaterThan":
		return GT_INT // Adjust opcode as per your design
	case "greaterThanOrEqual":
		return GTE_INT // Adjust opcode as per your design
	// Add other operators here
	default:
		return NOP // Or handle unsupported operators as needed
	}
}
