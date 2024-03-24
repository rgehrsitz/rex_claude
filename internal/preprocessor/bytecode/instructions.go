// preprocessor/bytecode/instructions.go

package bytecode

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
)

// Instruction represents a single bytecode instruction.
type Instruction struct {
	Opcode   Opcode // The operation code
	Operands []byte // The operands for the instruction
}
