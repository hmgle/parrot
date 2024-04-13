package code

type OpCode byte

const (
	OpPop OpCode = iota

	OpTrue
	OpFalse

	OpAnd
	OpOr

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

	HAVE_ARGUMENT // OpCodes from here have an argument

	OpConstant
)

// If op has an argument
func (op OpCode) HasArg() bool {
	return op > HAVE_ARGUMENT
}
