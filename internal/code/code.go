package code

import "encoding/binary"

type OpCode byte

const (
	OpPop OpCode = iota

	OpTrue
	OpFalse

	OpAnd
	OpOr

	// Prefix
	OpBang
	OpMinus
	OpIndex

	OpCmpEQ
	OpCmpNE
	OpCmpLT
	OpCmpLE
	OpCmpGT
	OpCmpGE

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod

	OpReturnValue

	OpCurrentClosure

	HAVE_ARGUMENT // OpCodes from here have an argument:

	OpConstant
	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
	OpGetFree

	OpList

	OpFunction
	OpCall
)

// If op has an argument
func (op OpCode) HasArg() bool {
	return op > HAVE_ARGUMENT
}

func ReadUint32(ins []byte) uint32 {
	return binary.BigEndian.Uint32(ins)
}
