package compile

import (
	"parrot/internal/code"
	"parrot/internal/object"
)

// Resolved or unresolved instruction stream
type Instructions []Instruction

// Add an instruction to the instructions
func (is *Instructions) Add(i Instruction) {
	*is = append(*is, i)
}

type Instruction interface {
	Output() []byte
}

type Op struct {
	Op code.OpCode
}

func (op *Op) Output() []byte {
	panic("not implemented") // TODO: Implement
}

type OpArg struct {
	Op  code.OpCode
	Arg uint32
}

func (oparg *OpArg) Output() []byte {
	panic("not implemented") // TODO: Implement
}

type Compilable interface {
	Compile(c *Compiler) error
}

type Compiler struct {
	constants []object.Object
	opCodes   Instructions
}

func (c *Compiler) OpArg(op code.OpCode, arg uint32) {
	if !op.HasArg() {
		panic("OpArg called with an instruction which doesn't take an Arg")
	}
	c.opCodes.Add(&OpArg{
		Op:  op,
		Arg: arg,
	})
}

func (c *Compiler) Op(op code.OpCode) {
	if op.HasArg() {
		panic("Op called with an instruction which takes an Arg")
	}
	c.opCodes.Add(&Op{
		Op: op,
	})
}

// Add constant, return the index into the Consts tuple.
func (c *Compiler) Const(o object.Object) uint32 {
	for i, v := range c.constants {
		if o.Type() == v.Type() && o.String() == v.String() {
			return uint32(i)
		}
	}
	c.constants = append(c.constants, o)
	return uint32(len(c.constants) - 1)
}

func (c *Compiler) Compile(prog Compilable) error {
	return prog.Compile(c)
}
