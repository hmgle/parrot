package vm

import (
	"fmt"
	"parrot/internal/code"
	"parrot/internal/compile"
	"parrot/internal/object"
)

const (
	StackSize  = 2048
	GlobalSize = 4096
)

var (
	True  = object.TRUEObj
	False = object.FALSEObj
	Null  = object.NULLObj
)

type VM struct {
	constants []object.Object
	stack     []object.Object
	globals   []object.Object
	ip        int // instruction pointer
	sp        int // Stack pointer: always points to the next free slot in the stack. Top of stack is stack[ip-1]
	opCodes   compile.Instructions
}

func New() *VM {
	return &VM{
		constants: []object.Object{},
		stack:     make([]object.Object, StackSize),
		globals:   make([]object.Object, GlobalSize),
		ip:        0,
		sp:        0,
		opCodes:   []compile.Instruction{},
	}
}

func (vm *VM) Next(constants []object.Object, opCodes compile.Instructions) {
	vm.constants = append(vm.constants, constants...)
	vm.opCodes = opCodes
	vm.ip = 0
}

func (vm *VM) Run() (err error) {
	// TODO
	for vm.ip < len(vm.opCodes) {
		op := vm.opCodes[vm.ip]
		vm.ip += 1
		switch o := op.(type) {
		case *compile.Op:
			switch o.Op {
			case code.OpPop:
				vm.pop()
			case code.OpTrue:
				err = vm.push(True)
			case code.OpFalse:
				err = vm.push(False)
			case code.OpAnd:
				vm.doAND()
			case code.OpOr:
				vm.doOR()
			case code.OpCmpEQ, code.OpCmpNE, code.OpCmpLE, code.OpCmpGE, code.OpCmpLT, code.OpCmpGT:
				vm.doCmp(o.Op)
			case code.OpAdd:
				vm.doAdd()
			case code.OpSub:
				vm.doSub()
			case code.OpMul:
				vm.doMul()
			case code.OpDiv:
				vm.doDiv()
			case code.OpMod:
				vm.doMod()
			case code.OpMinus:
				vm.doMinus()
			case code.OpBang:
				vm.doBang()
			case code.OpIndex:
				vm.doIndex()
			default:
				panic("not implemented") // TODO: Implement
			}
		case *compile.OpArg:
			switch o.Op {
			case code.OpConstant:
				vm.doLoadConst(int(o.Arg))
			case code.OpSetGlobal:
				vm.doStoreGlobal(int(o.Arg))
			case code.OpGetGlobal:
				vm.doGetGlobal(int(o.Arg))
			case code.OpList:
				vm.doList(int(o.Arg))
			default:
				panic("not implemented") // TODO: Implement
			}
		}
	}

	return err
}

func (vm *VM) doGetGlobal(index int) {
	_ = vm.push(vm.globals[index])
}

func (vm *VM) doStoreGlobal(index int) {
	vm.globals[index] = vm.pop()
}

func (vm *VM) doLoadConst(index int) {
	_ = vm.push(vm.constants[index])
}

func (vm *VM) doList(llen int) {
	var list []object.Object
	for i := range llen {
		list = append(list, vm.stack[vm.sp-llen+i])
	}
	l := object.NewList(list...)
	vm.sp -= llen
	_ = vm.push(l)
}

func (vm *VM) doIndex() {
	index := vm.pop()
	left := vm.Top()
	switch {
	case left.Type() == object.ListType && index.Type() == object.IntType:
		l := left.(*object.List)
		i := index.(*object.Integer)
		var o object.Object
		if int(*i) >= len(*l) {
			o = object.NewError("index out of range")
		} else {
			i := index.(*object.Integer)
			o = (*l)[*i]
		}
		vm.setTop(o)
	case left.Type() == object.StringType && index.Type() == object.IntType:
		s := left.(*object.String)
		i := index.(*object.Integer)
		var o object.Object
		if int(*i) >= len(*s) {
			o = object.NewError("index out of range")
		} else {
			o = object.NewString(string(string(*s)[*i]))
		}
		vm.setTop(o)
	default:
		panic("not implemented") // TODO: Implement
	}
}

func (vm *VM) doMinus() {
	a := vm.Top()
	switch a.Type() {
	case object.IntType:
		va := a.(*object.Integer)
		result := -*va
		vm.setTop(&result)
	default:
		panic("not implemented") // TODO: Implement
	}
}

func (vm *VM) doBang() {
	a := vm.Top()
	if a == False || a == object.NULLObj {
		vm.setTop(True)
	} else {
		vm.setTop(False)
	}
}

func (vm *VM) doMod() {
	b := vm.pop()
	a := vm.Top()
	switch a.Type() {
	case object.IntType:
		va := a.(*object.Integer)
		vb := b.(*object.Integer)
		result := *va % *vb
		vm.setTop(&result)
	default:
		panic("not implemented") // TODO: Implement
	}
}

func (vm *VM) doDiv() {
	b := vm.pop()
	a := vm.Top()
	switch a.Type() {
	case object.IntType:
		va := a.(*object.Integer)
		vb := b.(*object.Integer)
		result := *va / *vb
		vm.setTop(&result)
	default:
		panic("not implemented") // TODO: Implement
	}
}

func (vm *VM) doMul() {
	b := vm.pop()
	a := vm.Top()
	switch a.Type() {
	case object.IntType:
		va := a.(*object.Integer)
		vb := b.(*object.Integer)
		result := *va * *vb
		vm.setTop(&result)
	default:
		panic("not implemented") // TODO: Implement
	}
}

func (vm *VM) doSub() {
	b := vm.pop()
	a := vm.Top()
	switch a.Type() {
	case object.IntType:
		va := a.(*object.Integer)
		vb := b.(*object.Integer)
		result := *va - *vb
		vm.setTop(&result)
	default:
		panic("not implemented") // TODO: Implement
	}
}

func (vm *VM) doAdd() {
	b := vm.pop()
	a := vm.Top()
	switch a.Type() {
	case object.IntType:
		va := a.(*object.Integer)
		vb := b.(*object.Integer)
		result := *va + *vb
		vm.setTop(&result)
	case object.StringType:
		va := a.(*object.String)
		vb := b.(*object.String)
		result := *va + *vb
		vm.setTop(&result)
	default:
		panic("not implemented") // TODO: Implement
	}
}

func (vm *VM) doCmp(opCode code.OpCode) {
	b := vm.pop()
	a := vm.Top()
	result := False
	switch a.Type() {
	case object.BoolType:
		switch opCode {
		case code.OpCmpEQ:
			if a == b {
				result = True
			}
		case code.OpCmpNE:
			if a != b {
				result = True
			}
		case code.OpCmpLE:
			if a == b || (a == False) {
				result = True
			}
		case code.OpCmpGE:
			if a == b || (a == True) {
				result = True
			}
		case code.OpCmpLT:
			if a == False && b == True {
				result = True
			}
		case code.OpCmpGT:
			if a == True && b == False {
				result = True
			}
		}
	case object.IntType:
		va := a.(*object.Integer)
		vb := b.(*object.Integer)
		switch opCode {
		case code.OpCmpEQ:
			if *va == *vb {
				result = True
			}
		case code.OpCmpNE:
			if *va != *vb {
				result = True
			}
		case code.OpCmpLE:
			if *va <= *vb {
				result = True
			}
		case code.OpCmpGE:
			if *va >= *vb {
				result = True
			}
		case code.OpCmpLT:
			if *va < *vb {
				result = True
			}
		case code.OpCmpGT:
			if *va > *vb {
				result = True
			}
		}
	case object.StringType:
		va := a.(*object.String)
		vb := b.(*object.String)
		switch opCode {
		case code.OpCmpEQ:
			if *va == *vb {
				result = True
			}
		case code.OpCmpNE:
			if *va != *vb {
				result = True
			}
		case code.OpCmpLE:
			if *va <= *vb {
				result = True
			}
		case code.OpCmpGE:
			if *va >= *vb {
				result = True
			}
		case code.OpCmpLT:
			if *va < *vb {
				result = True
			}
		case code.OpCmpGT:
			if *va > *vb {
				result = True
			}
		}
	default:
		panic("not implemented") // TODO: Implement
	}
	vm.setTop(result)
}

func (vm *VM) doOR() {
	b := vm.pop()
	a := vm.Top()
	if a != True && b != True {
		vm.setTop(False)
	} else {
		vm.setTop(True)
	}
}

func (vm *VM) doAND() {
	b := vm.pop()
	a := vm.Top()
	if a == True && b == True {
		if a != True {
			vm.setTop(True)
		}
	}
	if a != False {
		vm.setTop(False)
	}
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) Top() object.Object {
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	if vm.sp > 0 {
		return vm.stack[vm.sp-1]
	}
	return vm.stack[vm.sp]
}

func (vm *VM) setTop(o object.Object) {
	vm.stack[vm.sp-1] = o
}
