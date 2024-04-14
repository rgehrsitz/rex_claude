// preprocessor/bytecode/instructions.go

package bytecode

import "fmt"

// Header defines the structure for bytecode metadata.
type Header struct {
	Version       uint16 // Version of the bytecode spec
	Checksum      uint32 // Checksum for integrity verification
	ConstPoolSize uint16 // Size of the constant pool
	NumRules      uint16 // Number of rules in the bytecode
	// ... other metadata fields
}

// Opcode represents the type of a bytecode instruction.
type Opcode byte

// Bytecode instructions
const (
	// Comparison instructions
	EQ_INT Opcode = iota
	NEQ_INT
	LT_INT
	LTE_INT
	GT_INT
	GTE_INT
	EQ_FLOAT
	NEQ_FLOAT
	LT_FLOAT
	LTE_FLOAT
	GT_FLOAT
	GTE_FLOAT
	EQ_STRING
	NEQ_STRING

	// Logical instructions
	AND
	OR
	NOT

	// Fact instructions
	LOAD_FACT
	STORE_FACT

	// Value instructions
	LOAD_CONST_INT
	LOAD_CONST_FLOAT
	LOAD_CONST_STRING
	LOAD_CONST_BOOL
	LOAD_VAR

	// Control flow instructions
	JUMP
	JUMP_IF_TRUE
	JUMP_IF_FALSE

	// Action instructions
	TRIGGER_ACTION
	UPDATE_FACT
	SEND_MESSAGE

	// Miscellaneous instructions
	NOP
	HALT
	ERROR

	// Optimization instructions
	INC
	DEC
	COMPARE_AND_JUMP

	// Label instruction
	LABEL

	RULE_START
	RULE_END // Add this instruction to mark the end of a rule

	COND_START
	COND_END // Add this instruction to mark the end of a condition

	ACTION_START
	ACTION_END // Add this instruction to mark the end of an action

)

// hasOperands returns true if the opcode requires operands.
func (op Opcode) HasOperands() bool {
	switch op {
	case LOAD_CONST_INT, LOAD_CONST_FLOAT, LOAD_CONST_STRING, LOAD_CONST_BOOL, LOAD_FACT, JUMP, JUMP_IF_TRUE, JUMP_IF_FALSE:
		return true
	default:
		return false
	}
}

// Instruction represents a single bytecode instruction.
type Instruction struct {
	Opcode           Opcode // The operation code
	Operands         []byte // The operands for the instruction
	BytecodePosition int    // Add this field to track the position in the bytecode
}

// String returns the string representation of the opcode.
func (op Opcode) String() string {
	switch op {
	case LOAD_CONST_INT:
		return "LOAD_CONST_INT"
	case LOAD_CONST_FLOAT:
		return "LOAD_CONST_FLOAT"
	case LOAD_CONST_STRING:
		return "LOAD_CONST_STRING"
	case LOAD_CONST_BOOL:
		return "LOAD_CONST_BOOL"
	case LOAD_FACT:
		return "LOAD_FACT"
	case EQ_INT:
		return "EQ_INT"
	case NEQ_INT:
		return "NEQ_INT"
	case LT_INT:
		return "LT_INT"
	case LTE_INT:
		return "LTE_INT"
	case GT_INT:
		return "GT_INT"
	case GTE_INT:
		return "GTE_INT"
	case EQ_FLOAT:
		return "EQ_FLOAT"
	case NEQ_FLOAT:
		return "NEQ_FLOAT"
	case LT_FLOAT:
		return "LT_FLOAT"
	case LTE_FLOAT:
		return "LTE_FLOAT"
	case GT_FLOAT:
		return "GT_FLOAT"
	case GTE_FLOAT:
		return "GTE_FLOAT"
	case EQ_STRING:
		return "EQ_STRING"
	case NEQ_STRING:
		return "NEQ_STRING"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"
	case JUMP:
		return "JUMP"
	case JUMP_IF_TRUE:
		return "JUMP_IF_TRUE"
	case JUMP_IF_FALSE:
		return "JUMP_IF_FALSE"
	case TRIGGER_ACTION:
		return "TRIGGER_ACTION"
	case UPDATE_FACT:
		return "UPDATE_FACT"
	case SEND_MESSAGE:
		return "SEND_MESSAGE"
	case NOP:
		return "NOP"
	case HALT:
		return "HALT"
	case ERROR:
		return "ERROR"
	case INC:
		return "INC"
	case DEC:
		return "DEC"
	case COMPARE_AND_JUMP:
		return "COMPARE_AND_JUMP"
	case LABEL:
		return "LABEL"
	case RULE_START:
		return "RULE_START"
	case RULE_END:
		return "RULE_END"
	case COND_START:
		return "COND_START"
	case COND_END:
		return "COND_END"
	case ACTION_START:
		return "ACTION_START"
	case ACTION_END:
		return "ACTION_END"
	default:
		return fmt.Sprintf("UNKNOWN_OPCODE(%d)", byte(op))
	}
}
