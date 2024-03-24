// runtime/runtime.go

package runtime

import (
	"encoding/binary"
	"fmt"
	"math"
	"rgehrsitz/rex/internal/preprocessor/bytecode"
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

	for vm.ip < len(vm.bytecode) {
		opcode := bytecode.Opcode(vm.bytecode[vm.ip])
		vm.ip++

		// Print the current instruction
		fmt.Printf("IP: %d, Opcode: %s", vm.ip, opcode.String())
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
			vm.stack = append(vm.stack, value)

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
			value, ok := vm.facts[factName]
			if !ok {
				return fmt.Errorf("undefined fact: %s", factName)
			}
			vm.stack = append(vm.stack, value)

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
			if err := vm.binaryOp(func(a, b interface{}) interface{} {
				return a.(int) > b.(int)
			}); err != nil {
				return err
			}

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
		value, m := decodeValue(bytecode)
		operands = append(operands, value)
		n += m
		bytecode = bytecode[m:]
	}
	return operands, n
}

func decodeValue(bytecode []byte) (interface{}, int) {
	switch bytecode[0] {
	case 0: // int
		value, m := decodeInt(bytecode[1:])
		return value, m + 1
	case 1: // float
		value, m := decodeFloat(bytecode[1:])
		return value, m + 1
	case 2: // string
		value, m := decodeString(bytecode[1:])
		return value, m + 1
	case 3: // bool
		return bytecode[1] == 1, 2
	default:
		return nil, 0
	}
}
