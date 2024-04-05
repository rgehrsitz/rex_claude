// File: compiler.go

package bytecode

import (
	"encoding/binary"
	"fmt"
	"math"
	"rgehrsitz/rex/internal/rules"
)

// Compiler compiles optimized rules into bytecode.
type Compiler struct {
	instructions       []Instruction
	labelOffsets       map[string]int
	instructionOffsets map[int]int
	labelCounter       int
}

// NewCompiler creates a new instance of the bytecode compiler.
func NewCompiler() *Compiler {
	return &Compiler{
		instructions:       []Instruction{},
		labelOffsets:       make(map[string]int),
		instructionOffsets: make(map[int]int),
		labelCounter:       0,
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
	bytecode, err := c.resolveLabelOffsets()
	if err != nil {
		return nil, err
	}

	return bytecode, nil
}

// generateUniqueLabel generates a unique label for use in the bytecode.
func (c *Compiler) generateUniqueLabel(base string) string {
	label := fmt.Sprintf("%s_%d", base, c.labelCounter)
	c.labelCounter++
	return label
}

// emitInstruction appends an instruction to the compiler's list of instructions.
func (c *Compiler) emitInstruction(opcode Opcode, operands ...byte) {
	c.instructions = append(c.instructions, Instruction{Opcode: opcode, Operands: operands})
}

// emitLabel emits a label instruction and records its offset.
func (c *Compiler) emitLabel(label string) {
	c.labelOffsets[label] = len(c.instructions)
	c.emitInstruction(LABEL, []byte(label)...)
}

// compileRule compiles a single rule into bytecode.
func (c *Compiler) compileRule(rule *rules.Rule) error {
	startLabel := c.generateUniqueLabel("rule_start")
	endLabel := c.generateUniqueLabel("rule_end")
	c.emitLabel(startLabel)

	if err := c.compileConditions(rule.Conditions, endLabel); err != nil {
		return err
	}

	c.emitLabel(endLabel)
	return nil
}

// compileConditions compiles conditions (including nested conditions) into bytecode.
func (c *Compiler) compileConditions(conditions rules.Conditions, endLabel string) error {
	for _, cond := range conditions.All {
		if err := c.compileCondition(cond, endLabel, false); err != nil {
			return err
		}
	}

	for _, cond := range conditions.Any {
		actionLabel := c.generateUniqueLabel("action")
		if err := c.compileCondition(cond, actionLabel, true); err != nil {
			return err
		}
		c.emitInstruction(JUMP, []byte(endLabel)...) // Jump to end if condition is true
	}

	return nil
}

// compileCondition compiles a single condition or nested block into bytecode.
func (c *Compiler) compileCondition(condition *rules.Condition, jumpLabel string, jumpIfTrue bool) error {
	// Handle nested `all` conditions
	if len(condition.All) > 0 {
		for _, nestedCond := range condition.All {
			// For `all`, we jump to end if false
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
			// For `any`, jump to action if true
			if err := c.compileCondition(&nestedCond, jumpLabel, true); err != nil {
				return err
			}
		}
		c.emitInstruction(JUMP, []byte(anyEndLabel)) // Skip remaining conditions if one is true
		c.emitLabel(anyEndLabel)
		return nil // All `any` conditions processed
	}

	// Compile simple condition based on `Fact`, `Operator`, `Value`
	factIndex := c.getFactIndex(condition.Fact) // Implement based on fact management
	c.emitInstruction(LOAD_FACT, byte(factIndex))
	c.emitLoadConstantInstruction(condition.Value) // Adjust for value type

	// Emit the comparison instruction based on `Operator`
	comparisonOpcode := c.getComparisonOpcode(condition.Operator)
	c.emitInstruction(comparisonOpcode)

	// Conditional jump based on the result
	if jumpIfTrue {
		c.emitInstruction(JUMP_IF_TRUE, []byte(jumpLabel)...)
	} else {
		c.emitInstruction(JUMP_IF_FALSE, []byte(jumpLabel)...)
	}

	return nil
}

// resolveLabelOffsets replaces label placeholders with actual instruction offsets.
func (c *Compiler) resolveLabelOffsets() ([]byte, error) {
	var bytecode []byte
	for _, instr := range c.instructions {
		if instr.Opcode == LABEL {
			// Skip labels as they don't directly translate to bytecode
			continue
		}
		bytecode = append(bytecode, byte(instr.Opcode))
		if instr.Opcode == JUMP || instr.Opcode == JUMP_IF_TRUE || instr.Opcode == JUMP_IF_FALSE {
			label := string(instr.Operands)
			offset, exists := c.labelOffsets[label]
			if !exists {
				return nil, fmt.Errorf("label %s not defined", label)
			}
			// Assuming offsets are stored as 2-byte integers. Adjust according to your bytecode specification.
			offsetBytes := make([]byte, 2)
			binary.BigEndian.PutUint16(offsetBytes, uint16(offset))
			bytecode = append(bytecode, offsetBytes...)
		} else {
			bytecode = append(bytecode, instr.Operands...)
		}
	}
	return bytecode, nil
}

// getFactIndex retrieves the index of a fact in the fact table, placeholder for actual implementation
func (c *Compiler) getFactIndex(factName string) int {
	// Implement based on your fact storage mechanism, returning the index of the fact
	return 0 // Placeholder
}

// emitLoadConstantInstruction emits instructions to load a constant value of various types.
func (c *Compiler) emitLoadConstantInstruction(value interface{}) {
	switch v := value.(type) {
	case int:
		// Convert int to 4-byte array and emit LOAD_CONST_INT
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(v))
		c.emitInstruction(LOAD_CONST_INT, buf...)
	case float64:
		// Convert float64 to 8-byte array and emit LOAD_CONST_FLOAT
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(v))
		c.emitInstruction(LOAD_CONST_FLOAT, buf...)
	case string:
		// Assume strings are UTF-8 encoded and fit in your bytecode design.
		// You might need to precede the string with its length or end with a null terminator based on your design.
		strBytes := []byte(v)
		length := len(strBytes)
		if length > 255 {
			// Simplified: Assuming a single byte to denote length, adjust as necessary.
			panic("String value too long for LOAD_CONST_STRING instruction")
		}
		// Emit length followed by string bytes
		c.emitInstruction(LOAD_CONST_STRING, append([]byte{byte(length)}, strBytes...)...)
	case bool:
		// Convert bool to byte and emit LOAD_CONST_BOOL
		var buf byte = 0x00
		if v {
			buf = 0x01
		}
		c.emitInstruction(LOAD_CONST_BOOL, buf)
	default:
		panic(fmt.Sprintf("Unsupported constant type: %T", value))
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
