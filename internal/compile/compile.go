package compile

import (
	"bytes"
	"encoding/binary"
	"parrot/internal/code"
	"parrot/internal/object"
)

type Instruction interface {
	Output() []byte
}

// Resolved or unresolved instruction stream
type Instructions []Instruction

// Add an instruction to the instructions
func (is *Instructions) Add(i Instruction) {
	*is = append(*is, i)
}

var _ Instruction = (*Instructions)(nil)

func (is *Instructions) Output() []byte {
	var buf bytes.Buffer
	for _, i := range *is {
		buf.Write(i.Output())
	}
	return buf.Bytes()
}

type Op struct {
	Op code.OpCode
}

func (op *Op) Output() []byte {
	return []byte{byte(op.Op)}
}

type OpArg struct {
	Op  code.OpCode
	Arg uint32
}

func (oparg *OpArg) Output() []byte {
	ret := [5]byte{
		byte(oparg.Op),
	}
	binary.BigEndian.PutUint32(ret[1:], oparg.Arg)
	return ret[:]
}

type Compilable interface {
	Compile(c *Compiler) error
}

type Compiler struct {
	Constants *[]object.Object
	OpCodes   Instructions
	*SymbolTable
}

func New() *Compiler {
	c := &Compiler{
		Constants:   &[]object.Object{},
		OpCodes:     []Instruction{},
		SymbolTable: NewSymbolTable(),
	}
	return c
}

func (c *Compiler) NewForFunction() *Compiler {
	nc := &Compiler{
		Constants:   c.Constants,
		OpCodes:     []Instruction{},
		SymbolTable: NewEnclosedSymbolTable(c.SymbolTable),
	}
	return nc
}

func (c *Compiler) OpArg(op code.OpCode, arg uint32) {
	if !op.HasArg() {
		panic("OpArg called with an instruction which doesn't take an Arg")
	}
	c.OpCodes.Add(&OpArg{
		Op:  op,
		Arg: arg,
	})
}

func (c *Compiler) Op(op code.OpCode) {
	if op.HasArg() {
		panic("Op called with an instruction which takes an Arg")
	}
	c.OpCodes.Add(&Op{
		Op: op,
	})
}

// Add constant, return the index into the Consts tuple.
func (c *Compiler) Const(o object.Object) uint32 {
loop:
	for i, v := range *c.Constants {
		switch o.Type() {
		case object.ListType, object.FunctionType, object.FunctionCompiledType:
			break loop
		}
		if o.Type() == v.Type() && o.String() == v.String() {
			return uint32(i)
		}
	}
	*c.Constants = append(*c.Constants, o)
	return uint32(len(*c.Constants) - 1)
}

func (c *Compiler) LoadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.OpArg(code.OpGetGlobal, uint32(s.Index))
	case LocalScope:
		c.OpArg(code.OpGetLocal, uint32(s.Index))
	case BuiltinScope:
		c.OpArg(code.OpGetBuiltin, uint32(s.Index))
	case FreeScope:
		c.OpArg(code.OpGetFree, uint32(s.Index))
	case FunctionScope:
		c.Op(code.OpCurrentClosure)
	}
}

func (c *Compiler) Compile(prog Compilable) error {
	return prog.Compile(c)
}
