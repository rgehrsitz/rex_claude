// runtime/runtime.go

package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"rgehrsitz/rex/internal/preprocessor/bytecode"
	"unsafe"

	"github.com/rs/zerolog/log"
)

// VM represents the virtual machine that executes bytecode.
type VM struct {
	bytecode []byte
	ip       int
	stack    []interface{}
	facts    map[string]interface{}
}

type VMError struct {
	Message string
	IP      int
}

func (e *VMError) Error() string {
	return fmt.Sprintf("VM error at IP %d: %s", e.IP, e.Message)
}

// NewVM creates a new instance of the virtual machine.
func NewVM(bytecode []byte) *VM {
	return &VM{
		bytecode: bytecode,
		ip:       0,
		stack:    make([]interface{}, 0),
		facts:    make(map[string]interface{}),
	}
}

// Run executes the bytecode in the virtual machine.
func (vm *VM) Run() error {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				panic(&VMError{Message: err.Error(), IP: vm.ip})
			}
			panic(r)
		}
	}()

	// Skip over the header bytes
	headerSize := readHeader(vm.bytecode)
	vm.ip = headerSize

	for vm.ip < len(vm.bytecode) {
		opcode := bytecode.Opcode(vm.bytecode[vm.ip])
		vm.ip++

		// Print the current instruction
		log.Debug().Int("IP", vm.ip).Str("Opcode", opcode.String()).Msg("Processing instruction")

		if opcode.HasOperands() {
			operands, n := decodeOperands(vm.bytecode[vm.ip:])
			vm.ip += n
			fmt.Printf(", Operands: %v", operands)
		}
		fmt.Println()

		switch opcode {
		case bytecode.LOAD_CONST_INT:
			value, n := decodeInt(vm.bytecode[vm.ip:])
			vm.ip += n
			log.Debug().Interface("StackBefore", vm.stack).Msg("Before LOAD_CONST_INT")
			vm.stack = append(vm.stack, value)
			log.Debug().Interface("StackAfter", vm.stack).Msg("After LOAD_CONST_INT")

		case bytecode.LOAD_CONST_FLOAT:
			value, n := decodeFloat(vm.bytecode[vm.ip:])
			vm.ip += n
			vm.stack = append(vm.stack, value)

		case bytecode.LOAD_CONST_STRING:
			value, n := decodeString(vm.bytecode[vm.ip:])
			vm.ip += n
			vm.stack = append(vm.stack, value)

		case bytecode.LOAD_FACT:
			factName, n := decodeString(vm.bytecode[vm.ip:])
			vm.ip += n
			fmt.Printf("Before LOAD_FACT: Stack = %v\n", vm.stack)
			value, ok := vm.facts[factName]
			if !ok {
				return fmt.Errorf("undefined fact: %s", factName)
			}
			vm.stack = append(vm.stack, value)
			fmt.Printf("After LOAD_FACT: Stack = %v\n", vm.stack)

		case bytecode.EQ_INT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(int) == b.(int)
			}); err != nil {
				return err
			}

		case bytecode.NEQ_INT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(int) != b.(int)
			}); err != nil {
				return err
			}

		case bytecode.LT_INT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(int) < b.(int)
			}); err != nil {
				return err
			}

		case bytecode.LTE_INT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(int) <= b.(int)
			}); err != nil {
				return err
			}

		case bytecode.GT_INT:
			fmt.Printf("Before GT_INT: Stack = %v\n", vm.stack)
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(int) > b.(int)
			}); err != nil {
				return err
			}
			fmt.Printf("After GT_INT: Stack = %v\n", vm.stack)

		case bytecode.GTE_INT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(int) >= b.(int)
			}); err != nil {
				return err
			}

		case bytecode.EQ_FLOAT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(float64) == b.(float64)
			}); err != nil {
				return err
			}

		case bytecode.NEQ_FLOAT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(float64) != b.(float64)
			}); err != nil {
				return err
			}

		case bytecode.LT_FLOAT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(float64) < b.(float64)
			}); err != nil {
				return err
			}

		case bytecode.LTE_FLOAT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(float64) <= b.(float64)
			}); err != nil {
				return err
			}

		case bytecode.GT_FLOAT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(float64) > b.(float64)
			}); err != nil {
				return err
			}

		case bytecode.GTE_FLOAT:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(float64) >= b.(float64)
			}); err != nil {
				return err
			}

		case bytecode.EQ_STRING:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(string) == b.(string)
			}); err != nil {
				return err
			}

		case bytecode.NEQ_STRING:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(string) != b.(string)
			}); err != nil {
				return err
			}

		case bytecode.AND:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(bool) && b.(bool)
			}); err != nil {
				return err
			}

		case bytecode.OR:
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(bool) || b.(bool)
			}); err != nil {
				return err
			}

		case bytecode.NOT:
			if err := vm.unaryOp(func(a interface{}) interface{} {
				return !a.(bool)
			}); err != nil {
				return err
			}

		case bytecode.JUMP:
			offset, n := decodeInt(vm.bytecode[vm.ip:])
			vm.ip += n
			vm.ip = offset

		case bytecode.JUMP_IF_TRUE:
			offset, n := decodeInt(vm.bytecode[vm.ip:])
			vm.ip += n
			a, err := vm.pop()
			if err != nil {
				return err
			}
			if a.(bool) {
				vm.ip = offset
			}

		case bytecode.JUMP_IF_FALSE:
			offset, n := decodeInt(vm.bytecode[vm.ip:])
			vm.ip += n
			a, err := vm.pop()
			if err != nil {
				return err
			}
			if !a.(bool) {
				vm.ip = offset
			}

		case bytecode.HALT:
			return nil

		default:
			return &VMError{Message: fmt.Sprintf("unknown opcode: %d", opcode), IP: vm.ip}
		}
	}

	return nil
}

func (vm *VM) binaryOp(op func(a, b interface{}) interface{}) error {
	b, err := vm.pop()
	if err != nil {
		return err
	}
	a, err := vm.pop()
	if err != nil {
		return err
	}
	vm.stack = append(vm.stack, op(a, b))
	return nil
}

func (vm *VM) unaryOp(op func(a interface{}) interface{}) error {
	a, err := vm.pop()
	if err != nil {
		return err
	}
	vm.stack = append(vm.stack, op(a))
	return nil
}

func (vm *VM) pop() (interface{}, error) {
	if len(vm.stack) == 0 {
		return nil, &VMError{Message: "pop from an empty stack", IP: vm.ip}
	}
	value := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return value, nil
}

func decodeInt(bytecode []byte) (int, int) {
	value, n := binary.Varint(bytecode)
	return int(value), n
}

func decodeFloat(bytecode []byte) (float64, int) {
	bits := binary.LittleEndian.Uint64(bytecode)
	value := math.Float64frombits(bits)
	return value, 8
}

func decodeString(bytecode []byte) (string, int) {
	var value string
	var n int
	for i := 0; i < len(bytecode); i++ {
		if bytecode[i] == 0 {
			value = string(bytecode[:i])
			n = i + 1
			break
		}
	}
	return value, n
}

func decodeOperands(bytecode []byte) ([]interface{}, int) {
	var operands []interface{}
	var n int
	for len(bytecode) > 0 {
		value, m := decodeValue(&bytecode)
		operands = append(operands, value)
		n += m
	}
	return operands, n
}

func decodeValue(bytecode *[]byte) (interface{}, int) {
	switch (*bytecode)[0] {
	case 0: // int
		value, m := decodeInt((*bytecode)[1:])
		*bytecode = (*bytecode)[m+1:]
		return value, m + 1
	case 1: // float
		value, m := decodeFloat((*bytecode)[1:])
		*bytecode = (*bytecode)[m+1:]
		return value, m + 1
	case 2: // string
		value, m := decodeString((*bytecode)[1:])
		*bytecode = (*bytecode)[m+1:]
		return value, m + 1
	case 3: // bool
		value := (*bytecode)[1] == 1
		*bytecode = (*bytecode)[2:]
		return value, 2
	default:
		return nil, 0
	}
}

// readHeader reads the header from the bytecode and returns the size of the header.
func readHeader(bytecode []byte) int {
	// Read the header fields
	var header Header
	err := binary.Read(bytes.NewReader(bytecode), binary.LittleEndian, &header)
	if err != nil {
		// Handle error reading header
		return 0
	}

	// Debug print: Print the header fields
	log.Debug().
		Uint16("Version", header.Version).
		Uint32("Checksum", header.Checksum).
		Uint16("ConstPoolSize", header.ConstPoolSize).
		Uint16("NumRules", header.NumRules).
		Msg("Bytecode header details")

	// Calculate the header size based on the struct size
	return int(unsafe.Sizeof(header))
}

type Header struct {
	Version       uint16 // Version of the bytecode spec
	Checksum      uint32 // Checksum for integrity verification
	ConstPoolSize uint16 // Size of the constant pool
	NumRules      uint16 // Number of rules in the bytecode
	// ... other metadata fields
}
