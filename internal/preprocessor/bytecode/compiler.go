// preprocessor/bytecode/compiler.go

package bytecode

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"math"
	"rgehrsitz/rex/internal/rules"
)

// Compiler compiles optimized rules into bytecode.
type Compiler struct {
	instructions       []Instruction
	labelOffsets       map[string]int
	instructionOffsets map[int]int
}

// NewCompiler creates a new instance of the bytecode compiler.
func NewCompiler() *Compiler {
	return &Compiler{
		instructions:       []Instruction{},
		labelOffsets:       make(map[string]int),
		instructionOffsets: make(map[int]int),
	}
}

// Compile the given optimized rules into bytecode.
func (c *Compiler) Compile(rules []*rules.Rule) ([]byte, error) {
	for _, rule := range rules {
		if err := c.compileRule(rule); err != nil {
			return nil, err
		}
	}

	// Resolve label offsets and generate bytecode
	bytecode := c.resolveLabelOffsets()

	// Generate header
	header := c.generateHeader(len(rules), len(bytecode))

	// Combine header and bytecode
	return append(header, bytecode...), nil
}

func (c *Compiler) generateHeader(numRules int, bytecodeSize int) []byte {
	var buf bytes.Buffer

	// Write header fields
	binary.Write(&buf, binary.LittleEndian, uint16(1))            // Version
	binary.Write(&buf, binary.LittleEndian, uint32(0))            // Checksum (placeholder)
	binary.Write(&buf, binary.LittleEndian, uint16(numRules))     // NumRules
	binary.Write(&buf, binary.LittleEndian, uint32(bytecodeSize)) // BytecodeSize

	// Calculate checksum
	checksum := crc32.ChecksumIEEE(buf.Bytes())

	// Update checksum in the header
	checksumBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(checksumBytes, checksum)
	copy(buf.Bytes()[2:6], checksumBytes)

	return buf.Bytes()
}

// Emit encodes the operands into a byte slice and appends the instruction.
func (c *Compiler) emit(opcode Opcode, operands ...interface{}) error {
	encodedOperands, err := encodeOperands(operands...)
	if err != nil {
		return err
	}
	c.instructions = append(c.instructions, Instruction{Opcode: opcode, Operands: encodedOperands})
	return nil
}

// encodeOperands encodes multiple operands into a byte slice.
func encodeOperands(operands ...interface{}) ([]byte, error) {
	var encoded []byte
	for _, operand := range operands {
		switch op := operand.(type) {
		case int:
			buf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutVarint(buf, int64(op))
			encoded = append(encoded, buf[:n]...)
		case int64:
			buf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutVarint(buf, op)
			encoded = append(encoded, buf[:n]...)
		case uint:
			buf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutUvarint(buf, uint64(op))
			encoded = append(encoded, buf[:n]...)
		case uint64:
			buf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutUvarint(buf, op)
			encoded = append(encoded, buf[:n]...)
		case string:
			encoded = append(encoded, []byte(op)...)
			encoded = append(encoded, 0) // Null-terminated string
		case float64:
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, math.Float64bits(op))
			encoded = append(encoded, buf...)
		case bool:
			if op {
				encoded = append(encoded, 1)
			} else {
				encoded = append(encoded, 0)
			}
		case []byte:
			// Directly append a byte slice. Consider if you need a length prefix or delimiter.
			encoded = append(encoded, op...)
		default:
			return nil, errors.New("unsupported operand type")
		}
	}
	return encoded, nil
}

func (c *Compiler) emitLabel(label string) {
	c.labelOffsets[label] = len(c.instructions)
	c.emit(LABEL, label)
}

// resolveLabelOffsets calculates label offsets and replaces label operands with calculated offsets.
//
// No parameters.
// Returns a byte slice.
func (c *Compiler) resolveLabelOffsets() []byte {
	var bytecode []byte
	var labelOffsets = make(map[string]int)

	// Calculate label offsets
	for i, inst := range c.instructions {
		if inst.Opcode == LABEL {
			label := string(inst.Operands)
			labelOffsets[label] = len(bytecode)
		} else {
			bytecode = append(bytecode, byte(inst.Opcode))
			bytecode = append(bytecode, inst.Operands...)
		}
		c.instructionOffsets[i] = len(bytecode)
	}

	// Replace label operands with calculated offsets
	for i, inst := range c.instructions {
		if inst.Opcode == JUMP || inst.Opcode == JUMP_IF_TRUE || inst.Opcode == JUMP_IF_FALSE {
			label := string(inst.Operands)
			offset, ok := labelOffsets[label]
			if !ok {
				panic(fmt.Sprintf("undefined label: %s", label))
			}
			c.replaceOperandAtOffset(i, offset)
		}
	}

	return bytecode
}

// replaceOperandAtOffset replaces the operand at the specified instruction offset with the given value.
//
// Parameters: instructionOffset int, value int
func (c *Compiler) replaceOperandAtOffset(instructionOffset, value int) {
	operand := make([]byte, 4)
	binary.LittleEndian.PutUint32(operand, uint32(value))
	c.instructions[instructionOffset].Operands = operand
}

func (c *Compiler) compileRule(rule *rules.Rule) error {
	// Emit a label for the start of the rule
	c.emitLabel(fmt.Sprintf("rule_%s_start", rule.Name))

	// Compile the rule conditions
	if err := c.compileConditions(rule.Conditions); err != nil {
		return err
	}

	// Emit a JUMP_IF_FALSE instruction to skip the actions if the conditions are not met
	jumpIfFalseOffset := len(c.instructions)
	if err := c.emit(JUMP_IF_FALSE, 0); err != nil {
		return err
	}

	// Compile the rule actions
	if err := c.compileActions(rule.Event.Actions); err != nil {
		return err
	}

	// Emit a HALT instruction to indicate the end of the rule
	if err := c.emit(HALT); err != nil {
		return err
	}

	// Calculate the offset for the JUMP_IF_FALSE instruction
	jumpOffset := len(c.instructions) - jumpIfFalseOffset - 1

	// Update the operand of the JUMP_IF_FALSE instruction with the calculated offset
	jumpIfFalseInstruction := &c.instructions[jumpIfFalseOffset]
	jumpIfFalseInstruction.Operands = make([]byte, 4)
	binary.LittleEndian.PutUint32(jumpIfFalseInstruction.Operands, uint32(jumpOffset))

	return nil
}

func (c *Compiler) compileConditions(conditions rules.Conditions) error {
	// Compile the "all" conditions
	if err := c.compileConditionList(conditions.All, AND); err != nil {
		return err
	}

	// Compile the "any" conditions
	if err := c.compileConditionList(conditions.Any, OR); err != nil {
		return err
	}

	return nil
}

func (c *Compiler) compileConditionList(conditions []rules.Condition, logicalOp Opcode) error {
	if len(conditions) == 0 {
		return nil
	}

	for i, condition := range conditions {
		if err := c.compileCondition(condition); err != nil {
			return err
		}

		// Emit the logical operator (AND or OR) between conditions
		if i < len(conditions)-1 {
			if err := c.emit(logicalOp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Compiler) compileCondition(condition rules.Condition) error {
	// Print condition details
	fmt.Printf("Compiling condition: Fact=%s, Operator=%s, Value=%v, ValueType=%s\n",
		condition.Fact, condition.Operator, condition.Value, condition.ValueType)

	// Compile the fact
	if err := c.emit(LOAD_FACT, condition.Fact); err != nil {
		return err
	}

	fmt.Printf("Emitted instruction: LOAD_FACT %s\n", condition.Fact)

	// Compile the value
	if err := c.compileValue(condition.Value, condition.ValueType); err != nil {
		return err
	}
	fmt.Printf("Emitted instructions for value: %v\n", condition.Value)

	// Compile the value based on the valueType
	switch condition.ValueType {
	case "int":
		value, ok := condition.Value.(float64)
		if !ok {
			return fmt.Errorf("invalid value for int type: %v", condition.Value)
		}
		if err := c.emit(LOAD_CONST_INT, int64(value)); err != nil {
			return err
		}
	case "float":
		value, ok := condition.Value.(float64)
		if !ok {
			return fmt.Errorf("invalid value for float type: %v", condition.Value)
		}
		if err := c.emit(LOAD_CONST_FLOAT, value); err != nil {
			return err
		}
	case "string":
		value, ok := condition.Value.(string)
		if !ok {
			return fmt.Errorf("invalid value for string type: %v", condition.Value)
		}
		if err := c.emit(LOAD_CONST_STRING, value); err != nil {
			return err
		}
	case "bool":
		value, ok := condition.Value.(bool)
		if !ok {
			return fmt.Errorf("invalid value for bool type: %v", condition.Value)
		}
		if err := c.emit(LOAD_CONST_BOOL, value); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported value type: %s", condition.ValueType)
	}

	// Emit the comparison instruction based on the operator and value type
	switch condition.ValueType {
	case "int", "float":
		switch condition.Operator {
		case "equal":
			if err := c.emit(EQ_INT); err != nil {
				return err
			}
		case "notEqual":
			if err := c.emit(NEQ_INT); err != nil {
				return err
			}
		case "lessThan":
			if err := c.emit(LT_INT); err != nil {
				return err
			}
		case "lessThanOrEqual":
			if err := c.emit(LTE_INT); err != nil {
				return err
			}
		case "greaterThan":
			if err := c.emit(GT_INT); err != nil {
				return err
			}
		case "greaterThanOrEqual":
			if err := c.emit(GTE_INT); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported operator for %s type: %s", condition.ValueType, condition.Operator)
		}
	case "string":
		switch condition.Operator {
		case "equal":
			if err := c.emit(EQ_STRING); err != nil {
				return err
			}
		case "notEqual":
			if err := c.emit(NEQ_STRING); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported operator for string type: %s", condition.Operator)
		}
	case "bool":
		switch condition.Operator {
		case "equal":
			if err := c.emit(EQ_INT); err != nil {
				return err
			}
		case "notEqual":
			if err := c.emit(NEQ_INT); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported operator for bool type: %s", condition.Operator)
		}
	default:
		return fmt.Errorf("unsupported value type: %s", condition.ValueType)
	}

	return nil
}

func (c *Compiler) compileValue(value interface{}, valueType string) error {
	switch valueType {
	case "int":
		switch v := value.(type) {
		case int:
			if err := c.emit(LOAD_CONST_INT, v); err != nil {
				return err
			}
		case float64:
			// Check if the float64 value is an integer
			if float64(int64(v)) == v {
				if err := c.emit(LOAD_CONST_INT, int64(v)); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("value %v is not an integer", v)
			}
		default:
			return fmt.Errorf("unsupported value type: %T", v)
		}
	case "float":
		if err := c.emit(LOAD_CONST_FLOAT, value.(float64)); err != nil {
			return err
		}
	case "string":
		if err := c.emit(LOAD_CONST_STRING, value.(string)); err != nil {
			return err
		}
	case "bool":
		if err := c.emit(LOAD_CONST_BOOL, value.(bool)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported value type: %s", valueType)
	}

	return nil
}

func (c *Compiler) compileActions(actions []rules.Action) error {
	for _, action := range actions {
		if err := c.compileAction(action); err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) compileAction(action rules.Action) error {
	switch action.Type {
	case "updateFact":
		// Emit the UPDATE_FACT instruction
		if err := c.emit(UPDATE_FACT, action.Target); err != nil {
			return err
		}

		// Compile the value based on its type
		switch value := action.Value.(type) {
		case int:
			if err := c.emit(LOAD_CONST_INT, int64(value)); err != nil {
				return err
			}
		case int64:
			if err := c.emit(LOAD_CONST_INT, value); err != nil {
				return err
			}
		case float64:
			if err := c.emit(LOAD_CONST_FLOAT, value); err != nil {
				return err
			}
		case string:
			if err := c.emit(LOAD_CONST_STRING, value); err != nil {
				return err
			}
		case bool:
			if err := c.emit(LOAD_CONST_BOOL, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported value type for updateFact action: %T", value)
		}

	case "sendMessage":
		// Emit the SEND_MESSAGE instruction
		if err := c.emit(SEND_MESSAGE, action.Target); err != nil {
			return err
		}

		// Compile the value based on its type
		switch value := action.Value.(type) {
		case int:
			if err := c.emit(LOAD_CONST_INT, int64(value)); err != nil {
				return err
			}
		case int64:
			if err := c.emit(LOAD_CONST_INT, value); err != nil {
				return err
			}
		case float64:
			if err := c.emit(LOAD_CONST_FLOAT, value); err != nil {
				return err
			}
		case string:
			if err := c.emit(LOAD_CONST_STRING, value); err != nil {
				return err
			}
		case bool:
			if err := c.emit(LOAD_CONST_BOOL, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported value type for sendMessage action: %T", value)
		}

	default:
		return fmt.Errorf("unsupported action type: %s", action.Type)
	}

	return nil
}
